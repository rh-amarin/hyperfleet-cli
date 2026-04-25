# Proposal: Phase 02 — Shared API Client & Resource Types

## Why

All command phases from Phase 03 onward (cluster, nodepool, adapter-status, tables,
database, maestro, pubsub) depend on three foundational packages that do not yet
exist: an HTTP client for the HyperFleet API, Go struct definitions for all API
resources, and an output formatter that dispatches across json/table/yaml. Without
these shared packages, every subsequent phase would have to duplicate HTTP plumbing,
type definitions, and rendering logic. Building them now as a single coherent layer
unblocks parallel development of Phases 03–13.

## What Changes

- **New `internal/api` package**: Generic HTTP client wrapping `net/http` with typed
  request/response methods (`Get[T]`, `Post[T]`, `Patch[T]`, `Delete`), RFC 7807
  error parsing, Bearer token auth, verbose request logging, configurable timeout,
  and context propagation.
- **New `internal/resource` package**: Go struct definitions for `Cluster`, `NodePool`,
  `AdapterStatus`, `Condition`, `CloudEvent`, and generic `ListResponse[T]`, all with
  JSON tags matching the HyperFleet API field names. `Spec` and `Labels` use
  `map[string]any` for extensibility.
- **New `internal/output` package**: Printer that dispatches on the `--output` flag
  (json/table/yaml), colored dot renderer for condition statuses (True=green,
  False=red, Unknown=yellow, absent=`-`), dynamic column builder for condition
  tables, and `--no-color` support.
- **No new CLI commands** — this phase produces only internal libraries consumed by
  later phases.

## Capabilities

### New Capabilities
- `api-client`: Shared HTTP client for all HyperFleet API operations, including
  RFC 7807 error handling, auth, verbose logging, and generic typed methods.
- `resource-types`: Go struct definitions for Cluster, NodePool, AdapterStatus,
  Condition, CloudEvent, and ListResponse — the data model layer.
- `output-formatting`: Multi-format output printer (json/table/yaml) with colored
  dot rendering and dynamic condition-column tables.

### Modified Capabilities
- `technical-architecture`: ADDED requirements for the api, resource, and output
  package contracts now that the concrete Go types and method signatures are defined.

## Canonical API Contract

All resource types MUST conform to the HyperFleet OpenAPI 3.0 specification at:
[openshift-hyperfleet/hyperfleet-api-spec/schemas/core/openapi.yaml](https://github.com/openshift-hyperfleet/hyperfleet-api-spec/blob/main/schemas/core/openapi.yaml)

Key alignment points verified against v1.0.7 of the spec:
- `Labels` → `map[string]string` (not `map[string]any`)
- `OwnerReferences` → single `ObjectReference` (not array)
- `ResourceCondition` vs `AdapterCondition` are distinct types with different fields and status enums
- `AdapterStatus` includes `metadata` object with job execution fields
- `Error` (RFC 9457) includes `errors: []ValidationError` for field-level validation
- Pagination fields (`page`, `size`, `total`) are `int32`

## Impact

- **New packages**: `internal/api/`, `internal/resource/`, `internal/output/`
- **Dependencies**: No new external dependencies — uses only `net/http`, `encoding/json`,
  `text/tabwriter`, `gopkg.in/yaml.v3` (already in go.mod), and `os` (ANSI colors).
- **Existing code**: `cmd/root.go` will be updated in Phase 03 to initialize the API
  client in `PersistentPreRunE` using `cfgStore.Cfg().Hyperfleet`; this phase does
  not modify `cmd/` files.
- **Testing**: Unit tests for JSON round-trip on resource types, RFC 7807 error
  parsing, and column-ordering algorithm. All testable without a live cluster.
