# PLAN: omnidevx-otel

## Phase 0 — Specs & repo

- [x] PRD/TRD/PLAN/ROADMAP (2026-07-16).
- [x] Repo scaffolded; git initialized (no commits yet by standing
  instruction).

## Phase 1 — Daemon relocation + ingest — DONE 2026-07-16

- [x] Daemon moved here from `omnidevx/cmd/omnidevx-otel` (old copy
  removed); ingest via `omnidevx-core/providers/genericotel`.
- [x] Rebuilt `~/go/bin/omnidevx-otel`; launchd service restarted and
  serving.

## Phase 2 — Query API — DONE 2026-07-16

- [x] `server/`: `/api/health`, `/api/series` (tokens, cost, activity,
  commits; day/week buckets; since/until; mode filter).
- [x] Unit tests against a fixture store; week-boundary bucket math
  (2026-07-18, alongside live verification against the real store).
- [x] Real-store verification: tokens/day and weekly ai_share sane.

## Phase 3 — Grafana — DONE 2026-07-16 (except PRISM panel)

- [x] Grafana installed (Homebrew, /usr/local prefix) + Infinity plugin.
- [x] Provisioning copied to /usr/local/etc/grafana/provisioning;
  grafana.ini `[paths] provisioning` pointed there (uncommented — survives
  brew upgrades, unlike homepath conf/).
- [x] Dashboard v0 live: http://localhost:3000/d/omnidevx-main (4 panels:
  tokens/day, cost/day, activity/day, weekly AI-commit share).
- [ ] PRISM capability-sparkline panel reading
  `grokify-ecosystem/prism/maturity/state.json`.

## Phase 4 — Codex OTel enablement

- [ ] Enable `[otel]` in `~/.codex/config.toml` → same endpoint; confirm
  codex-cli.jsonl receives otel-mode events; extend metric-name mapping if
  Codex names differ from the pre-registered guesses.

## Phase 5 — Later

- [ ] OTel logs/events channel (prompt.id correlation) — needs a
  content-stripping policy first.
- [ ] Derived SQLite index if dashboard latency demands it.
- [ ] Rename decision (omnidevxd?) before first publish.
