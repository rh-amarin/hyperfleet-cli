# Kubernetes Utilities — Delta Spec (Phase 10)

## Changes from openspec/specs/kubernetes/spec.md

### ADDED: Go Implementation via client-go

The Kubernetes utilities are implemented in Go using `k8s.io/client-go` — no `kubectl` binary dependency.

**Package `internal/kube`:**
- `BuildConfig(kubeconfigPath string) (*rest.Config, error)` — resolves: arg → `KUBECONFIG` env → `~/.kube/config`. `HF_KUBE_TOKEN` env overrides bearer token.
- `NewClientset(kubeconfigPath string) (kubernetes.Interface, error)`
- `CurrentContext(kubeconfigPath string) (string, error)` — reads current-context field from kubeconfig
- `PortForward{Service, LocalPort, RemotePort, PIDFile string, PID int}` — represents a port-forward entry
- `StartPortForward(kubeconfigPath, namespace, name, podPattern string, localPort, remotePort int) (*PortForward, error)` — forks self as `_pf-daemon`, writes PID file at `~/.config/hf/pf-<name>.pid`
- `StopPortForward(service string) error` — reads PID file, sends SIGTERM, removes file
- `ListPortForwards() ([]PortForward, error)` — reads all `~/.config/hf/pf-*.pid` files
- `IsProcessAlive(pid int) bool` — checks `/proc/<pid>`
- `FindRunningPod(ctx, cs, namespace, pattern string) (string, error)` — first running pod matching pattern substring
- `RunPortForwardDaemon(kubeconfigPath, namespace, service string, localPort, remotePort int) error` — SPDY port-forward loop, called by `_pf-daemon`
- `StreamLogs(ctx, cs, namespace, podPattern string, w io.Writer) error` — fans out log streaming, prefixes each line with `[pod-name]`
- `StreamLogsFiltered(ctx, cs, namespace, podPattern, clusterID string, w io.Writer) error` — filters by `cluster_id` logfmt field, skips JSON lines
- `RunCurlPod(ctx, kubeconfigPath, namespace string, curlArgs []string, w io.Writer) error` — creates/reuses `hf-curl` pod, execs curl via SPDY
- `CreateDebugPod(ctx, cs, namespace, pattern string) (string, error)` — creates debug pod from deployment template (probes removed, `restartPolicy: Never`)

**Commands:**
- `hf kube port-forward start [name | <service> <localPort:remotePort>]` — 0 args: start 4 predefined; 1 arg: start named predefined; 2 args: generic
- `hf kube port-forward stop [name]` — 0 args: stop all; 1 arg: stop named
- `hf kube port-forward status` — bullet list: `  ● name - localhost:port (PID: pid)` with ANSI green/red
- `hf kube context` — prints current context name
- `hf kube curl [--] [curl-flags...] <url>` — runs curl from ephemeral in-cluster pod; use `--` before curl flags that start with `-`
- `hf kube debug <deployment>` — creates debug pod, prints `kubectl exec` command
- `hf logs [pattern]` — uses `stern` if available, else streams logs via client-go
- `hf logs adapter [pattern] [--cluster-id <id>]` — streams adapter logs filtered by cluster-id

**Predefined port-forward services:**
| name           | pod pattern  | namespace    | local:remote |
|----------------|-------------|--------------|-------------|
| hyperfleet-api | hyperfleet-api | amarin-ns1 | 8000:8000   |
| postgresql     | postgresql   | amarin-ns1   | 5432:5432   |
| maestro-http   | maestro      | maestro      | 8100:8000   |
| maestro-grpc   | maestro      | maestro      | 8090:8090   |

Port values configurable via `cfg.PortForward.*` in `~/.config/hf/config.yaml`.

**Flags (persistent on `hf kube` and `hf logs`):**
- `--kubeconfig` — path to kubeconfig (default: `KUBECONFIG` env → `~/.kube/config`)
- `--namespace` / `-n` — Kubernetes namespace (default: `amarin-ns1`)

### MODIFIED: Port-Forward PID Storage

PID files are stored at `~/.config/hf/pf-<name>.pid` (reuses HyperFleet config dir).
Format: three lines — `<pid>\n<localPort>\n<remotePort>`.
The `name` field is used as the PID file key; `podPattern` is the pod search substring (may differ, e.g., `maestro-http`/`maestro-grpc` both use pod pattern `maestro`).
