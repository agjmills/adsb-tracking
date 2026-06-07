package main

import "strings"

type notablePattern struct {
	Manufacturer string
	Model        string
	Label        string
}

var notablePatterns = []notablePattern{
	// Warbirds / Vintage Military
	{Manufacturer: "Supermarine", Label: "Spitfire"},
	{Manufacturer: "Hawker", Model: "Hurricane", Label: "Hurricane"},
	{Manufacturer: "North American", Model: "P-51", Label: "P-51 Mustang"},
	{Manufacturer: "North American", Model: "P51", Label: "P-51 Mustang"},
	{Manufacturer: "Curtiss", Model: "P-40", Label: "P-40 Warhawk"},
	{Manufacturer: "Curtiss", Model: "C-46", Label: "C-46 Commando"},

	// Rare/old airliners
	{Manufacturer: "Douglas", Model: "DC-3", Label: "DC-3 Dakota"},
	{Manufacturer: "Douglas", Model: "C-47", Label: "C-47 Skytrain"},
	{Manufacturer: "Douglas", Model: "DC-6", Label: "DC-6"},
	{Manufacturer: "Douglas", Model: "DC-7", Label: "DC-7"},
	{Manufacturer: "Douglas", Model: "DC-8", Label: "DC-8"},
	{Manufacturer: "Douglas", Model: "DC-9", Label: "DC-9"},
	{Manufacturer: "McDonnell Douglas", Model: "DC-10", Label: "DC-10"},
	{Manufacturer: "McDonnell Douglas", Model: "MD-11", Label: "MD-11"},
	{Manufacturer: "Concorde", Label: "Concorde"},
	{Manufacturer: "Lockheed", Model: "L-1011", Label: "L-1011 Tristar"},
	{Manufacturer: "BAE", Model: "146", Label: "BAe 146/Avro RJ"},
	{Manufacturer: "Avro", Model: "RJ", Label: "Avro RJ"},
	{Manufacturer: "Fokker", Label: "Fokker"},

	// Eastern Bloc
	{Manufacturer: "Antonov", Label: "Antonov"},
	{Manufacturer: "Ilyushin", Label: "Ilyushin"},
	{Manufacturer: "Tupolev", Label: "Tupolev"},
	{Manufacturer: "Yakovlev", Label: "Yakovlev"},
	{Manufacturer: "Sukhoi", Label: "Sukhoi"},
	{Manufacturer: "Mikoyan", Label: "MiG"},
	{Manufacturer: "Beriev", Label: "Beriev"},

	// Special Military
	{Manufacturer: "Boeing", Model: "E-3", Label: "E-3 Sentry AWACS"},
	{Manufacturer: "Boeing", Model: "E-4", Label: "E-4 Nightwatch"},
	{Manufacturer: "Boeing", Model: "E-6", Label: "E-6 Mercury"},
	{Manufacturer: "Boeing", Model: "E-7", Label: "E-7 Wedgetail"},
	{Manufacturer: "Boeing", Model: "E-8", Label: "E-8 JSTARS"},
	{Manufacturer: "Boeing", Model: "KC-135", Label: "KC-135 Stratotanker"},
	{Manufacturer: "Boeing", Model: "KC-46", Label: "KC-46 Pegasus"},
	{Manufacturer: "Boeing", Model: "RC-135", Label: "RC-135 Rivet Joint"},
	{Manufacturer: "Boeing", Model: "C-17", Label: "C-17 Globemaster"},
	{Manufacturer: "Boeing", Model: "VC-25", Label: "VC-25 (Air Force One)"},
	{Manufacturer: "Boeing", Model: "C-32", Label: "C-32 (Air Force Two)"},
	{Manufacturer: "Boeing", Model: "C-40", Label: "C-40 Clipper"},
	{Manufacturer: "Boeing", Model: "P-8", Label: "P-8 Poseidon"},
	{Manufacturer: "Lockheed", Model: "C-5", Label: "C-5 Galaxy"},
	{Manufacturer: "Lockheed", Model: "C-130", Label: "C-130 Hercules"},
	{Manufacturer: "Lockheed", Model: "U-2", Label: "U-2 Dragon Lady"},
	{Manufacturer: "Lockheed", Model: "SR-71", Label: "SR-71 Blackbird"},
	{Manufacturer: "Lockheed", Model: "F-117", Label: "F-117 Nighthawk"},
	{Manufacturer: "Northrop", Model: "B-2", Label: "B-2 Spirit"},
	{Manufacturer: "Rockwell", Model: "B-1", Label: "B-1 Lancer"},
	{Manufacturer: "Sikorsky", Model: "CH-53", Label: "CH-53"},
	{Manufacturer: "Northrop Grumman", Model: "RQ-4", Label: "RQ-4 Global Hawk"},
	{Manufacturer: "Northrop Grumman", Model: "MQ-4", Label: "MQ-4 Triton"},
	{Manufacturer: "General Atomics", Label: "Drone/UAV"},

	// Tiltrotor
	{Manufacturer: "Bell", Model: "V-22", Label: "V-22 Osprey"},
	{Manufacturer: "Bell Boeing", Model: "V-22", Label: "V-22 Osprey"},

	// Display Teams / Trainers
	{Manufacturer: "BAE Systems", Model: "Hawk", Label: "Hawk T1/T2"},
	{Manufacturer: "British Aerospace", Model: "Hawk", Label: "Hawk T1/T2"},
	{Manufacturer: "Aermacchi", Model: "MB-339", Label: "MB-339 (Frecce Tricolori)"},
	{Manufacturer: "Alenia", Model: "MB-339", Label: "MB-339 (Frecce Tricolori)"},

	// Rare GA/Biz
	{Manufacturer: "Piaggio", Model: "P-180", Label: "Piaggio Avanti"},
	{Manufacturer: "Honda", Model: "HA-420", Label: "HondaJet"},
	{Manufacturer: "Epic", Model: "E1000", Label: "Epic E1000"},

	// Vintage Jets
	{Manufacturer: "De Havilland", Model: "Vampire", Label: "Vampire"},
	{Manufacturer: "De Havilland", Model: "Venom", Label: "Venom"},
	{Manufacturer: "Gloster", Model: "Meteor", Label: "Meteor"},
	{Manufacturer: "English Electric", Label: "Lightning"},
	{Manufacturer: "Hawker", Model: "Hunter", Label: "Hunter"},
	{Manufacturer: "Hawker Siddeley", Model: "Harrier", Label: "Harrier"},
	{Manufacturer: "McDonnell Douglas", Model: "AV-8", Label: "Harrier"},

	// Modern Fighters (stand out from civil traffic)
	{Manufacturer: "Dassault", Model: "Mirage", Label: "Mirage"},
	{Manufacturer: "Dassault", Model: "Rafale", Label: "Rafale"},
	{Manufacturer: "Eurofighter", Label: "Typhoon"},
	{Manufacturer: "Saab", Model: "Gripen", Label: "Gripen"},
	{Manufacturer: "Saab", Model: "JAS", Label: "Gripen"},
	{Manufacturer: "Saab", Model: "35", Label: "Saab 35 Draken"},
	{Manufacturer: "Saab", Model: "37", Label: "Saab 37 Viggen"},
	{Manufacturer: "Panavia", Model: "Tornado", Label: "Tornado"},
	{Manufacturer: "SEPECAT", Model: "Jaguar", Label: "Jaguar"},

	// Airships / Unusual
	{Manufacturer: "Zeppelin", Label: "Airship"},

	// Rare/old British props
	{Manufacturer: "Short", Model: "Skyvan", Label: "Skyvan"},
	{Manufacturer: "Shorts", Model: "Skyvan", Label: "Skyvan"},
	{Manufacturer: "Short", Model: "330", Label: "Shorts 330"},
	{Manufacturer: "Shorts", Model: "330", Label: "Shorts 330"},
	{Manufacturer: "Short", Model: "360", Label: "Shorts 360"},
	{Manufacturer: "Shorts", Model: "360", Label: "Shorts 360"},
	{Manufacturer: "Vickers", Label: "Vickers"},
	{Manufacturer: "Bristol", Label: "Bristol"},
	{Manufacturer: "Avro", Model: "Lancaster", Label: "Lancaster"},
	{Manufacturer: "De Havilland", Model: "Dove", Label: "Dove/Heron"},
	{Manufacturer: "De Havilland", Model: "Heron", Label: "Dove/Heron"},
	{Manufacturer: "Scottish Aviation", Label: "Twin Pioneer"},

	// Water bombers / firefighting (unusual to see)
	{Manufacturer: "Canadair", Model: "CL-215", Label: "Water Bomber"},
	{Manufacturer: "Canadair", Model: "CL-415", Label: "Water Bomber"},
	{Manufacturer: "Air Tractor", Model: "AT-802", Label: "Fire Boss"},
	{Manufacturer: "Beriev", Model: "Be-200", Label: "Be-200 Water Bomber"},
}

type notableResult struct {
	NotableLabel string  `json:"notable_label"`
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
	IsMilitary   bool    `json:"is_military"`
}

func isNotable(manufacturer, model string) (bool, string) {
	if manufacturer == "" && model == "" {
		return false, ""
	}
	mfrUpper := strings.ToUpper(manufacturer)
	modelUpper := strings.ToUpper(model)
	for _, p := range notablePatterns {
		mfrMatch := p.Manufacturer == "" || strings.Contains(mfrUpper, strings.ToUpper(p.Manufacturer))
		modelMatch := p.Model == "" || strings.Contains(modelUpper, strings.ToUpper(p.Model))
		if (p.Manufacturer != "" || p.Model != "") && mfrMatch && modelMatch {
			return true, p.Label
		}
	}
	return false, ""
}
