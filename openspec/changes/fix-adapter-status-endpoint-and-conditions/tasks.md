# Tasks: Fix adapter-status endpoint path and missing Finalized condition

- [x] 1. Fix cluster GET path: `adapter-statuses` → `statuses` in `cmd/cluster.go`
- [x] 2. Fix nodepool GET path: `adapter-statuses` → `statuses` in `cmd/nodepool.go`
- [x] 3. Add Finalized condition to `buildAdapterPostPayload` in `cmd/adapter.go`
- [x] 4. Update adapter_test.go condition-count assertions from 3 to 4
- [x] 5. Update `TestClusterAdapterPostStatus_AllThreeConditionTypes` to check all 4 types
- [x] 6. Run `go test ./cmd/...` and save output to `verification_proof/`
- [x] 7. Verify against live API and save output to `verification_proof/`
