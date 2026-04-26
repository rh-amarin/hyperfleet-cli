# Tasks: Phase 08 — Maestro

## 1. internal/maestro package

- [x] 1.1 `client.go` — Client struct with `New(httpEndpoint, consumer, token string) *Client`
- [x] 1.2 `client.go` — `List(ctx) ([]Resource, error)` — GET resources filtered by consumer
- [x] 1.3 `client.go` — `Get(ctx, name) (*Resource, error)` — GET single resource
- [x] 1.4 `client.go` — `Delete(ctx, name) error` — DELETE resource
- [x] 1.5 `client.go` — `ListBundles(ctx) ([]Bundle, error)` — GET resource-bundles
- [x] 1.6 `client.go` — `ListConsumers(ctx) ([]Consumer, error)` — GET consumers
- [x] 1.7 Resource, Bundle, Consumer, Manifest, Condition struct types

## 2. internal/maestro unit tests

- [x] 2.1 `TestClientList_FiltersConsumer` — verifies consumer_name query param
- [x] 2.2 `TestClientGet_ReturnsResource` — verifies GET /resources/<name>
- [x] 2.3 `TestClientDelete_SendsDELETE` — verifies DELETE method and path
- [x] 2.4 `TestClientListBundles_ReturnsItems` — verifies GET /resource-bundles
- [x] 2.5 `TestClientListConsumers_ReturnsItems` — verifies GET /consumers
- [x] 2.6 `TestClientList_NoConsumer_OmitsQueryParam` — consumer empty → no query param

## 3. cmd/maestro.go commands

- [x] 3.1 `hf maestro list` — table output (NAME, CONSUMER, VERSION, MANIFESTS)
- [x] 3.2 `hf maestro get <name>` — JSON output of single resource
- [x] 3.3 `hf maestro delete <name>` — Y/N confirmation + `--yes` flag
- [x] 3.4 `hf maestro bundles` — JSON output of bundle list
- [x] 3.5 `hf maestro consumers` — table output (ID, NAME)
- [x] 3.6 `hf maestro tui` — `syscall.Exec` into `maestro-cli tui --api-server=<endpoint>`

## 4. cmd/maestro unit tests

- [x] 4.1 `TestMaestroList_RendersTable`
- [x] 4.2 `TestMaestroGet_PrintsJSON`
- [x] 4.3 `TestMaestroDelete_WithYesFlag_CallsDELETE`
- [x] 4.4 `TestMaestroDelete_WithoutYesFlag_Cancels` (answers "N" via stdin mock)
- [x] 4.5 `TestMaestroBundles_PrintsJSON`
- [x] 4.6 `TestMaestroConsumers_RendersTable`
- [x] 4.7 `TestMaestroTUI_MissingBinary_ReturnsError`

## 5. Verify

- [x] (a) `go build ./...` succeeds → see `verification_proof/build.txt`
- [x] (b) `go vet ./...` reports no issues → see `verification_proof/vet.txt`
- [x] (c) `go test ./...` passes → `verification_proof/tests.txt`
- [ ] (d) Live verification → `verification_proof/maestro-connectivity.txt`
