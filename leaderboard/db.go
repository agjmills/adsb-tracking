package main

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

type DB struct {
	mu sync.Mutex
	*sql.DB
}

func openDB(path string) (*DB, error) {
	db, err := sql.Open("sqlite", path+"?_journal_mode=WAL&_busy_timeout=5000&_synchronous=NORMAL")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	d := &DB{DB: db}
	return d, d.migrate()
}

func (d *DB) migrate() error {
	ddl := `
	CREATE TABLE IF NOT EXISTS aircraft (
		icao24 TEXT PRIMARY KEY,
		registration TEXT DEFAULT '',
		manufacturer TEXT DEFAULT '',
		model TEXT DEFAULT '',
		operator_name TEXT DEFAULT '',
		operator_icao TEXT DEFAULT '',
		country TEXT DEFAULT '',
		country_flag TEXT DEFAULT '',
		is_military INTEGER DEFAULT 0,
		category TEXT DEFAULT '',
		first_seen INTEGER NOT NULL,
		last_seen INTEGER NOT NULL,
		last_callsign TEXT DEFAULT '',
		total_sightings INTEGER DEFAULT 1,
		total_flights INTEGER DEFAULT 1,
		min_distance_nm REAL,
		max_distance_nm REAL,
		min_altitude_ft INTEGER,
		max_altitude_ft INTEGER,
		max_ground_speed_kt REAL,
		last_lat REAL,
		last_lon REAL,
		last_altitude INTEGER,
		last_speed REAL,
		last_seen_pos INTEGER DEFAULT 0,
		enrichment_ts INTEGER DEFAULT 0,
		updated_at INTEGER NOT NULL
	);

	CREATE TABLE IF NOT EXISTS daily_stats (
		date TEXT PRIMARY KEY,
		unique_aircraft INTEGER DEFAULT 0,
		total_sightings INTEGER DEFAULT 0,
		military_aircraft INTEGER DEFAULT 0,
		max_distance_nm REAL DEFAULT 0,
		max_altitude_ft INTEGER DEFAULT 0,
		max_speed_kt REAL DEFAULT 0
	);

	CREATE TABLE IF NOT EXISTS sightings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		icao24 TEXT NOT NULL,
		callsign TEXT DEFAULT '',
		lat REAL,
		lon REAL,
		altitude_ft INTEGER,
		ground_speed_kt REAL,
		track REAL,
		distance_nm REAL,
		bearing_deg INTEGER,
		rssi REAL,
		category TEXT DEFAULT '',
		seen_at INTEGER NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_sightings_icao ON sightings(icao24);
	CREATE INDEX IF NOT EXISTS idx_sightings_seen ON sightings(seen_at);

	CREATE TABLE IF NOT EXISTS sightings_rollup (
		icao24 TEXT NOT NULL,
		callsign TEXT DEFAULT '',
		window_start INTEGER NOT NULL,
		sighting_count INTEGER DEFAULT 1,
		min_lat REAL, max_lat REAL, mean_lat REAL,
		min_lon REAL, max_lon REAL, mean_lon REAL,
		min_alt_ft INTEGER, max_alt_ft INTEGER, mean_alt_ft REAL,
		min_speed_kt REAL, max_speed_kt REAL, mean_speed_kt REAL,
		min_track REAL, max_track REAL,
		min_dist_nm REAL, max_dist_nm REAL,
		PRIMARY KEY (icao24, window_start)
	);
	CREATE INDEX IF NOT EXISTS idx_rollup_icao ON sightings_rollup(icao24);
	CREATE INDEX IF NOT EXISTS idx_rollup_window ON sightings_rollup(window_start);
	`
	_, err := d.Exec(ddl)
	if err != nil {
		return err
	}
	_, _ = d.Exec("ALTER TABLE aircraft ADD COLUMN min_ground_speed_kt REAL")
	_, _ = d.Exec("ALTER TABLE aircraft ADD COLUMN total_flights INTEGER DEFAULT 1")
	return nil
}

