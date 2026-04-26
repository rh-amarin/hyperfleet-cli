package cmd

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/rh-amarin/hyperfleet-cli/internal/db"
)

// mockDBClient implements db.DBClient for unit tests without a live database.
type mockDBClient struct {
	queryHeaders []string
	queryRows    [][]string
	queryErr     error
	execErr      error
	execCalled   bool
	execSQL      string
}

func (m *mockDBClient) Query(_ context.Context, _ string, _ ...any) ([]string, [][]string, error) {
	return m.queryHeaders, m.queryRows, m.queryErr
}

func (m *mockDBClient) Exec(_ context.Context, sql string, _ ...any) error {
	m.execCalled = true
	m.execSQL = sql
	return m.execErr
}

func (m *mockDBClient) Close() {}

// Compile-time assertion.
var _ db.DBClient = (*mockDBClient)(nil)

// withMockDB replaces dbClientFactory with one that returns mock, then restores.
func withMockDB(t *testing.T, mock *mockDBClient) func() {
	t.Helper()
	orig := dbClientFactory
	dbClientFactory = func(_ context.Context, _ string) (db.DBClient, error) {
		return mock, nil
	}
	return func() { dbClientFactory = orig }
}

// withMockDBError replaces dbClientFactory with one that returns an error.
func withMockDBError(t *testing.T, err error) func() {
	t.Helper()
	orig := dbClientFactory
	dbClientFactory = func(_ context.Context, _ string) (db.DBClient, error) {
		return nil, err
	}
	return func() { dbClientFactory = orig }
}

// withStdin replaces stdinReader with the provided string and returns a restore func.
func withStdin(t *testing.T, input string) func() {
	t.Helper()
	orig := stdinReader
	stdinReader = strings.NewReader(input)
	return func() { stdinReader = orig }
}

// runDbCmd runs a db subcommand with a temp config dir and no API server needed.
func runDbCmd(t *testing.T, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	cfgDir := t.TempDir()
	fullArgs := append([]string{"--config", cfgDir, "--no-color"}, args...)
	return runCmdRaw(t, fullArgs)
}

// ── config ────────────────────────────────────────────────────────────────────

func TestDbConfig_PrintsKeyValueTable(t *testing.T) {
	stdout, _, err := runDbCmd(t, "db", "config")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, want := range []string{"host", "port", "name", "user", "password"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("expected %q in output, got:\n%s", want, stdout)
		}
	}
	if !strings.Contains(stdout, "<set>") && !strings.Contains(stdout, "<not set>") {
		t.Errorf("expected masked password in output, got:\n%s", stdout)
	}
	if strings.Contains(stdout, "foobar") {
		t.Error("password must not appear in plain text")
	}
}

func TestDbConfig_DefaultPortShown(t *testing.T) {
	stdout, _, err := runDbCmd(t, "db", "config")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "5432") {
		t.Errorf("expected default port 5432 in output, got:\n%s", stdout)
	}
}

// ── query ─────────────────────────────────────────────────────────────────────

func TestDbQuery_PrintsTableWithRows(t *testing.T) {
	mock := &mockDBClient{
		queryHeaders: []string{"id", "adapter"},
		queryRows:    [][]string{{"1", "cl-deployment"}, {"2", "cl-ocm"}},
	}
	defer withMockDB(t, mock)()

	stdout, _, err := runDbCmd(t, "db", "query", "SELECT * FROM adapter_statuses")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "ID") || !strings.Contains(stdout, "ADAPTER") {
		t.Errorf("expected headers in output, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "cl-deployment") {
		t.Errorf("expected row data in output, got:\n%s", stdout)
	}
}

func TestDbQuery_EmptyResult_PrintsInfo(t *testing.T) {
	mock := &mockDBClient{
		queryHeaders: []string{"id"},
		queryRows:    nil,
	}
	defer withMockDB(t, mock)()

	_, stderr, err := runDbCmd(t, "db", "query", "SELECT * FROM empty_table")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "0 rows") {
		t.Errorf("expected '0 rows' info message, got: %s", stderr)
	}
}

func TestDbQuery_ConnectError_ReturnsError(t *testing.T) {
	defer withMockDBError(t, errors.New("connection refused"))()

	_, _, err := runDbCmd(t, "db", "query", "SELECT 1")
	if err == nil {
		t.Fatal("expected error when DB connection fails")
	}
}

func TestDbQuery_QueryError_ReturnsError(t *testing.T) {
	mock := &mockDBClient{queryErr: errors.New("relation does not exist")}
	defer withMockDB(t, mock)()

	_, _, err := runDbCmd(t, "db", "query", "SELECT * FROM nonexistent")
	if err == nil {
		t.Fatal("expected error when query fails")
	}
}

// ── delete ────────────────────────────────────────────────────────────────────

func TestDbDelete_ConfirmYes_ExecutesDelete(t *testing.T) {
	mock := &mockDBClient{
		queryHeaders: []string{"count"},
		queryRows:    [][]string{{"3"}},
	}
	defer withMockDB(t, mock)()
	defer withStdin(t, "y\n")()

	_, stderr, err := runDbCmd(t, "db", "delete", "clusters", "id='abc'")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !mock.execCalled {
		t.Error("expected Exec to be called after confirmation")
	}
	if !strings.Contains(stderr, "Deleted") {
		t.Errorf("expected 'Deleted' in stderr, got: %s", stderr)
	}
}

