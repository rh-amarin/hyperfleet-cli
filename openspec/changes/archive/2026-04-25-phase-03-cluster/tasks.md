# Tasks: Phase 03 — Cluster Lifecycle

## 1. OpenSpec Change Folder
- [x] 1.1 Create .openspec.yaml
- [x] 1.2 Write proposal.md
- [x] 1.3 Write design.md
- [x] 1.4 Write tasks.md (this file)

## 2. Shared Helper
- [x] 2.1 Add `api.FindClusterByName(c, ctx, name) ([]resource.Cluster, error)` to internal/api

## 3. cmd/cluster.go Subcommands
- [x] 3.1 `hf cluster create [name] [region] [version]` — defaults, duplicate guard, SetClusterID
- [x] 3.2 `hf cluster get [cluster_id]`
- [x] 3.3 `hf cluster list`
- [x] 3.4 `hf cluster search <name>` — sets cluster-id
- [x] 3.5 `hf cluster patch spec|labels [cluster_id]` — counter increment
- [x] 3.6 `hf cluster delete [cluster_id]`
- [x] 3.7 `hf cluster id` — print configured cluster-id
- [x] 3.8 `hf cluster conditions [-w] [cluster_id]`
- [x] 3.9 `hf cluster conditions table [cluster_id]` — tabwriter with colored dots
- [x] 3.10 `hf cluster statuses [-w] [cluster_id]`

## 4. Verify Against Real Cluster
- [x] 4.1 `hf cluster create` → cluster created, cluster-id written to state
- [x] 4.2 `hf cluster search <name>` → sets cluster-id
- [x] 4.3 `hf cluster get` → JSON output matches API response shape
- [x] 4.4 `hf cluster patch spec` → generation increments
- [x] 4.5 `hf cluster patch labels` → generation increments
- [x] 4.6 `hf cluster conditions` → Ready/Available conditions visible
- [x] 4.7 `hf cluster conditions table` → colored dot table
- [x] 4.8 `hf cluster statuses` → AdapterStatusList JSON
- [x] 4.9 `hf cluster delete` → cluster soft-deleted
- [x] 4.10 Duplicate create → `[WARN]` and skip
