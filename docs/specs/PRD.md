# PRD: omnidevx-otel — Local Telemetry Plane

**Scope:** This application repo. Ecosystem context lives in
`plexusone/devfolio/docs/specs/` (OmniDevX PRD/TRD/PLAN/ROADMAP).

## Problem

Three needs converge on one always-running local process:

1. **Live token capture.** Claude Code's local session history is retained
   only ~8 weeks; every week without OTel collection permanently loses token
   history. Codex CLI likewise exports OTel but stores history in internal
   formats.
2. **Operational visibility.** The event store (~140k+ events: sessions,
   tokens, commits, GitHub snapshots) has no live view. Grafana was selected
   as the dashboard app — also serving as the PRISM-maturity example app and
   bug-bounty familiarization target.
3. **One daemon, not three.** A solo-developer setup should run a single
   OmniDevX process, not an OTel Collector + Prometheus + exporter stack.

## Product

`omnidevx-otel` is the local OmniDevX telemetry plane:

- **Ingest:** OTLP/HTTP (JSON) receiver for Claude Code and Codex CLI
  metrics → canonical `ai.usage.recorded` events → the JSONL event store.
- **Query:** read-only HTTP API serving time-bucketed series (tokens, cost,
  sessions, prompts, tool calls, commits, AI-assisted share) aggregated
  from the store on demand.
- **Dashboards:** provisioned Grafana assets (Infinity datasource +
  dashboard JSON) reading the query API and the PRISM maturity state.

## Principles

1. **The store is canon; this daemon owns no data.** Ingest normalizes and
   appends; query aggregates on demand; nothing is cached authoritatively.
   Grafana visualizes projections and can always be rebuilt.
2. **Loopback only.** Binds 127.0.0.1; no auth in v0 because there is no
   network exposure. Multi-user/team is out of scope until the OmniDevX
   team phase.
3. **Metadata only.** Log/event bodies (which may contain prompt text) are
   acknowledged and discarded; only numeric metrics and identifiers are
   stored.
4. **Grafana is the cockpit, not the brochure.** Publishing/portfolio
   rendering stays with DevFolio/dashforge/plexusone.dev.

## Users

Solo developer (John), local machine, launchd-managed.

## Out of Scope (v0)

- Prometheus/OTel-Collector interop; OTLP protobuf; traces.
- Remote access, auth, multi-tenant.
- Storing Grafana state beyond provisioned assets.

## Naming note

"otel" names the ingest origin; the app has grown a query/dashboard role.
If it broadens further (e.g. scheduled collectors), rename candidate:
`omnidevxd`. Revisit before first publish.
