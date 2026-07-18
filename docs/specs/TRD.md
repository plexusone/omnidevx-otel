# TRD: omnidevx-otel

## Architecture

```text
Claude Code ──OTLP/JSON──┐
Codex CLI ───OTLP/JSON──┤
                         ▼
              omnidevx-otel (127.0.0.1:4318, launchd)
              ├── POST /v1/metrics   genericotel receiver (omnidevx-core)
              ├── POST /v1/logs     accepted, discarded
              ├── GET  /api/health
              └── GET  /api/series  read-only aggregation
                         │                    ▲
                         ▼                    │ reads
              ~/.plexusone/omnidevx/data (JSONL store, canon)

Grafana (localhost:3000) ── Infinity datasource ──► /api/series
                        └── Infinity (file/url) ──► grokify-ecosystem/prism/maturity/state.json
```

## Modules

- Module `github.com/plexusone/omnidevx-otel`; depends on
  `omnidevx-core` (receiver, store, event types). Local replace until core
  publishes. Stdlib-only otherwise.
- `server/` — HTTP wiring + the series aggregator.
- `cmd/omnidevx-otel/` — flags: `-addr` (default 127.0.0.1:4318),
  `-person`, `-store`.
- Ingest logic stays in `omnidevx-core/providers/genericotel` (library);
  this repo composes it. The former `omnidevx/cmd/omnidevx-otel` is
  superseded and removed.

## Query API

`GET /api/series?metric=<m>&bucket=<day|week>&since=<RFC3339>&until=<RFC3339>`

Response: `{"metric": m, "bucket": b, "rows": [{"t": "2026-07-17", ...}]}`

| metric | row fields | source events |
|---|---|---|
| tokens | input, output, cache_read, cache_creation | ai.message.completed + ai.usage.recorded (all modes) |
| cost | cost_usd | ai.usage.recorded (otel) |
| activity | sessions, prompts, tool_calls | ai.session.started, ai.prompt.submitted, ai.tool.completed |
| commits | commits, ai_assisted, ai_share | devx.change.committed |

Defaults: bucket=day, since=30 days ago. Aggregation is a full store scan
per request (~250ms at 140k events) — acceptable; add a derived SQLite
index only if dashboards outgrow it (reserved escape hatch, regenerable).

Double-count guard: token totals may count the same session via history
import AND otel; rows carry per-mode fields only when modes disagree is out
of v0 scope — v0 sums by (source, mode) and exposes `mode` as an optional
filter (`&mode=otel`).

## Grafana

- Install: Homebrew `grafana`; plugin `yesoreyeram-infinity-datasource`.
- Provisioning under `grafana/provisioning/` (datasource + dashboard
  providers) and `grafana/dashboards/omnidevx.json`; symlinked or copied
  into the Homebrew Grafana provisioning dir (documented in README).
- Panels v0: tokens/day stacked; cost/day; sessions+prompts/day; weekly
  AI-assisted commit share; capability sparklines (from PRISM state.json).

## Service

launchd `com.plexusone.omnidevx-otel` (RunAtLoad, KeepAlive), binary
`~/go/bin/omnidevx-otel`, logs `~/Library/Logs/omnidevx-otel.log`.

## Security

Loopback bind; no auth (v0); read-only query surface; 32MB ingest body cap;
no content-bearing fields stored (enforced by genericotel tests).
