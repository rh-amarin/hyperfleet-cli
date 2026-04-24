# HyperFleet CLI Requirements Specification

## Purpose

Complete requirements specification for the HyperFleet CLI tool, reverse-engineered from the shell scripts at [rh-amarin/hyperfleet-cli](https://github.com/rh-amarin/hyperfleet-cli). The CLI manages OpenShift HyperFleet cluster lifecycle operations via the HyperFleet API, Maestro, Kubernetes, and messaging infrastructure.

## Specification Index

Organized to match the [output index](https://github.com/rh-amarin/hyperfleet-cli/blob/main/scripts/output/00-index.json):

| # | Domain | Spec | Requirements | Scenarios | Scripts Covered |
|---|--------|------|-------------|-----------|-----------------|
| 01 | [Configuration](config/spec.md) | Config management, env profiles, diagnostics | 8 | 12 | hf.config.sh, hf.cluster.id.sh, hf.nodepool.id.sh |
| 02 | [Cluster Lifecycle](cluster-lifecycle/spec.md) | Cluster CRUD operations | 9 | 19 | hf.cluster.{create,search,get,patch,delete,conditions,conditions.table,statuses}.sh |
| 03 | [NodePool Lifecycle](nodepool-lifecycle/spec.md) | NodePool CRUD operations | 10 | 15 | hf.nodepool.{create,list,search,get,patch,delete,conditions,conditions.table,statuses,table}.sh |
| 04 | [Adapter Status](adapter-status/spec.md) | Adapter status posting and convergence model | 3 | 8 | hf.cluster.adapter.post.status.sh, hf.nodepool.adapter.post.status.sh |
| 05 | [Tables and Lists](tables-and-lists/spec.md) | Aggregated views and formatted tables | 4 | 7 | hf.cluster.{list,table}.sh, hf.nodepool.table.sh, hf.table.sh |
| 06 | [Database](database/spec.md) | Direct PostgreSQL operations | 6 | 9 | hf.db.{query,delete,delete.all,statuses,statuses.delete,config}.sh |
| 07 | [Maestro](maestro/spec.md) | Maestro resource management | 6 | 8 | hf.maestro.{list,bundles,consumers,get,delete,tui}.sh |
| 08 | [Pub/Sub & Messaging](pubsub/spec.md) | Event publishing to GCP Pub/Sub and RabbitMQ | 4 | 5 | hf.pubsub.{list,publish.cluster.change,publish.nodepool.change}.sh, hf.rabbitmq.publish.cluster.change.sh |
| 09 | [Kubernetes](kubernetes/spec.md) | Port-forwarding, debugging, log tailing | 5 | 8 | hf.kube.{port.forward,context,curl,debug.pod}.sh, hf.logs.sh, hf.logs.adapter.sh |
| 10 | [Repos](repos/spec.md) | GitHub repository status overview | 1 | 2 | hf.repos.sh |
| 11 | [Errors & Usage](errors-and-usage/spec.md) | Error handling, usage messages, edge cases | 7 | 9 | Cross-cutting across all commands |
| 12 | [Interactive Commands](interactive-commands/spec.md) | Interactive, streaming, and long-running commands | 8 | 13 | hf.workflow.sh, hf.workflow.api-only.sh, and watch/interactive modes |
| 13 | [Config Registry](config-registry/spec.md) | Configuration property registry and storage model | 7 | 11 | hf.lib.sh (shared library) |

**Totals: 77 requirements, 143 scenarios covering 50+ scripts**

## Environment Context

From the recording environment:

```json
{
  "api_url": "http://localhost:8000",
  "api_version": "v1",
  "context": "kind-kind",
  "namespace": "hyperfleet",
  "db_host": "localhost",
  "db_port": 5432,
  "db_name": "hyperfleet",
  "maestro_http": "http://localhost:8100",
  "maestro_grpc": "localhost:8090",
  "gcp_project": "hcm-hyperfleet"
}
```

## API Base Path

All HyperFleet API calls use: `/api/hyperfleet/v1/`

## Key Design Patterns

1. **Config-driven**: All scripts read from `~/.config/hf/` (one file per property)
2. **Composable**: Scripts build on each other (e.g., `cluster create` calls `cluster search`)
3. **Defaults over usage**: Create commands with no args use defaults, not usage display
4. **Generation tracking**: Resources track generation; adapters report observed_generation
5. **Convergence logic**: Ready becomes True when ALL required adapters report Available=True at current generation
6. **Watch mode**: Status/conditions commands support `-w` flag for live updates
7. **Confirmation prompts**: Destructive operations require typing `yes` to confirm
8. **RFC 7807 errors**: API errors follow Problem Details format
