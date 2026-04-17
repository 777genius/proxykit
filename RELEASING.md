# Releasing

This file describes the lightweight release process for `proxykit`.

## Goals

- keep releases predictable
- preserve public API trust
- avoid product-specific changes leaking into the module by accident

## Before tagging a release

Run:

```bash
go test ./...
go test -race ./...
npm ci
npm run docs:build
```

Then verify:

- public docs still match current exported behavior
- examples still compile
- changelog has a short summary of user-visible changes
- no app-specific DTOs, routes, or monitor protocols leaked into public packages

## Versioning rules

- patch release: bug fixes and docs-only corrections
- minor release: additive API growth, new helpers, new focused packages
- major release: breaking public API changes

## Release steps

1. update `CHANGELOG.md`
2. commit final release prep
3. create an annotated tag like `v0.1.2`
4. push the tag
5. publish GitHub release notes that summarize package-level impact

## Release note style

Good release notes answer:

- what changed
- which packages are affected
- whether the change is additive or breaking
- what downstream users should test
