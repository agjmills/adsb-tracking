package main

import (
	"strings"
	"unicode"
)

type operatorInfo struct {
	Name    string
	ICAO    string
	Country string
}

var callsignOperators = map[string]operatorInfo{
	"THY":  {"Turkish Airlines", "THY", "Turkey"},
	"PGT":  {"Pegasus Airlines", "PGT", "Turkey"},
	"SXS":  {"SunExpress", "SXS", "Turkey"},
	"RYR":  {"Ryanair", "RYR", "Ireland"},
	"EIN":  {"Aer Lingus", "EIN", "Ireland"},
	"STK":  {"Stobart Air", "STK", "Ireland"},
	"EZY":  {"easyJet", "EZY", "United Kingdom"},
	"EZS":  {"easyJet Switzerland", "EZS", "Switzerland"},
	"BAW":  {"British Airways", "BAW", "United Kingdom"},
	"SHT":  {"British Airways (Shuttle)", "SHT", "United Kingdom"},
	"VIR":  {"Virgin Atlantic", "VIR", "United Kingdom"},
	"EXS":  {"Jet2", "EXS", "United Kingdom"},
	"TOM":  {"TUI Airways", "TOM", "United Kingdom"},
	"LOG":  {"Loganair", "LOG", "United Kingdom"},
	"BEE":  {"flybe", "BEE", "United Kingdom"},
	"EZE":  {"Eastern Airways", "EZE", "United Kingdom"},
	"AUR":  {"Aurigny", "AUR", "United Kingdom"},
	"BMS":  {"Blue Islands", "BMS", "United Kingdom"},
	"CAW":  {"BA CityFlyer", "CAW", "United Kingdom"},
	"EDC":  {"Air Charter Scotland", "EDC", "United Kingdom"},
	"WUK":  {"Wizz Air UK", "WUK", "United Kingdom"},
	"IMC":  {"Eastern Airways", "IMC", "United Kingdom"},
	"WLK":  {"Sky Alps", "WLK", "Italy"},
	"AFR":  {"Air France", "AFR", "France"},
	"DLH":  {"Lufthansa", "DLH", "Germany"},
	"EWG":  {"Eurowings", "EWG", "Germany"},
	"CFG":  {"Condor", "CFG", "Germany"},
	"TUI":  {"TUIfly", "TUI", "Germany"},
	"BOX":  {"AeroLogic", "BOX", "Germany"},
	"GEC":  {"Lufthansa Cargo", "GEC", "Germany"},
	"KLM":  {"KLM", "KLM", "Netherlands"},
	"TRA":  {"Transavia", "TRA", "Netherlands"},
	"MPH":  {"Martinair", "MPH", "Netherlands"},
	"SWR":  {"Swiss International Air Lines", "SWR", "Switzerland"},
	"EDW":  {"Edelweiss Air", "EDW", "Switzerland"},
	"BEL":  {"Brussels Airlines", "BEL", "Belgium"},
	"JAF":  {"TUI fly Belgium", "JAF", "Belgium"},
	"TAY":  {"ASL Airlines Belgium", "TAY", "Belgium"},
	"AUA":  {"Austrian Airlines", "AUA", "Austria"},
	"OES":  {"easyJet Europe", "OES", "Austria"},
	"SAS":  {"Scandinavian Airlines", "SAS", "Sweden"},
	"NAX":  {"Norwegian Air Shuttle", "NAX", "Norway"},
	"NOZ":  {"Norwegian Air Sweden", "NOZ", "Sweden"},
	"NSZ":  {"Norwegian Air Sweden AOC", "NSZ", "Sweden"},
	"IBK":  {"Norwegian Air International", "IBK", "Ireland"},
	"FIN":  {"Finnair", "FIN", "Finland"},
	"ICE":  {"Icelandair", "ICE", "Iceland"},
	"WOW":  {"WOW air", "WOW", "Iceland"},
	"LOT":  {"LOT Polish Airlines", "LOT", "Poland"},
	"WZZ":  {"Wizz Air", "WZZ", "Hungary"},
	"WMT":  {"Wizz Air Malta", "WMT", "Malta"},
	"CSA":  {"Czech Airlines", "CSA", "Czech Republic"},
	"TAP":  {"TAP Air Portugal", "TAP", "Portugal"},
	"IBE":  {"Iberia", "IBE", "Spain"},
	"ANE":  {"Air Nostrum", "ANE", "Spain"},
	"VY":   {"Vueling", "VLG", "Spain"},
	"VLG":  {"Vueling", "VLG", "Spain"},
	"AEA":  {"Air Europa", "AEA", "Spain"},
	"SWT":  {"Swiftair", "SWT", "Spain"},
	"PVG":  {"Privilege Style", "PVG", "Spain"},
	"AZA":  {"ITA Airways", "ITY", "Italy"},
	"ITY":  {"ITA Airways", "ITY", "Italy"},
	"AEZ":  {"Aeroitalia", "AEZ", "Italy"},
	"ISS":  {"Meridiana", "ISS", "Italy"},
	"AEE":  {"Aegean Airlines", "AEE", "Greece"},
	"SEH":  {"Sky Express", "SEH", "Greece"},
	"ROT":  {"TAROM", "ROT", "Romania"},
	"BGH":  {"BH Air", "BGH", "Bulgaria"},
	"LZB":  {"Bulgaria Air", "LZB", "Bulgaria"},
	"CTN":  {"Croatia Airlines", "CTN", "Croatia"},
	"ADR":  {"Adria Airways", "ADR", "Slovenia"},
	"MAH":  {"Air Malta", "AMC", "Malta"},
	"AMC":  {"Air Malta", "AMC", "Malta"},
	"CLX":  {"Cargolux", "CLX", "Luxembourg"},
	"LGL":  {"Luxair", "LGL", "Luxembourg"},
	"MSR":  {"EgyptAir", "MSR", "Egypt"},
	"RJA":  {"Royal Jordanian", "RJA", "Jordan"},
	"MEA":  {"Middle East Airlines", "MEA", "Lebanon"},
	"ELY":  {"El Al", "ELY", "Israel"},
	"ISR":  {"Israir", "ISR", "Israel"},
	"ETD":  {"Etihad Airways", "ETD", "United Arab Emirates"},
	"UAE":  {"Emirates", "UAE", "United Arab Emirates"},
	"QTR":  {"Qatar Airways", "QTR", "Qatar"},
	"GFA":  {"Gulf Air", "GFA", "Bahrain"},
	"KAC":  {"Kuwait Airways", "KAC", "Kuwait"},
	"OMA":  {"Oman Air", "OMA", "Oman"},
	"IAW":  {"Iraqi Airways", "IAW", "Iraq"},
	"SVA":  {"Saudia", "SVA", "Saudi Arabia"},
	"ETH":  {"Ethiopian Airlines", "ETH", "Ethiopia"},
	"KQA":  {"Kenya Airways", "KQA", "Kenya"},
	"RAM":  {"Royal Air Maroc", "RAM", "Morocco"},
	"TAR":  {"Tunisair", "TAR", "Tunisia"},
	"DAH":  {"Air Algérie", "DAH", "Algeria"},
	"SAA":  {"South African Airways", "SAA", "South Africa"},
	"UAL":  {"United Airlines", "UAL", "United States"},
	"DAL":  {"Delta Air Lines", "DAL", "United States"},
	"AAL":  {"American Airlines", "AAL", "United States"},
	"ASA":  {"Alaska Airlines", "ASA", "United States"},
	"JBU":  {"JetBlue", "JBU", "United States"},
	"SWA":  {"Southwest Airlines", "SWA", "United States"},
	"FDX":  {"FedEx Express", "FDX", "United States"},
	"UPS":  {"UPS Airlines", "UPS", "United States"},
	"GTI":  {"Atlas Air", "GTI", "United States"},
	"SOO":  {"Southern Air", "SOO", "United States"},
	"ABX":  {"ABX Air", "ABX", "United States"},
	"CKS":  {"Kalitta Air", "CKS", "United States"},
	"ANA":  {"All Nippon Airways", "ANA", "Japan"},
	"JAL":  {"Japan Airlines", "JAL", "Japan"},
	"CPA":  {"Cathay Pacific", "CPA", "Hong Kong"},
	"CRK":  {"Cathay Dragon", "CRK", "Hong Kong"},
	"HKE":  {"Hong Kong Express", "HKE", "Hong Kong"},
	"CAL":  {"China Airlines", "CAL", "Taiwan"},
	"EVA":  {"EVA Air", "EVA", "Taiwan"},
	"CES":  {"China Eastern", "CES", "China"},
	"CCA":  {"Air China", "CCA", "China"},
	"CSN":  {"China Southern", "CSN", "China"},
	"CBJ":  {"Capital Airlines", "CBJ", "China"},
	"SIA":  {"Singapore Airlines", "SIA", "Singapore"},
	"SCO":  {"Scoot", "SCO", "Singapore"},
	"MAS":  {"Malaysia Airlines", "MAS", "Malaysia"},
	"AAX":  {"AirAsia X", "AXM", "Malaysia"},
	"AXM":  {"AirAsia", "AXM", "Malaysia"},
	"GIA":  {"Garuda Indonesia", "GIA", "Indonesia"},
	"PAL":  {"Philippine Airlines", "PAL", "Philippines"},
	"THA":  {"Thai Airways", "THA", "Thailand"},
	"QFA":  {"Qantas", "QFA", "Australia"},
	"JST":  {"Jetstar", "JST", "Australia"},
	"VOZ":  {"Virgin Australia", "VOZ", "Australia"},
	"ANZ":  {"Air New Zealand", "ANZ", "New Zealand"},
	"ACA":  {"Air Canada", "ACA", "Canada"},
	"ROU":  {"Air Canada Rouge", "ROU", "Canada"},
	"WJA":  {"WestJet", "WJA", "Canada"},
	"SWG":  {"Sunwing Airlines", "SWG", "Canada"},
	"TSC":  {"Air Transat", "TSC", "Canada"},
	"AMX":  {"Aeromexico", "AMX", "Mexico"},
	"SLI":  {"Aeromexico Connect", "SLI", "Mexico"},
	"AVA":  {"Avianca", "AVA", "Colombia"},
	"AIC":  {"Air India", "AIC", "India"},
	"IGO":  {"IndiGo", "IGO", "India"},
	"VTI":  {"Vistara", "VTI", "India"},
	"SEJ":  {"SpiceJet", "SEJ", "India"},
	"PIA":  {"Pakistan International Airlines", "PIA", "Pakistan"},
	"BBC":  {"Biman Bangladesh Airlines", "BBC", "Bangladesh"},
	"ALK":  {"SriLankan Airlines", "ALK", "Sri Lanka"},
	"AFL":  {"Aeroflot", "AFL", "Russia"},
	"SBI":  {"S7 Airlines", "SBI", "Russia"},
	"UTA":  {"Utair", "UTA", "Russia"},
	"PBD":  {"Pobeda", "PBD", "Russia"},
	"AZV":  {"Azur Air", "AZV", "Russia"},
	"NWS":  {"Nordwind Airlines", "NWS", "Russia"},
	"CKK":  {"China Cargo Airlines", "CKK", "China"},
	"AHK":  {"Air Hong Kong", "AHK", "Hong Kong"},
	"ABW":  {"AirBridgeCargo", "ABW", "Russia"},
	"PAC":  {"Polar Air Cargo", "PAC", "United States"},
	"BAH":  {"British Airways (Heavy)", "BAW", "United Kingdom"},
	"RCH":  {"US Air Force (AMC)", "RCH", "United States"},
	"CNV":  {"US Navy", "CNV", "United States"},
	"CFC":  {"RCAF", "CFC", "Canada"},
	"RRR":  {"Royal Air Force", "RRR", "United Kingdom"},
	"GAF":  {"German Air Force", "GAF", "Germany"},
	"FNY":  {"French Navy", "FNY", "France"},
	"CTM":  {"French Air Force", "CTM", "France"},
	"IAM":  {"Italian Air Force", "IAM", "Italy"},
	"AME":  {"Spanish Air Force", "AME", "Spain"},
	"BAF":  {"Belgian Air Force", "BAF", "Belgium"},
	"NAF":  {"Royal Netherlands Air Force", "NAF", "Netherlands"},
	"PLF":  {"Polish Air Force", "PLF", "Poland"},
	"HAF":  {"Hellenic Air Force", "HAF", "Greece"},
	"HUAF": {"Hungarian Air Force", "HUAF", "Hungary"},
	"SAF":  {"Swiss Air Force", "SAF", "Switzerland"},
	"SUI":  {"Swiss Air Force", "SUI", "Switzerland"},
	"PNY":  {"Portuguese Air Force", "PNY", "Portugal"},
	"ROKAF": {"Republic of Korea Air Force", "ROKAF", "South Korea"},
	"DAF":  {"Royal Danish Air Force", "DAF", "Denmark"},
	"SVK":  {"Slovak Air Force", "SVK", "Slovakia"},
	"NATO": {"NATO", "NATO", "Belgium"},
	"MMF":  {"NATO MMF", "MMF", "Netherlands"},
	"SAM":  {"USAF VIP", "SAM", "United States"},
	"SPAR": {"USAF Special Air Mission", "SPAR", "United States"},
	"AAC":  {"British Army Air Corps", "AAC", "United Kingdom"},
	"SHF":  {"UK Support Helicopter Force", "SHF", "United Kingdom"},
	"WAD":  {"UK Defence Helicopter School", "WAD", "United Kingdom"},
	"LOP":  {"US Navy Logistics", "LOP", "United States"},
	"NVY":  {"Royal Navy", "NVY", "United Kingdom"},
	"RAFAIR": {"Royal Air Force", "RAF", "United Kingdom"},
	"ACF":  {"RAuxAF", "ACF", "United Kingdom"},
	"AZG":  {"Silk Way West Airlines", "AZG", "Azerbaijan"},
	"AHY":  {"Azerbaijan Airlines", "AHY", "Azerbaijan"},
	"BLA":  {"Batik Air Malaysia", "BAT", "Malaysia"},
	"BTK":  {"Batik Air", "BTK", "Indonesia"},
	"CSZ":  {"Shenzhen Airlines", "CSZ", "China"},
	"CXA":  {"Xiamen Airlines", "CXA", "China"},
	"CQH":  {"Spring Airlines", "CQH", "China"},
	"CHH":  {"Hainan Airlines", "CHH", "China"},
	"DKH":  {"Juneyao Airlines", "DKH", "China"},
	"EPA":  {"Donghai Airlines", "EPA", "China"},
	"FJI":  {"Fiji Airways", "FJI", "Fiji"},
	"HVN":  {"Vietnam Airlines", "HVN", "Vietnam"},
	"BAV":  {"Bamboo Airways", "BAV", "Vietnam"},
	"VJC":  {"VietJet Air", "VJC", "Vietnam"},
	"KAL":  {"Korean Air", "KAL", "South Korea"},
	"AAR":  {"Asiana Airlines", "AAR", "South Korea"},
	"JJA":  {"Jeju Air", "JJA", "South Korea"},
	"ABL":  {"Air Busan", "ABL", "South Korea"},
	"TWB":  {"T'way Air", "TWB", "South Korea"},
	"JNA":  {"Jin Air", "JNA", "South Korea"},
	"ESR":  {"Eastar Jet", "ESR", "South Korea"},
	"GRL":  {"Air Greenland", "GRL", "Denmark"},
	"FLI":  {"Atlantic Airways", "FLI", "Faroe Islands"},
	"NVR":  {"Novair", "NVR", "Sweden"},
	"BLX":  {"TUIfly Nordic", "BLX", "Sweden"},
	"JTG":  {"Jet Time", "JTG", "Denmark"},
	"VKG":  {"Sunclass Airlines", "VKG", "Denmark"},
}

