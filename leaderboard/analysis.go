package main

import (
	"fmt"
	"math"
	"sort"
)

const (
	minSightingsForAnalysis = 5
	orbitMinCumulativeTurn  = 240.0
	orbitMaxBoundingNM      = 15.0
	orbitMinSpeedKt         = 80.0
	orbitMaxSpeedKt         = 450.0
	climbMinRateFPM         = 500.0
	descentMinRateFPM       = -500.0
	maxRealisticRateFPM     = 10000.0
	maxSingleStepAltDelta   = 5000.0
	transitMinSpeedKt       = 300.0
	transitMaxTrackVar      = 25.0
	loiterMaxSpeedKt        = 100.0
)

type patternResult struct {
	Type        string  `json:"type"`
	Confidence  float64 `json:"confidence"`
	Description string  `json:"description"`
}

type flightSummary struct {
	StartSeen int64   `json:"start_seen"`
	EndSeen   int64   `json:"end_seen"`
	Sightings int     `json:"sightings"`
	Callsign  string  `json:"callsign"`
	MinAlt    int     `json:"min_alt"`
	MaxAlt    int     `json:"max_alt"`
	MaxSpeed  float64 `json:"max_speed"`
}

type aircraftAnalysis struct {
	ICAO24            string          `json:"icao24"`
	SightingsUsed     int             `json:"sightings_used"`
	TimeSpanMin       float64         `json:"time_span_min"`
	Patterns          []patternResult `json:"patterns"`
	RecentFlights     []flightSummary `json:"recent_flights,omitempty"`
	HistoricalFlights int             `json:"historical_flights,omitempty"`
	PreviousCallsigns []callsignRow   `json:"previous_callsigns,omitempty"`
}

func rollupToSighting(r rollupRow) sightingRow {
	track := r.MinTrack
	if r.MaxTrack != r.MinTrack {
		track = (r.MinTrack + r.MaxTrack) / 2
	}
	dist := r.MinDistNM
	if r.MaxDistNM > dist {
		dist = r.MaxDistNM
	}
	return sightingRow{
		ICAO24:  r.ICAO24,
		Callsign: r.Callsign,
		Lat:     r.MeanLat,
		Lon:     r.MeanLon,
		AltFt:   int(r.MeanAltFt),
		SpeedKt: r.MeanSpeedKt,
		Track:   track,
		DistNM:  dist,
		SeenAt:  r.WindowStart,
	}
}

func analyzeSightings(icao24 string, rawSightings []sightingRow, rollups []rollupRow, callsigns []callsignRow) *aircraftAnalysis {
	totalSightings := len(rawSightings)
	if totalSightings < minSightingsForAnalysis && len(rollups) < minSightingsForAnalysis {
		return nil
	}
	aa := &aircraftAnalysis{ICAO24: icao24, SightingsUsed: totalSightings, PreviousCallsigns: callsigns}

	sort.Slice(rawSightings, func(i, j int) bool {
		return rawSightings[i].SeenAt < rawSightings[j].SeenAt
	})

	var allPoints []sightingRow
	seen := map[int64]bool{}
	for _, s := range rawSightings {
		allPoints = append(allPoints, s)
		seen[s.SeenAt] = true
	}
	for _, r := range rollups {
		if !seen[r.WindowStart] {
			allPoints = append(allPoints, rollupToSighting(r))
		}
	}
	sort.Slice(allPoints, func(i, j int) bool {
		return allPoints[i].SeenAt < allPoints[j].SeenAt
	})

	if len(allPoints) > 1 {
		aa.TimeSpanMin = float64(allPoints[len(allPoints)-1].SeenAt-allPoints[0].SeenAt) / 60.0
	}

	aa.RecentFlights = detectFlights(rawSightings)

	historicalFlights := detectFlights(allPoints)
	aa.HistoricalFlights = len(historicalFlights)

	if len(rawSightings) >= minSightingsForAnalysis {
		if orb := detectOrbit(rawSightings); orb != nil {
			aa.Patterns = append(aa.Patterns, *orb)
		}
		if cl := detectClimb(rawSightings); cl != nil {
			aa.Patterns = append(aa.Patterns, *cl)
		}
		if ds := detectDescent(rawSightings); ds != nil {
			aa.Patterns = append(aa.Patterns, *ds)
		}
		if tr := detectTransit(rawSightings); tr != nil {
			aa.Patterns = append(aa.Patterns, *tr)
		}
		if lo := detectLoiter(rawSightings); lo != nil {
			aa.Patterns = append(aa.Patterns, *lo)
		}
	}

	return aa
}

