# Design: Fix Combined Table Gaps

## Packages modified

| Package | File | Change |
|---------|------|--------|
| `internal/output` | `printer.go` | Add `GenCell(gen int32, deleted bool) string` method |
| `cmd` | `table.go` | Rewrite `renderCombinedTable`; add `adpCell` type and `adapterCells` helper |
| `cmd` | `table_test.go` | Update tests for removed columns; add adapter + deletion marker tests |

## Key decisions

### ID column
Add `ID` as the first element of the fixed header/row slices in `renderCombinedTable`. Nodepool IDs are prefixed with two spaces (`"  " + np.ID`).

### KIND/CLUSTER removal
Remove `"KIND"` and `"CLUSTER"` from headers and the corresponding cells from every row. Hierarchy is expressed by indentation only.

### Adapter columns
1. During the fetch phase, call `GET clusters/{id}/statuses` for each cluster and `GET clusters/{id}/nodepools/{id}/statuses` for each nodepool.
2. Collect unique adapter names across all responses; sort them alphabetically.
3. Append sorted adapter names to headers after condition columns.
4. For each row, call `adapterCells(statuses, deleted)` which picks the `Available` condition (or `Finalized` when `deleted_time != ""`), returning a `map[adapterName]adpCell{status, gen}`.
5. Render each cell with `p.DotWithGen(cell.status, cell.gen)`.

### Deletion marker
`Printer.GenCell(gen, deleted)` renders:
- `deleted=false` → `"<gen>"`
- `deleted=true, noColor=true` → `"<gen> ❌"`
- `deleted=true, noColor=false` → `"<gen> \033[31m❌\033[0m"`

The method is added to `internal/output/printer.go` and uses the package-level `colorRed`/`colorReset` constants from `dots.go`.

## Test strategy
- Update `TestTable_RendersClusterAndNodePoolRows`: mock `/statuses` endpoints (return empty list); assert ID in header; assert no KIND/CLUSTER; assert nodepool rows indented.
- Update `TestTable_EmptyCluster_NoNodePoolRows`: mock `/statuses`; remove NodePool KIND check.
- Add `TestTable_AdapterColumns`: two clusters with different adapters; assert adapter column headers and dot cells.
- Add `TestTable_DeletionMarker`: cluster with `deleted_time` set; assert `❌` in GEN cell and Finalized condition used for adapters.
