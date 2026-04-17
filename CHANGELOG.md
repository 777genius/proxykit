# Changelog

All notable changes to `proxykit` should be documented in this file.

The format is inspired by Keep a Changelog, adapted for a Go library with multiple focused packages.

## Unreleased

## v0.1.4

### Added

- `go vet ./...` in CI so public examples and exported usage stay under an extra correctness gate

### Changed

- fixed the README first-screen quick start to use the real `reverse.New` API instead of a non-existent helper

## v0.1.3

### Added

- release badge and discussions badge in the README
- explicit GitHub Discussions links in README, support docs, and issue template config
- extra GitHub topics for better discoverability in embedded and WebSocket-related searches

### Changed

- GitHub Discussions are now enabled for usage questions and design-fit discussion
- launch surface now points users to the right support path without sending them straight to issues

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
