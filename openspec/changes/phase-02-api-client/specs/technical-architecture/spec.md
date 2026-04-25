# Delta for technical-architecture

## MODIFIED Requirements

### Requirement: Shared API Client Package (internal/api)

The CLI SHALL provide a shared HTTP client for all HyperFleet API operations.

#### Scenario: API client capabilities

- GIVEN the `internal/api` package exists
- WHEN any command needs to call the HyperFleet API
- THEN the client MUST provide:
  - Base URL construction from config (`{api-url}/api/hyperfleet/{api-version}/`)
  - Generic typed methods: `Get[T]`, `Post[T]`, `Patch[T]`, `Delete[T]` using Go type parameters
  - Authentication via Bearer token from config (omitted when token is empty)
  - Automatic JSON marshaling/unmarshaling with `encoding/json`
  - RFC 7807 Problem Details error parsing with structured `APIError` type implementing `error`
  - Request/response logging when `--verbose` is set (format: `[DEBUG] METHOD URL → STATUS (DURATIONms)`)
  - Default timeout of 30 seconds via `http.Client.Timeout`
  - Context propagation for cancellation via `http.NewRequestWithContext`

#### Scenario: API error handling

- GIVEN the API returns a non-2xx response
- WHEN the client parses the response
- THEN it MUST return a structured `APIError` type containing code, detail, status, title, trace_id, type, timestamp
- AND the error MUST implement Go's `error` interface with format `[{status}] {title}: {detail}`
- AND commands MUST be able to output the raw error JSON (exit 0) or propagate as a Go error
- AND non-JSON error responses MUST be wrapped in an `APIError` with the raw body as `detail`

### Requirement: Shared Output Package (internal/output)

The CLI SHALL provide a shared output formatting package supporting multiple formats.

#### Scenario: Output format dispatch

- GIVEN the `--output` flag is set
- WHEN a command produces output
- THEN the output package MUST dispatch to the appropriate formatter:
  - `json`: pretty-printed JSON with 2-space indentation and trailing newline
  - `table`: formatted table with uppercase headers and aligned columns via `text/tabwriter`
  - `yaml`: YAML serialization via `gopkg.in/yaml.v3`
- AND the default format MUST be determined per command (table for list views, json for get views)

#### Scenario: Dynamic column table rendering

- GIVEN a table output is requested for cluster or nodepool resources
- WHEN conditions vary across resources
- THEN the table renderer MUST:
  - Collect all unique condition types across all items
  - Order columns: fixed columns first, then `Available`, then alphabetical adapter conditions, then `Ready` last
  - Render status values as colored dots: green `●`=True, red `●`=False, yellow `●`=Unknown, `-`=absent
  - Respect `--no-color` flag and `NO_COLOR` environment variable to disable ANSI colors
  - In no-color mode, render status as plain text: `True`, `False`, `Unknown`, `-`

#### Scenario: Status dot rendering

- GIVEN colored output is enabled
- WHEN a condition status is rendered
- THEN `True` MUST render as a green dot character (`●`)
- AND `False` MUST render as a red dot character (`●`)
- AND `Unknown` MUST render as a yellow dot character (`●`)
- AND absent conditions MUST render as `-`

### Requirement: Shared Resource Types Package (internal/resource)

The CLI SHALL define shared Go types for all HyperFleet resources.

#### Scenario: Core resource types

- GIVEN the `internal/resource` package exists
- WHEN resource types are defined
- THEN the package MUST include:
  - `Cluster` struct with fields: ID, Kind, Name, Generation, Labels, Spec, Status, CreatedBy, CreatedTime, UpdatedBy, UpdatedTime, DeletedBy, DeletedTime, Href
  - `NodePool` struct with fields: ID, Kind, Name, Generation, Labels, Spec, Status, OwnerReferences, CreatedBy, CreatedTime, UpdatedBy, UpdatedTime, DeletedBy, DeletedTime
  - `AdapterStatus` struct with fields: Adapter, Conditions, ObservedGeneration, LastReportTime, Data, CreatedTime
  - `Condition` struct with fields: Type, Status, Reason, Message, LastTransitionTime, ObservedGeneration
  - `CloudEvent` struct with fields: SpecVersion, Type, Source, ID, Data
  - Generic `ListResponse[T]` with fields: Items, Kind, Page, Size, Total
  - `ResourceStatus` struct with field: Conditions (shared between Cluster and NodePool)
  - `OwnerReference` struct with fields: Kind, ID
- AND all types MUST have JSON struct tags matching the API field names (snake_case)
- AND `Spec` and `Labels` MUST use `map[string]any` to support provider-specific content
- AND `ListResponse[T]` MUST use Go type parameters for type-safe list operations
