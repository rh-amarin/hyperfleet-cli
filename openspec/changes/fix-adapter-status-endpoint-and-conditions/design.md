# Design: Fix adapter-status endpoint path and missing Finalized condition

## Changes

### 1. cmd/cluster.go — GET endpoint path
Replace `"clusters/"+clusterID+"/adapter-statuses"` with `"clusters/"+clusterID+"/statuses"`.

### 2. cmd/nodepool.go — GET endpoint path
Replace `"clusters/"+clusterID+"/nodepools/"+nodepoolID+"/adapter-statuses"` with
`"clusters/"+clusterID+"/nodepools/"+nodepoolID+"/statuses"`.

### 3. cmd/adapter.go — add Finalized condition
Add `{Type: "Finalized", Status: status, Reason: "ManualStatusPost", Message: msg}` to
the `Conditions` slice in `buildAdapterPostPayload`.

### 4. cmd/adapter_test.go — update tests
- `TestClusterAdapterPostStatus_PayloadShape`: `len != 3` → `len != 4`
- `TestNodePoolAdapterPostStatus_PayloadShape`: `len != 3` → `len != 4`
- `TestClusterAdapterPostStatus_AllThreeConditionTypes`: rename to
  `TestClusterAdapterPostStatus_AllFourConditionTypes` and add `Finalized` to the
  expected-types list.

No struct changes needed — `ConditionRequest` already supports the required fields.
