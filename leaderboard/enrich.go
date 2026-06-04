package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

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
	if err != nil || len(list) == 0 {
		return
	}
	for _, icao := range list {
		e.mu.Lock()
		if e.active[icao] {
			e.mu.Unlock()
			continue
		}
		e.active[icao] = true
		e.mu.Unlock()

		go func(id string) {
			defer func() {
				e.mu.Lock()
				delete(e.active, id)
				e.mu.Unlock()
			}()
			info, err := e.lookup(id)
			if err != nil {
				e.db.updateEnrichment(id, "", "", "", "", "", "")
				return
			}
			e.db.updateEnrichment(id, info.Registration, info.Manufacturer, info.Model, info.Operator, info.OperatorICAO, info.CategoryDesc)
		}(icao)
	}
}

func (e *enricher) lookup(icao24 string) (*openskyAircraft, error) {
	url := fmt.Sprintf("https://opensky-network.org/api/metadata/aircraft/icao/%s", icao24)
	resp, err := e.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("opensky returned %d", resp.StatusCode)
	}
	var ac openskyAircraft
	if err := json.NewDecoder(resp.Body).Decode(&ac); err != nil {
		return nil, err
	}
	return &ac, nil
}
