// Package wsproxy provides a reusable WebSocket proxy handler with hook-based
// observation points and target resolution abstractions.
//
// The package intentionally does not own persistence, monitor payloads, or any
// application-specific storage schema. Callers can attach their own hooks and
// keep UI or backend compatibility layers outside the transport engine.
package wsproxy
