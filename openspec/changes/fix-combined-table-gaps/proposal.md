# Proposal: Fix Combined Table Gaps

## What

Fix four structural gaps in the `hf table` combined resources overview command so it matches the spec in `openspec/specs/tables-and-lists/spec.md`.

## Why

The current `hf table` output:
1. Is missing the `ID` column (first fixed column per spec)
2. Shows `KIND` and `CLUSTER` columns that the spec explicitly forbids
3. Has no adapter columns — the spec requires fetching `/statuses` per resource and rendering one column per unique adapter name
4. Has no deletion marker — spec requires the GEN cell to show `<gen> ❌` (red) when `deleted_time` is set

## Gaps

| # | Gap | Current | Required |
|---|-----|---------|----------|
| 1 | ID column | absent | first fixed column, full UUID |
| 2 | KIND/CLUSTER columns | present | must not appear |
| 3 | Adapter columns | absent | one per unique adapter name from `/statuses` |
| 4 | Deletion marker | absent | GEN cell appends red ❌ when deleted_time is set |
