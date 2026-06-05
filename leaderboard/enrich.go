package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"
)

var errOpenSkyNotFound = errors.New("opensky not found")

type openskyAircraft struct {
	ICAO24       string `json:"icao24"`
	Registration string `json:"registration"`
	Manufacturer string `json:"manufacturerName"`
	Model        string `json:"model"`
	Operator     string `json:"operator"`
	OperatorICAO string `json:"operatorIcao"`
	CategoryDesc string `json:"categoryDescription"`
}

type enricher struct {
	mu     sync.Mutex
	client *http.Client
	db     *DB
	active map[string]bool
}

func newEnricher(db *DB) *enricher {
	return &enricher{
		client: &http.Client{Timeout: 10 * time.Second},
		db:     db,
		active: make(map[string]bool),
	}
}

func (e *enricher) run(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			e.process()
		}
	}
}

func (e *enricher) process() {
	list, err := e.db.getUnenriched(5)
	if err == nil && len(list) > 0 {
		for _, icao := range list {
			e.enrichOne(icao)
		}
	}
	stale, err := e.db.getStaleEnriched(3)
	if err == nil && len(stale) > 0 {
		for _, icao := range stale {
			e.enrichOne(icao)
		}
	}
}

func (e *enricher) enrichOne(icao string) {
	e.mu.Lock()
	if e.active[icao] {
		e.mu.Unlock()
		return
	}
	e.active[icao] = true
	e.mu.Unlock()

	go func(id string) {
		defer func() {
			e.mu.Lock()
			delete(e.active, id)
			e.mu.Unlock()
		}()
		callsign, err := e.db.getCallsign(id)
		if err != nil {
			callsign = ""
		}
		info, err := e.lookup(id)
		if err != nil {
			if errors.Is(err, errOpenSkyNotFound) {
				e.db.updateEnrichment(id, "", "", "", "", "", "")
			} else {
				e.db.markEnrichmentAttempt(id)
			}
			if infOperator, infIcao := inferOperatorFromCallsign(callsign); infOperator != "" {
				e.db.updateEnrichment(id, "", "", "", infOperator, infIcao, "")
			}
			return
		}
		manufacturer := normalizeName(info.Manufacturer)
		operator := normalizeName(info.Operator)
		if operator == "" && info.OperatorICAO == "" {
			if infOperator, infIcao := inferOperatorFromCallsign(callsign); infOperator != "" {
				operator = infOperator
				if info.OperatorICAO == "" {
					info.OperatorICAO = infIcao
				}
			}
		}
		e.db.updateEnrichment(id, info.Registration, manufacturer, info.Model, operator, info.OperatorICAO, info.CategoryDesc)
	}(icao)
}

func (e *enricher) lookup(icao24 string) (*openskyAircraft, error) {
	url := fmt.Sprintf("https://opensky-network.org/api/metadata/aircraft/icao/%s", icao24)
	resp, err := e.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		if resp.StatusCode == 404 {
			return nil, errOpenSkyNotFound
		}
		return nil, fmt.Errorf("opensky returned %d", resp.StatusCode)
	}
	var ac openskyAircraft
	if err := json.NewDecoder(resp.Body).Decode(&ac); err != nil {
		return nil, err
	}
	return &ac, nil
}