func detectFlights(sightings []sightingRow) []flightSummary {
	if len(sightings) == 0 {
		return nil
	}
	var flights []flightSummary
	cur := flightSummary{
		StartSeen: sightings[0].SeenAt,
		EndSeen:   sightings[0].SeenAt,
		Sightings: 1,
		Callsign:  sightings[0].Callsign,
		MinAlt:    sightings[0].AltFt,
		MaxAlt:    sightings[0].AltFt,
		MaxSpeed:  sightings[0].SpeedKt,
	}
	csCounts := map[string]int{}
	if sightings[0].Callsign != "" {
		csCounts[sightings[0].Callsign] = 1
	}
	const flightGapSec = 1800

	for i := 1; i < len(sightings); i++ {
		s := sightings[i]
		if s.SeenAt-cur.EndSeen > flightGapSec {
			cur.Callsign = dominantCallsign(csCounts)
			flights = append(flights, cur)
			csCounts = map[string]int{}
			cur = flightSummary{
				StartSeen: s.SeenAt,
				EndSeen:   s.SeenAt,
				Sightings: 1,
				Callsign:  s.Callsign,
				MinAlt:    s.AltFt,
				MaxAlt:    s.AltFt,
				MaxSpeed:  s.SpeedKt,
			}
		} else {
			cur.EndSeen = s.SeenAt
			cur.Sightings++
			if s.AltFt < cur.MinAlt {
				cur.MinAlt = s.AltFt
			}
			if s.AltFt > cur.MaxAlt {
				cur.MaxAlt = s.AltFt
			}
			if s.SpeedKt > cur.MaxSpeed {
				cur.MaxSpeed = s.SpeedKt
			}
		}
		if s.Callsign != "" {
			csCounts[s.Callsign]++
		}
	}
	cur.Callsign = dominantCallsign(csCounts)
	flights = append(flights, cur)
	return flights
}

func dominantCallsign(counts map[string]int) string {
	best := ""
	bestN := 0
	for cs, n := range counts {
		if n > bestN {
			bestN = n
			best = cs
		}
	}
	return best
}

func detectOrbit(sightings []sightingRow) *patternResult {
	var orbitSeqs []struct {
		start, end   int
		totalTurn    float64
		altStd       float64
		boundingNM   float64
	}

	for window := 10; window <= 40; window += 10 {
		if window > len(sightings) {
			break
		}
		for start := 0; start+window <= len(sightings); start++ {
			end := start + window
			seq := sightings[start:end]

			var totalTurn float64
			prevTrack := seq[0].Track
			for i := 1; i < len(seq); i++ {
				delta := seq[i].Track - prevTrack
				for delta > 180 {
					delta -= 360
				}
				for delta < -180 {
					delta += 360
				}
				totalTurn += delta
				prevTrack = seq[i].Track
			}

			absTurn := math.Abs(totalTurn)
			if absTurn < orbitMinCumulativeTurn {
				continue
			}

			var sumAlt float64
			for _, s := range seq {
				sumAlt += float64(s.AltFt)
			}
			meanAlt := sumAlt / float64(len(seq))
			var sumSq float64
			for _, s := range seq {
				d := float64(s.AltFt) - meanAlt
				sumSq += d * d
			}
			altStd := math.Sqrt(sumSq / float64(len(seq)))

			if altStd > 2000 {
				continue
			}

			avgSpeed := 0.0
			for _, s := range seq {
				avgSpeed += s.SpeedKt
			}
			avgSpeed /= float64(len(seq))
			if avgSpeed < orbitMinSpeedKt || avgSpeed > orbitMaxSpeedKt {
				continue
			}

			var minLat, maxLat, minLon, maxLon float64
			minLat, maxLat = seq[0].Lat, seq[0].Lat
			minLon, maxLon = seq[0].Lon, seq[0].Lon
			for _, s := range seq {
				if s.Lat < minLat {
					minLat = s.Lat
				}
				if s.Lat > maxLat {
					maxLat = s.Lat
				}
				if s.Lon < minLon {
					minLon = s.Lon
				}
				if s.Lon > maxLon {
					maxLon = s.Lon
				}
			}
			diagNM := haversineNM(minLat, minLon, maxLat, maxLon)
			if diagNM > orbitMaxBoundingNM {
				continue
			}

			orbitSeqs = append(orbitSeqs, struct {
				start, end   int
				totalTurn    float64
				altStd       float64
				boundingNM   float64
			}{start, end, absTurn, altStd, diagNM})
		}
	}

	if len(orbitSeqs) == 0 {
		return nil
	}

	best := orbitSeqs[0]
	for _, o := range orbitSeqs {
		if o.totalTurn > best.totalTurn {
			best = o
		}
	}

	conf := math.Min(1.0, best.totalTurn/360.0)
	desc := "Likely holding/orbiting in a circular pattern"
	if best.totalTurn > 540 {
		desc = "Sustained orbit — possible tanker track, CAP, or holding pattern"
	} else if best.totalTurn > 400 {
		desc = "Completed full orbit — holding or circling"
	}
	desc += " (" + formatDuration(sightings[best.start].SeenAt, sightings[best.end-1].SeenAt) + ")"

	return &patternResult{Type: "orbit", Confidence: conf, Description: desc}
}

