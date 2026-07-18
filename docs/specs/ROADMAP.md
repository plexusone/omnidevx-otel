# ROADMAP: omnidevx-otel

| Version | Contents | Plan phase |
|---|---|---|
| v0.1.0 | Daemon (ingest + health), query API, Grafana provisioning + v0 dashboard | 1–3 |
| v0.2.0 | Codex OTel enabled end-to-end; metric mapping confirmed | 4 |
| v0.3.0 | OTel events/logs channel with content stripping; prompt.id correlation | 5 |
| v0.4.0 | Derived analytics index (SQLite/DuckDB) if needed; publish decision + possible rename | 5 |

Publish order dependency: requires `omnidevx-core` on GitHub first (drop
local replace). See ecosystem ROADMAP in devfolio/docs/specs.
