# Changelog

All notable changes to `proxykit` should be documented in this file.

The format is inspired by Keep a Changelog, adapted for a Go library with multiple focused packages.

## Unreleased

### Added

- VitePress documentation site with guide and reference sections
- live GitHub Pages docs deployment
- cookbook, migration, compatibility, and limits guides
- community files: contributing, security, support, code of conduct
- issue templates and PR template
- compile-checked examples for `proxyhttp`, `socketio`, and `mitm`

### Changed

- repo automation now includes CI, docs deploy, and dependency maintenance signals
- docs and README now explain project boundaries and stability expectations more explicitly

## v0.1.1

### Changed

- module path moved to `github.com/777genius/proxykit`

## v0.1.0

### Added

- initial public extraction of transport-focused packages:
  - `reverse`
  - `forward`
  - `connect`
  - `wsproxy`
  - `proxyruntime`
  - `observe`
  - `cookies`
  - `proxyhttp`
  - `socketio`
  - `mitm`
