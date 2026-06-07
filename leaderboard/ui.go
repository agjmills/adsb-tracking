package main

const uiHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1,user-scalable=no">
<title>ADS-B Leaderboard</title>
<style>
:root{
  --bg:#060d17;--surface:#0d1a2d;--surface2:#111e35;--border:#1a2d4a;
  --text:#dce4f0;--muted:#5e789c;--accent:#f0a030;--green:#00c896;
  --mil:#e05555;--warn:#f0a030;--blue:#3b8fd4;--white:#f4f6fa;
}
*{box-sizing:border-box;margin:0;padding:0}
body{
  font:13px/1.4 'SF Mono','Fira Code','JetBrains Mono',monospace;
  background:var(--bg);color:var(--text);min-height:100vh;
  display:flex;flex-direction:column;
  background-image:
    radial-gradient(ellipse at 50% 0%,rgba(0,180,140,0.04) 0%,transparent 60%),
    linear-gradient(rgba(13,26,45,0.4) 1px,transparent 1px),
    linear-gradient(90deg,rgba(13,26,45,0.4) 1px,transparent 1px);
  background-size:100% 100%,32px 32px,32px 32px;
}

/* HEADER */
#header{background:var(--surface);border-bottom:1px solid var(--border);flex-shrink:0;z-index:50}
.header-bar{display:flex;align-items:center;padding:10px 16px;gap:10px}
.header-bar .back-btn{
  background:none;border:1px solid var(--border);color:var(--muted);
  padding:4px 10px;border-radius:4px;cursor:pointer;font:11px monospace;
  display:none;
}
.header-bar .back-btn:hover{color:var(--text);border-color:var(--muted)}
.header-bar .radar-icon{
  width:28px;height:28px;border-radius:50%;flex-shrink:0;
  border:1.5px solid rgba(0,200,150,0.5);background:radial-gradient(circle,rgba(0,200,150,0.2) 0%,transparent 60%);
  position:relative;box-shadow:0 0 8px rgba(0,200,150,0.2);
}
.header-bar .radar-icon::after{
  content:'';position:absolute;top:50%;left:50%;width:50%;height:1px;
  background:linear-gradient(90deg,var(--green),transparent);
  transform-origin:0 50%;animation:radar 3s linear infinite;
}
@keyframes radar{from{transform:rotate(0deg)}to{transform:rotate(360deg)}}
.header-title{flex:1;min-width:0}
.header-title h1{font-size:15px;font-weight:700;color:var(--white);white-space:nowrap;overflow:hidden;text-overflow:ellipsis}
.header-title .sub{font-size:9px;color:var(--muted);text-transform:uppercase;letter-spacing:1px}
#headerSubTitle{display:none;font-size:12px;color:var(--muted);letter-spacing:1px;text-transform:uppercase}
.station-tag{
  background:rgba(240,160,48,0.12);color:var(--accent);padding:3px 10px;
  border-radius:3px;font-size:10px;font-weight:700;letter-spacing:2px;
  text-transform:uppercase;border:1px solid rgba(240,160,48,0.2);
}

/* MAIN AREAS */
#mainHome{flex:1;overflow-y:auto;padding:16px;max-width:1000px;margin:0 auto;width:100%}
.page{flex:1;overflow-y:auto;display:none}
.page.active{display:flex;flex-direction:column}

/* HOME - STATS GRID */
.stats-grid{display:grid;grid-template-columns:repeat(4,1fr);gap:10px;margin-bottom:20px}
.stat-tile{
  background:var(--surface);border:1px solid var(--border);border-radius:6px;
  padding:12px 14px;text-align:center;transition:all .15s;
}
.stat-tile:hover{border-color:var(--blue);box-shadow:0 0 12px rgba(59,143,212,0.1)}
.stat-tile .tile-label{font-size:9px;color:var(--muted);text-transform:uppercase;letter-spacing:1.5px;margin-bottom:4px}
.stat-tile .tile-value{font-size:22px;font-weight:700;color:var(--white);line-height:1.2}
.stat-tile .tile-value.g{color:var(--green)}
.stat-tile .tile-value.a{color:var(--accent)}
.stat-tile .tile-value.m{color:var(--mil)}
.stat-tile .tile-sub{font-size:9px;color:var(--muted);margin-top:2px}

/* HOME - NAV GRID */
.nav-grid{display:grid;grid-template-columns:repeat(3,1fr);gap:10px}
.nav-card{
  background:var(--surface);border:1px solid var(--border);border-radius:8px;
  padding:18px 16px;cursor:pointer;transition:all .2s;text-decoration:none;color:var(--text);
  display:flex;flex-direction:column;align-items:center;gap:8px;
}
.nav-card:hover{border-color:var(--accent);transform:translateY(-2px);box-shadow:0 4px 20px rgba(0,0,0,0.3)}
.nav-card .nav-icon{font-size:28px;line-height:1}
.nav-card .nav-label{font-size:12px;font-weight:600;letter-spacing:1px;text-transform:uppercase;color:var(--white)}
.nav-card .nav-desc{font-size:9px;color:var(--muted);text-align:center}
.nav-card.radar-card{border-color:rgba(59,143,212,0.3)}
.nav-card.radar-card:hover{border-color:var(--blue)}
.nav-card.mil-card:hover{border-color:var(--mil)}
.nav-card.near-card{border-color:rgba(0,200,150,0.3)}
.nav-card.near-card:hover{border-color:var(--green)}

/* RADAR PAGE */
#page-radar.active{display:flex}
#page-radar iframe{flex:1;border:none;background:#000}

/* NEAREST + NEAR ME PAGE */
.nearest-page{padding:16px;max-width:800px;margin:0 auto;width:100%}
.nearest-card{
  background:var(--surface);border:1px solid rgba(0,200,150,0.25);border-radius:8px;
  padding:20px 24px;box-shadow:0 0 20px rgba(0,200,150,0.04);
}
.nearest-card h2{font-size:11px;color:var(--muted);text-transform:uppercase;letter-spacing:2px;margin-bottom:12px}
.nearest-card .nc-grid{display:grid;grid-template-columns:1fr 1fr;gap:16px 24px}
.nearest-card .nc-label{font-size:9px;color:var(--muted);text-transform:uppercase;letter-spacing:2px;margin-bottom:2px}
.nearest-card .nc-callsign{font-size:28px;font-weight:700;color:var(--warn);grid-column:span 2;line-height:1.1}
.nearest-card .nc-reg{font-size:14px;color:var(--muted);grid-column:span 2;margin-top:-8px}
.nearest-card .nc-value{font-size:26px;font-weight:700;color:var(--green);line-height:1.1}
.nearest-card .nc-unit{font-size:11px;color:var(--muted)}
.nearest-card .nc-sub{font-size:11px;color:var(--muted);margin-top:2px}
.nearest-card .nc-mil{color:var(--mil);font-weight:700;font-size:12px;letter-spacing:2px;grid-column:span 2}
.nearest-card .nc-empty{text-align:center;color:var(--muted);padding:30px;font-style:italic;grid-column:span 2}
.gps-badge{
  display:inline-block;background:rgba(59,143,212,0.15);color:var(--blue);
  padding:2px 8px;border-radius:3px;font-size:9px;letter-spacing:1px;
  margin-left:8px;
}

