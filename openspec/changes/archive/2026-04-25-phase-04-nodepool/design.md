# Design: Phase 04 — NodePool Lifecycle

## Package Structure

```
internal/api/nodepool.go        — FindNodePoolByName helper
cmd/nodepool.go                 — all hf nodepool subcommands
internal/api/nodepool_test.go   — unit tests for FindNodePoolByName
cmd/nodepool_test.go            — command-level tests
```

## API Path Pattern

All nodepool endpoints are nested under clusters:
```
/api/hyperfleet/v1/clusters/{cluster_id}/nodepools
/api/hyperfleet/v1/clusters/{cluster_id}/nodepools/{nodepool_id}
/api/hyperfleet/v1/clusters/{cluster_id}/nodepools/{nodepool_id}/adapter-statuses
```

The cluster-id is always resolved via `config.ClusterID(cfgStore, "")` — it must be set in state.

## FindNodePoolByName

Mirrors `FindClusterByName` exactly, scoped to a cluster:
```go
path := "clusters/" + clusterID + "/nodepools?search=" + url.QueryEscape(fmt.Sprintf("name='%s'", name))
```
Client-side filter: exact name match AND empty DeletedTime.

## Create — Loop Design

```
for i := 1; i <= count; i++:
    name_i = fmt.Sprintf("%s-%d", name, i)
    payload = {name: name_i, labels: {counter: "i"}, spec: {counter: "i", platform: {type: instanceType}, replicas: 1}}
    nodepool, err = api.Post[resource.NodePool](...)
    last = nodepool
config.SetNodePoolID(cfgStore, last.ID)
printer().Print(last)   // prints the last created nodepool
```

## Table Command — Column Extraction

```go
// Collect all condition lists for DynamicColumns
allConditions := [][]output.Condition{}
for _, np := range items {
    var conds []output.Condition
    for _, c := range np.Status.Conditions { conds = append(conds, output.Condition{Type: c.Type}) }
    allConditions = append(allConditions, conds)
}
dynCols := output.DynamicColumns(allConditions)
headers := append([]string{"ID", "NAME", "REPLICAS", "TYPE", "GEN"}, dynCols...)

// Extract REPLICAS and TYPE from spec (map[string]any)
replicas := fmt.Sprintf("%v", np.Spec["replicas"])
platform, _ := np.Spec["platform"].(map[string]any)
instanceType, _ := platform["type"].(string)
```

## Statuses 404 Handling

Same pattern as cluster: 404 from adapter-statuses endpoint returns an empty AdapterStatusList instead of an error.

## watchLoop Reuse

`watchLoop` is defined in `cmd/cluster.go` — it is in the same `cmd` package so nodepool.go can call it directly without any changes.
