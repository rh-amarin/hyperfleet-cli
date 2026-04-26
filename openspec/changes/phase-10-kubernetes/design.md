# Design: Phase 10 — Kubernetes

## Command Tree

```
hf kube port-forward start <service> <localPort:remotePort>
hf kube port-forward stop <service>
hf kube port-forward status
hf kube context
hf kube curl <service> <path>
hf kube debug <deployment>
hf logs <pattern>
hf logs adapter <pattern>
```

## New Files

- `internal/kube/client.go` — package `kube`
- `cmd/kube.go` — package `cmd`, registers `kubeCmd` under `rootCmd`
- `cmd/logs.go` — package `cmd`, registers `logsCmd` under `rootCmd`

## Package: internal/kube

### Kubeconfig Resolution

All functions resolve kubeconfig via: flag arg → `KUBECONFIG` env → `~/.kube/config`

```go
func BuildConfig(kubeconfigPath string) (*rest.Config, error)
func NewClientset(kubeconfigPath string) (*kubernetes.Clientset, error)
func CurrentContext(kubeconfigPath string) (string, error)
```

### Port-Forward Management

PID files stored at `~/.config/hf/pf-<service>.pid`.

```go
type PortForward struct {
    Service    string
    LocalPort  int
    RemotePort int
    PIDFile    string
    PID        int
}

func StartPortForward(ctx context.Context, cs *kubernetes.Clientset, namespace, service string, localPort, remotePort int) (*PortForward, error)
func StopPortForward(service string) error
func ListPortForwards() ([]PortForward, error)
```

`StartPortForward` locates a pod matching `service` (substring), starts an SPDY port-forward tunnel via client-go, forks a subprocess that runs the tunnel, and writes the subprocess PID to the PID file.

`StopPortForward` reads the PID file, sends SIGTERM, and removes the PID file.

`ListPortForwards` reads all `pf-*.pid` files from `~/.config/hf/`, parses the service name and PID, and checks if the process is alive.

### Log Streaming

```go
func StreamLogs(ctx context.Context, cs *kubernetes.Clientset, namespace, podPattern string, w io.Writer) error
```

Lists pods in `namespace`, filters by substring match on `podPattern`, then fans out log streaming across matched pods. Each line is prefixed with the pod name. Uses `Follow: true` for streaming.

## Package: cmd

### cmd/kube.go

Persistent flags (via `kubeCmd`):
- `--kubeconfig` — path to kubeconfig file
- `--namespace` / `-n` — default `amarin-ns1`

Pattern: package-level `var kubeCmd`, registered in `init()` via `rootCmd.AddCommand(kubeCmd)`.

`hf kube port-forward start <service> <localPort:remotePort>`:
- Parse `localPort:remotePort` from arg
- Call `kube.StartPortForward`

`hf kube port-forward stop <service>`:
- Call `kube.StopPortForward(service)`

`hf kube port-forward status`:
- Call `kube.ListPortForwards()`, print table: SERVICE | LOCAL_PORT | REMOTE_PORT | PID | STATUS

`hf kube context`:
- Call `kube.CurrentContext(kubeconfigPath)`, print to stdout

`hf kube curl <service> <path>`:
- Start port-forward, run HTTP GET against `localhost:<localPort>/<path>`, print response body

`hf kube debug <deployment>`:
- List pods matching deployment name, exec shell into first running pod

### cmd/logs.go

`hf logs <pattern>`:
- Call `kube.StreamLogs(ctx, cs, namespace, pattern, os.Stdout)`

`hf logs adapter <pattern>`:
- Same but additionally filters log lines matching cluster-id from config

## ID Resolution

kubeconfig: `--kubeconfig` flag → `KUBECONFIG` env → `~/.kube/config` (resolved in `BuildConfig`)

## Key Decisions

| Decision | Choice | Rationale |
|---|---|---|
| No kubectl dependency | client-go SPDY port-forward | Self-contained binary requirement |
| PID file path | `~/.config/hf/pf-<service>.pid` | Matches existing HyperFleet config dir |
| Default namespace | `amarin-ns1` | Matches live cluster namespace |
| Log prefix | `[pod-name]` prefix per line | Matches stern behavior the bash scripts relied on |
| Unit tests | `k8s.io/client-go/kubernetes/fake` | No live cluster required |
