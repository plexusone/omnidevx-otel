# omnidevx-otel

The **OmniDevX local telemetry plane**: a loopback daemon that ingests
OpenTelemetry metrics from AI coding agents (Claude Code, Codex CLI) into
the canonical [OmniDevX](https://github.com/plexusone/omnidevx-core) event
store, serves a read-only query API over it, and ships provisioned Grafana
dashboards.

```text
Claude Code ──OTLP/JSON──┐
Codex CLI ───OTLP/JSON──┤
                         ▼
              omnidevx-otel (127.0.0.1:4318)
              ├── POST /v1/metrics   → ai.usage.recorded events
              ├── POST /v1/logs     accepted, discarded (never stored)
              ├── GET  /api/health
              └── GET  /api/series  tokens|cost|activity|commits × day|week
                         │
              ~/.plexusone/omnidevx/data (JSONL store, canon)
                         ▲
              Grafana ── Infinity datasource ──┘
```

Live token capture matters because agents' local session history is
subject to retention: without it, token history is permanently lost.

## Run

```bash
go install github.com/plexusone/omnidevx-otel/cmd/omnidevx-otel@latest
omnidevx-otel -person person:you
```

Point Claude Code at it via `~/.claude/settings.json`:

```json
{
  "env": {
    "CLAUDE_CODE_ENABLE_TELEMETRY": "1",
    "OTEL_METRICS_EXPORTER": "otlp",
    "OTEL_EXPORTER_OTLP_PROTOCOL": "http/json",
    "OTEL_EXPORTER_OTLP_ENDPOINT": "http://127.0.0.1:4318"
  }
}
```

Only the OTLP **JSON** encoding is supported (stdlib-only by design; no
protobuf/gRPC dependencies). Log payloads are acknowledged and discarded —
prompt-bearing bodies are never stored.

For macOS persistence, run it under launchd (RunAtLoad + KeepAlive) with
the binary at a stable path.

## Grafana

Copy `grafana/provisioning/` into your Grafana provisioning directory
(set the dashboards provider `path` to this repository's
`grafana/dashboards`, and install the
`yesoreyeram-infinity-datasource` plugin). The provisioned dashboard
charts tokens/day, cost/day, session activity, and weekly AI-assisted
commit share against the query API.

## Specs

Product and technical specs live in [docs/specs](docs/specs/PRD.md).

## License

MIT
