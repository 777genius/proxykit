# Compatibility and Versioning

This page explains what stability means in `proxykit`.

## Versioning model

`proxykit` uses normal Go module semantic versioning.

That means:

- additive changes should land in new minor releases
- breaking public API changes should land only in a new major version
- patch releases should fix bugs without changing the intended API shape

## What counts as public API

For this repository, public API includes:

- exported Go identifiers in public packages
- documented package behavior described in this docs site
- compile-checked examples
- module path and package layout

It does **not** include:

- internal package details
- undocumented incidental behavior
- product-specific route names from applications that embed `proxykit`

## Compatibility promise

The project aims to grow by:

- adding new packages
- adding new `Options` fields
- adding new helper functions
- tightening docs around already intended behavior

The project should avoid:

- breaking constructor signatures without a major release
- collapsing multiple focused packages into a monolith
- adding app-specific DTOs, route contracts, or monitor protocols to the public surface

## Stability expectations by package

### Core packages

These are the main foundation packages and should evolve conservatively:

- `reverse`
- `forward`
- `connect`
- `wsproxy`
- `observe`
- `proxyruntime`

### Supporting packages

These are also public, but narrower in scope:

- `cookies`
- `proxyhttp`
- `socketio`
- `mitm`

They should still be stable, but they are intentionally more focused helpers, not the main entry points for most users.

## How to adopt safely

If your application depends on `proxykit`, the safest pattern is:

1. pin a released version in `go.mod`
2. keep your own adapter layer between product contracts and transport packages
3. upgrade one minor release at a time
4. run your integration tests against the new version before shipping

## What this repo will not promise

`proxykit` will not promise stability for:

- your application's route names
- your storage schema
- your frontend delivery protocol
- your admin API shapes

Those belong in your application layer, not in the module.
