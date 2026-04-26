# Tasks: Phase 04 — NodePool Lifecycle

## 1. OpenSpec Change Folder
- [x] 1.1 Create .openspec.yaml
- [x] 1.2 Write proposal.md
- [x] 1.3 Write design.md
- [x] 1.4 Write tasks.md (this file)

## 2. Shared Helper
- [x] 2.1 Create `internal/api/nodepool.go` with `FindNodePoolByName(c, ctx, clusterID, name)` — exact name + non-deleted filter, path scoped to cluster

## 3. cmd/nodepool.go Subcommands
- [x] 3.1 `hf nodepool create [name] [count] [instance-type]` — defaults my-nodepool/1/m4, loop N posts, persists last ID
- [x] 3.2 `hf nodepool list [cluster_id]`
- [x] 3.3 `hf nodepool search <name>` — sets nodepool-id
- [x] 3.4 `hf nodepool get [nodepool_id]`
- [x] 3.5 `hf nodepool patch spec|labels [nodepool_id]` — counter increment, returns error (not os.Exit)
- [x] 3.6 `hf nodepool delete [nodepool_id]`
- [x] 3.7 `hf nodepool id` — print configured nodepool-id
- [x] 3.8 `hf nodepool conditions [-w] [nodepool_id]`
- [x] 3.9 `hf nodepool conditions table [nodepool_id]`
- [x] 3.10 `hf nodepool statuses [-w] [nodepool_id]` — 404 → empty AdapterStatusList
- [x] 3.11 `hf nodepool table [cluster_id]` — dynamic condition columns with DynamicColumns

## 4. Unit Tests — internal/api/nodepool_test.go
- [x] 4.1 TestFindNodePoolByName_ExactMatch
- [x] 4.2 TestFindNodePoolByName_FiltersDeletedNodePools
- [x] 4.3 TestFindNodePoolByName_FiltersPartialNameMatches
- [x] 4.4 TestFindNodePoolByName_NoMatch
- [x] 4.5 TestFindNodePoolByName_MultipleActiveMatches
- [x] 4.6 TestFindNodePoolByName_SearchQueryEncodesClusterID
- [x] 4.7 TestFindNodePoolByName_APIError

## 5. Unit Tests — cmd/nodepool_test.go
- [x] 5.1 TestNodePoolCreate_DefaultArgs — single nodepool with defaults
- [x] 5.2 TestNodePoolCreate_MultipleNodePools — count=2, names suffixed -1/-2, last ID persisted
- [x] 5.3 TestNodePoolList_OutputsJSON
- [x] 5.4 TestNodePoolSearch_SetsNodePoolID
- [x] 5.5 TestNodePoolSearch_NoMatch_WarnAndEmptyArray
- [x] 5.6 TestNodePoolGet_OutputsJSON
- [x] 5.7 TestNodePoolPatch_NoArgs_ExitsNonZero
- [x] 5.8 TestNodePoolPatch_SpecCounter_IncrementsAndPatches
- [x] 5.9 TestNodePoolPatch_LabelsCounter_IncrementsAndPatches
- [x] 5.10 TestNodePoolDelete_OutputsDeletedNodePool
- [x] 5.11 TestNodePoolConditions_OutputsGenerationAndConditions
- [x] 5.12 TestNodePoolConditionsTable_RendersRows
- [x] 5.13 TestNodePoolStatuses_EmptyList
- [x] 5.14 TestNodePoolStatuses_404_TreatedAsEmptyList
- [x] 5.15 TestNodePoolTable_RendersWithDynamicColumns

## 6. Verify
- [x] 6.1 `go build ./...` succeeds
- [x] 6.2 `go vet ./...` reports no issues
- [x] 6.3 `go test ./...` passes — captured to verification_proof/tests.txt (39 tests passing)
- [ ] 6.4 Live verification against real cluster — capture output to verification_proof/<task>.txt
