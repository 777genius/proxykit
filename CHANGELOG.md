# Changelog

All notable changes to `proxykit` should be documented in this file.

The format is inspired by Keep a Changelog, adapted for a Go library with multiple focused packages.

## Unreleased

## v0.1.2

### Added

- first-screen README quick start example
- capability maps in README and docs homepage
- architecture diagram showing `proxykit` as the reusable core under `flutter_network_debugger`
- VitePress documentation site with guide and reference sections
- live GitHub Pages docs deployment
- cookbook, migration, compatibility, and limits guides
- community files: contributing, security, support, code of conduct
- issue templates and PR template
- compile-checked examples for `proxyhttp`, `socketio`, and `mitm`

### Changed

- README and docs homepage now explain the `goproxy` / `oxy` / `Martian` positioning directly, not only in deeper docs
- release surface is now explicitly oriented around launch clarity for new users
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
