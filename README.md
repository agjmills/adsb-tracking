# ADS-B Tracking

Personal ADS-B receiver with leaderboard, metrics, and military aircraft detection.

## Services

| Service | Purpose | URL |
|---|---|---|
| ultrafeeder | SDR receiver + tar1090 map | `ADSB_HOST` |
| piaware | FlightAware feeder | — |
| adsbx | ADSB Exchange feeder | — |
| leaderboard | Metrics dashboard | `LEADERBOARD_HOST` |

## Leaderboard Features

- **Radar** — embedded tar1090 map with coverage splat
- **Nearest** — closest low-altitude aircraft (last 3 min, <12k ft, <10nm)
- **Leaderboard** — most seen aircraft ranked by sighting count
- **Military** — dedicated military aircraft tracker (hex range + callsign detection)
- **Distance / Altitude / Speed** — all-time extremes (furthest, closest, highest, lowest, fastest, slowest)
- **Recent** — live sighting feed
- **Daily** — per-day aggregate stats
- **Aircraft detail** — click any ICAO24 for full stats + links to PlaneSpotters, FR24, ADSBx
- **OpenSky enrichment** — automatic lookup of registration, type, manufacturer, operator

### Compact page

`/nearest` — dark theme, auto-refreshing page for embedded displays (Home Assistant, Raspberry Pi).

## Setup

```bash
cp .env.example .env
# edit .env with your location and hostnames
docker compose up -d
```

### Environment variables

| Variable | Description | Default |
|---|---|---|
| `HOME_LAT` | Receiver latitude | — |
| `HOME_LON` | Receiver longitude | — |
| `SITE_NAME` | Display name in dashboard | `Unknown` |
| `SITE_ALT` | Receiver altitude | `0m` |
| `ADSB_HOST` | tar1090 hostname | — |
| `LEADERBOARD_HOST` | Dashboard hostname | — |
| `POLL_INTERVAL` | Data polling interval (seconds) | `5` |
| `CLEANUP_HOURS` | Retention for raw sightings | `48` |
| `PIAWARE_FEEDER_ID` | FlightAware feeder ID | — |
| `ADSX_UUID` | ADSB Exchange UUID | — |

## Architecture

```
ultrafeeder:/data/aircraft.json
        │
        ▼
   leaderboard (Go)
        │
   ┌────┴────┐
   ▼         ▼
 SQLite   HTTP API (:5000)
   │         │
   │    ┌────┴─────────────┐
   │    ▼                  ▼
   │  Dashboard (/ )   API (/api/*)
   │    │
   ▼    ▼
 OpenSky API (enrichment)
```

- **SQLite** with WAL mode stores sightings, per-aircraft stats, and daily aggregates
- **Poller** fetches `/data/aircraft.json` every `POLL_INTERVAL` seconds
- **Enricher** queries OpenSky Network API for aircraft metadata on first sighting
- **Military detection** uses known ICAO24 hex ranges + callsign prefix matching