func detectClimb(sightings []sightingRow) *patternResult {
	return detectVerticalChange(sightings, true)
}

func detectDescent(sightings []sightingRow) *patternResult {
	return detectVerticalChange(sightings, false)
}

func detectVerticalChange(sightings []sightingRow, climbing bool) *patternResult {
	var bestRate float64
	var bestStart, bestEnd int
	var bestDuration float64

	for window := 10; window <= 30; window += 5 {
		if window > len(sightings) {
			break
		}
		for start := 0; start+window <= len(sightings); start++ {
			end := start + window
			startAlt := sightings[start].AltFt
			endAlt := sightings[end-1].AltFt

			if startAlt == 0 || endAlt == 0 {
				continue
			}

			altChange := float64(endAlt - startAlt)
			deltaSec := float64(sightings[end-1].SeenAt - sightings[start].SeenAt)
			if deltaSec < 60 {
				continue
			}
			rateFPM := altChange / deltaSec * 60

			absRate := math.Abs(rateFPM)
			if absRate > maxRealisticRateFPM || absRate < climbMinRateFPM {
				continue
			}

			if climbing && rateFPM < 0 {
				continue
			}
			if !climbing && rateFPM > 0 {
				continue
			}

			monotonic := 0
			badStep := false
			for i := start + 1; i < end; i++ {
				if sightings[i].AltFt == 0 || sightings[i-1].AltFt == 0 {
					badStep = true
					break
				}
				stepDelta := math.Abs(float64(sightings[i].AltFt - sightings[i-1].AltFt))
				stepSec := float64(sightings[i].SeenAt - sightings[i-1].SeenAt)
				if stepSec > 0 && stepDelta/stepSec*60 > maxRealisticRateFPM*1.5 {
					badStep = true
					break
				}
				if climbing && sightings[i].AltFt >= sightings[i-1].AltFt {
					monotonic++
				} else if !climbing && sightings[i].AltFt <= sightings[i-1].AltFt {
					monotonic++
				}
			}
			if badStep {
				continue
			}

			steps := end - start - 1
			if float64(monotonic)/float64(steps) < 0.75 {
				continue
			}

			totalChange := math.Abs(float64(endAlt - startAlt))
			if totalChange > maxSingleStepAltDelta*float64(steps)/3 {
				continue
			}

			if absRate > bestRate {
				bestRate = absRate
				bestStart = start
				bestEnd = end
				bestDuration = deltaSec
			}
		}
	}

	if bestRate == 0 {
		return nil
	}

	action := "descending"
	if climbing {
		action = "climbing"
	}

	totalChange := sightings[bestEnd-1].AltFt - sightings[bestStart].AltFt
	conf := math.Min(1.0, bestRate/5000.0)
	desc := "Sustained " + action + " at " + formatFPM(bestRate) +
		" (" + formatAlt(totalChange) + " over " + formatSec(bestDuration) + ")"

	return &patternResult{Type: action, Confidence: conf, Description: desc}
}

