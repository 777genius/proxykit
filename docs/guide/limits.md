# Limits and Non-Goals

Strong open-source projects are clearer when they say what they are **not** trying to be.

## Non-goals

`proxykit` is intentionally not:

- a gateway platform
- a service discovery system
- a ready-made admin API
- a session explorer backend
- a storage engine
- a replay product
- a frontend protocol or monitor room server

## What stays in your app

You should keep these responsibilities outside `proxykit`:

- REST and WebSocket route naming
- admin authentication and loopback policy
- settings persistence
- capture visibility rules
- session query language and pagination
- preview shaping for UI clients
- tags, annotations, HAR workflows, or replay flows

## Current trade-offs

Compared with older incumbents, `proxykit` still has some honest limitations:

- fewer examples than long-standing projects
- less ecosystem recognition
- fewer ready-made integrations
- no built-in policy DSL

Those trade-offs are deliberate. The goal is to stay smaller, more embeddable, and easier to version.

## When another project is a better fit

Choose a different tool if you need:

- a production ingress or gateway platform such as Traefik, Skipper, or Caddy
- a richer built-in routing and policy model
- a single large programmable proxy object with deeper built-in mutation systems
- a product backend that already owns storage, admin APIs, and UI delivery contracts

See [Comparisons](/guide/comparisons) for where those tools are stronger.

## What "complete" means for this repo

`proxykit` does not become better by absorbing more and more backend code.

It becomes better when:

- the public API is clear
- examples compile
- docs explain the scope quickly
- adapter boundaries stay honest
- the package family solves real embedding scenarios without product baggage