/* TABLE PAGES */
.table-page{padding:16px;max-width:1400px;margin:0 auto;width:100%}
.table-wrap{overflow-x:auto;-webkit-overflow-scrolling:touch}
table{width:100%;border-collapse:collapse}
th,td{padding:8px 12px;text-align:left;font-size:11px;border-bottom:1px solid var(--border);white-space:nowrap}
th{color:var(--muted);font-weight:600;text-transform:uppercase;letter-spacing:0.5px;font-size:10px;background:var(--surface2);position:sticky;top:0}
tr:hover{background:rgba(240,160,48,0.03)}
.icao-link{color:var(--blue);cursor:pointer;font-weight:600}
.icao-link:hover{text-decoration:underline}
.callsign-link{color:var(--warn);font-weight:600}
.mil-tag{color:var(--mil);font-weight:700;font-size:10px;margin-left:4px}
.notable-tag{display:inline-block;color:var(--accent);font-weight:700;font-size:10px;margin-left:4px;padding:1px 6px;border:1px solid rgba(240,160,48,0.3);border-radius:3px;letter-spacing:1px}
.notable-card{border-color:rgba(240,160,48,0.35)}
.notable-card:hover{border-color:var(--accent)}
.cat-tag{display:inline-block;padding:1px 5px;border-radius:3px;font-size:9px;background:rgba(59,143,212,0.1);color:var(--blue);font-weight:600}
.bright{color:var(--white);font-weight:600}
.dim{color:var(--muted)}
.empty{text-align:center;padding:40px;color:var(--muted);font-style:italic}
.loading-row td{text-align:center;padding:30px;color:var(--muted)}
.hide-mob{}
.hide-desk{display:none}

/* MODAL */
.modal-overlay{display:none;position:fixed;inset:0;background:rgba(0,0,0,.75);z-index:200;justify-content:center;align-items:center;backdrop-filter:blur(4px)}
.modal-overlay.open{display:flex}
.modal{
  background:var(--surface);border:1px solid var(--border);border-radius:8px;
  max-width:580px;width:92%;max-height:82vh;overflow-y:auto;box-shadow:0 0 40px rgba(0,0,0,0.6);
}
.modal-header{padding:12px 18px;border-bottom:1px solid var(--border);background:var(--surface2);display:flex;justify-content:space-between;align-items:center}
.modal-header h3{font-size:14px;color:var(--white)}
.modal-close{background:none;border:1px solid var(--border);color:var(--muted);font-size:13px;cursor:pointer;padding:4px 10px;border-radius:4px}
.modal-close:hover{color:var(--white)}
.modal-body{padding:16px;font-size:12px}
.modal-body .m-row{display:flex;justify-content:space-between;padding:6px 0;border-bottom:1px solid rgba(26,45,74,0.5)}
.modal-body .m-label{color:var(--muted);text-transform:uppercase;font-size:9px;letter-spacing:1px}
.modal-body .m-value{color:var(--text);text-align:right;font-weight:500}
.modal-body .m-value.big{font-size:18px;font-weight:700;color:var(--warn)}
.modal-body .m-value.gr{color:var(--green);font-weight:700}
.modal-links{margin-top:12px;padding-top:12px;border-top:1px solid var(--border);display:flex;gap:8px;flex-wrap:wrap}
.ext-link{color:var(--blue);text-decoration:none;font-size:10px;padding:3px 8px;border:1px solid rgba(59,143,212,0.25);border-radius:4px}
.ext-link:hover{background:rgba(59,143,212,0.1)}

