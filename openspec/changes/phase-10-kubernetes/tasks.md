# Tasks: Phase 10 — Kubernetes

## 1. OpenSpec Artifacts
- [x] 1.1 .openspec.yaml
- [x] 1.2 proposal.md
- [x] 1.3 design.md
- [x] 1.4 tasks.md
- [x] 1.5 specs/kubernetes/spec.md (delta)

## 2. Dependencies
- [x] 2.1 go get k8s.io/client-go@v0.33.0
- [x] 2.2 go get k8s.io/api@v0.33.0
- [x] 2.3 go get k8s.io/apimachinery@v0.33.0

## 3. internal/kube/client.go
- [x] 3.1 BuildConfig(kubeconfigPath string) (*rest.Config, error)
- [x] 3.2 NewClientset(kubeconfigPath string) (kubernetes.Interface, error)
- [x] 3.3 CurrentContext(kubeconfigPath string) (string, error)
- [x] 3.4 PortForward struct and pfPIDFile helpers (format: pid\nlocalPort\nremotePort)
- [x] 3.5 StartPortForward(kubeconfigPath, namespace, name, podPattern, localPort, remotePort)
- [x] 3.6 StopPortForward(service string) error
- [x] 3.7 ListPortForwards() ([]PortForward, error)
- [x] 3.8 StreamLogs(ctx, cs, namespace, podPattern, w) error
- [x] 3.9 StreamLogsFiltered(ctx, cs, namespace, podPattern, clusterID, w) error
- [x] 3.10 parseLogfmt(line string) map[string]string
- [x] 3.11 FindRunningPod(ctx, cs, namespace, pattern) (string, error)
- [x] 3.12 RunPortForwardDaemon(kubeconfigPath, namespace, service, localPort, remotePort) error
- [x] 3.13 RunCurlPod(ctx, kubeconfigPath, namespace, curlArgs, w) error
- [x] 3.14 CreateDebugPod(ctx, cs, namespace, pattern) (string, error)
- [x] 3.15 waitForPodPhase helper
- [x] 3.16 execInPod helper (SPDY remotecommand)

## 4. internal/kube/client_test.go
- [x] 4.1 TestBuildConfig_ExplicitPath
- [x] 4.2 TestBuildConfig_EnvFallback
- [x] 4.3 TestBuildConfig_DefaultPath
- [x] 4.4 TestListPortForwards_Empty
- [x] 4.5 TestListPortForwards_ParsesPIDFiles (validates LocalPort in new format)
- [x] 4.6 TestListPortForwards_IgnoresMalformedPIDFiles
- [x] 4.7 TestStreamLogs_NoMatchingPods (fake client)
- [x] 4.8 TestStreamLogs_EmptyPattern_MatchesAll
- [x] 4.9 TestStreamLogs_PatternFilters
- [x] 4.10 TestPIDFilePath_Format
- [x] 4.11 TestParseLogfmt_BasicFields
- [x] 4.12 TestParseLogfmt_QuotedMsgWithSpaces
- [x] 4.13 TestStreamLogsFiltered_SkipsJsonLines

## 5. cmd/kube.go
- [x] 5.1 kubeCmd with --kubeconfig and --namespace/-n flags
- [x] 5.2 hf kube context
- [x] 5.3 hf kube port-forward start [name | <service> <localPort:remotePort>]
- [x] 5.4 hf kube port-forward stop [name]
- [x] 5.5 hf kube port-forward status (bullet format with ANSI color)
- [x] 5.6 hf kube curl [--] [curl-flags...] <url> (in-cluster curl pod)
- [x] 5.7 hf kube debug <deployment> (debug pod from deployment template)
- [x] 5.8 hf kube _pf-daemon (hidden, used by StartPortForward)
- [x] 5.9 predefinedPFs() — 4 services: hyperfleet-api, postgresql, maestro-http, maestro-grpc

## 6. cmd/kube_test.go
- [x] 6.1 TestKubeContext_MissingKubeconfig
- [x] 6.2 TestKubePortForwardStatus_Empty (checks "Port Forward Status" header)
- [x] 6.3 TestKubePFStartCmd_InvalidPorts

## 7. cmd/logs.go
- [x] 7.1 logsCmd with --namespace/-n and --kubeconfig flags
- [x] 7.2 hf logs [pattern] (stern if available, else StreamLogs)
- [x] 7.3 hf logs adapter [pattern] [--cluster-id] (StreamLogsFiltered, auto-resolves clusterID)

## 8. cmd/logs_test.go
- [x] 8.1 TestLogsCmd_MissingKubeconfig
- [x] 8.2 TestLogsAdapterCmd_MissingKubeconfig
- [x] 8.3 TestLogsCmd_NoArgs_UsesEmptyPattern
- [x] 8.4 TestLogsAdapterCmd_PatternPrefixed

## Verify
- [x] (a) `go build ./...` succeeds → verification_proof/build.txt
- [x] (b) `go vet ./...` no issues → verification_proof/vet.txt
- [x] (c) `go test ./...` passes → verification_proof/tests.txt
- [x] (d) Live cluster verification → verification_proof/*.txt
  - hf kube context: prints cluster context name ✓
  - hf kube port-forward start: starts all 4 predefined (hyperfleet-api, postgresql, maestro-http, maestro-grpc) ✓
  - hf kube port-forward status: shows 4 entries with ● bullets, localhost:port, PID ✓
  - hf kube port-forward stop: stops all 4 ✓
  - hf kube curl -- -s <url>: exec curl in hf-curl pod, returns in-cluster response ✓
  - hf kube debug hyperfleet-api: creates hf-debug-* pod, prints exec command ✓
  - hf logs adapter: connects to cluster, streams filtered adapter logs ✓
