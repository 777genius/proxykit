# Contributing

Thanks for contributing to `proxykit`.

## Local workflow

### Verify Go packages

```bash
go test ./...
go test -race ./...
```

### Verify docs

```bash
npm ci
npm run docs:build
```

### Useful local commands

```bash
npm run docs:dev
go test ./... -run Example
```

## Architecture guardrails

Keep `proxykit` focused on reusable transport foundations.

Good additions:

- transport handlers
- listener lifecycle helpers
- neutral observation contracts
- focused helpers that support transport seams

Bad additions:

- app-specific REST DTOs
- admin auth policy
- UI preview models
- monitor room protocol
- product-specific route names
- storage ownership or spool lifecycle policy

## API design expectations

- prefer small packages with explicit responsibility
- prefer additive `Options` fields over breaking redesigns
- prefer concrete types over exported interfaces unless polymorphism is essential
- keep examples compile-checked
- document non-obvious behavior in the docs site, not only in code comments

## Commits

This repository uses Conventional Commits.

Examples:

- `docs: add cookbook and migration guides`
- `feat(reverse): add redirect rewrite helper`
- `fix(connect): normalize hijack error handling`

## Pull requests

Before opening a PR:

1. run the Go test suite
2. run race tests
3. build the docs site
4. make sure the change still respects the transport vs app boundary
