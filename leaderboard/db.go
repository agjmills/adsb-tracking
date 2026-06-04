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
	`
	_, err := d.Exec(ddl)
	if err != nil {
		return err
	}
	_, _ = d.Exec("ALTER TABLE aircraft ADD COLUMN min_ground_speed_kt REAL")
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
			total_sightings, min_distance_nm, max_distance_nm, min_altitude_ft, max_altitude_ft, max_ground_speed_kt, min_ground_speed_kt,
			last_lat, last_lon, last_altitude, last_speed, last_seen_pos, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			icao24, ci.Name, ci.Flag, boolToInt(isMil), category, seenAt, seenAt, callsign,
			distNM, distNM, alt, alt, gs, gs,
			lat, lon, alt, gs, seenAt, seenAt)
	} else {
		_, err = tx.Exec(`UPDATE aircraft SET
			last_seen = ?, last_callsign = CASE WHEN ? != '' THEN ? ELSE last_callsign END,
			total_sightings = total_sightings + 1,
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
	if dsExists == 0 {
		_, err = tx.Exec(`INSERT INTO daily_stats (date, unique_aircraft, total_sightings, military_aircraft, max_distance_nm, max_altitude_ft, max_speed_kt)
			VALUES (?, 0, 1, 0, ?, ?, ?)`,
			date, distNM, alt, gs)
	} else {
		_, err = tx.Exec(`UPDATE daily_stats SET
			total_sightings = total_sightings + 1,
			max_distance_nm = MAX(max_distance_nm, ?),
			max_altitude_ft = MAX(max_altitude_ft, ?),
			max_speed_kt = MAX(max_speed_kt, ?)
			WHERE date = ?`,
			distNM, alt, gs, date)
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

type globalStats struct {
	SiteName         string  `json:"site_name"`
	SiteLat          float64 `json:"site_lat"`
	SiteLon          float64 `json:"site_lon"`
	SiteAlt          string  `json:"site_alt"`
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
		first_seen, last_seen, total_sightings,
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
		first_seen, last_seen, total_sightings,
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
		&a.FirstSeen, &a.LastSeen, &a.TotalSightings,
		&a.MinDistNM, &a.MaxDistNM, &a.MinAltFt, &a.MaxAltFt, &a.MaxSpeedKt,
		&a.LastLat, &a.LastLon, &a.LastAlt, &a.LastSpeed,
		&a.LastSeenPos, &a.UpdatedAt)
	a.IsMilitary = mil == 1
	if err != nil {
		return nil, err
	}
	return &a, nil
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

func (d *DB) uniqueAircraftToday() int {
	var c int
	d.QueryRow(`SELECT COUNT(*) FROM aircraft WHERE date(last_seen, 'unixepoch') = date('now')`).Scan(&c)
	return c
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
			&a.FirstSeen, &a.LastSeen, &a.TotalSightings,
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
