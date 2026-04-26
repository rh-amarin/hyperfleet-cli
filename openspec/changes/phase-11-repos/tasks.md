# Tasks: Phase 11 — Repos

## 1. Config Extension (additive)

- [x] 1.1 Add `Token string` to `RegistryConfig` in `internal/config/types.go`
- [x] 1.2 Add `registry.token` to `getField`, `setField`, `AllPaths`, and `secretPaths` in `internal/config/fields.go`

## 2. Go Dependency

- [x] 2.1 `go get github.com/google/go-github/v60` → updates go.mod and go.sum

## 3. internal/repos

- [x] 3.1 `client.go`: `Client`, `New(token)`, `RepoInfo`, `ListRepos`, `SetBaseURL`
- [x] 3.2 `client_test.go`: httptest-based tests for `ListRepos`

## 4. cmd/repos.go

- [x] 4.1 `reposCmd` with `--registry`, `--watch/-w`, `--interval`, `--github-api-url` (hidden) flags
- [x] 4.2 Token resolution: `GITHUB_TOKEN` env → `registry.token` config
- [x] 4.3 Table output: NAME, OPEN PRS, DEFAULT BRANCH, CI STATUS
- [x] 4.4 Watch mode with configurable interval
- [x] 4.5 `repos_test.go`: httptest-based cmd tests

## 5. Verify

- [x] (a) `go build ./...` succeeds
- [x] (b) `go vet ./...` no issues
- [x] (c) `go test ./...` passes → `verification_proof/tests.txt`
- [x] (d) Live: `hf repos --registry rh-amarin` → `verification_proof/repos-list.txt`
- [x] (d) Live: watch mode 2 cycles → `verification_proof/repos-watch.txt`
