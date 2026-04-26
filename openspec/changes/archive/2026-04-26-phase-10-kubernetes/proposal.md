# Proposal: Phase 10 — Kubernetes

## Intent

Port-forward lifecycle, pod log streaming, debug pod exec, and kube context — all via `client-go`. No `kubectl` required.

## Scope In

- `internal/kube` package: `BuildConfig`, `NewClientset`, `CurrentContext`, port-forward PID helpers, `StreamLogs`
- `cmd/kube.go`: `hf kube port-forward start/stop/status`, `hf kube context`, `hf kube curl`, `hf kube debug`
- `cmd/logs.go`: `hf logs <pattern>`, `hf logs adapter <pattern>`

## Scope Out

- Any API (HyperFleet) interactions
- Database or Maestro commands
- Shell completions

## Testing Scope

| Package | Test Cases |
|---|---|
| `internal/kube` | `BuildConfig` path resolution (arg → env → default), `ListPortForwards` PID file parsing, `StreamLogs` pod pattern matching using fake client |
| `cmd` | `hf kube context`, `hf kube port-forward status` (empty), `hf logs <pattern>` with fake pod list |

## Live Verification (requires kubeconfig)

- `hf kube context` — prints cluster context name
- `hf kube port-forward status` — lists running port-forwards (or empty table)
- `hf logs adapter "" --namespace amarin-ns1` — streams/lists adapter pods
