# Design: Remove Legacy File-Per-Property Config Detection

## internal/config/store.go

Remove the `warnLegacy` call site and the method itself.

**In `NewStore`** — delete the comment and call:

```go
// before
	// Warn about legacy file-per-property layout.
	s.warnLegacy()

	// Load config.yaml once; unmarshal into both cfg (with defaults) and rawCfg (zero base).

// after
	// Load config.yaml once; unmarshal into both cfg (with defaults) and rawCfg (zero base).
```

**Method** — delete entirely:

```go
// remove
func (s *Store) warnLegacy() {
	// Detect old file-per-property files (flat, no extension, not yaml files).
	legacyKeys := []string{"api-url", "api-version", "token", "cluster-id"}
	for _, k := range legacyKeys {
		if _, err := os.Stat(filepath.Join(s.dir, k)); err == nil {
			fmt.Fprintf(os.Stderr,
				"[WARN] Legacy file-per-property config detected in %s. "+
					"Run 'hf config migrate' to convert to YAML format.\n", s.dir)
			return
		}
	}
}
```

After removal, check whether the `fmt` import is still used elsewhere in `store.go`. It is not
— `fmt` is only used by `warnLegacy`. Remove it from the import block.

## openspec/specs/config-model/spec.md

**Purpose paragraph** — remove the backwards-compatibility clause:

```
# before
…This replaces the file-per-property model from the shell scripts while maintaining backwards
compatibility during migration.

# after
…This replaces the file-per-property model from the shell scripts.
```

**Migration requirement** — delete the section from `### Requirement: Migration from
File-Per-Property` through the end of the "No legacy config" scenario (lines 221–238),
leaving "### Requirement: Config File Path Override" as the next section after "Config Show
with Source Annotation":

```
# remove
### Requirement: Migration from File-Per-Property

The CLI SHALL support one-time migration from the legacy file-per-property config format.

#### Scenario: Auto-detect legacy config

- GIVEN individual property files exist in `~/.config/hf/` (e.g., `api-url`, `cluster-id`)
- WHEN the CLI starts and no `config.yaml` exists
- THEN the CLI MUST detect the legacy format
- AND offer to migrate by reading all property files and writing `config.yaml` and `state.yaml`
- AND upon successful migration, rename the legacy files to `~/.config/hf/legacy/` for backup

#### Scenario: No legacy config

- GIVEN no legacy files exist and no `config.yaml` exists
- WHEN the CLI starts
- THEN the CLI MUST create `config.yaml` with defaults
- AND create an empty `state.yaml`
```

## openspec/specs/config-registry/spec.md

**Historical note** — remove the parenthetical line below the "Configuration Storage" requirement:

```
# before
(Previously: file-per-property storage at `~/.config/hf/<key>`. Superseded by config-model/spec.md.)

# after
(line deleted)
```
