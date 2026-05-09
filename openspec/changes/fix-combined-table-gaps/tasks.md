# Tasks: Fix Combined Table Gaps

- [x] 1. Add `GenCell` method to `internal/output/printer.go`
- [x] 2. Rewrite `renderCombinedTable` in `cmd/table.go` (ID column, remove KIND/CLUSTER, adapter columns, deletion marker)
- [x] 3. Add `adpCell` type and `adapterCells` helper to `cmd/table.go`
- [x] 4. Update `TestTable_RendersClusterAndNodePoolRows` for new column layout
- [x] 5. Update `TestTable_EmptyCluster_NoNodePoolRows` for new column layout
- [x] 6. Add `TestTable_AdapterColumns` test
- [x] 7. Add `TestTable_DeletionMarker` test
- [x] 8. Run `go test ./...` and verify all tests pass
- [x] 9. Build and verify against live API
