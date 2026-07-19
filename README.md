# OmniDevX-oTel

[![Go CI][go-ci-svg]][go-ci-url]
[![Go Lint][go-lint-svg]][go-lint-url]
[![Go SAST][go-sast-svg]][go-sast-url]
[![Docs][docs-godoc-svg]][docs-godoc-url]
[![Visualization][viz-svg]][viz-url]
[![License][license-svg]][license-url]

 [go-ci-svg]: https://github.com/plexusone/omnidevx-otel/actions/workflows/go-ci.yaml/badge.svg?branch=main
 [go-ci-url]: https://github.com/plexusone/omnidevx-otel/actions/workflows/go-ci.yaml
 [go-lint-svg]: https://github.com/plexusone/omnidevx-otel/actions/workflows/go-lint.yaml/badge.svg?branch=main
 [go-lint-url]: https://github.com/plexusone/omnidevx-otel/actions/workflows/go-lint.yaml
 [go-sast-svg]: https://github.com/plexusone/omnidevx-otel/actions/workflows/go-sast-codeql.yaml/badge.svg?branch=main
 [go-sast-url]: https://github.com/plexusone/omnidevx-otel/actions/workflows/go-sast-codeql.yaml
 [docs-godoc-svg]: https://pkg.go.dev/badge/github.com/plexusone/omnidevx-otel
 [docs-godoc-url]: https://pkg.go.dev/github.com/plexusone/omnidevx-otel
 [docs-mkdoc-svg]: https://img.shields.io/badge/docs-guide-blue.svg
 [docs-mkdoc-url]: https://plexusone.github.io/omnidevx-otel
 [viz-svg]: https://img.shields.io/badge/repo-visualization-blue.svg
 [viz-url]: https://mango-dune-07a8b7110.1.azurestaticapps.net/?repo=plexusone%2Fomnidevx-otel
 [license-svg]: https://img.shields.io/badge/license-MIT-blue.svg
 [license-url]: https://github.com/plexusone/omnidevx-otel/blob/main/LICENSE

The **OmniDevX local telemetry plane**: a loopback daemon that ingests OpenTelemetry metrics from AI coding agents (Claude Code, Codex CLI) into the canonical [OmniDevX](https://github.com/plexusone/omnidevx-core) event store, serves a read-only query API over it, and ships provisioned Grafana dashboards.

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

Only the OTLP **JSON** encoding is supported (stdlib-only by design; no protobuf/gRPC dependencies). Log payloads are acknowledged and discarded — prompt-bearing bodies are never stored.

For macOS persistence, run it under launchd (RunAtLoad + KeepAlive) with the binary at a stable path.

## Grafana

Copy `grafana/provisioning/` into your Grafana provisioning directory (set the dashboards provider `path` to this repository's `grafana/dashboards`, and install the `yesoreyeram-infinity-datasource` plugin). The provisioned dashboard charts tokens/day, cost/day, session activity, and weekly AI-assisted commit share against the query API.

## Specs

Product and technical specs live in [docs/specs](docs/specs/PRD.md).

## License

MIT
