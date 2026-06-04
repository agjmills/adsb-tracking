package main

import "strings"

type countryInfo struct {
	Name string
	Flag string
}

type hexRange struct {
	Start, End uint32
	Military   bool
	Name       string
}

var militaryHexRanges = []hexRange{
	// USA military
	{0xAE0000, 0xAEFFFF, true, "US Air Force"},
	{0xAD0000, 0xAFFFFF, true, "US Navy/Marines"},
	{0xAF0000, 0xAFFFFF, true, "US Military"},

	// UK military
	{0x43C000, 0x43CFFF, true, "RAF"},
	{0x43E000, 0x43EFFF, true, "RAF"},

	// German military
	{0x3E8000, 0x3EFFFF, true, "Luftwaffe"},
	{0x3F4000, 0x3FFFFF, true, "Luftwaffe"},

	// French military
	{0x3A0000, 0x3A3FFF, true, "Armée de l'Air"},
	{0x3B7000, 0x3B7FFF, true, "Armée de l'Air"},

	// Italian military
	{0x33F000, 0x33FFFF, true, "Aeronautica Militare"},

	// Belgian military
	{0x44F000, 0x44FFFF, true, "Belgian Air Component"},

	// Dutch military
	{0x480800, 0x480FFF, true, "Royal Netherlands Air Force"},

	// Spanish military
	{0x350000, 0x351FFF, true, "Ejército del Aire"},

	// Polish military
	{0x48D800, 0x48DFFF, true, "Polish Air Force"},

	// Canadian military
	{0xC2B000, 0xC2BFFF, true, "RCAF"},
	{0xC2E000, 0xC2EFFF, true, "RCAF"},

	// Australian military
	{0x7CF000, 0x7CFFFF, true, "RAAF"},

	// NATO / AWACS / MMF
	{0x4D0300, 0x4D03FF, true, "NATO AWACS"},
	{0x480C40, 0x480C8F, true, "NATO MMF"},

	// Various AWACS/tanker
	{0x447F00, 0x447FFF, true, "NATO"},

	// Russian military
	{0x140000, 0x14FFFF, true, "Russian Air Force"},
	{0x1B0000, 0x1BFFFF, true, "Russian Military"},
}

func isMilitaryHex(icao24 string) bool {
	if len(icao24) != 6 {
		return false
	}
	val := hexToUint32(icao24)
	for _, r := range militaryHexRanges {
		if val >= r.Start && val <= r.End {
			return true
		}
	}
	return false
}

func lookupCountry(icao24 string) countryInfo {
	if len(icao24) != 6 {
		return countryInfo{}
	}
	val := hexToUint32(icao24)
	for _, r := range militaryHexRanges {
		if val >= r.Start && val <= r.End {
			if r.Military {
				return countryInfo{Name: r.Name + " (Military)", Flag: ""}
			}
		}
	}
	return countryInfo{}
}

func hexToUint32(icao24 string) uint32 {
	val := uint32(0)
	for _, c := range icao24 {
		val <<= 4
		switch {
		case c >= '0' && c <= '9':
			val |= uint32(c - '0')
		case c >= 'A' && c <= 'F':
			val |= uint32(c - 'A' + 10)
		case c >= 'a' && c <= 'f':
			val |= uint32(c - 'a' + 10)
		default:
			return 0
		}
	}
	return val
}

var militaryCallsignPrefixes = []string{
	"RCH", "CNV", "CFC", "RRR", "NAF", "NVY",
	"BAF", "PLF", "HUAF", "SVK", "ROKAF",
	"AAC", "SHF", "WAD", "LOP", "IAM",
	"AFP", "AME", "FNY", "CTM", "GAF",
	"HAF", "IAF", "JAF", "KAF", "NYB",
	"PNY", "RSF", "SAF", "TAF", "TUAF",
	"VIPER", "HAWK", "EAGLE", "FALCON", "RAVEN",
	"HORNET", "LIGHTNING", "TYPHOON", "TORNADO",
	"APACHE", "CHINOOK", "PUMA", "MERLIN",
	"SENTRY", "SENTINEL", "SHADOW", "RIVET",
	"ASCOT", "COBRA", "DAWG", "DEMON",
	"SPAR", "SAM", "VENOM", "VULTURE",
	"DRAGN", "WOLF", "TIGER", "LION",
	"RAFAIR", "NAVY", "ARMY", "GUARD",
	"COMET", "MACE", "SCALP", "DAGGR",
}

func isMilitaryCallsign(callsign string) bool {
	if callsign == "" {
		return false
	}
	upper := strings.ToUpper(strings.TrimSpace(callsign))
	for _, p := range militaryCallsignPrefixes {
		if strings.HasPrefix(upper, p) {
			return true
		}
	}
	return false
}

var emitterCategories = map[string]string{
	"A0": "No info", "A1": "Light (<7031kg)", "A2": "Small (7031-34019kg)",
	"A3": "Large (34019-136078kg)", "A4": "High vortex (B757)", "A5": "Heavy (>136078kg)",
	"A6": "High perf (>5g,>400kt)", "A7": "Rotorcraft",
	"B0": "No info", "B1": "Glider/sailplane", "B2": "Lighter than air",
	"B3": "Parachutist", "B4": "Ultralight/glider", "B5": "Reserved",
	"B6": "UAV/drone", "B7": "Space/trans-atmo",
	"C0": "Surface vehicle", "C1": "Surface emerg", "C2": "Surface service",
	"C3": "Point obstacle", "C4": "Cluster obstacle", "C5": "Line obstacle",
	"C6": "Reserved", "C7": "Reserved",
}

func categoryDesc(code string) string {
	if desc, ok := emitterCategories[code]; ok {
		return code + " (" + desc + ")"
	}
	return code
}
