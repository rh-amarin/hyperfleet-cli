package cmd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/rh-amarin/hyperfleet-cli/internal/db"
	out "github.com/rh-amarin/hyperfleet-cli/internal/output"
	"github.com/spf13/cobra"
)

// dbClientFactory creates a DB client; replaced in tests to inject mocks.
var dbClientFactory = func(ctx context.Context, dsn string) (db.DBClient, error) {
	return db.New(ctx, dsn)
}

// stdinReader is the source for confirmation prompts; replaced in tests.
var stdinReader io.Reader = os.Stdin

func init() {
	rootCmd.AddCommand(dbCmd)
}

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Database operations",
}

func init() {
	dbCmd.AddCommand(dbQueryCmd)
	dbCmd.AddCommand(dbDeleteCmd)
	dbCmd.AddCommand(dbDeleteAllCmd)
	dbCmd.AddCommand(dbStatusesCmd)
	dbCmd.AddCommand(dbStatusesDeleteCmd)
	dbCmd.AddCommand(dbConfigCmd)
}

// dbDSN builds a postgres DSN from the resolved config.
func dbDSN() string {
	cfg := cfgStore.Cfg().Database
	port := cfg.Port
	if port == 0 {
		port = 5432
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s", cfg.User, cfg.Password, cfg.Host, port, cfg.Name)
}

// dbConnect opens a DB connection using the current config.
func dbConnect(ctx context.Context) (db.DBClient, error) {
	return dbClientFactory(ctx, dbDSN())
}

// confirmPrompt writes a prompt to stderr and reads one line from stdinReader.
// Returns true only when the user types "y" or "yes".
func confirmPrompt(msg string) bool {
	fmt.Fprintf(os.Stderr, "%s [y/N]: ", msg)
	scanner := bufio.NewScanner(stdinReader)
	if !scanner.Scan() {
		return false
	}
	answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
	return answer == "y" || answer == "yes"
}

// ── query ─────────────────────────────────────────────────────────────────────

var dbQueryCmd = &cobra.Command{
	Use:          "query <sql>",
	Short:        "Run a SELECT query and print results as a table",
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		client, err := dbConnect(ctx)
		if err != nil {
			return fmt.Errorf("connect to database: %w", err)
		}
		defer client.Close()

		headers, rows, err := client.Query(ctx, args[0])
		if err != nil {
			return err
		}
		if len(rows) == 0 {
			out.Info("Query returned 0 rows")
			return nil
		}
		return printer().PrintTable(headers, rows)
	},
}

// ── delete ────────────────────────────────────────────────────────────────────

var dbDeleteCmd = &cobra.Command{
	Use:          "delete <table> <where>",
	Short:        "Delete rows from a table matching a WHERE clause",
	Args:         cobra.ExactArgs(2),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		table, where := args[0], args[1]
		ctx := context.Background()

		client, err := dbConnect(ctx)
		if err != nil {
			return fmt.Errorf("connect to database: %w", err)
		}
		defer client.Close()

		_, rows, err := client.Query(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", table, where))
		if err != nil {
			return err
		}
		count := "0"
		if len(rows) > 0 && len(rows[0]) > 0 {
			count = rows[0][0]
		}

		if !confirmPrompt(fmt.Sprintf("Delete %s row(s) from %s WHERE %s?", count, table, where)) {
			out.Info("Aborted")
			return nil
		}

		if err := client.Exec(ctx, fmt.Sprintf("DELETE FROM %s WHERE %s", table, where)); err != nil {
			return err
		}
		out.Info(fmt.Sprintf("Deleted %s row(s) from %s", count, table))
		return nil
	},
}

// ── delete-all ────────────────────────────────────────────────────────────────

var dbDeleteAllCmd = &cobra.Command{
	Use:          "delete-all <table>",
	Short:        "Delete all rows from a table",
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		table := args[0]
		ctx := context.Background()

		client, err := dbConnect(ctx)
		if err != nil {
			return fmt.Errorf("connect to database: %w", err)
		}
		defer client.Close()

		_, rows, err := client.Query(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s", table))
		if err != nil {
			return err
		}
		count := "0"
		if len(rows) > 0 && len(rows[0]) > 0 {
			count = rows[0][0]
		}

		if !confirmPrompt(fmt.Sprintf("Delete all %s row(s) from %s?", count, table)) {
			out.Info("Aborted")
			return nil
		}

		if err := client.Exec(ctx, fmt.Sprintf("DELETE FROM %s", table)); err != nil {
			return err
		}
		out.Info(fmt.Sprintf("Deleted all rows from %s", table))
		return nil
	},
}

// ── statuses ──────────────────────────────────────────────────────────────────

var dbStatusesCmd = &cobra.Command{
	Use:          "statuses",
	Short:        "Show all adapter statuses from the database",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		client, err := dbConnect(ctx)
		if err != nil {
			return fmt.Errorf("connect to database: %w", err)
		}
		defer client.Close()

		headers, rows, err := client.Query(ctx, "SELECT * FROM adapter_statuses")
		if err != nil {
			return err
		}
		if len(rows) == 0 {
			out.Info("No adapter statuses found")
			return nil
		}
		return printer().PrintTable(headers, rows)
	},
}

// ── statuses-delete ───────────────────────────────────────────────────────────

var dbStatusesDeleteCmd = &cobra.Command{
	Use:          "statuses-delete",
	Short:        "Delete all adapter statuses from the database",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		client, err := dbConnect(ctx)
		if err != nil {
			return fmt.Errorf("connect to database: %w", err)
		}
		defer client.Close()

		_, rows, err := client.Query(ctx, "SELECT COUNT(*) FROM adapter_statuses")
		if err != nil {
			return err
		}
		count := "0"
		if len(rows) > 0 && len(rows[0]) > 0 {
			count = rows[0][0]
		}

		if !confirmPrompt(fmt.Sprintf("Delete all %s adapter status record(s)?", count)) {
			out.Info("Aborted")
			return nil
		}

		if err := client.Exec(ctx, "DELETE FROM adapter_statuses"); err != nil {
			return err
		}
		out.Info(fmt.Sprintf("Deleted %s adapter status record(s)", count))
		return nil
	},
}

// ── config ────────────────────────────────────────────────────────────────────

var dbConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Show the resolved database connection config",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := cfgStore.Cfg().Database
		port := cfg.Port
		if port == 0 {
			port = 5432
		}
		password := "<not set>"
		if cfg.Password != "" {
			password = "<set>"
		}
		headers := []string{"key", "value"}
		rows := [][]string{
			{"host", cfg.Host},
			{"port", fmt.Sprintf("%d", port)},
			{"name", cfg.Name},
			{"user", cfg.User},
			{"password", password},
		}
		return printer().PrintTable(headers, rows)
	},
}
