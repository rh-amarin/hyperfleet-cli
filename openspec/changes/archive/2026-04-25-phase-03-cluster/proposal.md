# Proposal: Phase 03 — Cluster Lifecycle

## Intent

Implement all `hf cluster` subcommands — the core CRUD loop for clusters.

## Why

Cluster management is the primary user-facing operation of the HyperFleet CLI. All downstream phases (nodepool, adapter-status, tables, workflow) depend on cluster state. This is the first phase that exercises the API client and config system together under real API conditions.

## Scope In

- `cmd/cluster.go` — all cluster subcommands wired into the Cobra tree
- `internal/api` — `FindClusterByName` shared helper (used by create and search)
- Cluster-scoped API endpoints: `GET|POST|PATCH|DELETE /clusters[/{id}]` and `/clusters/{id}/adapter-statuses`

## Scope Out

- NodePool commands (phase-04)
- Adapter status posting (phase-05)
- Watch mode and combined tables (phase-06)
- Any non-cluster resource types
