# Tasks: Phase 05 — Adapter Status

## 1. OpenSpec Change Folder
- [x] 1.1 Create .openspec.yaml
- [x] 1.2 Write proposal.md
- [x] 1.3 Write design.md
- [x] 1.4 Write tasks.md (this file)

## 2. Resource Type Updates

- [x] 2.1 Extend `resource.ConditionRequest` in `internal/resource/condition.go`:
      add `LastTransitionTime string \`json:"last_transition_time,omitempty"\``
- [x] 2.2 Extend `resource.AdapterStatusCreateRequest` in `internal/resource/adapter.go`:
      add `CreatedTime string \`json:"created_time,omitempty"\`` and `LastReportTime string \`json:"last_report_time,omitempty"\``
      NOTE: also changed `Data` tag from `omitempty` to required (empty map must be serialized as `{}`)
- [x] 2.3 Confirm existing unit tests in `internal/resource/types_test.go` still pass after struct additions

## 3. cmd/adapter.go

- [x] 3.1 Create `cmd/adapter.go` with package declaration and imports
- [x] 3.2 Implement `buildAdapterPostPayload(adapter, status string, generation int32) resource.AdapterStatusCreateRequest`
      — constructs payload with 3 conditions (Available, Applied, Health), `data: {}`, `created_time`, `last_report_time` all set to `time.Now().UTC().Format(time.RFC3339)`
- [x] 3.3 Implement `clusterAdapterCmd` and `clusterAdapterPostStatusCmd` (`RangeArgs(2,3)`)
      — validates status in {True, False, Unknown}, resolves cluster-id, POSTs, prints response
- [x] 3.4 Implement `nodepoolAdapterCmd` and `nodepoolAdapterPostStatusCmd` (`RangeArgs(2,4)`)
      — validates status, resolves cluster-id + nodepool-id (4th arg or state), POSTs, prints response
- [x] 3.5 Add `init()` registering both subtrees:
      `clusterCmd.AddCommand(clusterAdapterCmd)` and `nodepoolCmd.AddCommand(nodepoolAdapterCmd)`

## 4. Unit Tests

- [x] 4.1 `cmd/adapter_test.go`: `TestClusterAdapterPostStatus_PostsCorrectURL`
      — httptest.Server; assert POST URL, adapter field, 3 conditions with status="True", `data:{}`
- [x] 4.2 `cmd/adapter_test.go`: `TestNodePoolAdapterPostStatus_PostsCorrectURL`
      — httptest.Server; assert POST URL includes `/nodepools/{nid}/adapter-statuses`
- [x] 4.3 `cmd/adapter_test.go`: `TestAdapterPostStatus_InvalidStatus_ReturnsError`
      — pass status="maybe"; assert error returned, no HTTP request made
- [x] 4.4 `cmd/adapter_test.go`: additional validation tests (invalid generation, no cluster-id, no nodepool-id)

## 5. Delta Spec

- [x] 5.1 Write `specs/adapter-status/spec.md` (MODIFIED) documenting:
      - Go command signatures
      - Payload contract (3-condition array, `created_time`, `last_report_time`)
      - Status validation rule
      - ID resolution rules
      - Struct changes to `ConditionRequest` and `AdapterStatusCreateRequest`

## 6. Verify

- [x] 6.1 `go build ./...` succeeds
- [x] 6.2 `go vet ./...` reports no issues
- [x] 6.3 `go test ./...` passes — output saved to `verification_proof/tests.txt`
- [x] 6.4 `hf cluster adapter post-status cl-deployment True 1` → 201 response with conditions
      — captured to `verification_proof/6.4-cluster-post-true.txt`
- [x] 6.5 `hf cluster adapter post-status cl-job False 1` → 201 with False conditions
      — captured to `verification_proof/6.5-cluster-post-false.txt`
- [x] 6.6 `hf cluster adapter post-status cl-maestro Unknown 1` → 204 No Content (API behaviour for Unknown)
      — captured to `verification_proof/6.6-cluster-post-unknown.txt`
- [x] 6.7 `hf nodepool adapter post-status np-configmap True 1` → nodepool adapter reported; Ready converges
      — captured to `verification_proof/6.7-nodepool-post-true.txt`
- [x] 6.8 `hf cluster conditions table` after posting all required adapters True → Ready=True, Available=True
      — captured to `verification_proof/6.8-conditions-table.txt`
- [x] 6.9 `hf cluster adapter post-status cl-deployment maybe` → error "must be one of: True, False, Unknown"
      — captured to `verification_proof/6.9-invalid-status.txt`