func (d *DB) upsertSighting(icao24, callsign, category string, lat, lon float64, alt, rssi int, gs, track float64, distNM float64, bearing int, seenAt int64, isMil bool) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	tx, err := d.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`INSERT INTO sightings (icao24, callsign, lat, lon, altitude_ft, ground_speed_kt, track, distance_nm, bearing_deg, rssi, category, seen_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		icao24, callsign, lat, lon, alt, gs, track, distNM, bearing, rssi, category, seenAt)
	if err != nil {
		return err
	}

	var exists int
	err = tx.QueryRow("SELECT COUNT(*) FROM aircraft WHERE icao24 = ?", icao24).Scan(&exists)
	if err != nil {
		return err
	}
	if exists == 0 {
		ci := lookupCountry(icao24)
		if isMil {
			if ci.Name == "" {
				ci.Name = "Military"
			} else if !strings.Contains(ci.Name, "(Military)") {
				ci.Name += " (Military)"
			}
		}
		_, err = tx.Exec(`INSERT INTO aircraft (icao24, country, country_flag, is_military, category, first_seen, last_seen, last_callsign,
			total_sightings, total_flights, min_distance_nm, max_distance_nm, min_altitude_ft, max_altitude_ft, max_ground_speed_kt, min_ground_speed_kt,
			last_lat, last_lon, last_altitude, last_speed, last_seen_pos, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, 1, 1, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			icao24, ci.Name, ci.Flag, boolToInt(isMil), category, seenAt, seenAt, callsign,
			distNM, distNM, alt, alt, gs, gs,
			lat, lon, alt, gs, seenAt, seenAt)
	} else {
		_, err = tx.Exec(`UPDATE aircraft SET
			last_seen = ?, last_callsign = CASE WHEN ? != '' THEN ? ELSE last_callsign END,
			total_sightings = total_sightings + 1,
			total_flights = total_flights + CASE WHEN (? - last_seen) > 1800 THEN 1 ELSE 0 END,
			category = CASE WHEN ? != '' THEN ? ELSE category END,
			min_distance_nm = MIN(COALESCE(min_distance_nm, ?), ?),
			max_distance_nm = MAX(COALESCE(max_distance_nm, ?), ?),
			min_altitude_ft = MIN(COALESCE(min_altitude_ft, ?), ?),
			max_altitude_ft = MAX(COALESCE(max_altitude_ft, ?), ?),
			max_ground_speed_kt = MAX(COALESCE(max_ground_speed_kt, 0), ?),
			min_ground_speed_kt = CASE WHEN ? > 0 THEN MIN(COALESCE(min_ground_speed_kt, ?), ?) ELSE min_ground_speed_kt END,
			last_lat = ?, last_lon = ?, last_altitude = ?, last_speed = ?,
			last_seen_pos = ?, updated_at = ?
			WHERE icao24 = ?`,
			seenAt, callsign, callsign,
			seenAt,
			category, category,
			distNM, distNM, distNM, distNM,
			alt, alt, alt, alt,
			gs,
			gs, gs, gs,
			lat, lon, alt, gs,
			seenAt, seenAt,
			icao24)
	}
	if err != nil {
		return err
	}

	date := time.Unix(seenAt, 0).UTC().Format("2006-01-02")
	var dsExists int
	err = tx.QueryRow("SELECT COUNT(*) FROM daily_stats WHERE date = ?", date).Scan(&dsExists)
	if err != nil {
		return err
	}

	var firstToday int
	_ = tx.QueryRow(`SELECT COUNT(*) FROM sightings WHERE icao24 = ? AND date(seen_at, 'unixepoch') = ?`, icao24, date).Scan(&firstToday)

	if dsExists == 0 {
		_, err = tx.Exec(`INSERT INTO daily_stats (date, unique_aircraft, total_sightings, military_aircraft, max_distance_nm, max_altitude_ft, max_speed_kt)
			VALUES (?, 1, 1, ?, ?, ?, ?)`,
			date, boolToInt(isMil), distNM, alt, gs)
	} else {
		milInc := 0
		if isMil && firstToday == 0 {
			milInc = 1
		}
		unqInc := 0
		if firstToday == 0 {
			unqInc = 1
		}
		_, err = tx.Exec(`UPDATE daily_stats SET
			total_sightings = total_sightings + 1,
			unique_aircraft = unique_aircraft + ?,
			military_aircraft = military_aircraft + ?,
			max_distance_nm = MAX(max_distance_nm, ?),
			max_altitude_ft = MAX(max_altitude_ft, ?),
			max_speed_kt = MAX(max_speed_kt, ?)
			WHERE date = ?`,
			unqInc, milInc, distNM, alt, gs, date)
	}
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (d *DB) updateEnrichment(icao24, reg, manufacturer, model, operator, operatorIcao, category string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now().Unix()
	_, err := d.Exec(`UPDATE aircraft SET registration=?, manufacturer=?, model=?, operator_name=?, operator_icao=?, category=?, enrichment_ts=? WHERE icao24=?`,
		reg, manufacturer, model, operator, operatorIcao, category, now, icao24)
	return err
}

func (d *DB) markEnrichmentAttempt(icao24 string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	now := time.Now().Unix()
	_, err := d.Exec(`UPDATE aircraft SET enrichment_ts = ? WHERE icao24 = ? AND enrichment_ts = 0`, now, icao24)
	return err
}

