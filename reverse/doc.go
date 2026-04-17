// Package reverse provides a reusable HTTP reverse proxy handler with
// transport-level mutation and observation hooks.
//
// The package owns target resolution, hop-by-hop header handling, request
// forwarding, and neutral observation callbacks. It does not own application
// routes, persistence, preview JSON schemas, or UI-facing DTOs.
package reverse
