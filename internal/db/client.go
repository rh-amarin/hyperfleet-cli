package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DBClient is the interface for database operations, enabling mock injection in tests.
type DBClient interface {
	Query(ctx context.Context, sql string, args ...any) (headers []string, rows [][]string, err error)
	Exec(ctx context.Context, sql string, args ...any) error
	Close()
}

// Client wraps a pgxpool.Pool and implements DBClient.
type Client struct {
	pool *pgxpool.Pool
}

// New connects to the database using the given DSN and returns a ready Client.
func New(ctx context.Context, dsn string) (*Client, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return &Client{pool: pool}, nil
}

// Query runs a SELECT and returns column headers and string-formatted rows.
func (c *Client) Query(ctx context.Context, sql string, args ...any) ([]string, [][]string, error) {
	rows, err := c.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	fields := rows.FieldDescriptions()
	headers := make([]string, len(fields))
	for i, f := range fields {
		headers[i] = string(f.Name)
	}

	var result [][]string
	for rows.Next() {
		vals, err := rows.Values()
		if err != nil {
			return nil, nil, fmt.Errorf("scan row: %w", err)
		}
		row := make([]string, len(vals))
		for i, v := range vals {
			if v == nil {
				row[i] = ""
			} else {
				row[i] = fmt.Sprintf("%v", v)
			}
		}
		result = append(result, row)
	}
	return headers, result, rows.Err()
}

// Exec runs a non-SELECT SQL statement.
func (c *Client) Exec(ctx context.Context, sql string, args ...any) error {
	_, err := c.pool.Exec(ctx, sql, args...)
	return err
}

// Close closes the connection pool.
func (c *Client) Close() {
	c.pool.Close()
}
