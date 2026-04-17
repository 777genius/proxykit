// Package mitm provides optional TLS interception primitives for proxykit.
//
// The package focuses on certificate authority lifecycle, per-host certificate
// issuance, interception policy, and PEM helpers. It does not own transport
// handlers, persistence, or application-specific MITM workflows.
package mitm