func TestDbDelete_ConfirmNo_SkipsDelete(t *testing.T) {
	mock := &mockDBClient{
		queryHeaders: []string{"count"},
		queryRows:    [][]string{{"5"}},
	}
	defer withMockDB(t, mock)()
	defer withStdin(t, "n\n")()

	_, stderr, err := runDbCmd(t, "db", "delete", "clusters", "id='abc'")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.execCalled {
		t.Error("Exec must not be called when confirmation is denied")
	}
	if !strings.Contains(stderr, "Aborted") {
		t.Errorf("expected 'Aborted' in stderr, got: %s", stderr)
	}
}

func TestDbDelete_ConnectError_ReturnsError(t *testing.T) {
	defer withMockDBError(t, errors.New("no route to host"))()

	_, _, err := runDbCmd(t, "db", "delete", "clusters", "id='x'")
	if err == nil {
		t.Fatal("expected error when DB connection fails")
	}
}

// ── delete-all ────────────────────────────────────────────────────────────────

func TestDbDeleteAll_ConfirmYes_ExecutesDelete(t *testing.T) {
	mock := &mockDBClient{
		queryHeaders: []string{"count"},
		queryRows:    [][]string{{"10"}},
	}
	defer withMockDB(t, mock)()
	defer withStdin(t, "yes\n")()

	_, stderr, err := runDbCmd(t, "db", "delete-all", "adapter_statuses")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !mock.execCalled {
		t.Error("expected Exec to be called after confirmation")
	}
	if !strings.Contains(mock.execSQL, "adapter_statuses") {
		t.Errorf("Exec SQL should reference the table, got: %s", mock.execSQL)
	}
	_ = stderr
}

func TestDbDeleteAll_ConfirmNo_SkipsDelete(t *testing.T) {
	mock := &mockDBClient{
		queryHeaders: []string{"count"},
		queryRows:    [][]string{{"2"}},
	}
	defer withMockDB(t, mock)()
	defer withStdin(t, "N\n")()

	_, _, err := runDbCmd(t, "db", "delete-all", "clusters")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.execCalled {
		t.Error("Exec must not be called when confirmation is denied")
	}
}

// ── statuses ──────────────────────────────────────────────────────────────────

func TestDbStatuses_PrintsTable(t *testing.T) {
	mock := &mockDBClient{
		queryHeaders: []string{"id", "adapter", "cluster_id"},
		queryRows:    [][]string{{"s-001", "cl-deployment", "c-001"}},
	}
	defer withMockDB(t, mock)()

	stdout, _, err := runDbCmd(t, "db", "statuses")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "cl-deployment") {
		t.Errorf("expected adapter name in output, got:\n%s", stdout)
	}
}

func TestDbStatuses_EmptyResult_PrintsInfo(t *testing.T) {
	mock := &mockDBClient{
		queryHeaders: []string{"id"},
		queryRows:    nil,
	}
	defer withMockDB(t, mock)()

	_, stderr, err := runDbCmd(t, "db", "statuses")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "No adapter statuses") {
		t.Errorf("expected info message, got: %s", stderr)
	}
}

// ── statuses-delete ───────────────────────────────────────────────────────────

func TestDbStatusesDelete_ConfirmYes_ExecutesDelete(t *testing.T) {
	mock := &mockDBClient{
		queryHeaders: []string{"count"},
		queryRows:    [][]string{{"7"}},
	}
	defer withMockDB(t, mock)()
	defer withStdin(t, "y\n")()

	_, stderr, err := runDbCmd(t, "db", "statuses-delete")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !mock.execCalled {
		t.Error("expected Exec to be called after confirmation")
	}
	if !strings.Contains(stderr, "Deleted") {
		t.Errorf("expected 'Deleted' in stderr, got: %s", stderr)
	}
}

func TestDbStatusesDelete_ConfirmNo_SkipsDelete(t *testing.T) {
	mock := &mockDBClient{
		queryHeaders: []string{"count"},
		queryRows:    [][]string{{"4"}},
	}
	defer withMockDB(t, mock)()
	defer withStdin(t, "n\n")()

	_, _, err := runDbCmd(t, "db", "statuses-delete")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.execCalled {
		t.Error("Exec must not be called when confirmation is denied")
	}
}

// TestDbCmd_NoArgs_ShowsHelp verifies the db parent command prints help.
func TestDbCmd_NoArgs_ShowsHelp(t *testing.T) {
	cfgDir := t.TempDir()
	stdout, _, _ := runCmdRaw(t, []string{"--config", cfgDir, "db", "--help"})
	for _, want := range []string{"query", "delete", "statuses", "config"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("expected %q in help output, got:\n%s", want, stdout)
		}
	}
}

// TestDbQuery_RequiresSQLArg verifies that omitting the SQL arg returns an error.
func TestDbQuery_RequiresSQLArg_ReturnsError(t *testing.T) {
	// Ensure the factory returns a mock so we don't try to connect.
	mock := &mockDBClient{}
	defer withMockDB(t, mock)()

	cfgDir := t.TempDir()
	oldOut := os.Stdout
	oldErr := os.Stderr
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout, os.Stderr = wOut, wErr

	rootCmd.SetArgs([]string{"--config", cfgDir, "db", "query"})
	err := rootCmd.Execute()

	wOut.Close()
	wErr.Close()
	os.Stdout, os.Stderr = oldOut, oldErr
	rOut.Close()
	rErr.Close()

	if err == nil {
		t.Fatal("expected error when SQL arg is missing")
	}
}
