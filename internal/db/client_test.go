package db_test

import (
	"context"
	"errors"
	"testing"

	"github.com/rh-amarin/hyperfleet-cli/internal/db"
)

// mockClient is a test double that implements DBClient.
type mockClient struct {
	queryHeaders []string
	queryRows    [][]string
	queryErr     error
	execErr      error
	closed       bool
	execSQL      string
}

func (m *mockClient) Query(ctx context.Context, sql string, args ...any) ([]string, [][]string, error) {
	return m.queryHeaders, m.queryRows, m.queryErr
}

func (m *mockClient) Exec(ctx context.Context, sql string, args ...any) error {
	m.execSQL = sql
	return m.execErr
}

func (m *mockClient) Close() {
	m.closed = true
}

// Compile-time assertion: mockClient implements DBClient.
var _ db.DBClient = (*mockClient)(nil)

func TestMockClient_Query_ReturnsHeadersAndRows(t *testing.T) {
	mock := &mockClient{
		queryHeaders: []string{"id", "name"},
		queryRows:    [][]string{{"1", "alice"}, {"2", "bob"}},
	}

	headers, rows, err := mock.Query(context.Background(), "SELECT id, name FROM users")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(headers) != 2 || headers[0] != "id" || headers[1] != "name" {
		t.Errorf("headers = %v, want [id name]", headers)
	}
	if len(rows) != 2 {
		t.Errorf("rows count = %d, want 2", len(rows))
	}
	if rows[0][1] != "alice" {
		t.Errorf("rows[0][1] = %q, want alice", rows[0][1])
	}
}

func TestMockClient_Query_PropagatesError(t *testing.T) {
	want := errors.New("connection refused")
	mock := &mockClient{queryErr: want}

	_, _, err := mock.Query(context.Background(), "SELECT 1")
	if !errors.Is(err, want) {
		t.Errorf("err = %v, want %v", err, want)
	}
}

func TestMockClient_Exec_ReturnsNilOnSuccess(t *testing.T) {
	mock := &mockClient{}
	if err := mock.Exec(context.Background(), "DELETE FROM users WHERE id=1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMockClient_Exec_PropagatesError(t *testing.T) {
	want := errors.New("exec failed")
	mock := &mockClient{execErr: want}

	err := mock.Exec(context.Background(), "DELETE FROM users")
	if !errors.Is(err, want) {
		t.Errorf("err = %v, want %v", err, want)
	}
}

func TestMockClient_Close_SetsFlag(t *testing.T) {
	mock := &mockClient{}
	if mock.closed {
		t.Fatal("expected closed=false before Close()")
	}
	mock.Close()
	if !mock.closed {
		t.Error("expected closed=true after Close()")
	}
}
