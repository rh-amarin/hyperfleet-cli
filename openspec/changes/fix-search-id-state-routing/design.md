# Design: Fix search no-args state routing and id WARN exit-0

## clusterSearchCmd / nodepoolSearchCmd

Remove `Args: cobra.ExactArgs(1)`. In `RunE`, branch on `len(args) == 0`:

```
if len(args) == 0 {
    id := cfgStore.State().ClusterID   // or .NodePoolID
    if id == "" {
        out.Errorf("<message>")
        return fmt.Errorf("<message>")   // cobra exits 1; [ERROR] already on stderr
    }
    // behave identically to the get command
    cluster, err := api.Get[resource.Cluster](c, ctx, "clusters/"+id)
    ...
    return printer().Print(cluster)
}
// existing search-by-name path unchanged
name := args[0]
...
```

For `nodepoolSearchCmd` the no-args get path also requires `config.ClusterID` (same
guard as `nodepoolGetCmd`).

## clusterIDCmd / nodepoolIDCmd

Read state directly instead of going through `config.ClusterID` (which errors on miss):

```
id := cfgStore.State().ClusterID   // or .NodePoolID
if id == "" {
    out.Warn("No cluster-id set in state")
    return nil   // exit 0
}
fmt.Println(id)
return nil
```

## Error format

`[ERROR]` messages use `out.Errorf(msg)` (which writes `[ERROR] <msg>\n` to stderr).
Cobra also prints `Error: <msg>` after because `rootCmd.SilenceErrors` is not set — this
matches existing behaviour in `clusterPatchCmd` and is acceptable per the test suite.

`[WARN]` messages use `out.Warn(msg)`.
