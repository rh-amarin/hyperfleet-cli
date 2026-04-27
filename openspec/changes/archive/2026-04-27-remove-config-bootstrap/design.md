# Design: Remove hf config bootstrap from Technical Architecture Spec

## openspec/specs/technical-architecture/spec.md

**Module layout** — update the `config.go` comment to remove `bootstrap`:

```
# before
│   ├── config.go           # hf config [show|set|clear|doctor|bootstrap|env]

# after
│   ├── config.go           # hf config [show|set|clear|doctor|env]
```

**Cobra command tree** — remove `bootstrap [env-name]` and add `new [name]` under `env`:

```
# before
├── config
│   ├── show      [env-name]
│   ├── set       <key> <value>
│   ├── clear     <key|all>
│   ├── doctor
│   ├── bootstrap [env-name]
│   └── env
│       ├── list
│       ├── show     <name>
│       └── activate <name>

# after
├── config
│   ├── show      [env-name]
│   ├── set       <key> <value>
│   ├── clear     <key|all>
│   ├── doctor
│   └── env
│       ├── new      [name]
│       ├── list
│       ├── show     <name>
│       └── activate <name>
```