func (d *DB) getCallsign(icao24 string) (string, error) {
	var cs string
	err := d.QueryRow(`SELECT last_callsign FROM aircraft WHERE icao24 = ?`, icao24).Scan(&cs)
	if err != nil {
		return "", err
	}
	return cs, nil
}

func (d *DB) getUnenriched(count int) ([]string, error) {
	rows, err := d.Query(`SELECT icao24 FROM aircraft WHERE enrichment_ts = 0 LIMIT ?`, count)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (d *DB) getStaleEnriched(count int) ([]string, error) {
	cutoff := time.Now().Add(-6 * time.Hour).Unix()
	rows, err := d.Query(`SELECT icao24 FROM aircraft WHERE enrichment_ts > 0 AND enrichment_ts < ? AND (operator_name = '' OR manufacturer = '') ORDER BY enrichment_ts ASC LIMIT ?`, cutoff, count)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

type globalStats struct {
	SiteName         string  `json:"site_name"`
	SiteLat          float64 `json:"site_lat"`
	SiteLon          float64 `json:"site_lon"`
	SiteAlt          string  `json:"site_alt"`
	FlightAwareSiteID string  `json:"flightaware_site_id"`
	TotalAircraft    int     `json:"total_aircraft"`
	TotalSightings   int64   `json:"total_sightings"`
	MilitaryAircraft int     `json:"military_aircraft"`
	FirstSeen        int64   `json:"first_seen"`
	LastSeen         int64   `json:"last_seen"`
	FurthestNM       float64 `json:"furthest_nm"`
	FurthestIcao     string  `json:"furthest_icao"`
	FurthestCall     string  `json:"furthest_call"`
	ClosestNM        float64 `json:"closest_nm"`
	ClosestIcao      string  `json:"closest_icao"`
	ClosestCall      string  `json:"closest_call"`
	HighestFt        int     `json:"highest_ft"`
	HighestIcao      string  `json:"highest_icao"`
	HighestCall      string  `json:"highest_call"`
	LowestFt         int     `json:"lowest_ft"`
	LowestIcao       string  `json:"lowest_icao"`
	LowestCall       string  `json:"lowest_call"`
	FastestKt        float64 `json:"fastest_kt"`
	FastestIcao      string  `json:"fastest_icao"`
	FastestCall      string  `json:"fastest_call"`
	SlowestKt        float64 `json:"slowest_kt"`
	SlowestIcao      string  `json:"slowest_icao"`
	SlowestCall      string  `json:"slowest_call"`
}

func (d *DB) getGlobalStats() (*globalStats, error) {
	s := &globalStats{}
	err := d.QueryRow("SELECT COUNT(*) FROM aircraft").Scan(&s.TotalAircraft)
	if err != nil {
		return nil, err
	}
	err = d.QueryRow("SELECT COALESCE(SUM(total_sightings),0) FROM aircraft").Scan(&s.TotalSightings)
	if err != nil {
		return nil, err
	}
	err = d.QueryRow("SELECT COUNT(*) FROM aircraft WHERE is_military=1").Scan(&s.MilitaryAircraft)
	if err != nil {
		return nil, err
	}
	d.QueryRow("SELECT MIN(first_seen) FROM aircraft").Scan(&s.FirstSeen)
	d.QueryRow("SELECT MAX(last_seen) FROM aircraft").Scan(&s.LastSeen)

	row := d.QueryRow(`SELECT icao24, last_callsign, max_distance_nm FROM aircraft ORDER BY max_distance_nm DESC LIMIT 1`)
	row.Scan(&s.FurthestIcao, &s.FurthestCall, &s.FurthestNM)

	row = d.QueryRow(`SELECT icao24, last_callsign, min_distance_nm FROM aircraft WHERE min_distance_nm > 0 ORDER BY min_distance_nm ASC LIMIT 1`)
	row.Scan(&s.ClosestIcao, &s.ClosestCall, &s.ClosestNM)

	row = d.QueryRow(`SELECT icao24, last_callsign, max_altitude_ft FROM aircraft ORDER BY max_altitude_ft DESC LIMIT 1`)
	row.Scan(&s.HighestIcao, &s.HighestCall, &s.HighestFt)

	row = d.QueryRow(`SELECT icao24, last_callsign, min_altitude_ft FROM aircraft WHERE min_altitude_ft > 0 ORDER BY min_altitude_ft ASC LIMIT 1`)
	row.Scan(&s.LowestIcao, &s.LowestCall, &s.LowestFt)

	row = d.QueryRow(`SELECT icao24, last_callsign, max_ground_speed_kt FROM aircraft ORDER BY max_ground_speed_kt DESC LIMIT 1`)
	row.Scan(&s.FastestIcao, &s.FastestCall, &s.FastestKt)

	row = d.QueryRow(`SELECT icao24, last_callsign, min_ground_speed_kt FROM aircraft WHERE min_ground_speed_kt > 0 ORDER BY min_ground_speed_kt ASC LIMIT 1`)
	row.Scan(&s.SlowestIcao, &s.SlowestCall, &s.SlowestKt)

	return s, nil
}

type aircraftRow struct {
	ICAO24         string  `json:"icao24"`
	Registration   string  `json:"reg"`
	Manufacturer   string  `json:"manufacturer"`
	Model          string  `json:"model"`
	Operator       string  `json:"operator"`
	OperatorICAO   string  `json:"operator_icao"`
	Country        string  `json:"country"`
	CountryFlag    string  `json:"country_flag"`
	IsMilitary     bool    `json:"is_military"`
	Category       string  `json:"category"`
	LastCallsign   string  `json:"last_callsign"`
	FirstSeen      int64   `json:"first_seen"`
	LastSeen       int64   `json:"last_seen"`
	TotalSightings int     `json:"total_sightings"`
	TotalFlights   int     `json:"total_flights"`
	MinDistNM      float64 `json:"min_dist_nm"`
	MaxDistNM      float64 `json:"max_dist_nm"`
	MinAltFt       int     `json:"min_alt_ft"`
	MaxAltFt       int     `json:"max_alt_ft"`
	MaxSpeedKt     float64 `json:"max_speed_kt"`
	LastLat        float64 `json:"last_lat"`
	LastLon        float64 `json:"last_lon"`
	LastAlt        int     `json:"last_alt"`
	LastSpeed      float64 `json:"last_speed"`
	LastSeenPos    int64   `json:"last_seen_pos"`
	UpdatedAt      int64   `json:"updated_at"`
}

func (d *DB) getLeaderboard(sortBy string, militaryOnly bool, limit int) ([]aircraftRow, error) {
	if limit <= 0 {
		limit = 50
	}
	milFilter := ""
	if militaryOnly {
		milFilter = "WHERE is_military = 1"
	}

	orderBy := "total_sightings DESC"
	switch sortBy {
	case "distance":
		orderBy = "max_distance_nm DESC"
	case "altitude":
		orderBy = "max_altitude_ft DESC"
	case "speed":
		orderBy = "max_ground_speed_kt DESC"
	case "recent":
		orderBy = "last_seen DESC"
	default:
		orderBy = "total_sightings DESC"
	}

	query := fmt.Sprintf(`SELECT icao24, registration, manufacturer, model, operator_name, operator_icao,
		country, country_flag, is_military, category, last_callsign,
		first_seen, last_seen, total_sightings, COALESCE(total_flights,1),
		COALESCE(min_distance_nm,0), COALESCE(max_distance_nm,0),
		COALESCE(min_altitude_ft,0), COALESCE(max_altitude_ft,0),
		COALESCE(max_ground_speed_kt,0),
		COALESCE(last_lat,0), COALESCE(last_lon,0),
		COALESCE(last_altitude,0), COALESCE(last_speed,0),
		last_seen_pos, updated_at
		FROM aircraft %s ORDER BY %s LIMIT ?`, milFilter, orderBy)

	rows, err := d.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAircraft(rows)
}

func (d *DB) getAircraft(icao24 string) (*aircraftRow, error) {
	row := d.QueryRow(`SELECT icao24, registration, manufacturer, model, operator_name, operator_icao,
		country, country_flag, is_military, category, last_callsign,
		first_seen, last_seen, total_sightings, COALESCE(total_flights,1),
		COALESCE(min_distance_nm,0), COALESCE(max_distance_nm,0),
		COALESCE(min_altitude_ft,0), COALESCE(max_altitude_ft,0),
		COALESCE(max_ground_speed_kt,0),
		COALESCE(last_lat,0), COALESCE(last_lon,0),
		COALESCE(last_altitude,0), COALESCE(last_speed,0),
		last_seen_pos, updated_at
		FROM aircraft WHERE icao24 = ?`, icao24)
	var a aircraftRow
	var mil int
	err := row.Scan(&a.ICAO24, &a.Registration, &a.Manufacturer, &a.Model, &a.Operator, &a.OperatorICAO,
		&a.Country, &a.CountryFlag, &mil, &a.Category, &a.LastCallsign,
		&a.FirstSeen, &a.LastSeen, &a.TotalSightings, &a.TotalFlights,
		&a.MinDistNM, &a.MaxDistNM, &a.MinAltFt, &a.MaxAltFt, &a.MaxSpeedKt,
		&a.LastLat, &a.LastLon, &a.LastAlt, &a.LastSpeed,
		&a.LastSeenPos, &a.UpdatedAt)
	a.IsMilitary = mil == 1
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (d *DB) getAircraftSightings(icao24 string, limit int) ([]sightingRow, error) {
	if limit <= 0 {
		limit = 200
	}
	rows, err := d.Query(`SELECT icao24, callsign, lat, lon, altitude_ft, ground_speed_kt, track, distance_nm, bearing_deg, rssi, category, seen_at
		FROM sightings WHERE icao24 = ? ORDER BY seen_at DESC LIMIT ?`, icao24, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []sightingRow
	for rows.Next() {
		var s sightingRow
		err := rows.Scan(&s.ICAO24, &s.Callsign, &s.Lat, &s.Lon, &s.AltFt, &s.SpeedKt, &s.Track, &s.DistNM, &s.Bearing, &s.RSSI, &s.Category, &s.SeenAt)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

type callsignRow struct {
	Callsign string `json:"callsign"`
	FirstSeen int64 `json:"first_seen"`
	LastSeen  int64 `json:"last_seen"`
	Count     int   `json:"count"`
}

func (d *DB) getAircraftCallsigns(icao24 string) ([]callsignRow, error) {
	rows, err := d.Query(`SELECT callsign, MIN(seen_at), MAX(seen_at), COUNT(*)
		FROM sightings WHERE icao24 = ? AND callsign != ''
		GROUP BY callsign ORDER BY MAX(seen_at) DESC LIMIT 20`, icao24)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []callsignRow
	for rows.Next() {
		var c callsignRow
		if err := rows.Scan(&c.Callsign, &c.FirstSeen, &c.LastSeen, &c.Count); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

type rollupRow struct {
	ICAO24       string  `json:"icao24"`
	Callsign     string  `json:"callsign"`
	WindowStart  int64   `json:"window_start"`
	SightingCount int    `json:"sighting_count"`
	MeanLat      float64 `json:"mean_lat"`
	MeanLon      float64 `json:"mean_lon"`
	MinAltFt     int     `json:"min_alt_ft"`
	MaxAltFt     int     `json:"max_alt_ft"`
	MeanAltFt    float64 `json:"mean_alt_ft"`
	MinSpeedKt   float64 `json:"min_speed_kt"`
	MaxSpeedKt   float64 `json:"max_speed_kt"`
	MeanSpeedKt  float64 `json:"mean_speed_kt"`
	MinTrack     float64 `json:"min_track"`
	MaxTrack     float64 `json:"max_track"`
	MinDistNM    float64 `json:"min_dist_nm"`
	MaxDistNM    float64 `json:"max_dist_nm"`
}

func (d *DB) getAircraftRollups(icao24 string, limit int) ([]rollupRow, error) {
	if limit <= 0 {
		limit = 1000
	}
	rows, err := d.Query(`SELECT icao24, callsign, window_start, sighting_count,
		COALESCE(mean_lat,0), COALESCE(mean_lon,0),
		COALESCE(min_alt_ft,0), COALESCE(max_alt_ft,0), COALESCE(mean_alt_ft,0),
		COALESCE(min_speed_kt,0), COALESCE(max_speed_kt,0), COALESCE(mean_speed_kt,0),
		COALESCE(min_track,0), COALESCE(max_track,0),
		COALESCE(min_dist_nm,0), COALESCE(max_dist_nm,0)
		FROM sightings_rollup WHERE icao24 = ? ORDER BY window_start DESC LIMIT ?`, icao24, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []rollupRow
	for rows.Next() {
		var r rollupRow
		if err := rows.Scan(&r.ICAO24, &r.Callsign, &r.WindowStart, &r.SightingCount,
			&r.MeanLat, &r.MeanLon, &r.MinAltFt, &r.MaxAltFt, &r.MeanAltFt,
			&r.MinSpeedKt, &r.MaxSpeedKt, &r.MeanSpeedKt,
			&r.MinTrack, &r.MaxTrack, &r.MinDistNM, &r.MaxDistNM); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (d *DB) getRecentSightings(limit int) ([]sightingRow, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := d.Query(`SELECT icao24, callsign, lat, lon, altitude_ft, ground_speed_kt, track, distance_nm, bearing_deg, rssi, category, seen_at
		FROM sightings ORDER BY seen_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []sightingRow
	for rows.Next() {
		var s sightingRow
		err := rows.Scan(&s.ICAO24, &s.Callsign, &s.Lat, &s.Lon, &s.AltFt, &s.SpeedKt, &s.Track, &s.DistNM, &s.Bearing, &s.RSSI, &s.Category, &s.SeenAt)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (d *DB) getRecentMilitarySightings(limit int) ([]militarySightingRow, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := d.Query(`SELECT s.icao24, s.callsign, s.lat, s.lon, s.altitude_ft, s.ground_speed_kt, s.track,
		s.distance_nm, s.bearing_deg, s.category, s.seen_at,
		a.registration, a.manufacturer, a.model, a.operator_name, a.country, a.country_flag
		FROM sightings s
		JOIN aircraft a ON s.icao24 = a.icao24
		JOIN (SELECT icao24, MAX(id) AS max_id FROM sightings GROUP BY icao24) latest
			ON s.icao24 = latest.icao24 AND s.id = latest.max_id
		WHERE a.is_military = 1
		ORDER BY s.seen_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []militarySightingRow
	for rows.Next() {
		var s militarySightingRow
		err := rows.Scan(&s.ICAO24, &s.Callsign, &s.Lat, &s.Lon, &s.AltFt, &s.SpeedKt, &s.Track,
			&s.DistNM, &s.Bearing, &s.Category, &s.SeenAt,
			&s.Registration, &s.Manufacturer, &s.Model, &s.Operator, &s.Country, &s.CountryFlag)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (d *DB) getDailyStats(limit int) ([]dailyRow, error) {
	if limit <= 0 {
		limit = 30
	}
	rows, err := d.Query(`SELECT d.date,
		COALESCE((SELECT COUNT(DISTINCT s.icao24) FROM sightings s WHERE date(s.seen_at, 'unixepoch') = d.date), 0),
		d.total_sightings,
		COALESCE((SELECT COUNT(DISTINCT s.icao24) FROM sightings s JOIN aircraft a ON s.icao24 = a.icao24 WHERE date(s.seen_at, 'unixepoch') = d.date AND a.is_military = 1), 0),
		d.max_distance_nm, d.max_altitude_ft, d.max_speed_kt
		FROM daily_stats d ORDER BY d.date DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []dailyRow
	for rows.Next() {
		var dr dailyRow
		err := rows.Scan(&dr.Date, &dr.Unique, &dr.Total, &dr.Military, &dr.MaxDist, &dr.MaxAlt, &dr.MaxSpeed)
		if err != nil {
			return nil, err
		}
		out = append(out, dr)
	}
	return out, rows.Err()
}

func (d *DB) getNearest(maxAlt int, maxDistNM float64, seenSecs int) (*nearestInfo, error) {
	cutoff := time.Now().Unix() - int64(seenSecs)
	row := d.QueryRow(`SELECT s.icao24, s.callsign, s.lat, s.lon, s.altitude_ft, s.ground_speed_kt, s.track, s.distance_nm, s.bearing_deg, s.seen_at,
		a.registration, a.manufacturer, a.model, a.operator_name, a.country, a.country_flag, a.is_military
		FROM sightings s LEFT JOIN aircraft a ON s.icao24 = a.icao24
		WHERE s.seen_at > ? AND s.altitude_ft <= ? AND s.distance_nm <= ? AND s.distance_nm > 0
		ORDER BY s.distance_nm ASC LIMIT 1`, cutoff, maxAlt, maxDistNM)
	var n nearestInfo
	var mil int
	err := row.Scan(&n.ICAO24, &n.Callsign, &n.Lat, &n.Lon, &n.AltitudeFt, &n.SpeedKt, &n.Track, &n.DistNM, &n.Bearing, &n.SeenAt,
		&n.Registration, &n.Manufacturer, &n.Model, &n.Operator, &n.Country, &n.CountryFlag, &mil)
	n.IsMilitary = mil == 1
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &n, nil
}

func (d *DB) fixMilitaryFlags() {
	rows, err := d.Query("SELECT icao24 FROM aircraft")
	if err != nil {
		return
	}
	var all []string
	for rows.Next() {
		var icao string
		if err := rows.Scan(&icao); err != nil {
			continue
		}
		all = append(all, icao)
	}
	rows.Close()

	for _, icao := range all {
		shouldBeMil := boolToInt(isMilitaryHex(icao))
		d.Exec("UPDATE aircraft SET is_military = ? WHERE icao24 = ? AND is_military != ?", shouldBeMil, icao, shouldBeMil)
	}
}

func (d *DB) backfillCountries() {
	rows, err := d.Query("SELECT icao24 FROM aircraft WHERE country = ''")
	if err != nil {
		return
	}
	var all []string
	for rows.Next() {
		var icao string
		if err := rows.Scan(&icao); err != nil {
			continue
		}
		all = append(all, icao)
	}
	rows.Close()

	for _, icao := range all {
		ci := lookupCountry(icao)
		if ci.Name != "" {
			d.Exec("UPDATE aircraft SET country = ?, country_flag = ? WHERE icao24 = ? AND country = ''", ci.Name, ci.Flag, icao)
		}
	}
}

func (d *DB) backfillOperators() {
	rows, err := d.Query("SELECT icao24, last_callsign, operator_icao FROM aircraft WHERE operator_name = '' AND last_callsign != ''")
	if err != nil {
		return
	}
	type row struct {
		icao, call, opIcao string
	}
	var empty []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.icao, &r.call, &r.opIcao); err != nil {
			continue
		}
		empty = append(empty, r)
	}
	rows.Close()

	for _, r := range empty {
		name, icao := inferOperatorFromCallsign(r.call)
		if name == "" {
			name, icao = inferOperatorFromIcao(r.opIcao)
		}
		if name != "" {
			d.Exec("UPDATE aircraft SET operator_name = ?, operator_icao = CASE WHEN operator_icao = '' THEN ? ELSE operator_icao END WHERE icao24 = ? AND operator_name = ''", name, icao, r.icao)
		}
	}

	rows2, err := d.Query("SELECT icao24, operator_name, manufacturer FROM aircraft WHERE operator_name != '' OR manufacturer != ''")
	if err != nil {
		return
	}
	type row2 struct {
		icao, op, mfr string
	}
	var normRows []row2
	for rows2.Next() {
		var r row2
		if err := rows2.Scan(&r.icao, &r.op, &r.mfr); err != nil {
			continue
		}
		normRows = append(normRows, r)
	}
	rows2.Close()

	for _, r := range normRows {
		normOp := normalizeName(r.op)
		normMfr := normalizeName(r.mfr)
		if normOp != r.op || normMfr != r.mfr {
			d.Exec("UPDATE aircraft SET operator_name = ?, manufacturer = ? WHERE icao24 = ?", normOp, normMfr, r.icao)
		}
	}
}

func (d *DB) getNearestByCoords(maxAlt int, seenSecs int, userLat, userLon float64) (*nearestInfo, error) {
	cutoff := time.Now().Unix() - int64(seenSecs)
	rows, err := d.Query(`SELECT s.icao24, s.callsign, s.lat, s.lon, s.altitude_ft, s.ground_speed_kt, s.track, s.seen_at,
		a.registration, a.manufacturer, a.model, a.operator_name, a.country, a.country_flag, a.is_military
		FROM sightings s LEFT JOIN aircraft a ON s.icao24 = a.icao24
		WHERE s.seen_at > ? AND s.altitude_ft <= ? AND s.lat IS NOT NULL AND s.lon IS NOT NULL
		ORDER BY s.distance_nm ASC LIMIT 50`, cutoff, maxAlt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var best *nearestInfo
	var bestDist float64 = 1e9
	for rows.Next() {
		var n nearestInfo
		var mil int
		var lat, lon float64
		err := rows.Scan(&n.ICAO24, &n.Callsign, &lat, &lon, &n.AltitudeFt, &n.SpeedKt, &n.Track, &n.SeenAt,
			&n.Registration, &n.Manufacturer, &n.Model, &n.Operator, &n.Country, &n.CountryFlag, &mil)
		if err != nil {
			continue
		}
		n.IsMilitary = mil == 1
		n.Lat = lat
		n.Lon = lon
		dist := haversineNM(userLat, userLon, lat, lon)
		if dist < bestDist {
			bestDist = dist
			n.DistNM = dist
			n.Bearing = int(initialBearing(userLat, userLon, lat, lon))
			cp := n
			best = &cp
		}
	}
	return best, rows.Err()
}

func (d *DB) uniqueAircraftToday() int {
	var c int
	d.QueryRow(`SELECT COUNT(*) FROM aircraft WHERE date(last_seen, 'unixepoch') = date('now')`).Scan(&c)
	return c
}

func (d *DB) compressAndClean(retentionHours int) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	cutoff := time.Now().Add(-time.Duration(retentionHours) * time.Hour).Unix()

	tx, err := d.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	const windowSec = 300
	_, err = tx.Exec(`INSERT OR REPLACE INTO sightings_rollup
		(icao24, callsign, window_start, sighting_count,
		 min_lat, max_lat, mean_lat, min_lon, max_lon, mean_lon,
		 min_alt_ft, max_alt_ft, mean_alt_ft,
		 min_speed_kt, max_speed_kt, mean_speed_kt,
		 min_track, max_track, min_dist_nm, max_dist_nm)
		SELECT
			icao24,
			(SELECT callsign FROM sightings s2 WHERE s2.icao24 = s.icao24 AND (s2.seen_at / ?) = (s.seen_at / ?) AND s2.callsign != '' GROUP BY s2.callsign ORDER BY COUNT(*) DESC LIMIT 1),
			(seen_at / ?) * ?,
			COUNT(*),
			MIN(lat), MAX(lat), AVG(lat),
			MIN(lon), MAX(lon), AVG(lon),
			MIN(altitude_ft), MAX(altitude_ft), AVG(altitude_ft),
			MIN(ground_speed_kt), MAX(ground_speed_kt), AVG(ground_speed_kt),
			MIN(track), MAX(track),
			MIN(distance_nm), MAX(distance_nm)
		FROM sightings s
		WHERE seen_at < ? AND lat IS NOT NULL AND lon IS NOT NULL
		GROUP BY icao24, seen_at / ?`,
		windowSec, windowSec, windowSec, windowSec, cutoff, windowSec)
	if err != nil {
		return err
	}

	if _, err = tx.Exec("DELETE FROM sightings WHERE seen_at < ?", cutoff); err != nil {
		return err
	}

	return tx.Commit()
}

func (d *DB) cleanOldSightings(retentionHours int) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	cutoff := time.Now().Add(-time.Duration(retentionHours) * time.Hour).Unix()
	_, err := d.Exec("DELETE FROM sightings WHERE seen_at < ?", cutoff)
	return err
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func scanAircraft(rows *sql.Rows) ([]aircraftRow, error) {
	var out []aircraftRow
	for rows.Next() {
		var a aircraftRow
		var mil int
		err := rows.Scan(&a.ICAO24, &a.Registration, &a.Manufacturer, &a.Model, &a.Operator, &a.OperatorICAO,
			&a.Country, &a.CountryFlag, &mil, &a.Category, &a.LastCallsign,
			&a.FirstSeen, &a.LastSeen, &a.TotalSightings, &a.TotalFlights,
			&a.MinDistNM, &a.MaxDistNM, &a.MinAltFt, &a.MaxAltFt, &a.MaxSpeedKt,
			&a.LastLat, &a.LastLon, &a.LastAlt, &a.LastSpeed,
			&a.LastSeenPos, &a.UpdatedAt)
		if err != nil {
			return nil, err
		}
		a.IsMilitary = mil == 1
		out = append(out, a)
	}
	return out, rows.Err()
}

type sightingRow struct {
	ICAO24   string  `json:"icao24"`
	Callsign string  `json:"callsign"`
	Lat      float64 `json:"lat"`
	Lon      float64 `json:"lon"`
	AltFt    int     `json:"alt_ft"`
	SpeedKt  float64 `json:"speed_kt"`
	Track    float64 `json:"track"`
	DistNM   float64 `json:"dist_nm"`
	Bearing  int     `json:"bearing"`
	RSSI     float64 `json:"rssi"`
	Category string  `json:"category"`
	SeenAt   int64   `json:"seen_at"`
}

type militarySightingRow struct {
	ICAO24       string  `json:"icao24"`
	Callsign     string  `json:"callsign"`
	Lat          float64 `json:"lat"`
	Lon          float64 `json:"lon"`
	AltFt        int     `json:"alt_ft"`
	SpeedKt      float64 `json:"speed_kt"`
	Track        float64 `json:"track"`
	DistNM       float64 `json:"dist_nm"`
	Bearing      int     `json:"bearing"`
	Category     string  `json:"category"`
	SeenAt       int64   `json:"seen_at"`
	Registration string  `json:"reg"`
	Manufacturer string  `json:"manufacturer"`
	Model        string  `json:"model"`
	Operator     string  `json:"operator"`
	Country      string  `json:"country"`
	CountryFlag  string  `json:"country_flag"`
}

type dailyRow struct {
	Date      string  `json:"date"`
	Unique    int     `json:"unique_aircraft"`
	Total     int     `json:"total_sightings"`
	Military  int     `json:"military_aircraft"`
	MaxDist   float64 `json:"max_distance_nm"`
	MaxAlt    int     `json:"max_altitude_ft"`
	MaxSpeed  float64 `json:"max_speed_kt"`
}

type nearestInfo struct {
	ICAO24       string   `json:"icao24"`
	Callsign     string   `json:"callsign"`
	Lat          float64  `json:"lat"`
	Lon          float64  `json:"lon"`
	AltitudeFt   int      `json:"altitude_ft"`
	SpeedKt      float64  `json:"speed_kt"`
	Track        float64  `json:"track"`
	DistNM       float64  `json:"dist_nm"`
	Bearing      int      `json:"bearing"`
	SeenAt       int64    `json:"seen_at"`
	Registration *string  `json:"reg"`
	Manufacturer *string  `json:"manufacturer"`
	Model        *string  `json:"model"`
	Operator     *string  `json:"operator"`
	Country      *string  `json:"country"`
	CountryFlag  *string  `json:"country_flag"`
	IsMilitary   bool     `json:"is_military"`
}