func detectTransit(sightings []sightingRow) *patternResult {
	var bestSpeed float64
	var bestTrackVar float64
	var bestStart, bestEnd int

	for window := 8; window <= 30; window += 8 {
		if window > len(sightings) {
			break
		}
		for start := 0; start+window <= len(sightings); start++ {
			end := start + window
			seq := sightings[start:end]

			var sumSpeed float64
			for _, s := range seq {
				sumSpeed += s.SpeedKt
			}
			avgSpeed := sumSpeed / float64(len(seq))
			if avgSpeed < transitMinSpeedKt {
				continue
			}

			var trackVals []float64
			for _, s := range seq {
				trackVals = append(trackVals, s.Track)
			}
			tv := trackVariance(trackVals)
			if tv > transitMaxTrackVar {
				continue
			}

			if avgSpeed > bestSpeed {
				bestSpeed = avgSpeed
				bestTrackVar = tv
				bestStart = start
				bestEnd = end
			}
		}
	}

	if bestSpeed == 0 {
		return nil
	}

	conf := math.Min(1.0, bestSpeed/500.0) * (1.0 - bestTrackVar/transitMaxTrackVar)

	distNM := 0.0
	for i := bestStart; i < bestEnd-1; i++ {
		distNM += haversineNM(sightings[i].Lat, sightings[i].Lon, sightings[i+1].Lat, sightings[i+1].Lon)
	}

	desc := "High-speed transit at " + fmt.Sprintf("%.0f", bestSpeed) + " kt, covering ~" +
		fmt.Sprintf("%.1f", distNM) + " nm on a steady heading"

	return &patternResult{Type: "transit", Confidence: conf, Description: desc}
}

func detectLoiter(sightings []sightingRow) *patternResult {
	var bestConf float64
	var bestStart, bestEnd int

	for window := 6; window <= 30; window += 6 {
		if window > len(sightings) {
			break
		}
		for start := 0; start+window <= len(sightings); start++ {
			end := start + window
			seq := sightings[start:end]

			var sumSpeed float64
			fast := 0
			for _, s := range seq {
				sumSpeed += s.SpeedKt
				if s.SpeedKt > loiterMaxSpeedKt {
					fast++
				}
			}
			avgSpeed := sumSpeed / float64(len(seq))
			if fast > len(seq)/3 {
				continue
			}

			var trackVals []float64
			for _, s := range seq {
				trackVals = append(trackVals, s.Track)
			}
			tv := trackVariance(trackVals)

			conf := (1.0 - avgSpeed/loiterMaxSpeedKt) * math.Min(1.0, tv/90.0)
			if conf > bestConf {
				bestConf = conf
				bestStart = start
				bestEnd = end
			}
		}
	}

	if bestConf < 0.4 {
		return nil
	}

	avgAlt := 0
	for i := bestStart; i < bestEnd; i++ {
		avgAlt += sightings[i].AltFt
	}
	avgAlt /= (bestEnd - bestStart)

	desc := "Slow movement, likely loitering or hovering"
	if avgAlt < 2000 {
		desc += " at low altitude"
		bestConf = math.Min(1.0, bestConf+0.15)
	}

	return &patternResult{Type: "loiter", Confidence: math.Min(1.0, bestConf), Description: desc}
}

func trackVariance(vals []float64) float64 {
	if len(vals) < 2 {
		return 0
	}
	mean := 0.0
	for _, v := range vals {
		mean += v
	}
	mean /= float64(len(vals))

	var sumSq float64
	for _, v := range vals {
		diff := v - mean
		for diff > 180 {
			diff -= 360
		}
		for diff < -180 {
			diff += 360
		}
		sumSq += diff * diff
	}
	return math.Sqrt(sumSq / float64(len(vals)))
}

func formatDuration(start, end int64) string {
	s := end - start
	if s < 60 {
		return fmt.Sprintf("%ds", s)
	}
	if s < 3600 {
		return fmt.Sprintf("%dm%ds", s/60, s%60)
	}
	return fmt.Sprintf("%dh%dm", s/3600, (s%3600)/60)
}

func formatFPM(rate float64) string {
	return fmt.Sprintf("%.0f ft/min", rate)
}

func formatAlt(ft int) string {
	if ft < 0 {
		return fmt.Sprintf("%d ft", ft)
	}
	return fmt.Sprintf("+%d ft", ft)
}

func formatSec(s float64) string {
	sec := int64(s)
	return formatDuration(0, sec)
}