/* DETAIL PAGE */
.detail-page{padding:16px;max-width:900px;margin:0 auto;width:100%}
.detail-hero{background:var(--surface);border:1px solid var(--border);border-radius:8px;padding:20px 24px;margin-bottom:16px}
.detail-hero .dh-reg{font-size:28px;font-weight:700;color:var(--warn);line-height:1.1}
.detail-hero .dh-type{font-size:14px;color:var(--muted);margin-top:4px}
.detail-hero .dh-op{font-size:11px;color:var(--muted);margin-top:2px}
.detail-hero .dh-mil{display:inline-block;color:var(--mil);font-weight:700;font-size:11px;letter-spacing:2px;margin-top:6px;padding:2px 8px;border:1px solid rgba(224,85,85,0.3);border-radius:3px}
.detail-hero .dh-row{display:flex;justify-content:space-between;align-items:flex-start;flex-wrap:wrap;gap:8px}
.detail-stats{display:grid;grid-template-columns:repeat(4,1fr);gap:10px;margin-bottom:16px}
.detail-stat{background:var(--surface);border:1px solid var(--border);border-radius:6px;padding:12px 14px;text-align:center}
.detail-stat .ds-label{font-size:9px;color:var(--muted);text-transform:uppercase;letter-spacing:1px;margin-bottom:4px}
.detail-stat .ds-value{font-size:18px;font-weight:700;color:var(--white);line-height:1.2}
.detail-stat .ds-value.g{color:var(--green)}
.detail-stat .ds-value.a{color:var(--accent)}
.detail-stat .ds-sub{font-size:9px;color:var(--muted);margin-top:2px}
.detail-section{background:var(--surface);border:1px solid var(--border);border-radius:8px;padding:18px 20px;margin-bottom:14px}
.detail-section h3{font-size:10px;color:var(--muted);text-transform:uppercase;letter-spacing:2px;margin-bottom:12px}
.pattern-badge{display:inline-flex;align-items:center;gap:6px;padding:6px 12px;border-radius:4px;margin:3px;font-size:11px}
.pattern-badge.orbit{background:rgba(59,143,212,0.12);border:1px solid rgba(59,143,212,0.25);color:var(--blue)}
.pattern-badge.climbing{background:rgba(0,200,150,0.12);border:1px solid rgba(0,200,150,0.25);color:var(--green)}
.pattern-badge.descending{background:rgba(240,160,48,0.12);border:1px solid rgba(240,160,48,0.25);color:var(--accent)}
.pattern-badge.transit{background:rgba(224,85,85,0.12);border:1px solid rgba(224,85,85,0.25);color:var(--mil)}
.pattern-badge.loiter{background:rgba(160,130,200,0.12);border:1px solid rgba(160,130,200,0.25);color:#a082c8}
.pattern-conf{font-size:9px;opacity:0.7;margin-left:2px}
.pattern-desc{font-size:11px;color:var(--muted);margin:8px 0;line-height:1.5}
.flight-row{display:flex;justify-content:space-between;padding:6px 0;border-bottom:1px solid rgba(26,45,74,0.4);font-size:11px;align-items:center}
.flight-row .fr-idx{color:var(--muted);width:24px;flex-shrink:0}
.flight-row .fr-dur{color:var(--text);font-weight:500}
.flight-row .fr-stats{color:var(--muted)}
.photo-grid{display:grid;grid-template-columns:repeat(3,1fr);gap:8px}
.photo-link{display:flex;align-items:center;justify-content:center;padding:14px 10px;border:1px solid var(--border);border-radius:6px;text-decoration:none;color:var(--text);font-size:11px;font-weight:500;transition:all 0.15s;text-align:center;letter-spacing:0.5px}
.photo-link:hover{border-color:var(--blue);background:rgba(59,143,212,0.06);color:var(--white)}
.photo-link .pl-icon{font-size:18px;margin-right:6px}
.analysis-loading{text-align:center;padding:20px;color:var(--muted);font-style:italic}
.analysis-btn{background:rgba(59,143,212,0.12);border:1px solid rgba(59,143,212,0.25);color:var(--blue);padding:8px 16px;border-radius:4px;cursor:pointer;font-size:11px;letter-spacing:1px;text-transform:uppercase;transition:all 0.15s}
.analysis-btn:hover{background:rgba(59,143,212,0.2);color:var(--white)}
.analysis-empty{text-align:center;padding:16px;color:var(--muted);font-style:italic;font-size:11px}
@media(max-width:768px){
 .detail-stats{grid-template-columns:repeat(2,1fr);gap:8px}
 .detail-stat{padding:10px 8px}
 .detail-stat .ds-value{font-size:16px}
 .detail-hero .dh-reg{font-size:22px}
 .photo-grid{grid-template-columns:repeat(2,1fr)}
}
@media(max-width:480px){
 .detail-stats{grid-template-columns:repeat(2,1fr);gap:6px}
 .photo-grid{grid-template-columns:1fr}
}
footer{padding:12px;text-align:center;color:var(--muted);font-size:9px;letter-spacing:1px;border-top:1px solid var(--border);flex-shrink:0}
footer a{color:var(--muted)}

/* RESPONSIVE */
@media(max-width:768px){
  .stats-grid{grid-template-columns:repeat(3,1fr);gap:8px}
  .stat-tile{padding:10px 8px}
  .stat-tile .tile-value{font-size:18px}
  .nav-grid{grid-template-columns:repeat(2,1fr);gap:8px}
  .nav-card{padding:14px 10px;gap:4px}
  .nav-card .nav-icon{font-size:22px}
  .nav-card .nav-label{font-size:10px}
  .nav-card .nav-desc{display:none}
  th,td{padding:6px 8px;font-size:10px}
  .hide-mob{display:none}
  .hide-desk{display:inline}
  .nearest-card .nc-callsign{font-size:20px}
  .nearest-card .nc-value{font-size:20px}
  .nearest-card .nc-grid{grid-template-columns:1fr 1fr;gap:10px 12px}
  .table-page{padding:8px}
  .nearest-page{padding:8px}
  #mainHome{padding:12px}
}
@media(max-width:480px){
  .stats-grid{grid-template-columns:repeat(2,1fr);gap:6px}
  .stat-tile{padding:8px 6px}
  .stat-tile .tile-value{font-size:16px}
  .stat-tile .tile-label{font-size:8px;letter-spacing:0.5px}
  .stat-tile .tile-sub{font-size:8px}
  .nav-grid{grid-template-columns:repeat(2,1fr);gap:6px}
  .nav-card{padding:12px 8px}
  .header-bar{padding:8px 10px}
  .header-title h1{font-size:13px}
  .station-tag{font-size:8px;padding:2px 6px}
  .nearest-card{padding:14px}
  .nearest-card .nc-callsign{font-size:18px}
  .nearest-card .nc-value{font-size:18px}
}
</style>
</head>
<body>
<header id="header">
 <div class="header-bar">
  <button class="back-btn" id="backBtn" onclick="goHome()">&larr; HOME</button>
  <div class="radar-icon"></div>
  <div class="header-title">
   <h1 id="headerTitle">ADS-B LEADERBOARD</h1>
   <div class="sub" id="headerSub">Loading...</div>
  </div>
  <div id="headerSubTitle"></div>
  <div class="station-tag" id="stationTag">LOCAL</div>
 </div>
</header>

<main id="mainHome">
 <div class="stats-grid" id="statsGrid"></div>
 <div class="nav-grid">
  <div class="nav-card radar-card" onclick="navigate('radar')"><div class="nav-icon">▣</div><div class="nav-label">Radar</div><div class="nav-desc">Live map &amp; coverage</div></div>
  <div class="nav-card near-card" onclick="navigate('nearest')"><div class="nav-icon">✈</div><div class="nav-label">Nearest</div><div class="nav-desc">Closest low-altitude</div></div>
  <div class="nav-card near-card" onclick="navigate('nearme')"><div class="nav-icon">📍</div><div class="nav-label">Near Me</div><div class="nav-desc">What's overhead?</div></div>
  <div class="nav-card" onclick="navigate('leaderboard')"><div class="nav-icon">📋</div><div class="nav-label">Leaderboard</div><div class="nav-desc">Most seen planes</div></div>
  <div class="nav-card mil-card" onclick="navigate('military')"><div class="nav-icon">⚔</div><div class="nav-label">Military</div><div class="nav-desc">Recent military activity</div></div>
  <div class="nav-card notable-card" onclick="navigate('notable')"><div class="nav-icon">💎</div><div class="nav-label">Notable</div><div class="nav-desc">Rare &amp; unusual types</div></div>
  <div class="nav-card" onclick="navigate('distance')"><div class="nav-icon">📏</div><div class="nav-label">Distance</div><div class="nav-desc">Furthest tracked</div></div>
  <div class="nav-card" onclick="navigate('altitude')"><div class="nav-icon">⛰</div><div class="nav-label">Altitude</div><div class="nav-desc">Highest tracked</div></div>
  <div class="nav-card" onclick="navigate('speed')"><div class="nav-icon">⚡</div><div class="nav-label">Speed</div><div class="nav-desc">Fastest tracked</div></div>
  <div class="nav-card" onclick="navigate('recent')"><div class="nav-icon">🕐</div><div class="nav-label">Recent</div><div class="nav-desc">Live sightings</div></div>
   <div class="nav-card" onclick="navigate('daily')"><div class="nav-icon">📅</div><div class="nav-label">Daily</div><div class="nav-desc">Per-day stats</div></div>
   <div class="nav-card" id="faCard" style="display:none"><div class="nav-icon">🛰</div><div class="nav-label">FlightAware</div><div class="nav-desc">My feeder stats</div></div>
 </div>
</main>

<!-- SUB PAGES -->
<div class="page" id="page-radar"><iframe id="radarFrame" src="" allow="geolocation"></iframe></div>

<div class="page" id="page-nearest"><div class="nearest-page"><div class="nearest-card" id="nearestCard"><div class="empty">Loading...</div></div></div></div>

<div class="page" id="page-nearme"><div class="nearest-page"><div class="nearest-card" id="nearmeCard"><div class="empty">Requesting location...</div></div></div></div>

<div class="page" id="page-leaderboard"><div class="table-page"><div class="table-wrap"><table>
 <thead><tr><th>ICAO24</th><th class="hide-mob">Reg</th><th class="hide-mob">Type</th><th>Callsign</th><th class="hide-mob">Country</th><th class="bright">Seen</th><th>Dist</th><th class="hide-mob">Alt</th><th class="hide-mob">Spd</th><th>Last</th></tr></thead>
 <tbody id="lbBody"><tr class="loading-row"><td colspan="10">Loading...</td></tr></tbody>
</table></div></div></div>

<div class="page" id="page-military"><div class="table-page"><div class="table-wrap"><table>
 <thead><tr><th>ICAO24</th><th>Callsign</th><th>Reg</th><th class="hide-mob">Type</th><th>Alt</th><th class="hide-mob">Spd</th><th class="hide-mob">Trk</th><th>Dist</th><th class="hide-mob">Brg</th><th>Seen</th></tr></thead>
 <tbody id="milBody"><tr class="loading-row"><td colspan="10">Loading...</td></tr></tbody>
</table></div></div></div>

<div class="page" id="page-notable"><div class="table-page"><div class="table-wrap"><table>
 <thead><tr><th>Tag</th><th>ICAO24</th><th>Callsign</th><th>Reg</th><th class="hide-mob">Type</th><th>Alt</th><th class="hide-mob">Spd</th><th>Dist</th><th>Op</th><th>Seen</th></tr></thead>
 <tbody id="ntbBody"><tr class="loading-row"><td colspan="10">Loading...</td></tr></tbody>
</table></div></div></div>

<div class="page" id="page-distance"><div class="table-page"><div class="table-wrap"><table>
 <thead><tr><th>ICAO24</th><th class="hide-mob">Reg</th><th>Callsign</th><th class="bright">Max Dist</th><th class="hide-mob">Min Dist</th><th>Seen</th><th>Last</th></tr></thead>
 <tbody id="distBody"><tr class="loading-row"><td colspan="7">Loading...</td></tr></tbody>
</table></div></div></div>

<div class="page" id="page-altitude"><div class="table-page"><div class="table-wrap"><table>
 <thead><tr><th>ICAO24</th><th class="hide-mob">Reg</th><th>Callsign</th><th class="bright">Max Alt</th><th class="hide-mob">Min Alt</th><th>Seen</th><th>Last</th></tr></thead>
 <tbody id="altBody"><tr class="loading-row"><td colspan="7">Loading...</td></tr></tbody>
</table></div></div></div>

<div class="page" id="page-speed"><div class="table-page"><div class="table-wrap"><table>
 <thead><tr><th>ICAO24</th><th class="hide-mob">Reg</th><th>Callsign</th><th class="bright">Max Speed</th><th class="hide-mob">Max Alt</th><th>Seen</th><th>Last</th></tr></thead>
 <tbody id="spdBody"><tr class="loading-row"><td colspan="7">Loading...</td></tr></tbody>
</table></div></div></div>

<div class="page" id="page-recent"><div class="table-page"><div class="table-wrap"><table>
 <thead><tr><th>ICAO24</th><th>Callsign</th><th>Cat</th><th>Alt</th><th class="hide-mob">Spd</th><th class="hide-mob">Trk</th><th>Dist</th><th class="hide-mob">Brg</th><th class="hide-mob">Pos</th><th>Seen</th></tr></thead>
 <tbody id="recBody"><tr class="loading-row"><td colspan="10">Loading...</td></tr></tbody>
</table></div></div></div>

<div class="page" id="page-daily"><div class="table-page"><div class="table-wrap"><table>
 <thead><tr><th>Date</th><th class="bright">Unique</th><th>Sightings</th><th>Military</th><th class="hide-mob">Max Dist</th><th class="hide-mob">Max Alt</th><th class="hide-mob">Max Speed</th></tr></thead>
 <tbody id="dayBody"><tr class="loading-row"><td colspan="7">Loading...</td></tr></tbody>
</table></div></div></div>

<div class="page" id="page-detail"><div class="detail-page" id="detailContent"><div class="empty">Select an aircraft</div></div></div>

<footer><span id="footerTime"></span> &middot; <a href="https://github.com/agjmills/adsb-tracking" target="_blank">GitHub</a></footer>

<!-- MODAL -->
<div class="modal-overlay" id="modalOverlay">
 <div class="modal">
  <div class="modal-header"><h3 id="modalTitle"></h3><button class="modal-close" onclick="closeModal()">CLOSE</button></div>
  <div class="modal-body" id="modalBody"></div>
 </div>
</div>

<script>
var currentPage='home';
var tabData={};

function fmt(v,d){return v!=null?v.toFixed(d||1):'--'}
function ts(t){return new Date(t*1000).toLocaleString()}
function ago(t){var s=Math.floor(Date.now()/1000-t);if(s<60)return s+'s';if(s<3600)return Math.floor(s/60)+'m';if(s<86400)return Math.floor(s/3600)+'h';return Math.floor(s/86400)+'d';}
function cardDir(b){var d=['N','NE','E','SE','S','SW','W','NW'];return d[Math.round(((b||0)+360)%360/45)%8]}
function flag(f){return f?f+' ':''}

var stats=null;
function loadStats(){
 fetch('/api/stats').then(function(r){return r.json()}).then(function(s){
  stats=s;
  var fa=document.getElementById('faCard');
   if(s.flightaware_site_id){
    fa.style.display='flex';
    fa.onclick=function(){window.open('https://flightaware.com/adsb/stats/user/'+encodeURIComponent(s.flightaware_site_id),'_blank')};
  }
  document.getElementById('headerSub').textContent=s.site_name+' \u00b7 '+fmt(s.site_lat,3)+'\u00b0N '+fmt(Math.abs(s.site_lon),3)+'\u00b0'+(s.site_lon<0?'W':'E')+' \u00b7 '+s.site_alt+' AMSL';
  document.getElementById('footerTime').textContent='Updated '+new Date().toLocaleTimeString();
  updateStatsGrid();
 }).catch(function(){});
}

function updateStatsGrid(){
 if(!stats||currentPage!=='home')return;
 var s=stats;
 var tiles=[
  ['AIRCRAFT',s.total_aircraft.toLocaleString(),'g','total unique'],
  ['SIGHTINGS',(s.total_sightings||0).toLocaleString(),'a','total pings'],
  ['MILITARY',(s.military_aircraft||0).toLocaleString(),'m','unique'],
  ['FURTHEST',fmt(s.furthest_nm,1)+'nm','',s.furthest_icao],
  ['CLOSEST',fmt(s.closest_nm,1)+'nm','g',s.closest_icao],
  ['HIGHEST',fmt(s.highest_ft,0)+'ft','',s.highest_icao],
  ['LOWEST',fmt(s.lowest_ft,0)+'ft','g',s.lowest_icao],
  ['FASTEST',fmt(s.fastest_kt,1)+'kt','',s.fastest_icao],
  ['SLOWEST',fmt(s.slowest_kt,1)+'kt','g',s.slowest_icao]
 ];
 var h='';
 for(var i=0;i<tiles.length;i++){
  var t=tiles[i];
  h+='<div class="stat-tile"><div class="tile-label">'+t[0]+'</div><div class="tile-value '+t[2]+'">'+t[1]+'</div><div class="tile-sub">'+t[3]+'</div></div>';
 }
 document.getElementById('statsGrid').innerHTML=h;
}

function navigate(page){
 if(page==='nearme'){loadNearMe();}
 else if(page==='radar'){
  document.getElementById('radarFrame').src='/radar/';
 }
 currentPage=page;
 window.location.hash=page;
 document.getElementById('mainHome').style.display='none';
 document.querySelectorAll('.page').forEach(function(p){p.classList.remove('active')});
 var el=document.getElementById('page-'+page);
 if(el){el.classList.add('active')}
 document.getElementById('backBtn').style.display='inline-block';
 document.getElementById('headerSub').style.display='none';
 var titles={
  radar:'RADAR',nearest:'NEAREST',nearme:'NEAR ME',leaderboard:'LEADERBOARD',
  military:'MILITARY',notable:'NOTABLE',distance:'DISTANCE',altitude:'ALTITUDE',speed:'SPEED',
  recent:'RECENT',daily:'DAILY',detail:'AIRCRAFT DETAIL'
 };
 document.getElementById('headerSubTitle').style.display='block';
 document.getElementById('headerSubTitle').textContent=titles[page]||'';
 loadPage(page);
}

function goHome(){
 currentPage='home';
 window.location.hash='';
 document.getElementById('mainHome').style.display='block';
 document.querySelectorAll('.page').forEach(function(p){p.classList.remove('active')});
 document.getElementById('backBtn').style.display='none';
 document.getElementById('headerSub').style.display='block';
 document.getElementById('headerSubTitle').style.display='none';
 updateStatsGrid();
}

function loadPage(page){
 switch(page){
  case'nearest':loadNearest();break;
  case'nearme':loadNearMe();break;
  case'leaderboard':loadTable('lbBody','/api/leaderboard?sort=sightings&limit=100',leaderboardRows);break;
  case'military':loadTable('milBody','/api/military?limit=200',militaryRows);break;
  case'notable':loadTable('ntbBody','/api/notable?age=1800&limit=100',notableRows);break;
  case'distance':loadTable('distBody','/api/leaderboard?sort=distance&limit=100',distanceRows);break;
  case'altitude':loadTable('altBody','/api/leaderboard?sort=altitude&limit=100',altitudeRows);break;
  case'speed':loadTable('spdBody','/api/leaderboard?sort=speed&limit=100',speedRows);break;
  case'recent':loadTable('recBody','/api/recent?limit=100',recentRows);break;
  case'daily':loadTable('dayBody','/api/daily?limit=30',dailyRows);break;
 }
}

function loadTable(bodyId,url,rowFn){
 var body=document.getElementById(bodyId);
 fetch(url).then(function(r){return r.json()}).then(function(data){
  if(!data||!data.length){body.innerHTML='<tr><td colspan="12" class="empty">No data yet</td></tr>';return}
  body.innerHTML=rowFn(data);
 }).catch(function(){body.innerHTML='<tr><td colspan="12" class="empty">Error loading</td></tr>'});
}

function icaoL(icao){return'<span class="icao-link" onclick="showAircraft(\''+icao+'\')">'+icao+'</span>'}
function milB(m){return m?' <span class="mil-tag">MIL</span>':''}
function csL(c){return c?'<span class="callsign-link">'+c+'</span>':'<span class="dim">-</span>'}
function dC(v){return v>0?fmt(v,1)+' nm':'<span class="dim">-</span>'}

function leaderboardRows(data){var h='';
 for(var i=0;i<data.length;i++){var a=data[i];
  h+='<tr><td>'+icaoL(a.icao24)+milB(a.is_military)+'</td><td class="hide-mob">'+(a.reg||'<span class="dim">-</span>')+'</td><td class="hide-mob">'+([a.manufacturer,a.model].filter(Boolean).join(' ')||'<span class="dim">-</span>')+'</td><td>'+csL(a.last_callsign)+'</td><td class="hide-mob">'+flag(a.country_flag)+(a.country||'<span class="dim">-</span>')+'</td><td class="bright">'+a.total_sightings.toLocaleString()+'</td><td>'+dC(a.min_dist_nm)+' / '+dC(a.max_dist_nm)+'</td><td class="hide-mob">'+(a.min_alt_ft||'-')+'/'+(a.max_alt_ft||'-')+'</td><td class="hide-mob">'+fmt(a.max_speed_kt,0)+' kt</td><td>'+ago(a.last_seen)+'</td></tr>';
 }return h;}

function militaryRows(data){var h='';
 for(var i=0;i<data.length;i++){var a=data[i];
  h+='<tr><td>'+icaoL(a.icao24)+' <span class="mil-tag">MIL</span></td><td>'+csL(a.callsign)+'</td><td>'+(a.reg||'<span class="dim">-</span>')+'</td><td class="hide-mob">'+([a.manufacturer,a.model].filter(Boolean).join(' ')||'<span class="dim">-</span>')+'</td><td class="bright">'+(a.alt_ft||'-').toLocaleString()+' ft</td><td class="hide-mob">'+fmt(a.speed_kt,0)+' kt</td><td class="hide-mob">'+fmt(a.track,0)+'°</td><td>'+fmt(a.dist_nm,1)+' nm</td><td class="hide-mob">'+(a.bearing||'-')+'°</td><td>'+ago(a.seen_at)+'</td></tr>';
 }return h;}

function notableRows(data){var h='';
 for(var i=0;i<data.length;i++){var a=data[i];
  h+='<tr><td><span class="notable-tag">'+a.notable_label+'</span></td><td>'+icaoL(a.icao24)+(a.is_military?' <span class="mil-tag">MIL</span>':'')+'</td><td>'+csL(a.callsign)+'</td><td>'+(a.reg||'<span class="dim">-</span>')+'</td><td class="hide-mob">'+([a.manufacturer,a.model].filter(Boolean).join(' ')||'<span class="dim">-</span>')+'</td><td class="bright">'+(a.alt_ft||'-').toLocaleString()+' ft</td><td class="hide-mob">'+fmt(a.speed_kt,0)+' kt</td><td>'+fmt(a.dist_nm,1)+' nm</td><td class="hide-mob">'+(a.operator||'<span class="dim">-</span>')+'</td><td>'+ago(a.seen_at)+'</td></tr>';
 }return h;}

function distanceRows(data){var h='';
 for(var i=0;i<data.length;i++){var a=data[i];
  h+='<tr><td>'+icaoL(a.icao24)+milB(a.is_military)+'</td><td class="hide-mob">'+(a.reg||'<span class="dim">-</span>')+'</td><td>'+csL(a.last_callsign)+'</td><td class="bright">'+dC(a.max_dist_nm)+'</td><td class="hide-mob">'+dC(a.min_dist_nm)+'</td><td>'+a.total_sightings.toLocaleString()+'</td><td>'+ago(a.last_seen)+'</td></tr>';
 }return h;}

function altitudeRows(data){var h='';
 for(var i=0;i<data.length;i++){var a=data[i];
  h+='<tr><td>'+icaoL(a.icao24)+milB(a.is_military)+'</td><td class="hide-mob">'+(a.reg||'<span class="dim">-</span>')+'</td><td>'+csL(a.last_callsign)+'</td><td class="bright">'+(a.max_alt_ft||'-').toLocaleString()+' ft</td><td class="hide-mob">'+(a.min_alt_ft||'-').toLocaleString()+' ft</td><td>'+a.total_sightings.toLocaleString()+'</td><td>'+ago(a.last_seen)+'</td></tr>';
 }return h;}

function speedRows(data){var h='';
 for(var i=0;i<data.length;i++){var a=data[i];
  h+='<tr><td>'+icaoL(a.icao24)+milB(a.is_military)+'</td><td class="hide-mob">'+(a.reg||'<span class="dim">-</span>')+'</td><td>'+csL(a.last_callsign)+'</td><td class="bright">'+fmt(a.max_speed_kt,0)+' kt</td><td class="hide-mob">'+(a.max_alt_ft||'-').toLocaleString()+' ft</td><td>'+a.total_sightings.toLocaleString()+'</td><td>'+ago(a.last_seen)+'</td></tr>';
 }return h;}

function recentRows(data){var h='';
 for(var i=0;i<data.length;i++){var s=data[i];
  h+='<tr><td>'+icaoL(s.icao24)+'</td><td>'+csL(s.callsign)+'</td><td><span class="cat-tag">'+(s.category||'-')+'</span></td><td>'+(s.alt_ft||'-').toLocaleString()+'</td><td class="hide-mob">'+fmt(s.speed_kt,0)+'</td><td class="hide-mob">'+fmt(s.track,0)+'\u00b0</td><td>'+fmt(s.dist_nm,1)+' nm</td><td class="hide-mob">'+(s.bearing||'-')+'\u00b0</td><td class="hide-mob">'+fmt(s.lat,3)+' '+fmt(s.lon,3)+'</td><td>'+ago(s.seen_at)+'</td></tr>';
 }return h;}

function dailyRows(data){var h='';
 for(var i=0;i<data.length;i++){var d=data[i];
  h+='<tr><td>'+d.date+'</td><td class="bright">'+d.unique_aircraft.toLocaleString()+'</td><td>'+d.total_sightings.toLocaleString()+'</td><td>'+d.military_aircraft.toLocaleString()+'</td><td class="hide-mob">'+fmt(d.max_distance_nm,1)+' nm</td><td class="hide-mob">'+(d.max_altitude_ft||'-').toLocaleString()+' ft</td><td class="hide-mob">'+fmt(d.max_speed_kt,1)+' kt</td></tr>';
 }return h;}

function loadNearest(){
 var card=document.getElementById('nearestCard');
 fetch('/api/nearest?seen=180').then(function(r){return r.json()}).then(function(n){
  if(!n){card.innerHTML='<h2>NEAREST AIRCRAFT</h2><div class="nc-grid"><div class="nc-empty">No low-altitude aircraft in range</div></div>';return}
  card.innerHTML=buildNearestCard(n,'NEAREST AIRCRAFT');
 }).catch(function(){card.innerHTML='<h2>NEAREST AIRCRAFT</h2><div class="nc-grid"><div class="nc-empty">Error loading</div></div>'});
}

var nearMePos=null;
function loadNearMe(){
 var card=document.getElementById('nearmeCard');
 if(nearMePos){
  fetchNearMe(card);
  return;
 }
 card.innerHTML='<h2>NEAR ME <span class="gps-badge">GPS</span></h2><div class="nc-grid"><div class="nc-empty">Requesting location...</div></div>';
 if('geolocation' in navigator){
  navigator.geolocation.getCurrentPosition(function(pos){
   nearMePos={lat:pos.coords.latitude,lon:pos.coords.longitude};
   fetchNearMe(card);
  },function(){
   card.innerHTML='<h2>NEAR ME <span class="gps-badge">GPS</span></h2><div class="nc-grid"><div class="nc-empty">Location denied. Allow GPS access.</div></div>';
  },{enableHighAccuracy:true,timeout:10000});
 }else{
  card.innerHTML='<h2>NEAR ME <span class="gps-badge">GPS</span></h2><div class="nc-grid"><div class="nc-empty">Geolocation not supported</div></div>';
 }
}

function fetchNearMe(card){
 if(!nearMePos)return;
 fetch('/api/nearme?lat='+nearMePos.lat+'&lon='+nearMePos.lon+'&seen=180').then(function(r){return r.json()}).then(function(n){
  if(!n){card.innerHTML='<h2>NEAR ME <span class="gps-badge">GPS</span></h2><div class="nc-grid"><div class="nc-empty">No low-altitude aircraft nearby</div></div>';return}
  card.innerHTML=buildNearestCard(n,'NEAR ME <span class="gps-badge">GPS</span>');
 }).catch(function(){card.innerHTML='<h2>NEAR ME</h2><div class="nc-grid"><div class="nc-empty">Error loading</div></div>'});
}

function buildNearestCard(n,title){
 var cs=n.callsign||n.icao24||'--';
 var reg=n.reg||'';
 var typ=[n.manufacturer,n.model].filter(Boolean).join(' ')||'';
 var op=n.operator||'';
 return '<h2>'+title+'</h2>'+
  (cs!=='--'?'<div class="nc-callsign">'+cs+'</div>':'')+
  (n.is_notable?'<div class="nc-mil" style="color:var(--accent);border-color:rgba(240,160,48,0.3)">NOTABLE: '+n.notable_label+'</div>':'')+
  (reg||typ||op?'<div class="nc-reg">'+[reg,typ,op].filter(Boolean).join(' \u00b7 ')+'</div>':'')+
  (n.is_military?'<div class="nc-mil">MILITARY</div>':'')+
  '<div class="nc-grid">'+
  '<div><div class="nc-label">DISTANCE</div><div class="nc-value">'+fmt(n.dist_nm,1)+' <span class="nc-unit">NM</span></div></div>'+
  '<div><div class="nc-label">ALTITUDE</div><div class="nc-value">'+(n.altitude_ft||0).toLocaleString()+' <span class="nc-unit">FT</span></div></div>'+
  '<div><div class="nc-label">BEARING</div><div class="nc-value">'+(n.bearing||'0')+'\u00b0</div><div class="nc-sub">'+cardDir(n.bearing)+'</div></div>'+
  '<div><div class="nc-label">SPEED / TRACK</div><div class="nc-value">'+fmt(n.speed_kt,0)+' <span class="nc-unit">KT</span></div><div class="nc-sub">TRK '+fmt(n.track,0)+'\u00b0</div></div>'+
  '</div>'+
  '<div style="margin-top:8px"><span class="nc-label">POSITION</span> <span class="nc-sub">'+fmt(n.lat,4)+', '+fmt(n.lon,4)+'</span></div>'+
  '<div><span class="nc-label">SEEN</span> <span class="nc-sub">'+ago(n.seen_at)+' ago \u00b7 '+ts(n.seen_at)+'</span></div>';
}

var currentDetailIcao=null;

function showAircraft(icao){
 currentDetailIcao=icao;
 window.location.hash='detail/'+icao;
 navigateDetail(icao);
}

function navigateDetail(icao){
 currentPage='detail';
 document.getElementById('mainHome').style.display='none';
 document.querySelectorAll('.page').forEach(function(p){p.classList.remove('active')});
 var el=document.getElementById('page-detail');
 if(el){el.classList.add('active')}
 document.getElementById('backBtn').style.display='inline-block';
 document.getElementById('headerSub').style.display='none';
 document.getElementById('headerSubTitle').style.display='block';
 document.getElementById('headerSubTitle').textContent='AIRCRAFT DETAIL';
 document.getElementById('detailContent').innerHTML='<div class="empty">Loading...</div>';
 loadAircraftDetail(icao);
}

function loadAircraftDetail(icao){
 fetch('/api/aircraft/'+icao).then(function(r){return r.json()}).then(function(a){
  renderDetailInfo(a,icao);
  loadAnalysis(icao);
 }).catch(function(){
  document.getElementById('detailContent').innerHTML='<div class="empty">Failed to load</div>';
 });
}

function renderDetailInfo(a,icao){
 var typ=[a.manufacturer,a.model].filter(Boolean).join(' ')||'-';
 var reg=a.reg||'';
 var cs=a.last_callsign||'';
 var totalFlights=a.total_flights||0;

 var h='';
 h+='<div class="detail-hero">';
 h+='<div class="dh-row"><div>';
 h+='<div class="dh-reg">'+(reg||cs||a.icao24)+'</div>';
 h+='<div class="dh-type">'+typ+'</div>';
 if(a.operator)h+='<div class="dh-op">'+a.operator+'</div>';
 h+='</div><div style="text-align:right">';
 h+='<div style="font-size:12px;color:var(--muted)">'+a.icao24+'</div>';
 h+='<div style="font-size:12px;color:var(--text);margin-top:2px">'+flag(a.country_flag)+(a.country||'-')+'</div>';
 if(a.is_military)h+='<div class="dh-mil">MILITARY</div>';
 h+='</div></div></div>';

 h+='<div class="detail-stats">';
 h+=statTile('FLIGHTS',totalFlights.toLocaleString(),'g','unique visits');
 h+=statTile('SIGHTINGS',(a.total_sightings||0).toLocaleString(),'a','total pings');
 h+=statTile('FIRST SEEN','', '', ts(a.first_seen));
 h+=statTile('LAST SEEN','', '', ago(a.last_seen)+' ago');
 h+=statTile('DISTANCE',dC(a.min_dist_nm)+' / '+dC(a.max_dist_nm),'', 'min / max');
 h+=statTile('ALTITUDE',(a.min_alt_ft||'-').toLocaleString()+' / '+(a.max_alt_ft||'-').toLocaleString()+' ft','','min / max');
 h+=statTile('MAX SPEED',fmt(a.max_speed_kt,0)+' kt','','');
 h+=statTile('EMITTER',a.category||'-','','');
 h+='</div>';

 h+='<div class="detail-section"><h3>LAST POSITION</h3>';
 h+='<div style="font-size:12px;color:var(--text)">'+fmt(a.last_lat,4)+', '+fmt(a.last_lon,4);
 h+=' &middot; Alt: '+(a.last_alt||'-').toLocaleString()+' ft';
 h+=' &middot; Spd: '+fmt(a.last_speed,0)+' kt';
 h+=' &middot; '+ts(a.last_seen_pos)+'</div>';
 h+='</div>';

 h+='<div class="detail-section"><h3>ACTIVITY ANALYSIS</h3>';
 h+='<div id="analysisSection"><div class="analysis-loading">Analyzing flight patterns...</div></div>';
 h+='</div>';

 h+='<div class="detail-section"><h3>PHOTOS</h3>';
 h+='<div class="photo-grid">';
 var q=reg||typ||a.icao24;
 h+='<a class="photo-link" target="_blank" href="https://www.planespotters.net/search?q='+encodeURIComponent(reg||typ)+'"><span class="pl-icon">&#9992;</span> Planespotters</a>';
 h+='<a class="photo-link" target="_blank" href="https://www.jetphotos.com/photo/keyword/'+encodeURIComponent(reg||typ)+'"><span class="pl-icon">&#128247;</span> JetPhotos</a>';
 h+='<a class="photo-link" target="_blank" href="https://www.airliners.net/search?keywords='+encodeURIComponent(reg||a.icao24)+'"><span class="pl-icon">&#9992;</span> Airliners</a>';
 h+='</div></div>';

 h+='<div class="detail-section"><h3>EXTERNAL LINKS</h3>';
 h+='<div class="photo-grid">';
 h+='<a class="photo-link" target="_blank" href="https://globe.adsbexchange.com/?icao='+icao+'"><span class="pl-icon">&#127758;</span> ADSBx Globe</a>';
 if(reg)h+='<a class="photo-link" target="_blank" href="https://www.flightradar24.com/data/aircraft/'+encodeURIComponent(reg)+'"><span class="pl-icon">&#128064;</span> FlightRadar24</a>';
 h+='<a class="photo-link" target="_blank" href="https://opensky-network.org/aircraft-profile?icao24='+icao+'"><span class="pl-icon">&#128225;</span> OpenSky</a>';
 h+='</div></div>';

 document.getElementById('detailContent').innerHTML=h;
}

function loadAnalysis(icao){
 fetch('/api/aircraft/'+icao+'/analysis').then(function(r){return r.json()}).then(function(an){
  if(an.error){
   document.getElementById('analysisSection').innerHTML='<div class="analysis-empty">'+'Not enough data &mdash; need at least 5 recent position reports'+'</div>';
   return;
  }
  var h='';

  if(an.previous_callsigns&&an.previous_callsigns.length>0){
   h+='<div style="margin-bottom:12px"><div style="font-size:10px;color:var(--muted);text-transform:uppercase;letter-spacing:1px;margin-bottom:6px">Previous Callsigns</div>';
   h+='<div style="display:flex;flex-wrap:wrap;gap:6px">';
   for(var i=0;i<an.previous_callsigns.length;i++){
    var cs=an.previous_callsigns[i];
    h+='<span style="background:rgba(240,160,48,0.08);border:1px solid rgba(240,160,48,0.2);color:var(--warn);padding:3px 8px;border-radius:3px;font-size:11px;font-weight:500">'+cs.callsign+'</span>';
   }
   h+='</div></div>';
  }

  if(an.patterns&&an.patterns.length>0){
   h+='<div style="margin-bottom:10px">';
   for(var i=0;i<an.patterns.length;i++){
    var p=an.patterns[i];
    h+='<div class="pattern-badge '+p.type+'">';
    if(p.type==='orbit')h+='&#8635; ';
    else if(p.type==='climbing')h+='&#8593; ';
    else if(p.type==='descending')h+='&#8595; ';
    else if(p.type==='transit')h+='&#10145; ';
    else if(p.type==='loiter')h+='&#8634; ';
    h+=p.type.toUpperCase();
    h+='<span class="pattern-conf">'+Math.round(p.confidence*100)+'%</span>';
    h+='</div>';
   }
   h+='</div>';
   for(var i=0;i<an.patterns.length;i++){
    h+='<div class="pattern-desc">'+an.patterns[i].description+'</div>';
   }
  }else{
   h+='<div class="analysis-empty">No distinctive patterns detected in the recent '+an.sightings_used+' sightings</div>';
  }

  if(an.recent_flights&&an.recent_flights.length>0){
   h+='<div style="margin-top:12px"><div style="font-size:10px;color:var(--muted);text-transform:uppercase;letter-spacing:1px;margin-bottom:6px">Recent Movements ('+an.recent_flights.length+' flights'+(an.historical_flights?' &middot; '+an.historical_flights+' total all-time':'')+')</div>';
   for(var i=0;i<Math.min(an.recent_flights.length,8);i++){
    var f=an.recent_flights[i];
    var dur=f.end_seen-f.start_seen;
    var durStr=dur<60?dur+'s':dur<3600?Math.floor(dur/60)+'m':Math.floor(dur/3600)+'h'+Math.floor((dur%3600)/60)+'m';
    h+='<div class="flight-row">';
    h+='<span class="fr-idx">#'+(an.recent_flights.length-i)+'</span>';
    h+='<span class="fr-dur">'+(f.callsign?f.callsign:durStr)+'</span>';
    h+='<span class="fr-stats">'+f.sightings+' pings &middot; '+durStr+' &middot; '+f.min_alt.toLocaleString()+'&ndash;'+f.max_alt.toLocaleString()+' ft &middot; '+fmt(f.max_speed,0)+' kt</span>';
    h+='</div>';
   }
   if(an.recent_flights.length>8)h+='<div style="font-size:10px;color:var(--muted);text-align:center;padding:6px">+ '+(an.recent_flights.length-8)+' more flights</div>';
   h+='</div>';
  }

  h+='<div style="margin-top:10px;font-size:9px;color:var(--muted)">Analyzed '+an.sightings_used+' sightings over '+fmt(an.time_span_min,1)+' minutes</div>';

  document.getElementById('analysisSection').innerHTML=h;
 }).catch(function(){
  document.getElementById('analysisSection').innerHTML='<div class="analysis-empty">Analysis unavailable</div>';
 });
}

function statTile(label,value,cls,sub){
 return '<div class="detail-stat"><div class="ds-label">'+label+'</div><div class="ds-value '+cls+'">'+value+'</div>'+(sub?'<div class="ds-sub">'+sub+'</div>':'')+'</div>';
}

function closeModal(){document.getElementById('modalOverlay').classList.remove('open')}
document.getElementById('modalOverlay').addEventListener('click',function(e){if(e.target===this)closeModal()});

var hash=window.location.hash.replace('#','');
if(hash){
 if(hash.startsWith('detail/')){
  var icao=hash.split('/')[1];
  if(icao)navigateDetail(icao);
 }else if(document.getElementById('page-'+hash)){navigate(hash)}
}
window.addEventListener('hashchange',function(){
 var h=window.location.hash.replace('#','');
 if(h.startsWith('detail/')){
  var icao=h.split('/')[1];
  if(icao)navigateDetail(icao);
 }else if(h&&document.getElementById('page-'+h)){navigate(h)}
 else if(!h){goHome()}
});

loadStats();
setInterval(loadStats,10000);
</script>
</body>
</html>`

const nearestHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>Nearest Aircraft</title>
<style>
*{box-sizing:border-box;margin:0;padding:0}
body{background:#060d17;color:#dce4f0;font:14px monospace;display:flex;align-items:center;justify-content:center;min-height:100vh;padding:16px;background-image:radial-gradient(ellipse at 50% 0%,rgba(0,180,140,0.06) 0%,transparent 60%)}
.card{background:#0d1a2d;border:1px solid rgba(0,200,150,0.25);border-radius:8px;padding:24px;max-width:400px;width:100%;box-shadow:0 0 24px rgba(0,200,150,0.06)}
h2{font-size:10px;color:#5e789c;text-transform:uppercase;letter-spacing:3px;margin-bottom:12px;text-align:center}
.callsign{font-size:34px;font-weight:700;color:#f0a030;margin-bottom:2px;text-align:center}
.reg{font-size:18px;color:#5e789c;text-align:center;margin-bottom:16px}
.grid{display:grid;grid-template-columns:1fr 1fr;gap:14px 20px}
.label{font-size:9px;color:#5e789c;text-transform:uppercase;letter-spacing:2px;margin-bottom:2px}
.value{font-size:24px;font-weight:700;color:#00c896}
.unit{font-size:11px;color:#5e789c}
.sub{font-size:12px;color:#7a8ba6;margin-top:2px}
.mil{color:#e05555;font-weight:700;font-size:13px;text-align:center;margin-top:8px;letter-spacing:2px}
.empty{text-align:center;color:#3a4e6b;font-style:italic;padding:40px 0;font-size:16px}
.foot{color:#3a4e6b;font-size:10px;text-align:center;margin-top:18px;letter-spacing:1px}
</style>
<meta http-equiv="refresh" content="10">
</head>
<body>
<div class="card" id="card">
 <h2>NEAREST AIRCRAFT</h2>
 <div class="empty">SCANNING...</div>
 <div class="foot">REFRESH 10s</div>
</div>
<script>
function fmt(v,d){return v!=null?v.toFixed(d||1):'--'}
function ago(t){var s=Math.floor(Date.now()/1000-t);if(s<60)return s+'s';if(s<3600)return Math.floor(s/60)+'m';if(s<86400)return Math.floor(s/3600)+'h';return Math.floor(s/86400)+'d';}
function cardDir(b){var d=['N','NE','E','SE','S','SW','W','NW'];return d[Math.round(((b||0)+360)%360/45)%8]}
fetch('/api/nearest?seen=180').then(function(r){return r.json()}).then(function(n){
 var card=document.getElementById('card');
 if(!n){card.innerHTML='<h2>NEAREST AIRCRAFT</h2><div class="empty">NOTHING IN RANGE</div><div class="foot">REFRESH 10s</div>';return}
 var cs=n.callsign||n.icao24||'--';
 var reg=n.reg||'';
 var typ=[n.manufacturer,n.model].filter(Boolean).join(' ')||'';
 card.innerHTML='<h2>NEAREST AIRCRAFT</h2>'+
  '<div class="callsign">'+cs+'</div>'+
  (reg?'<div class="reg">'+reg+'</div>':'')+
  (typ?'<div class="sub" style="text-align:center;margin-bottom:12px">'+typ+'</div>':'')+
  (n.is_military?'<div class="mil">MILITARY</div>':'')+
  '<div class="grid">'+
  '<div><div class="label">DISTANCE</div><div class="value">'+fmt(n.dist_nm,1)+' <span class="unit">NM</span></div></div>'+
  '<div><div class="label">ALTITUDE</div><div class="value">'+(n.altitude_ft||0).toLocaleString()+' <span class="unit">FT</span></div></div>'+
  '<div><div class="label">BEARING</div><div class="value">'+n.bearing+'&deg;</div><div class="sub">'+cardDir(n.bearing)+'</div></div>'+
  '<div><div class="label">SPEED</div><div class="value">'+fmt(n.speed_kt,0)+' <span class="unit">KT</span></div><div class="sub">TRK '+fmt(n.track,0)+'&deg;</div></div>'+
  '</div>'+
  '<div style="margin-top:12px;text-align:center"><div class="label">POSITION</div><div class="sub">'+fmt(n.lat,4)+', '+fmt(n.lon,4)+'</div></div>'+
  '<div class="foot">SEEN '+ago(n.seen_at)+' AGO &middot; REFRESH 10s</div>';
}).catch(function(){
 document.getElementById('card').innerHTML='<h2>NEAREST AIRCRAFT</h2><div class="empty">ERROR</div><div class="foot">REFRESH 10s</div>';
});
</script>
</body>
</html>`
