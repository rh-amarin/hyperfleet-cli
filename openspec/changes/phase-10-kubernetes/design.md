# Design: Phase 10 — Kubernetes

## Command Tree

```
hf kube port-forward start [name | <service> <localPort:remotePort>]
hf kube port-forward stop [name]
hf kube port-forward status
hf kube context
hf kube curl [--] [curl-flags...] <url>
hf kube debug <deployment>
hf logs [pattern]
hf logs adapter [pattern] [--cluster-id <id>]
```

## New Files

- `internal/kube/client.go` — package `kube`
- `cmd/kube.go` — package `cmd`, registers `kubeCmd` under `rootCmd`
- `cmd/logs.go` — package `cmd`, registers `logsCmd` under `rootCmd`

## Package: internal/kube

### Kubeconfig Resolution

All functions resolve kubeconfig via: flag arg → `KUBECONFIG` env → `~/.kube/config`.
`HF_KUBE_TOKEN` env overrides bearer token (bypasses gke-gcloud-auth-plugin exec provider).

```go
func BuildConfig(kubeconfigPath string) (*rest.Config, error)
func NewClientset(kubeconfigPath string) (*kubernetes.Clientset, error)
func CurrentContext(kubeconfigPath string) (string, error)
```

### Port-Forward Management

PID files stored at `~/.config/hf/pf-<name>.pid`.
Format: `"<pid>\n<localPort>\n<remotePort>"` (three lines).

```go
type PortForward struct {
    Service    string
    LocalPort  int
    RemotePort int
    PIDFile    string
    PID        int
}

// name = PID file key and display name
// podPattern = pod search substring (falls back to name if empty)
func StartPortForward(kubeconfigPath, namespace, name, podPattern string, localPort, remotePort int) (*PortForward, error)
func StopPortForward(service string) error
func ListPortForwards() ([]PortForward, error)
func IsProcessAlive(pid int) bool
```

`StartPortForward` forks the current binary as `hf kube _pf-daemon <podPattern> <localPort:remotePort>`.
The daemon uses SPDY port-forward via `k8s.io/client-go/tools/portforward`.

`ListPortForwards` reads all `pf-*.pid` files and reports alive/dead status.

Predefined services (in `cmd/kube.go`):
| name          | podPattern   | namespace    | localPort | remotePort |
|---------------|-------------|--------------|-----------|------------|
| hyperfleet-api | hyperfleet-api | `amarin-ns1` | 8000 | 8000 |
| postgresql     | postgresql   | `amarin-ns1` | 5432 | 5432 |
| maestro-http   | maestro      | `maestro`    | 8100 | 8000 |
| maestro-grpc   | maestro      | `maestro`    | 8090 | 8090 |

Port values override via `cfg.PortForward.*` from `~/.config/hf/config.yaml`.

### Log Streaming

```go
func StreamLogs(ctx context.Context, cs kubernetes.Interface, namespace, podPattern string, w io.Writer) error
func StreamLogsFiltered(ctx context.Context, cs kubernetes.Interface, namespace, podPattern, clusterID string, w io.Writer) error
func parseLogfmt(line string) map[string]string
```

`StreamLogsFiltered` skips JSON/OTel lines (starting with `{`), filters by `cluster_id=<clusterID>`,
and formats matching lines as `[pod] time  LEVEL  msg`.

### In-Cluster Curl Pod

```go
func RunCurlPod(ctx context.Context, kubeconfigPath, namespace string, curlArgs []string, w io.Writer) error
```

Creates or reuses a persistent pod named `hf-curl` running `curlimages/curl:latest` with `sleep 3600`.
Execs the curl command via SPDY remotecommand. If the pod is stale (not Running), it is deleted and recreated.

### Debug Pod

```go
func CreateDebugPod(ctx context.Context, cs kubernetes.Interface, namespace, pattern string) (string, error)
```

Finds a deployment matching `pattern` (substring), deep-copies its pod template,
removes liveness/readiness probes, sets `restartPolicy: Never`, creates pod
`hf-debug-<deployment-name>-<unix-timestamp>`, waits up to 3 minutes for Running.

## Package: cmd

### cmd/kube.go

Persistent flags (via `kubeCmd`):
- `--kubeconfig` — path to kubeconfig file
- `--namespace` / `-n` — default `amarin-ns1`

`hf kube port-forward start [name | <service> <localPort:remotePort>]`:
- 0 args → start all 4 predefined services
- 1 arg → start named predefined service
- 2 args → generic: `kube.StartPortForward(kubeconfig, namespace, service, "", localPort, remotePort)`

`hf kube port-forward stop [name]`:
- 0 args → stop all active port-forwards
- 1 arg → `kube.StopPortForward(name)`

`hf kube port-forward status`:
- `kube.ListPortForwards()`, print bullet list with ANSI color (green=alive, red=dead)
- Format: `  ● name - localhost:port (PID: pid)`

`hf kube context`:
- `kube.CurrentContext(kubeconfigPath)`, print to stdout

`hf kube curl [--] [curl-flags...] <url>`:
- Creates/reuses `hf-curl` pod, execs curl inside it
- Curl flags (e.g., `-s`, `-H`, `-v`) must be preceded by `--` to avoid Cobra flag parsing

`hf kube debug <deployment>`:
- `kube.CreateDebugPod(ctx, cs, namespace, args[0])`
- Prints exec command: `kubectl exec -it <pod> -n <ns> -- /bin/sh`

### cmd/logs.go

`hf logs [pattern]`:
- Uses `stern` if available (passes `pattern`, `-n`, `--kubeconfig` to stern)
- Falls back to `kube.StreamLogs(ctx, cs, namespace, pattern, os.Stdout)`

`hf logs adapter [pattern] [--cluster-id <id>]`:
- Search pattern prefixed with `"adapter"` or `"adapter-<pattern>"`
- cluster-id: `--cluster-id` flag → `cfgStore.State().ClusterID` (active cluster)
- `kube.StreamLogsFiltered(ctx, cs, namespace, searchPattern, clusterID, os.Stdout)`

## ID Resolution

kubeconfig: `--kubeconfig` flag → `KUBECONFIG` env → `~/.kube/config` (resolved in `BuildConfig`)

## Key Decisions

| Decision | Choice | Rationale |
|---|---|---|
| No kubectl dependency | client-go SPDY port-forward + remotecommand | Self-contained binary requirement |
| PID file path | `~/.config/hf/pf-<name>.pid` | Matches existing HyperFleet config dir |
| PID file format | `pid\nlocalPort\nremotePort` | Carries enough info for status display |
| name vs podPattern | Separate fields in StartPortForward | maestro-http and maestro-grpc share same pod |
| Port-forward daemon | Fork self as `_pf-daemon` subcommand | No external binary required |
| Curl pod | Persistent `hf-curl` with `sleep 3600` | Avoids pod startup latency on repeated calls |
| `--` for curl flags | Standard pflag flag terminator | Avoids cobra flag conflicts without DisableFlagParsing |
| Default namespace | `amarin-ns1` | Matches live cluster namespace |
| Log prefix | `[pod-name]` prefix per line | Matches stern behavior |
| stern integration | `exec.LookPath("stern")` then exec | Better multi-pod UX when tool is available |
| Unit tests | `k8s.io/client-go/kubernetes/fake` | No live cluster required |
