# Tasks: Phase 02 — Shared API Client & Resource Types

## 1. internal/api — Client core

- [ ] 1.1 Create `internal/api/client.go`: `Client` struct with `baseURL`, `token`, `httpClient`, `verbose` fields
- [ ] 1.2 Implement `NewClient(apiURL, apiVersion, token string, verbose bool) *Client` constructor that builds base URL as `{apiURL}/api/hyperfleet/{apiVersion}/`
- [ ] 1.3 Implement `Do(ctx, method, path string, body any) (*http.Response, error)` core method: marshal body, build request with context, set headers (`Content-Type`, `Accept`, `Authorization`), send, log if verbose
- [ ] 1.4 Implement verbose logging to stderr: `[DEBUG] METHOD URL → STATUS (DURATIONms)`

## 2. internal/api — Generic typed methods

- [ ] 2.1 Create `internal/api/methods.go`: `Get[T any](c *Client, ctx context.Context, path string) (*T, error)`
- [ ] 2.2 Implement `Post[T any](c *Client, ctx context.Context, path string, body any) (*T, error)`
- [ ] 2.3 Implement `Patch[T any](c *Client, ctx context.Context, path string, body any) (*T, error)`
- [ ] 2.4 Implement `Delete[T any](c *Client, ctx context.Context, path string) (*T, error)` (returns deleted resource)

## 3. internal/api — RFC 7807 error handling

- [ ] 3.1 Create `internal/api/errors.go`: `APIError` struct with fields `Code`, `Detail`, `Instance`, `Status`, `Title`, `TraceID`, `Type`, `Timestamp` and JSON tags
- [ ] 3.2 Implement `APIError.Error() string` returning `[{status}] {title}: {detail}`
- [ ] 3.3 Implement `parseError(resp *http.Response) *APIError`: try JSON unmarshal as RFC 7807; fall back to raw body with HTTP status text
- [ ] 3.4 Integrate `parseError` into `Do()`: return `*APIError` for any non-2xx response
- [ ] 3.5 Implement `IsAPIError(err error) (*APIError, bool)` helper for callers to type-assert

## 4. internal/api — Unit tests

- [ ] 4.1 Create `internal/api/client_test.go` using `httptest.NewServer`
- [ ] 4.2 Test `Get[T]` with 200 response: verify JSON unmarshaling into typed result
- [ ] 4.3 Test `Post[T]` with 201 response: verify request body and typed response
- [ ] 4.4 Test RFC 7807 error parsing: 404 response with valid JSON → `APIError` with all fields
- [ ] 4.5 Test non-JSON error response: 500 with plain text → `APIError` with raw body as detail
- [ ] 4.6 Test verbose logging: capture stderr, verify `[DEBUG]` line format
- [ ] 4.7 Test no-auth: client without token sends no `Authorization` header
- [ ] 4.8 Test context cancellation: cancelled context returns `context.Canceled`

## 5. internal/resource — Cluster and NodePool types

- [ ] 5.1 Create `internal/resource/cluster.go`: `Cluster` struct with all fields and JSON tags
- [ ] 5.2 Create `internal/resource/nodepool.go`: `NodePool` struct with `OwnerReferences` field, `OwnerReference` struct
- [ ] 5.3 Create `internal/resource/condition.go`: `Condition` struct, `ResourceStatus` struct

## 6. internal/resource — AdapterStatus, CloudEvent, ListResponse

- [ ] 6.1 Create `internal/resource/adapter.go`: `AdapterStatus` struct with all fields
- [ ] 6.2 Create `internal/resource/event.go`: `CloudEvent` struct with `specversion` defaulting to `"1.0"`
- [ ] 6.3 Create `internal/resource/list.go`: `ListResponse[T any]` generic struct with `Items`, `Kind`, `Page`, `Size`, `Total`

## 7. internal/resource — Unit tests

- [ ] 7.1 Create `internal/resource/types_test.go`
- [ ] 7.2 Test Cluster JSON round-trip: unmarshal real API JSON fixture → re-marshal → compare
- [ ] 7.3 Test NodePool JSON round-trip: verify `owner_references`, `spec.platform.type` preserved
- [ ] 7.4 Test Condition JSON round-trip: verify all fields including `observed_generation` omitempty
- [ ] 7.5 Test ListResponse[Cluster] JSON round-trip: verify `items` as empty slice (not nil) when empty
- [ ] 7.6 Test AdapterStatus JSON round-trip: verify `data` field accepts arbitrary JSON

## 8. internal/output — Printer and format dispatch

- [ ] 8.1 Create `internal/output/printer.go`: `Printer` struct with `format`, `noColor`, `writer` fields
- [ ] 8.2 Implement `NewPrinter(format string, noColor bool) *Printer` (defaults writer to stdout)
- [ ] 8.3 Implement `Print(v any) error`: dispatch to JSON (indented), table, or YAML based on format
- [ ] 8.4 Implement JSON output: `encoding/json.MarshalIndent` with 2-space indent + trailing newline
- [ ] 8.5 Implement YAML output: `gopkg.in/yaml.v3.Encoder` to writer

## 9. internal/output — Table rendering

- [ ] 9.1 Create `internal/output/table.go`: `PrintTable(headers []string, rows [][]string) error` using `text/tabwriter`
- [ ] 9.2 Render headers in uppercase with tab separation
- [ ] 9.3 Align columns using tabwriter with minimum width and padding

## 10. internal/output — Colored dot renderer

- [ ] 10.1 Create `internal/output/dots.go`: `Dot(status string) string` function
- [ ] 10.2 Implement color mapping: True→green `●`, False→red `●`, Unknown→yellow `●`, absent→`-`
- [ ] 10.3 Implement no-color mode: True→`True`, False→`False`, Unknown→`Unknown`, absent→`-`
- [ ] 10.4 Respect `NO_COLOR` environment variable in addition to `--no-color` flag

## 11. internal/output — Dynamic column builder

- [ ] 11.1 Create `internal/output/columns.go`: `DynamicColumns(conditions [][]Condition) []string`
- [ ] 11.2 Implement collection of unique condition types across all resource conditions
- [ ] 11.3 Implement ordering: `Available` first, alphabetical middle, `Ready` last
- [ ] 11.4 Return empty slice when no conditions are present

## 12. internal/output — Message helpers

- [ ] 12.1 Implement `Warn(msg string)`: writes `[WARN] msg` to stderr
- [ ] 12.2 Implement `Info(msg string)`: writes `[INFO] msg` to stderr
- [ ] 12.3 Implement `Errorf(msg string)`: writes `[ERROR] msg` to stderr

## 13. internal/output — Unit tests

- [ ] 13.1 Create `internal/output/output_test.go`
- [ ] 13.2 Test JSON output: verify indentation and trailing newline
- [ ] 13.3 Test dot rendering: verify all 4 status mappings in color and no-color modes
- [ ] 13.4 Test dynamic column ordering: mixed conditions → Available first, Ready last, alpha middle
- [ ] 13.5 Test dynamic columns with no conditions → empty slice
- [ ] 13.6 Test dynamic columns with only Ready/Available → correct ordering maintained

## 14. Verify

- [ ] 14.1 `go build ./...` succeeds with all three new packages
- [ ] 14.2 `go test ./internal/api/...` — all API client tests pass
- [ ] 14.3 `go test ./internal/resource/...` — all resource type tests pass
- [ ] 14.4 `go test ./internal/output/...` — all output tests pass
- [ ] 14.5 `go vet ./...` reports no issues
