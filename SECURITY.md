# Security Policy

## Reporting a vulnerability

If you believe you found a security issue in `proxykit`, do not open a public issue first.

Preferred path:

1. use GitHub's private vulnerability reporting for the repository if available
2. if that is not possible, contact the maintainer privately and include:
   - affected package
   - impact
   - reproduction steps
   - any proposed mitigation

## Scope

Security reports are especially relevant for:

- CONNECT tunneling
- TLS interception helpers
- header handling
- request routing and target resolution
- cookie rewriting behavior

## What helps a report

Please include:

- exact version or commit
- whether the issue affects default usage or only custom integrations
- proof of concept or reduced reproducer
- whether the issue is exploitable remotely

## Disclosure

The goal is coordinated disclosure:

- confirm the issue
- prepare a fix
- publish the fix and release notes
- then disclose details publicly
