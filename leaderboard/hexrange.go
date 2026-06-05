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
	{0xAE0000, 0xAEFFFF, true, "US Air Force"},
	{0xAD0000, 0xAFFFFF, true, "US Navy/Marines"},
	{0xAF0000, 0xAFFFFF, true, "US Military"},
	{0x43C000, 0x43CFFF, true, "RAF"},
	{0x43E000, 0x43EFFF, true, "RAF"},
	{0x3E8000, 0x3EFFFF, true, "Luftwaffe"},
	{0x3F4000, 0x3FFFFF, true, "Luftwaffe"},
	{0x3A0000, 0x3A3FFF, true, "Armée de l'Air"},
	{0x3B7000, 0x3B7FFF, true, "Armée de l'Air"},
	{0x33F000, 0x33FFFF, true, "Aeronautica Militare"},
	{0x44F000, 0x44FFFF, true, "Belgian Air Component"},
	{0x480800, 0x480FFF, true, "Royal Netherlands Air Force"},
	{0x350000, 0x351FFF, true, "Ejército del Aire"},
	{0x48D800, 0x48DFFF, true, "Polish Air Force"},
	{0xC2B000, 0xC2BFFF, true, "RCAF"},
	{0xC2E000, 0xC2EFFF, true, "RCAF"},
	{0x7CF000, 0x7CFFFF, true, "RAAF"},
	{0x4D0300, 0x4D03FF, true, "NATO AWACS"},
	{0x480C40, 0x480C8F, true, "NATO MMF"},
	{0x447F00, 0x447FFF, true, "NATO"},
	{0x140000, 0x14FFFF, true, "Russian Air Force"},
	{0x1B0000, 0x1BFFFF, true, "Russian Military"},
}

type civilCountryRange struct {
	Start, End uint32
	Name       string
	Flag       string
}

