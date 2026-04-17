// Package observe defines transport-neutral observation contracts shared by
// proxykit packages.
//
// The package is intentionally narrow. It models sessions, HTTP round trips,
// WebSocket frames, and hook functions without owning storage, route shapes,
// preview JSON schemas, or application-specific projections.
package observe
