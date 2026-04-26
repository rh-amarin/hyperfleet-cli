# Design: Phase 08 — Maestro

## Package Layout

```
internal/maestro/
├── client.go       ← HTTP client + types (Resource, Bundle, Consumer)
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

| Method | Path | Notes |
|--------|------|-------|
| List | GET `/api/maestro/v1/resources?consumer_name=<consumer>` | consumer param omitted when empty |
| Get | GET `/api/maestro/v1/resources/<name>` | |
| Delete | DELETE `/api/maestro/v1/resources/<name>` | |
| ListBundles | GET `/api/maestro/v1/resource-bundles` | per canonical spec |
| ListConsumers | GET `/api/maestro/v1/consumers` | |

Note: The task description mentioned `/api/maestro/v1/manifest-bundles` but the canonical spec (`openspec/specs/maestro/spec.md`) specifies `/api/maestro/v1/resource-bundles`. Canonical spec takes precedence.

### Response Types

```go
type Resource struct {
    ID            string      `json:"id"`
    Kind          string      `json:"kind"`
    Name          string      `json:"name"`
    ConsumerName  string      `json:"consumer_name"`
    Version       int         `json:"version"`
    ManifestCount int         `json:"manifest_count"`
    Manifests     []Manifest  `json:"manifests"`
    Conditions    []Condition `json:"conditions"`
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

type Bundle struct {
    ID              string            `json:"id"`
    Kind            string            `json:"kind"`
    Name            string            `json:"name"`
    Labels          map[string]string `json:"labels"`
    Manifests       []any             `json:"manifests"`
    ManifestConfigs []any             `json:"manifest_configs"`
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
| list | table (NAME, CONSUMER, VERSION, MANIFESTS, CONDITIONS) |
| get | JSON |
| delete | success message to stderr |
| bundles | JSON (pass-through from API) |
| consumers | table (ID, NAME) |
| tui | process replaced via syscall.Exec |

### Delete Confirmation

```
Delete Maestro resource '<name>'? [y/N]: 
```

Read from `os.Stdin`. The `--yes` / `-y` flag skips the prompt.

### TUI Exec

```go
path, err := exec.LookPath("maestro-cli")
if err != nil {
    return fmt.Errorf("maestro-cli not found in PATH; install with:\n  oc apply -f .../maestro-cli-deployment.yaml")
}
endpoint := cfgStore.Cfg().Maestro.HTTPEndpoint
return syscall.Exec(path, []string{"maestro-cli", "tui", "--api-server=" + endpoint}, os.Environ())
```

## Test Strategy

cmd tests use a helper `runMaestroCmd` that writes a `config.yaml` to the temp dir with `maestro.http-endpoint` set to the test server URL, then calls `runCmdRaw`.

## Key Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Separate HTTP client | `internal/maestro.Client` | Maestro has its own base URL, independent auth model |
| Bundles endpoint | `/resource-bundles` | Canonical spec (`openspec/specs/maestro/spec.md`) takes precedence over task description |
| TUI exec method | `syscall.Exec` | Replaces the process cleanly; no zombie parent process |
| Delete confirmation | `--yes` flag to skip | Enables scriptability and testability |
| Error format | Plain `fmt.Errorf("HTTP %d: ...")` | Maestro error shape differs from HyperFleet RFC 7807 |
