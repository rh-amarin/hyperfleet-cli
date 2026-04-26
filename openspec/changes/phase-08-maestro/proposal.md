# Proposal: Phase 08 — Maestro

## Intent

Add `hf maestro` subcommands for managing Maestro resources via the Maestro HTTP API. Maestro is a separate service (different base URL from the HyperFleet API) used by HyperFleet adapters to deploy Kubernetes manifests to managed clusters. The `hf maestro tui` command delegates to the `maestro-cli` binary via `syscall.Exec`.

## Scope In

- `internal/maestro/` — a self-contained HTTP client package (net/http, no reuse of internal/api) with its own base URL
- `cmd/maestro.go` — six subcommands: list, get, delete, bundles, consumers, tui
- Unit tests for both packages using httptest.NewServer

## Scope Out

- Maestro gRPC operations (separate endpoint, future phase)
- Interactive selection menus for get/delete with no name arg (prompt-style TUI interaction deferred)
- Maestro namespace/deployment management

## Config Keys

| Key | Type | Used For |
|-----|------|---------|
| `maestro.http-endpoint` | string | Maestro API base URL |
| `maestro.consumer` | string | Consumer filter for resource listing |
| `hyperfleet.token` | string | Bearer auth token (shared) |

## Testing Scope

### internal/maestro

- `TestClientList_FiltersConsumer` — GET with consumer_name query param
- `TestClientGet_ReturnsResource` — GET /resources/<name>
- `TestClientDelete_SendsDELETE` — DELETE /resources/<name>
- `TestClientListBundles_ReturnsItems` — GET /resource-bundles
- `TestClientListConsumers_ReturnsItems` — GET /consumers

### cmd/maestro

- `TestMaestroList_RendersTable`
- `TestMaestroGet_PrintsJSON`
- `TestMaestroDelete_WithYesFlag_Deletes`
- `TestMaestroDelete_NoBinaryDeletion_Cancelled`
- `TestMaestroBundles_RendersTable`
- `TestMaestroConsumers_RendersTable`
- `TestMaestroTUI_MissingBinary_ReturnsError`

## Verification Steps Requiring Live Access

- `hf maestro consumers` — requires a reachable Maestro endpoint
- `hf maestro list` — requires configured consumer
- `hf maestro bundles` — requires configured endpoint
