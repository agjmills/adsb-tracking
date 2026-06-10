package main

import (
	"context"
	"encoding/json"
	"log"
	"math"
	"net/http"
	"time"
)

type dumpAircraft struct {
	Lat      *float64 `json:"lat"`
	Lon      *float64 `json:"lon"`
	AltBaro  *int     `json:"alt_baro"`
	AltGeom  *int     `json:"alt_geom"`
	Flight   *string  `json:"flight"`
	Hex      *string  `json:"hex"`
	Seen     *float64 `json:"seen"`
	GS       *float64 `json:"gs"`
	Track    *float64 `json:"track"`
	RSSI     *float64 `json:"rssi"`
	Category *string  `json:"category"`
}

type dumpResp struct {
	Aircraft []dumpAircraft `json:"aircraft"`
	Now      *float64       `json:"now"`
}

func runPoller(ctx context.Context, db *DB, dumpURL string, homeLat, homeLon float64, interval time.Duration) {
	client := &http.Client{Timeout: 5 * time.Second}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			dr, err := fetchDump(client, dumpURL)
			if err != nil {
				log.Printf("fetch error: %v", err)
				continue
			}
			now := time.Now().Unix()
			for _, a := range dr.Aircraft {
				if a.Hex == nil {
					continue
				}
				hex := *a.Hex

				callsign := ""
				if a.Flight != nil {
					callsign = *a.Flight
				}

				isMil := isMilitaryHex(hex) || isMilitaryCallsign(callsign)

				if a.Lat == nil || a.Lon == nil {
					if !isMil {
						continue
					}
				}

				lat := 0.0
				lon := 0.0
				if a.Lat != nil {
					lat = *a.Lat
				}
				if a.Lon != nil {
					lon = *a.Lon
				}

				alt := coalesceInt(a.AltBaro, a.AltGeom)
				if alt == nil {
					altVal := 0
					alt = &altVal
				}
				gs := 0.0
				if a.GS != nil {
					gs = *a.GS
				}
				track := 0.0
				if a.Track != nil {
					track = *a.Track
				}
				rssi := 0
				if a.RSSI != nil {
					rssi = int(*a.RSSI)
				}
				cat := ""
				if a.Category != nil {
					cat = *a.Category
				}

				dist := 0.0
				bearing := 0
				if lat != 0 || lon != 0 {
					dist = haversineNM(homeLat, homeLon, lat, lon)
					bearing = int(math.Round(initialBearing(homeLat, homeLon, lat, lon)))
				}

				if err := db.upsertSighting(hex, callsign, cat, lat, lon, *alt, rssi, gs, track, dist, bearing, now, isMil); err != nil {
					log.Printf("db upsert error for %s: %v", hex, err)
				}
			}
		}
	}
}

func fetchDump(c *http.Client, url string) (*dumpResp, error) {
	resp, err := c.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var dr dumpResp
	if err := json.NewDecoder(resp.Body).Decode(&dr); err != nil {
		return nil, err
	}
	return &dr, nil
}

func haversineNM(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 3440.065
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180
	φ1 := lat1 * math.Pi / 180
	φ2 := lat2 * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(φ1)*math.Cos(φ2)*math.Sin(dLon/2)*math.Sin(dLon/2)
	return 2 * R * math.Asin(math.Sqrt(a))
}

func initialBearing(lat1, lon1, lat2, lon2 float64) float64 {
	φ1 := lat1 * math.Pi / 180
	φ2 := lat2 * math.Pi / 180
	Δλ := (lon2 - lon1) * math.Pi / 180
	y := math.Sin(Δλ) * math.Cos(φ2)
	x := math.Cos(φ1)*math.Sin(φ2) - math.Sin(φ1)*math.Cos(φ2)*math.Cos(Δλ)
	θ := math.Atan2(y, x) * 180 / math.Pi
	if θ < 0 {
		θ += 360
	}
	return θ
}

func coalesceInt(a, b *int) *int {
	if a != nil {
		return a
	}
	return b
}
