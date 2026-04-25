# Design: Phase 02 — Shared API Client & Resource Types

## Context

Phase 00 (bootstrap) and Phase 01 (config) are complete. The CLI compiles,
`hf version` works, and the full config system (split YAML, env profiles,
precedence chain, state management) is operational. The `internal/config.Store`
is initialized in `cmd/root.go`'s `PersistentPreRunE` and is accessible to all
subcommands.

This phase builds the three shared internal packages that every data-producing
command in Phases 03–13 will import. No new CLI commands are added — only
libraries and their unit tests.

**Existing integration points:**
- `cfgStore.Cfg().Hyperfleet.APIURL` — base URL for the API
- `cfgStore.Cfg().Hyperfleet.APIVersion` — version path segment (e.g., `v1`)
- `cfgStore.Cfg().Hyperfleet.Token` — Bearer token for auth
- Global flags: `--output`, `--no-color`, `--verbose` (defined in `cmd/root.go`)

## Goals / Non-Goals

**Goals:**
- Provide a type-safe, tested HTTP client that handles all HyperFleet API
  conventions (base URL, auth, RFC 7807 errors, verbose logging)
- Define Go structs for all HyperFleet resources that faithfully match the API
  JSON schema, enabling `encoding/json` round-trip without data loss
- Provide an output printer that dispatches on `--output json|table|yaml` with
  colored dot rendering and dynamic column ordering for condition tables
- Keep the public API surface minimal and idiomatic Go

**Non-Goals:**
- Adding CLI commands (deferred to Phase 03+)
- Watch mode / live refresh (Phase 06)
- Kubernetes client-go integration (Phase 10)
- Streaming or WebSocket support

## Decisions

### 1. Generic typed methods via Go generics

**Decision:** Use Go 1.18+ type parameters for `Get[T]`, `Post[T]`, `Patch[T]`.

**Rationale:** Every API call follows the same pattern: build URL, set headers,
marshal request body, send, check status, unmarshal response body into a typed
result. Generics eliminate boilerplate without sacrificing type safety.

**Alternative considered:** Interface-based approach with `json.RawMessage` and
caller-side unmarshaling. Rejected because it pushes type assertions to every
call site and loses compile-time safety.

```go
func Get[T any](c *Client, path string) (*T, error)
func Post[T any](c *Client, path string, body any) (*T, error)
func Patch[T any](c *Client, path string, body any) (*T, error)
func (c *Client) Delete(path string) error
```

`Delete` is not generic because delete responses vary (some return the deleted
object, some return empty). Callers that need the response body can use `Get[T]`
semantics with a DELETE-specific helper.

### 2. RFC 7807 error handling as a first-class type

**Decision:** Parse every non-2xx response as `APIError` and implement `error`.

**Rationale:** The HyperFleet API consistently returns RFC 7807 Problem Details.
Making this a concrete Go type lets commands choose between outputting the raw
error JSON (exit 0, matching bash behavior) or wrapping it in a Go error chain.

```go
type APIError struct {
    Code      string `json:"code"`
    Detail    string `json:"detail"`
    Instance  string `json:"instance"`
    Status    int    `json:"status"`
    Title     string `json:"title"`
    TraceID   string `json:"trace_id"`
    Type      string `json:"type"`
    Timestamp string `json:"timestamp"`
}

func (e *APIError) Error() string { return fmt.Sprintf("[%d] %s: %s", e.Status, e.Title, e.Detail) }
```

Commands that want bash-compatible behavior call `output.Print(apiErr)` and
exit 0. Commands that want Go-style error handling return the error.

### 3. `map[string]any` for Spec and Labels

**Decision:** `Spec` and `Labels` fields use `map[string]any` instead of typed
structs.

**Rationale:** The HyperFleet API allows arbitrary keys in spec and labels.
Different adapters inject different fields. Typed structs would require
updating the CLI for every new adapter field. `map[string]any` keeps the CLI
forward-compatible.

**Trade-off:** No compile-time safety for spec/label field access. Mitigated
by helper functions for commonly accessed fields (e.g., `resource.SpecCounter`,
`resource.SpecRegion`).

### 4. Output printer as a standalone package

**Decision:** `internal/output.Printer` accepts the format string and no-color
flag, and provides `Print(v any)`, `PrintTable(headers, rows)`, and
`PrintDot(status string)` methods.

**Rationale:** Separating output from both the API client and the command layer
allows unit testing of formatting logic without HTTP or Cobra dependencies.

**Column ordering algorithm for condition tables:**
1. Fixed columns first (e.g., `ID`, `NAME`, `GEN`)
2. `Available` column
3. All other condition types, sorted alphabetically
4. `Ready` column last

This matches the bash script behavior where `Available` is the leading
indicator and `Ready` is the aggregated final status.

### 5. No new external dependencies

**Decision:** Use only stdlib (`net/http`, `encoding/json`, `text/tabwriter`,
`fmt`, `os`) plus `gopkg.in/yaml.v3` (already in go.mod).

**Rationale:** Phase 02 is a foundational layer. Adding HTTP client libraries
(resty, req) or logging frameworks (logrus, zap) here would propagate
transitively to every command. The stdlib is sufficient for a REST client with
JSON payloads.

**Alternative considered:** `go-resty/resty` for retry/middleware. Deferred —
retry logic is a non-functional concern that can be layered on later without
changing the `internal/api` interface.

### 6. Delete returns the deleted resource

**Decision:** Provide `DeleteWithResponse[T]` alongside bare `Delete`.

**Rationale:** The HyperFleet API returns the full resource object on DELETE
(with `deleted_by`, `deleted_time`, incremented `generation`). The bash scripts
output this response. A typed delete method preserves this behavior.

