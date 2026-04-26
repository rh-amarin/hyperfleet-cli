# Design: Phase 08 — Maestro

## Package Layout

```
internal/maestro/
├── client.go       ← HTTP client + types (Resource, Consumer)
└── client_test.go  ← httptest-based unit tests

cmd/
├── maestro.go      ← six subcommands registered to rootCmd
└── maestro_test.go ← command-level unit tests
```

## internal/maestro.Client

The Maestro HTTP client is completely separate from `internal/api` because Maestro has its own base URL.

```go
type Client struct {
    httpClient *http.Client
    baseURL    string   // "<http-endpoint>/api/maestro/v1/"
    consumer   string   // maestro.consumer config value
    token      string   // hyperfleet.token config value
}

func New(httpEndpoint, consumer, token string) *Client
```

### Endpoints

Verified against a live Maestro deployment (EU cluster, 2026-04-26). The `/resources` endpoint does not exist; all resource operations use `/resource-bundles`.

| Method | Path | Notes |
|--------|------|-------|
| List | GET `/api/maestro/v1/resource-bundles?search=consumer_name = '<consumer>'` | SQL-like search; param omitted when consumer is empty |
| Get | GET `/api/maestro/v1/resource-bundles/<id>` | id is a UUID |
| Delete | DELETE `/api/maestro/v1/resource-bundles/<id>` | id is a UUID |
| ListBundles | GET `/api/maestro/v1/resource-bundles` | unfiltered, used by `maestro bundles` |
| ListConsumers | GET `/api/maestro/v1/consumers` | |

### Response Types

Resource name is stored in `metadata["name"]`, not a top-level field. This matches the real Maestro API shape.

```go
type Resource struct {
    ID            string            `json:"id"`
    Kind          string            `json:"kind"`
    Metadata      map[string]string `json:"metadata"`   // name at Metadata["name"]
    ConsumerName  string            `json:"consumer_name"`
    Version       int               `json:"version"`
    ManifestCount int               `json:"manifest_count"`
    Manifests     []Manifest        `json:"manifests"`
    Conditions    []Condition       `json:"conditions"`
}

type Manifest struct {
    Kind      string `json:"kind"`
    Name      string `json:"name"`
    Namespace string `json:"namespace"`
}

type Condition struct {
    Type   string `json:"type"`
    Status string `json:"status"`
    Reason string `json:"reason"`
}

type Consumer struct {
    ID   string `json:"id"`
    Kind string `json:"kind"`
    Name string `json:"name"`
}
```

### Error Handling

HTTP non-2xx responses return `fmt.Errorf("HTTP %d: %s", statusCode, body)`. This is intentionally simpler than the RFC 7807 parsing in `internal/api` since Maestro may use different error shapes.

## cmd/maestro.go

### Command Registration

```go
func init() { rootCmd.AddCommand(maestroCmd) }

func init() {
    maestroCmd.AddCommand(maestroListCmd)
    maestroCmd.AddCommand(maestroGetCmd)
    maestroCmd.AddCommand(maestroDeleteCmd)
    maestroCmd.AddCommand(maestroBundlesCmd)
    maestroCmd.AddCommand(maestroConsumersCmd)
    maestroCmd.AddCommand(maestroTUICmd)
    maestroDeleteCmd.Flags().BoolP("yes", "y", false, "skip confirmation prompt")
}
```

### Client Construction

```go
cfg := cfgStore.Cfg()
c := maestro.New(cfg.Maestro.HTTPEndpoint, cfg.Maestro.Consumer, cfg.Hyperfleet.Token)
```

### Output Behavior

| Command | Default Output |
|---------|---------------|
| list | table (NAME, CONSUMER, VERSION, MANIFESTS) |
| get | JSON |
| delete | success message to stderr |
| bundles | JSON (pass-through from API, unfiltered) |
| consumers | table (ID, NAME) |
| tui | process replaced via syscall.Exec |

### Delete Confirmation

```
Delete Maestro resource bundle '<id>'? [y/N]: 
```

Read from `os.Stdin`. The `--yes` / `-y` flag skips the prompt.

### TUI Exec

`maestro-cli` reads its endpoint from persistent flags `--http-endpoint` and `--grpc-endpoint`, which can also be set via the env vars `MAESTRO_HTTP_ENDPOINT` and `MAESTRO_GRPC_ENDPOINT`. We inject the env vars so the user does not need to pass flags manually.

```go
cfg := cfgStore.Cfg().Maestro
env := append(os.Environ(),
    "MAESTRO_HTTP_ENDPOINT="+cfg.HTTPEndpoint,
    "MAESTRO_GRPC_ENDPOINT="+cfg.GRPCEndpoint,
)
return syscall.Exec(maestroCLI, []string{"maestro-cli", "tui"}, env)
```

## Test Strategy

cmd tests use a helper `runMaestroCmd` that writes a `config.yaml` to the temp dir with `maestro.http-endpoint` set to the test server URL, then calls `runCmdRaw`.

## Key Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Separate HTTP client | `internal/maestro.Client` | Maestro has its own base URL, independent auth model |
| Resource endpoint | `/resource-bundles` | `/resources` does not exist in real deployments; verified live |
| Consumer filter | SQL-like `search=consumer_name = 'X'` | Required by Maestro API; simple `?consumer_name=X` returns 400 |
| Resource name | `metadata["name"]` | Top-level `name` field is absent in real API responses |
| TUI env vars | `MAESTRO_HTTP_ENDPOINT` / `MAESTRO_GRPC_ENDPOINT` | maestro-cli reads these natively; avoids passing unknown flags |
| TUI exec method | `syscall.Exec` | Replaces the process cleanly; no zombie parent process |
| Delete confirmation | `--yes` flag to skip | Enables scriptability and testability |
| Error format | Plain `fmt.Errorf("HTTP %d: ...")` | Maestro error shape differs from HyperFleet RFC 7807 |
