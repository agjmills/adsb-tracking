# AGENTS.md

## Build & Test

```bash
cd leaderboard && go build ./... && go vet ./...
```

Rebuild container:
```bash
docker compose build leaderboard && docker compose up -d leaderboard
```

## Architecture

Single Go binary (`leaderboard/`) with embedded SQLite (WAL mode), no external web framework.

```
ultrafeeder:/data/aircraft.json → poller.go → db.go → SQLite
                                              → main.go → HTTP :5000 → ui.go (SPA)
                                              → enrich.go → OpenSky API
```

## Key Files

| File | Purpose |
|---|---|
| `main.go` | Entry point, HTTP server, API routes, env config |
| `db.go` | SQLite schema, all CRUD, upsert logic, rollup compression |
| `poller.go` | Fetches `aircraft.json`, computes distance/bearing, detects military |
| `enrich.go` | OpenSky Network API client for aircraft metadata |
| `ui.go` | Single-file SPA: HTML, CSS, JS all in one Go string constant |
| `analysis.go` | Rule-based pattern detection (orbit, climb, descent, transit, loiter) |
| `hexrange.go` | Country hex ranges, military hex ranges, military callsign prefixes, emitter categories |
| `callsigns.go` | Callsign-to-operator mapping, name normalization |

## Database

Three main tables + one rollup:
- **`aircraft`** — one row per ICAO24, permanent aggregate stats
- **`sightings`** — raw position reports, pruned after `CLEANUP_HOURS` (default 72h)
- **`daily_stats`** — per-day aggregate rollup, permanent
- **`sightings_rollup`** — 5-min window compression, permanent (populated on cleanup)

### Adding a column

Add to the `CREATE TABLE` DDL and add an `ALTER TABLE` migration with ignored error:
```go
_, _ = d.Exec("ALTER TABLE aircraft ADD COLUMN new_col TYPE DEFAULT val")
```

### Adding an API endpoint

1. Add route in `main.go` with handler func
2. Add query method in `db.go`
3. Add row struct + scan logic
4. Use `writeJSON(w, data)` for response

## UI Conventions

- `ui.go` contains a single Go string constant `uiHTML` — all HTML/CSS/JS inline
- Dark theme with CSS custom properties (`--bg`, `--surface`, `--text`, `--accent`, `--mil`, `--green`, `--blue`)
- SPA navigation via `window.location.hash` with `navigate(page)` function
- Table pages use `loadTable(bodyId, url, rowFn)` pattern
- Responsive breakpoints: 768px (tablet), 480px (mobile)
- Mobile columns hidden with `.hide-mob` class, desktop-only with `.hide-desk`

## Military Detection

Two-phase:
1. ICAO24 hex range check (`hexrange.go:militaryHexRanges`) — covers US, UK, DE, FR, IT, NATO, RU
2. Callsign prefix match (`hexrange.go:militaryCallsignPrefixes`) — NATO/ICAO codes + tactical names

Logic in `poller.go:80`: `isMilitaryHex(hex) || isMilitaryCallsign(callsign)`

## Analysis System

- `analysis.go` runs rule-based heuristics on sighting data
- Patterns detected: orbit (holding), climb, descent, high-speed transit, loiter
- Flight detection: groupings with >30min gap = new flight session
- Sanity checks: rejects altitude=0, caps climb/descent at 10k ft/min, requires 75% monotonicity
- Rollup data supplements raw sightings for long-term flight counting

## Environment Variables

| Variable | Default | Purpose |
|---|---|---|
| `HOME_LAT` | 0 | Receiver latitude |
| `HOME_LON` | 0 | Receiver longitude |
| `SITE_NAME` | Unknown | Dashboard display name |
| `POLL_INTERVAL` | 5 | Data polling interval (seconds) |
| `CLEANUP_HOURS` | 72 | Raw sighting retention (hours) |
| `PORT` | 5000 | HTTP listen port |
| `DB_PATH` | /data/leaderboard.db | SQLite database path |

## Commits

Use [conventional commits](https://www.conventionalcommits.org/):
```
feat: add pattern analysis for aircraft movements
fix: reject altitude=0 in climb/descent detection
refactor: extract rollup compression into compressAndClean
chore: bump CLEANUP_HOURS default to 72
```