```go
func Delete[T any](c *Client, path string) (*T, error)
```

## Package Layout

```
internal/
├── api/
│   ├── client.go      — Client struct, NewClient, Do (core HTTP)
│   ├── methods.go     — Get[T], Post[T], Patch[T], Delete[T]
│   ├── errors.go      — APIError type, RFC 7807 parser
│   └── client_test.go — unit tests with httptest.Server fixtures
├── resource/
│   ├── cluster.go     — Cluster, ClusterStatus structs
│   ├── nodepool.go    — NodePool, NodePoolStatus, OwnerReference
│   ├── condition.go   — Condition struct
│   ├── adapter.go     — AdapterStatus struct
│   ├── event.go       — CloudEvent struct
│   ├── list.go        — ListResponse[T] generic wrapper
│   └── types_test.go  — JSON round-trip tests
└── output/
    ├── printer.go     — Printer, Print (json/table/yaml dispatch)
    ├── table.go       — table renderer with tabwriter
    ├── dots.go        — colored dot renderer (True/False/Unknown/absent)
    ├── columns.go     — dynamic column builder and ordering
    └── output_test.go — unit tests for column ordering and dot rendering
```

## Key Go Types

### internal/api

```go
type Client struct {
    baseURL    string
    token      string
    httpClient *http.Client
    verbose    bool
}

func NewClient(apiURL, apiVersion, token string, verbose bool) *Client
```

### internal/resource

```go
type Cluster struct {
    ID              string         `json:"id"`
    Kind            string         `json:"kind"`
    Name            string         `json:"name"`
    Generation      int            `json:"generation"`
    Labels          map[string]any `json:"labels"`
    Spec            map[string]any `json:"spec"`
    Status          ResourceStatus `json:"status"`
    CreatedBy       string         `json:"created_by"`
    CreatedTime     string         `json:"created_time"`
    UpdatedBy       string         `json:"updated_by"`
    UpdatedTime     string         `json:"updated_time"`
    DeletedBy       string         `json:"deleted_by,omitempty"`
    DeletedTime     string         `json:"deleted_time,omitempty"`
    Href            string         `json:"href"`
}

type ResourceStatus struct {
    Conditions []Condition `json:"conditions"`
}

type NodePool struct {
    ID              string           `json:"id"`
    Kind            string           `json:"kind"`
    Name            string           `json:"name"`
    Generation      int              `json:"generation"`
    Labels          map[string]any   `json:"labels"`
    Spec            map[string]any   `json:"spec"`
    Status          ResourceStatus   `json:"status"`
    OwnerReferences []OwnerReference `json:"owner_references"`
    CreatedBy       string           `json:"created_by"`
    CreatedTime     string           `json:"created_time"`
    UpdatedBy       string           `json:"updated_by"`
    UpdatedTime     string           `json:"updated_time"`
    DeletedBy       string           `json:"deleted_by,omitempty"`
    DeletedTime     string           `json:"deleted_time,omitempty"`
}

type Condition struct {
    Type               string `json:"type"`
    Status             string `json:"status"`
    Reason             string `json:"reason"`
    Message            string `json:"message"`
    LastTransitionTime string `json:"last_transition_time"`
    ObservedGeneration int    `json:"observed_generation,omitempty"`
}

type AdapterStatus struct {
    Adapter            string      `json:"adapter"`
    Conditions         []Condition `json:"conditions"`
    ObservedGeneration int         `json:"observed_generation"`
    LastReportTime     string      `json:"last_report_time"`
    Data               any         `json:"data"`
    CreatedTime        string      `json:"created_time"`
}

type CloudEvent struct {
    SpecVersion string `json:"specversion"`
    Type        string `json:"type"`
    Source      string `json:"source"`
    ID          string `json:"id"`
    Data        any    `json:"data"`
}

type OwnerReference struct {
    Kind string `json:"kind"`
    ID   string `json:"id"`
}

type ListResponse[T any] struct {
    Items []T    `json:"items"`
    Kind  string `json:"kind"`
    Page  int    `json:"page"`
    Size  int    `json:"size"`
    Total int    `json:"total"`
}
```

### internal/output

```go
type Printer struct {
    format  string // "json", "table", "yaml"
    noColor bool
    writer  io.Writer
}

func NewPrinter(format string, noColor bool) *Printer
func (p *Printer) Print(v any) error
func (p *Printer) PrintTable(headers []string, rows [][]string) error
func (p *Printer) PrintDot(status string) string
func DynamicColumns(conditions [][]Condition) []string
```

## Verbose Logging

When `--verbose` is set, the API client logs to stderr:

```
[DEBUG] GET http://localhost:8000/api/hyperfleet/v1/clusters → 200 (45ms)
[DEBUG] POST http://localhost:8000/api/hyperfleet/v1/clusters → 201 (128ms)
[DEBUG] GET http://localhost:8000/api/hyperfleet/v1/clusters/abc-123 → 404 (12ms)
```

Format: `[DEBUG] <METHOD> <URL> → <STATUS> (<DURATION>)`

## Risks / Trade-offs

- **[Risk] Generic methods require Go 1.18+** → The go.mod already targets
  1.22+, so this is not a concern. Documented for clarity.
- **[Risk] `map[string]any` loses type safety for spec/labels** → Mitigated by
  helper functions for commonly accessed fields. Acceptable because the API
  schema for these fields is intentionally open.
- **[Risk] No retry/backoff in HTTP client** → Acceptable for Phase 02. Retry
  logic can be layered on without changing the interface. The bash scripts had
  no retry either.
- **[Trade-off] Colored output depends on ANSI terminal** → `--no-color` flag
  provides an escape hatch. The `NO_COLOR` environment variable convention will
  also be respected.