func inferOperatorFromCallsign(callsign string) (name, icao string) {
	if callsign == "" {
		return "", ""
	}
	upper := strings.ToUpper(strings.TrimSpace(callsign))
	for i := 3; i <= 5; i++ {
		if i > len(upper) {
			break
		}
		prefix := upper[:i]
		if info, ok := callsignOperators[prefix]; ok {
			return info.Name, info.ICAO
		}
	}
	if len(upper) >= 3 {
		prefix := upper[:3]
		if info, ok := callsignOperators[prefix]; ok {
			return info.Name, info.ICAO
		}
	}
	return "", ""
}

func normalizeName(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	norm := map[string]string{
		"BOEING":                  "The Boeing Company",
		"BOEING COMPANY":          "The Boeing Company",
		"THE BOEING COMPANY":      "The Boeing Company",
		"AIRBUS":                  "Airbus",
		"AIRBUS SAS":              "Airbus",
		"AIRBUS INDUSTRIE":        "Airbus",
		"EMBRAER":                 "Embraer",
		"EMBRAER S.A.":            "Embraer",
		"BOMBARDIER":              "Bombardier",
		"BOMBARDIER INC":           "Bombardier",
		"EASYJET":                 "easyJet",
		"EASYJET AIRLINE":         "easyJet",
		"RYANAIR":                 "Ryanair",
		"RYANAIR LTD":             "Ryanair",
		"RYANAIR DAC":             "Ryanair",
		"BRITISH AIRWAYS PLC":     "British Airways",
		"BA CITYFLYER LTD":        "BA CityFlyer",
		"LUFTHANSA GERMAN AIRLINES": "Lufthansa",
		"DEUTSCHE LUFTHANSA AG":   "Lufthansa",
		"AIR FRANCE K.L.M.":       "Air France",
		"K.L.M. ROYAL DUTCH AIRLINES": "KLM",
		"KLM ROYAL DUTCH AIRLINES": "KLM",
		"SWISS":                   "Swiss International Air Lines",
		"SWISS INTERNATIONAL AIR LINES LTD": "Swiss International Air Lines",
		"TURK HAVA YOLLARI A.O.":  "Turkish Airlines",
		"PEGASUS HAVA TASIMACILIGI A.S.": "Pegasus Airlines",
		"JET2.COM":                "Jet2",
		"JET2.COM LTD":            "Jet2",
		"TUI AIRWAYS LTD":         "TUI Airways",
		"WIZZ AIR HUNGARY LTD":    "Wizz Air",
		"WIZZ AIR MALTA LTD":      "Wizz Air Malta",
		"VIRGIN ATLANTIC AIRWAYS LTD": "Virgin Atlantic",
		"LOGANAIR LTD":            "Loganair",
		"EASTERN AIRWAYS LTD":     "Eastern Airways",
		"ASL AIRLINES":            "ASL Airlines Belgium",
		"ASL AIRLINES BELGIUM":    "ASL Airlines Belgium",
	}
	upper := strings.ToUpper(s)
	if mapped, ok := norm[upper]; ok {
		return mapped
	}
	return toTitleCase(s)
}

func inferOperatorFromIcao(opIcao string) (name, icao string) {
	if opIcao == "" {
		return "", ""
	}
	upper := strings.ToUpper(strings.TrimSpace(opIcao))
	for _, info := range callsignOperators {
		if info.ICAO == upper {
			return info.Name, info.ICAO
		}
	}
	return "", ""
}

func toTitleCase(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}
