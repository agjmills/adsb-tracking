package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"
)

func main() {
	homeLat := getenvFloat("HOME_LAT", 0)
	homeLon := getenvFloat("HOME_LON", 0)
	siteName := os.Getenv("SITE_NAME")
	if siteName == "" {
		siteName = "Unknown"
	}
	siteAlt := os.Getenv("SITE_ALT")
	if siteAlt == "" {
		siteAlt = "0m"
	}
	dumpURL := os.Getenv("DUMP1090_URL")
	if dumpURL == "" {
		dumpURL = "http://localhost/data/aircraft.json"
	}
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "/data/leaderboard.db"
	}
	pollInterval := getenvInt("POLL_INTERVAL", 5)
	cleanupHours := getenvInt("CLEANUP_HOURS", 24)
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	db, err := openDB(dbPath)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	go func() {
		time.Sleep(3 * time.Second)
		db.fixMilitaryFlags()
	}()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	go runPoller(ctx, db, dumpURL, homeLat, homeLon, time.Duration(pollInterval)*time.Second)

	enr := newEnricher(db)
	go enr.run(ctx)

	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				db.cleanOldSightings(cleanupHours)
			}
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/", serveUI)
	mux.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		stats, err := db.getGlobalStats()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		stats.SiteName = siteName
		stats.SiteLat = homeLat
		stats.SiteLon = homeLon
		stats.SiteAlt = siteAlt
		stats.FeederID = os.Getenv("PIAWARE_FEEDER_ID")
		writeJSON(w, stats)
	})
	mux.HandleFunc("/api/leaderboard", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		sortBy := q.Get("sort")
		mil := q.Get("military") == "1"
		limit := getQueryInt(q, "limit", 50)
		rows, err := db.getLeaderboard(sortBy, mil, limit)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		writeJSON(w, rows)
	})
	mux.HandleFunc("/api/aircraft/", func(w http.ResponseWriter, r *http.Request) {
		icao := strings.TrimPrefix(r.URL.Path, "/api/aircraft/")
		if icao == "" {
			http.Error(w, "missing icao", 400)
			return
		}
		ac, err := db.getAircraft(icao)
		if err != nil {
			http.Error(w, "not found", 404)
			return
		}
		writeJSON(w, ac)
	})
	mux.HandleFunc("/api/recent", func(w http.ResponseWriter, r *http.Request) {
		limit := getQueryInt(r.URL.Query(), "limit", 100)
		rows, err := db.getRecentSightings(limit)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		writeJSON(w, rows)
	})
	mux.HandleFunc("/api/nearest", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		maxAlt := getQueryInt(q, "max_alt", 12000)
		maxDist := getenvFloat("NEAREST_MAX_NM", 10)
		seenSecs := getQueryInt(q, "seen", 180)
		nearest, err := db.getNearest(maxAlt, maxDist, seenSecs)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		writeJSON(w, nearest)
	})
	mux.HandleFunc("/api/daily", func(w http.ResponseWriter, r *http.Request) {
		limit := getQueryInt(r.URL.Query(), "limit", 30)
		rows, err := db.getDailyStats(limit)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		writeJSON(w, rows)
	})
	mux.HandleFunc("/api/nearme", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		maxAlt := getQueryInt(q, "max_alt", 12000)
		seenSecs := getQueryInt(q, "seen", 180)
		userLat := getQueryFloat(q, "lat", 0)
		userLon := getQueryFloat(q, "lon", 0)
		if userLat == 0 && userLon == 0 {
			http.Error(w, "provide lat and lon", 400)
			return
		}
		nearest, err := db.getNearestByCoords(maxAlt, seenSecs, userLat, userLon)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		writeJSON(w, nearest)
	})
	mux.HandleFunc("/nearest", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		w.Write([]byte(nearestHTML))
	})

	radarURL, _ := url.Parse("http://ultrafeeder")
	radarProxy := httputil.NewSingleHostReverseProxy(radarURL)
	mux.HandleFunc("/radar/", func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = "/" + strings.TrimPrefix(r.URL.Path, "/radar/")
		r.URL.RawPath = ""
		radarProxy.ServeHTTP(w, r)
	})
	mux.HandleFunc("/radar", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/radar/", http.StatusMovedPermanently)
	})

	addr := ":" + port
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func serveUI(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Write([]byte(uiHTML))
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	json.NewEncoder(w).Encode(v)
}

func getenvFloat(key string, def float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return def
}

func getenvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}

func getQueryInt(q map[string][]string, key string, def int) int {
	if v, ok := q[key]; ok && len(v) > 0 {
		if i, err := strconv.Atoi(v[0]); err == nil {
			return i
		}
	}
	return def
}

func getQueryFloat(q map[string][]string, key string, def float64) float64 {
	if v, ok := q[key]; ok && len(v) > 0 {
		if f, err := strconv.ParseFloat(v[0], 64); err == nil {
			return f
		}
	}
	return def
}
