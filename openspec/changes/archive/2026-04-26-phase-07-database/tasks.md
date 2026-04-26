# Tasks: Phase 07 — Database

## 1. internal/db Package

- [x] 1.1 Define `DBClient` interface (Query, Exec, Close)
- [x] 1.2 Implement `Client` struct wrapping `pgxpool.Pool`
- [x] 1.3 `New(ctx, dsn)` — connect and ping
- [x] 1.4 `Query` — return headers + string-formatted rows
- [x] 1.5 `Exec` — run non-SELECT statements
- [x] 1.6 `Close` — close pool

## 2. internal/db Unit Tests

- [x] 2.1 `mockClient` implements `DBClient` (compile-time assertion)
- [x] 2.2 `TestMockClient_Query` — verifies Query contract
- [x] 2.3 `TestMockClient_Exec` — verifies Exec contract
- [x] 2.4 `TestMockClient_Close` — verifies Close contract

## 3. cmd/db.go Commands

- [x] 3.1 `hf db query <sql>` — connect, Query, PrintTable
- [x] 3.2 `hf db delete <table> <where>` — count + confirm + Exec
- [x] 3.3 `hf db delete-all <table>` — count + confirm + Exec
- [x] 3.4 `hf db config` — print resolved DB config (password masked)

## 4. cmd/db Unit Tests

- [x] 4.1 `TestDbConfig_PrintsTable` — no DB needed
- [x] 4.2 `TestDbQuery_PrintsTable` — mock returns rows
- [x] 4.3 `TestDbQuery_EmptyResult` — prints info, no table
- [x] 4.4 `TestDbQuery_ConnectError` — returns error
- [x] 4.5 `TestDbDelete_ConfirmYes` — mock confirm Y, Exec called
- [x] 4.6 `TestDbDelete_ConfirmNo` — mock confirm N, Exec not called
- [x] 4.7 `TestDbDeleteAll_ConfirmYes` — mock confirm Y

## 5. Verify

- [x] (a) `go build ./...` succeeds
- [x] (b) `go vet ./...` no issues
- [x] (c) `go test ./...` passes — output to `verification_proof/tests.txt`
- [x] (d) Live DB verification — output to `verification_proof/db-connectivity.txt`
