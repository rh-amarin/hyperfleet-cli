# Tasks: Fix search no-args state routing and id WARN exit-0

- [x] 1. Fix `clusterSearchCmd`: remove `Args: cobra.ExactArgs(1)`, add no-args state routing
- [x] 2. Fix `clusterIDCmd`: print [WARN] and exit 0 when no cluster-id in state
- [x] 3. Fix `nodepoolSearchCmd`: remove `Args: cobra.ExactArgs(1)`, add no-args state routing
- [x] 4. Fix `nodepoolIDCmd`: print [WARN] and exit 0 when no nodepool-id in state
- [x] 5. Add tests: cluster search no-args with state ID (behaves like get)
- [x] 6. Add tests: cluster search no-args no state (ERROR exit 1)
- [x] 7. Add tests: cluster id no state (WARN exit 0)
- [x] 8. Add tests: cluster id with state (prints ID)
- [x] 9. Add tests: nodepool search no-args with state ID (behaves like get)
- [x] 10. Add tests: nodepool search no-args no state (ERROR exit 1)
- [x] 11. Add tests: nodepool id no state (WARN exit 0)
- [x] 12. Add tests: nodepool id with state (prints ID)
- [x] 13. Run `go test ./...` and `go vet ./...` — must pass with zero failures