var civilCountryRanges = []civilCountryRange{
	{0x400000, 0x40FFFF, "United Kingdom", "\U0001F1EC\U0001F1E7"},
	{0x407000, 0x407FFF, "United Kingdom", "\U0001F1EC\U0001F1E7"},
	{0x4CA000, 0x4CAFFF, "Ireland", "\U0001F1EE\U0001F1EA"},
	{0x380000, 0x38FFFF, "France", "\U0001F1EB\U0001F1F7"},
	{0x390000, 0x39FFFF, "France", "\U0001F1EB\U0001F1F7"},
	{0x3A0000, 0x3AFFFF, "France", "\U0001F1EB\U0001F1F7"},
	{0x3B0000, 0x3BFFFF, "France", "\U0001F1EB\U0001F1F7"},
	{0x3C0000, 0x3CFFFF, "Germany", "\U0001F1E9\U0001F1EA"},
	{0x3D0000, 0x3DFFFF, "Germany", "\U0001F1E9\U0001F1EA"},
	{0x3E0000, 0x3E7FFF, "Germany", "\U0001F1E9\U0001F1EA"},
	{0x3F0000, 0x3F3FFF, "Germany", "\U0001F1E9\U0001F1EA"},
	{0x300000, 0x30FFFF, "Italy", "\U0001F1EE\U0001F1F9"},
	{0x310000, 0x31FFFF, "Italy", "\U0001F1EE\U0001F1F9"},
	{0x320000, 0x32FFFF, "Italy", "\U0001F1EE\U0001F1F9"},
	{0x330000, 0x33FFFF, "Italy", "\U0001F1EE\U0001F1F9"},
	{0x340000, 0x34FFFF, "Spain", "\U0001F1EA\U0001F1F8"},
	{0x350000, 0x351FFF, "Spain", "\U0001F1EA\U0001F1F8"},
	{0x480000, 0x4807FF, "Netherlands", "\U0001F1F3\U0001F1F1"},
	{0x484000, 0x484FFF, "Netherlands", "\U0001F1F3\U0001F1F1"},
	{0x44E000, 0x44EFFF, "Belgium", "\U0001F1E7\U0001F1EA"},
	{0x44F000, 0x44FFFF, "Belgium", "\U0001F1E7\U0001F1EA"},
	{0x4B0000, 0x4B1FFF, "Switzerland", "\U0001F1E8\U0001F1ED"},
	{0x440000, 0x440FFF, "Austria", "\U0001F1E6\U0001F1F9"},
	{0x4A0000, 0x4A2FFF, "Sweden", "\U0001F1F8\U0001F1EA"},
	{0x450000, 0x45FFFF, "Denmark", "\U0001F1E9\U0001F1F0"},
	{0x470000, 0x47FFFF, "Norway", "\U0001F1F3\U0001F1F4"},
	{0x460000, 0x46FFFF, "Finland", "\U0001F1EB\U0001F1EE"},
	{0x490000, 0x491FFF, "Portugal", "\U0001F1F5\U0001F1F9"},
	{0x498000, 0x498FFF, "Czech Republic", "\U0001F1E8\U0001F1FF"},
	{0x487000, 0x489FFF, "Poland", "\U0001F1F5\U0001F1F1"},
	{0x468000, 0x469FFF, "Greece", "\U0001F1EC\U0001F1F7"},
	{0x4B8000, 0x4B9FFF, "Turkey", "\U0001F1F9\U0001F1F7"},
	{0x4A4000, 0x4A7FFF, "Romania", "\U0001F1F7\U0001F1F4"},
	{0x451000, 0x452FFF, "Bulgaria", "\U0001F1E7\U0001F1EC"},
	{0x4D0000, 0x4D03FF, "Luxembourg", "\U0001F1F1\U0001F1FA"},
	{0x4C8000, 0x4C8FFF, "Cyprus", "\U0001F1E8\U0001F1FE"},
	{0x4D2000, 0x4D2FFF, "Malta", "\U0001F1F2\U0001F1F9"},
	{0x500000, 0x50FFFF, "Croatia", "\U0001F1ED\U0001F1F7"},
	{0x502C00, 0x502CFF, "Latvia", "\U0001F1F1\U0001F1FB"},
	{0x503C00, 0x503CFF, "Lithuania", "\U0001F1F1\U0001F1F9"},
	{0x511000, 0x511FFF, "Estonia", "\U0001F1EA\U0001F1EA"},
	{0x506000, 0x506FFF, "Slovenia", "\U0001F1F8\U0001F1EE"},
	{0x505C00, 0x505CFF, "Slovakia", "\U0001F1F8\U0001F1F0"},
	{0x475000, 0x475FFF, "Hungary", "\U0001F1ED\U0001F1FA"},
	{0x010000, 0x017FFF, "Egypt", "\U0001F1EA\U0001F1EC"},
	{0x020000, 0x02FFFF, "Morocco", "\U0001F1F2\U0001F1E6"},
	{0x030000, 0x03FFFF, "Algeria", "\U0001F1E9\U0001F1FF"},
	{0x050000, 0x057FFF, "Tunisia", "\U0001F1F9\U0001F1F3"},
	{0x008000, 0x00FFFF, "South Africa", "\U0001F1FF\U0001F1E6"},
	{0x710000, 0x71BFFF, "Saudi Arabia", "\U0001F1F8\U0001F1E6"},
	{0x740000, 0x740FFF, "Jordan", "\U0001F1EF\U0001F1F4"},
	{0x896000, 0x896FFF, "United Arab Emirates", "\U0001F1E6\U0001F1EA"},
	{0x06A000, 0x06AFFF, "Qatar", "\U0001F1F6\U0001F1E6"},
	{0x760000, 0x761FFF, "Pakistan", "\U0001F1F5\U0001F1F0"},
	{0x800000, 0x800FFF, "India", "\U0001F1EE\U0001F1F3"},
	{0x840000, 0x84FFFF, "Japan", "\U0001F1EF\U0001F1F5"},
	{0x780000, 0x78FFFF, "China", "\U0001F1E8\U0001F1F3"},
	{0x7C0000, 0x7FFFFF, "Australia", "\U0001F1E6\U0001F1FA"},
	{0xC80000, 0xC81FFF, "New Zealand", "\U0001F1F3\U0001F1FF"},
	{0x76B000, 0x76CFFF, "Singapore", "\U0001F1F8\U0001F1EC"},
	{0x750000, 0x75FFFF, "Malaysia", "\U0001F1F2\U0001F1FE"},
	{0x880000, 0x88FFFF, "Thailand", "\U0001F1F9\U0001F1ED"},
	{0x8A0000, 0x8A0FFF, "Indonesia", "\U0001F1EE\U0001F1E9"},
	{0x700000, 0x70FFFF, "Philippines", "\U0001F1F5\U0001F1ED"},
	{0x71C000, 0x71FFFF, "South Korea", "\U0001F1F0\U0001F1F7"},
	{0x899000, 0x899FFF, "Hong Kong", "\U0001F1ED\U0001F1F0"},
	{0x040000, 0x04FFFF, "Ethiopia", "\U0001F1EA\U0001F1F9"},
	{0x04C000, 0x04CFFF, "Kenya", "\U0001F1F0\U0001F1EA"},
	{0x152000, 0x152FFF, "Nigeria", "\U0001F1F3\U0001F1EC"},
	{0x06C000, 0x06CFFF, "Bahrain", "\U0001F1E7\U0001F1ED"},
	{0x70C000, 0x70CFFF, "Kuwait", "\U0001F1F0\U0001F1FC"},
	{0x70E000, 0x70EEFF, "Oman", "\U0001F1F4\U0001F1F2"},
	{0x494000, 0x494FFF, "Iceland", "\U0001F1EE\U0001F1F8"},
	{0x738000, 0x738FFF, "Israel", "\U0001F1EE\U0001F1F1"},
	{0x4D4000, 0x4D4FFF, "Lebanon", "\U0001F1F1\U0001F1E7"},
	{0xA00000, 0xAFFFFF, "United States", "\U0001F1FA\U0001F1F8"},
	{0xC00000, 0xCFFFFF, "Canada", "\U0001F1E8\U0001F1E6"},
	{0xE80000, 0xE8FFFF, "Mexico", "\U0001F1F2\U0001F1FD"},
	{0xAC0000, 0xAC0FFF, "Colombia", "\U0001F1E8\U0001F1F4"},
	{0xE00000, 0xE0FFFF, "Brazil", "\U0001F1E7\U0001F1F7"},
	{0xE20000, 0xE2FFFF, "Argentina", "\U0001F1E6\U0001F1F7"},
	{0x180000, 0x1FFFFF, "Russia", "\U0001F1F7\U0001F1FA"},
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
	for _, r := range civilCountryRanges {
		if val >= r.Start && val <= r.End {
			return countryInfo{Name: r.Name, Flag: r.Flag}
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
