# Kubernetes Utilities — Delta Spec (Phase 10)

## Changes from openspec/specs/kubernetes/spec.md

### ADDED: Go Implementation via client-go

The Kubernetes utilities are implemented in Go using `k8s.io/client-go` — no `kubectl` binary dependency.

**Package `internal/kube`:**
- `BuildConfig(kubeconfigPath string) (*rest.Config, error)` — resolves: arg → `KUBECONFIG` env → `~/.kube/config`
- `NewClientset(kubeconfigPath string) (*kubernetes.Clientset, error)`
- `CurrentContext(kubeconfigPath string) (string, error)` — reads current-context field from kubeconfig
- `PortForward{Service, LocalPort, RemotePort, PIDFile string, PID int}` — represents an active port-forward
- `StartPortForward(ctx, cs, namespace, service string, localPort, remotePort int) (*PortForward, error)` — starts SPDY port-forward, writes PID to `~/.config/hf/pf-<service>.pid`
- `StopPortForward(service string) error` — reads PID file, sends SIGTERM, removes file
- `ListPortForwards() ([]PortForward, error)` — reads all `~/.config/hf/pf-*.pid` files
- `StreamLogs(ctx, cs, namespace, podPattern string, w io.Writer) error` — fans out log streaming across pods matching pattern (substring), prefixes each line with pod name

**Commands:**
- `hf kube port-forward start <service> <localPort:remotePort>`
- `hf kube port-forward stop <service>`
- `hf kube port-forward status` — table: SERVICE | LOCAL_PORT | REMOTE_PORT | PID | STATUS
- `hf kube context` — prints current context name
- `hf kube curl <service> <path>` — starts port-forward, GETs localhost:<port>/<path>, prints body
- `hf kube debug <deployment>` — execs shell into first running pod matching deployment name
- `hf logs <pattern>` — streams logs from pods matching pattern in configured namespace
- `hf logs adapter <pattern>` — same but filters for cluster-id in log lines

**Flags (persistent on `hf kube` and `hf logs`):**
- `--kubeconfig` — path to kubeconfig (default: `KUBECONFIG` env → `~/.kube/config`)
- `--namespace` / `-n` — Kubernetes namespace (default: `amarin-ns1`)

### MODIFIED: Port-Forward PID Storage

PID files are stored at `~/.config/hf/pf-<service>.pid` (reuses HyperFleet config dir).
Format: plain text, decimal PID.
