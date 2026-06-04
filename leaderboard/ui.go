package main

const uiHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>ADS-B Leaderboard</title>
<style>
:root{
  --bg:#060d17;
  --surface:#0d1a2d;
  --surface2:#111e35;
  --border:#1a2d4a;
  --text:#dce4f0;
  --muted:#5e789c;
  --accent:#f0a030;
  --green:#00c896;
  --green-dim:#007a5a;
  --mil:#e05555;
  --warn:#f0a030;
  --blue:#3b8fd4;
  --white:#f4f6fa;
  --radar:rgba(0,200,150,0.15);
  --glow-accent:0 0 12px rgba(240,160,48,0.4);
  --glow-green:0 0 10px rgba(0,200,150,0.3);
}
*{box-sizing:border-box;margin:0;padding:0}
body{
  font:13px/1.5 'SF Mono','Fira Code','Cascadia Code','JetBrains Mono',monospace;
  background:var(--bg);
  color:var(--text);
  display:flex;flex-direction:column;height:100vh;
  background-image:
    radial-gradient(ellipse at 50% 0%, rgba(0,180,140,0.04) 0%, transparent 60%),
    radial-gradient(ellipse at 80% 20%, rgba(240,160,48,0.03) 0%, transparent 50%),
    linear-gradient(rgba(13,26,45,0.5) 1px, transparent 1px),
    linear-gradient(90deg, rgba(13,26,45,0.5) 1px, transparent 1px);
  background-size: 100% 100%, 100% 100%, 40px 40px, 40px 40px;
}

/* HEADER */
header{
  background:linear-gradient(180deg, var(--surface) 0%, var(--surface2) 100%);
  border-bottom:1px solid var(--border);
  padding:16px 24px;flex-shrink:0;
  box-shadow:0 1px 20px rgba(0,0,0,0.4);
}
.header-top{display:flex;align-items:center;gap:16px;margin-bottom:12px}
.radar{
  width:42px;height:42px;border-radius:50%;
  background:radial-gradient(circle, rgba(0,200,150,0.25) 0%, transparent 65%);
  border:2px solid rgba(0,200,150,0.5);
  position:relative;flex-shrink:0;
  box-shadow:var(--glow-green);
}
.radar::after{
  content:'';position:absolute;top:50%;left:50%;width:50%;height:1px;
  background:linear-gradient(90deg, var(--green) 0%, transparent 100%);
  transform-origin:0 50%;animation:radar-sweep 3s linear infinite;
  box-shadow:0 0 4px var(--green);
}
@keyframes radar-sweep{from{transform:rotate(0deg)}to{transform:rotate(360deg)}}
.radar-dot{position:absolute;width:4px;height:4px;border-radius:50%;background:var(--green);}
.radar-dot.d1{top:8px;left:20px;animation:blip 2s ease-in-out infinite}
.radar-dot.d2{top:24px;right:10px;animation:blip 3s ease-in-out 0.7s infinite}
.radar-dot.d3{bottom:10px;left:14px;animation:blip 2.5s ease-in-out 1.5s infinite}
@keyframes blip{0%,100%{opacity:0.2;box-shadow:none}50%{opacity:1;box-shadow:0 0 6px var(--green)}}

.header-title h1{font-size:18px;font-weight:700;color:var(--white);letter-spacing:1px}
.header-title .sub{font-size:10px;color:var(--muted);text-transform:uppercase;letter-spacing:3px}
.station-tag{
  background:rgba(240,160,48,0.12);color:var(--accent);
  padding:2px 10px;border-radius:3px;font-size:10px;font-weight:700;
  letter-spacing:2px;text-transform:uppercase;margin-left:auto;
  border:1px solid rgba(240,160,48,0.25);
}

/* STATS BAR */
.stats-bar{display:flex;gap:12px;flex-wrap:wrap}
.stat{
  background:var(--surface);border:1px solid var(--border);
  border-radius:4px;padding:6px 12px;display:flex;align-items:center;gap:8px;
  font-size:11px;min-width:0;
}
.stat .label{color:var(--muted);text-transform:uppercase;font-size:9px;letter-spacing:1px;white-space:nowrap}
.stat .value{font-size:15px;font-weight:700;color:var(--white);white-space:nowrap}
.stat .value.accent{color:var(--accent)}
.stat .value.green{color:var(--green)}
.stat .value.mil{color:var(--mil)}
.stat .ext{font-size:9px;color:var(--muted);white-space:nowrap;overflow:hidden;text-overflow:ellipsis;max-width:100px}

/* NAV */
nav{
  display:flex;gap:0;background:var(--surface);border-bottom:2px solid var(--border);
  padding:0 24px;flex-shrink:0;overflow-x:auto;
}
nav button{
  background:none;border:none;color:var(--muted);padding:12px 18px;cursor:pointer;
  font:12px monospace;border-bottom:3px solid transparent;white-space:nowrap;
  transition:all .2s;text-transform:uppercase;letter-spacing:1px;
}
nav button:hover{color:var(--text);border-color:var(--muted)}
nav button.active{color:var(--accent);border-color:var(--accent);text-shadow:0 0 8px rgba(240,160,48,0.3)}
nav button.mil-tab{color:var(--mil)}
nav button.mil-tab.active{border-color:var(--mil);text-shadow:0 0 8px rgba(224,85,85,0.3)}
nav button.nearest-tab{color:var(--green)}
nav button.nearest-tab.active{border-color:var(--green);text-shadow:0 0 8px rgba(0,200,150,0.3)}
nav button.radar-tab{color:var(--blue)}
nav button.radar-tab.active{border-color:var(--blue);text-shadow:0 0 8px rgba(59,143,212,0.3)}

/* MAIN */
main{flex:1;min-height:0;overflow-y:auto;padding:24px;max-width:1440px;margin:0 auto;width:100%}
main.full{padding:0;max-width:none;overflow:hidden}

/* RADAR IFRAME */
.radar-frame{width:100%;height:100%;border:none;background:#000;display:block}

/* PANEL */
.panel{
  background:var(--surface);border:1px solid var(--border);border-radius:6px;
  overflow:hidden;margin-bottom:20px;box-shadow:0 2px 12px rgba(0,0,0,0.3);
}
.panel-header{
  padding:12px 20px;border-bottom:1px solid var(--border);
  display:flex;justify-content:space-between;align-items:center;
  background:var(--surface2);font-size:11px;color:var(--muted);
  text-transform:uppercase;letter-spacing:2px;
}
.panel-header h2{font-size:12px;color:var(--white);font-weight:600;letter-spacing:1px}

/* NEAREST CARD (ATIS-style) */
.nearest-panel{border-color:rgba(0,200,150,0.3);box-shadow:0 0 20px rgba(0,200,150,0.06)}
.nearest-panel .panel-header{background:rgba(0,200,150,0.06);border-color:rgba(0,200,150,0.2)}
.nearest-panel .panel-header h2{color:var(--green)}
.atis-card{display:grid;grid-template-columns:repeat(auto-fit,minmax(200px,1fr));gap:0;padding:0}
.atis-section{padding:16px 20px;border-right:1px solid var(--border)}
.atis-section:last-child{border-right:none}
.atis-label{font-size:9px;color:var(--muted);text-transform:uppercase;letter-spacing:2px;margin-bottom:4px}
.atis-callsign{font-size:28px;font-weight:700;color:var(--warn);line-height:1.1}
.atis-reg{font-size:14px;color:var(--muted);margin-bottom:6px}
.atis-value{font-size:26px;font-weight:700;color:var(--green);line-height:1.1}
.atis-unit{font-size:12px;color:var(--muted)}
.atis-sub{font-size:11px;color:var(--muted);margin-top:2px}
.atis-mil{color:var(--mil);font-weight:700;font-size:12px;margin-top:4px;letter-spacing:2px}
.atis-compass{
  display:inline-block;width:50px;height:50px;border-radius:50%;
  border:2px solid var(--border);position:relative;vertical-align:middle;margin-right:10px;
}
.atis-compass::after{
  content:'';position:absolute;top:50%;left:50%;width:1px;height:40%;
  background:var(--accent);transform-origin:top center;
}
.atis-pos{font-size:11px;font-family:monospace;color:var(--muted)}
.atis-typo{font-size:10px;color:var(--muted);margin-top:4px}

/* EMPTY STATE */
.empty-state{text-align:center;padding:40px 20px;color:var(--muted)}
.empty-state .big{font-size:48px;opacity:0.15;margin-bottom:8px}
.empty-state p{font-size:13px;font-style:italic}

/* TABLE */
.table-wrap{overflow-x:auto}
table{width:100%;border-collapse:collapse}
thead tr{background:var(--surface2)}
th{
  padding:10px 14px;text-align:left;font-size:10px;text-transform:uppercase;
  letter-spacing:1px;color:var(--muted);border-bottom:2px solid var(--border);
  position:sticky;top:0;white-space:nowrap;
}
td{padding:9px 14px;font-size:12px;border-bottom:1px solid rgba(26,45,74,0.6);white-space:nowrap}
tbody tr{transition:background .15s}
tbody tr:hover{background:rgba(240,160,48,0.04)}

/* CELL STYLES */
.icao{color:var(--blue);cursor:pointer;font-weight:600;letter-spacing:0.5px}
.icao:hover{text-decoration:underline;text-shadow:0 0 6px rgba(59,143,212,0.3)}
.callsign-cell{color:var(--warn);font-weight:600}
.callsign-cell:hover{text-decoration:underline;cursor:pointer}
.mil-badge{color:var(--mil);font-size:10px;font-weight:700;letter-spacing:1px;margin-left:6px}
.cat-badge{
  display:inline-block;padding:1px 6px;border-radius:3px;
  font-size:10px;font-weight:600;letter-spacing:0.5px;
  background:rgba(59,143,212,0.1);color:var(--blue);
}
.reg-cell{color:var(--text);font-weight:500}
.op-cell{color:var(--muted);font-size:11px}
.country-cell{color:var(--muted);font-size:11px}
.dim{color:var(--muted)}
.bright{color:var(--white);font-weight:600}

/* LOADING SPINNER */
.loader{display:inline-block;width:10px;height:10px;border:2px solid var(--muted);border-top-color:var(--accent);border-radius:50%;animation:spin .6s linear infinite;margin-left:8px}
@keyframes spin{to{transform:rotate(360deg)}}

/* MODAL */
.modal-overlay{display:none;position:fixed;inset:0;background:rgba(0,0,0,.75);z-index:200;justify-content:center;align-items:center;backdrop-filter:blur(4px)}
.modal-overlay.open{display:flex}
.modal{
  background:var(--surface);border:1px solid var(--border);border-radius:8px;
  max-width:620px;width:92%;max-height:82vh;overflow-y:auto;
  box-shadow:0 0 40px rgba(0,0,0,0.6);
}
.modal-header{
  padding:14px 20px;border-bottom:1px solid var(--border);
  display:flex;justify-content:space-between;align-items:center;
  background:var(--surface2);
}
.modal-header h3{font-size:15px;color:var(--white);font-weight:600;letter-spacing:1px}
.modal-close{
  background:none;border:1px solid var(--border);color:var(--muted);
  font-size:16px;cursor:pointer;padding:4px 10px;border-radius:4px;
}
.modal-close:hover{color:var(--white);border-color:var(--muted)}
.modal-body{padding:20px}
.modal-grid{display:grid;grid-template-columns:1fr 1fr;gap:16px}
.modal-field{margin-bottom:12px}
.modal-label{font-size:9px;color:var(--muted);text-transform:uppercase;letter-spacing:2px;margin-bottom:2px}
.modal-value{font-size:14px;color:var(--text);font-weight:500}
.modal-value.big{font-size:22px;font-weight:700;color:var(--warn)}
.modal-value.green{color:var(--green);font-weight:700;font-size:18px}
.modal-links{margin-top:16px;padding-top:12px;border-top:1px solid var(--border);display:flex;gap:12px;flex-wrap:wrap}
.ext-link{color:var(--blue);text-decoration:none;font-size:11px;padding:4px 10px;border:1px solid rgba(59,143,212,0.25);border-radius:4px;transition:all .15s}
.ext-link:hover{background:rgba(59,143,212,0.1);border-color:var(--blue)}

/* FOOTER */
footer{padding:20px;text-align:center;color:var(--muted);font-size:10px;letter-spacing:1px;border-top:1px solid var(--border);margin-top:20px}

@media(max-width:700px){
  .stats-bar{gap:8px}
  .stat{flex:1 1 45%;min-width:0}
  th,td{padding:7px 10px;font-size:11px}
  .atis-section{border-right:none;border-bottom:1px solid var(--border)}
  .modal-grid{grid-template-columns:1fr}
}
</style>
</head>
<body>
<header>
 <div class="header-top">
  <div class="radar"><div class="radar-dot d1"></div><div class="radar-dot d2"></div><div class="radar-dot d3"></div></div>
  <div class="header-title">
   <h1>ADS-B LEADERBOARD</h1>
   <div class="sub">Droitwich, UK &middot; 52.254&deg;N 2.150&deg;W &middot; 50m AMSL</div>
  </div>
  <div class="station-tag">LOCAL</div>
 </div>
 <div class="stats-bar" id="statsBar">
  <div class="stat"><div><div class="label">Aircraft</div><div class="value green" id="stTotal">--</div></div></div>
  <div class="stat"><div><div class="label">Sightings</div><div class="value accent" id="stSightings">--</div></div></div>
  <div class="stat"><div><div class="label">Military</div><div class="value mil" id="stMil">--</div></div></div>
  <div class="stat"><div><div class="label">Furthest</div><div class="value" id="stFurthest">--</div><div class="ext" id="stFurthestExt"></div></div></div>
  <div class="stat"><div><div class="label">Highest</div><div class="value" id="stHighest">--</div><div class="ext" id="stHighestExt"></div></div></div>
  <div class="stat"><div><div class="label">Fastest</div><div class="value" id="stFastest">--</div><div class="ext" id="stFastestExt"></div></div></div>
  <div class="stat"><div><div class="label">Slowest</div><div class="value green" id="stSlowest">--</div><div class="ext" id="stSlowestExt"></div></div></div>
  <div class="stat"><div><div class="label">Closest</div><div class="value green" id="stClosest">--</div><div class="ext" id="stClosestExt"></div></div></div>
  <div class="stat"><div><div class="label">Lowest</div><div class="value green" id="stLowest">--</div><div class="ext" id="stLowestExt"></div></div></div>
 </div>
</header>
<nav>
 <button class="radar-tab active" data-tab="radar">Radar</button>
 <button class="nearest-tab" data-tab="nearest">Nearest</button>
 <button data-tab="leaderboard">Leaderboard</button>
 <button class="mil-tab" data-tab="military">Military</button>
 <button data-tab="distance">Distance</button>
 <button data-tab="altitude">Altitude</button>
 <button data-tab="speed">Speed</button>
 <button data-tab="recent">Recent</button>
 <button data-tab="daily">Daily</button>
</nav>
<main>
<!-- RADAR -->
<div id="tab-radar" class="tab-content" style="height:100%">
 <iframe class="radar-frame" src="https://ads-b.asdfx.us" allow="geolocation"></iframe>
</div>
<!-- NEAREST -->
<div id="tab-nearest" class="tab-content" style="display:none">
 <div class="panel nearest-panel">
  <div class="panel-header"><h2>NEAREST AIRCRAFT &lt;12,000ft / &lt;10nm / last 3m</h2></div>
  <div id="nearestCard"><div class="empty-state"><div class="big">&#9992;</div><p>Scanning for low-altitude aircraft...</p></div></div>
 </div>
</div>
<!-- LEADERBOARD -->
<div id="tab-leaderboard" class="tab-content" style="display:none">
 <div class="panel">
  <div class="panel-header"><h2>MOST SEEN</h2></div>
  <div class="table-wrap"><table>
   <thead><tr><th>ICAO24</th><th>Reg</th><th>Type</th><th>Operator</th><th>Country</th><th>Callsign</th><th class="bright">Seen</th><th>Dist min/max</th><th>Alt min/max</th><th>Max Spd</th><th>Last</th></tr></thead>
   <tbody id="lbBody"><tr><td colspan="11"><div class="empty-state"><p>Loading...</p></div></td></tr></tbody>
  </table></div>
 </div>
</div>
<!-- MILITARY -->
<div id="tab-military" class="tab-content" style="display:none">
 <div class="panel" style="border-color:rgba(224,85,85,0.2)">
  <div class="panel-header" style="background:rgba(224,85,85,0.05);border-color:rgba(224,85,85,0.15)"><h2 style="color:var(--mil)">MILITARY</h2></div>
  <div class="table-wrap"><table>
   <thead><tr><th>ICAO24</th><th>Reg</th><th>Type</th><th>Operator</th><th>Country</th><th>Callsign</th><th class="bright">Seen</th><th>Dist min/max</th><th>Max Spd</th><th>Last</th></tr></thead>
   <tbody id="milBody"><tr><td colspan="10"><div class="empty-state"><p>Loading...</p></div></td></tr></tbody>
  </table></div>
 </div>
</div>
<!-- DISTANCE -->
<div id="tab-distance" class="tab-content" style="display:none">
 <div class="panel">
  <div class="panel-header"><h2>FURTHEST HORIZONTALLY (NM)</h2></div>
  <div class="table-wrap"><table>
   <thead><tr><th>ICAO24</th><th>Reg</th><th>Type</th><th>Operator</th><th>Country</th><th>Callsign</th><th class="bright">Max Dist</th><th>Min Dist</th><th>Seen</th><th>Last</th></tr></thead>
   <tbody id="distBody"><tr><td colspan="10"><div class="empty-state"><p>Loading...</p></div></td></tr></tbody>
  </table></div>
 </div>
</div>
<!-- ALTITUDE -->
<div id="tab-altitude" class="tab-content" style="display:none">
 <div class="panel">
  <div class="panel-header"><h2>HIGHEST ALTITUDE (FT)</h2></div>
  <div class="table-wrap"><table>
   <thead><tr><th>ICAO24</th><th>Reg</th><th>Type</th><th>Operator</th><th>Country</th><th>Callsign</th><th class="bright">Max Alt</th><th>Min Alt</th><th>Seen</th><th>Last</th></tr></thead>
   <tbody id="altBody"><tr><td colspan="10"><div class="empty-state"><p>Loading...</p></div></td></tr></tbody>
  </table></div>
 </div>
</div>
<!-- SPEED -->
<div id="tab-speed" class="tab-content" style="display:none">
 <div class="panel">
  <div class="panel-header"><h2>FASTEST GROUND SPEED (KT)</h2></div>
  <div class="table-wrap"><table>
   <thead><tr><th>ICAO24</th><th>Reg</th><th>Type</th><th>Operator</th><th>Country</th><th>Callsign</th><th class="bright">Max Speed</th><th>Max Alt</th><th>Seen</th><th>Last</th></tr></thead>
   <tbody id="spdBody"><tr><td colspan="10"><div class="empty-state"><p>Loading...</p></div></td></tr></tbody>
  </table></div>
 </div>
</div>
<!-- RECENT -->
<div id="tab-recent" class="tab-content" style="display:none">
 <div class="panel">
  <div class="panel-header"><h2>RECENT SIGHTINGS</h2></div>
  <div class="table-wrap"><table>
   <thead><tr><th>ICAO24</th><th>Callsign</th><th>Cat</th><th>Alt</th><th>Spd</th><th>Trk</th><th>Dist</th><th>Brg</th><th>Lat/Lon</th><th>Seen</th></tr></thead>
   <tbody id="recBody"><tr><td colspan="10"><div class="empty-state"><p>Loading...</p></div></td></tr></tbody>
  </table></div>
 </div>
</div>
<!-- DAILY -->
<div id="tab-daily" class="tab-content" style="display:none">
 <div class="panel">
  <div class="panel-header"><h2>DAILY STATISTICS</h2></div>
  <div class="table-wrap"><table>
   <thead><tr><th>Date</th><th class="bright">Unique</th><th>Sightings</th><th>Military</th><th>Max Dist</th><th>Max Alt</th><th>Max Speed</th></tr></thead>
   <tbody id="dayBody"><tr><td colspan="7"><div class="empty-state"><p>Loading...</p></div></td></tr></tbody>
  </table></div>
 </div>
</div>
</main>
<div class="modal-overlay" id="modalOverlay">
 <div class="modal">
  <div class="modal-header"><h3 id="modalTitle">AIRCRAFT DETAIL</h3><button class="modal-close" onclick="closeModal()">CLOSE</button></div>
  <div class="modal-body" id="modalBody"></div>
 </div>
</div>
<footer>UPDATED <span id="footerTime"></span> &middot; POLL 10s &middot; OPENSKY ENRICHMENT</footer>
<script>
var activeTab='radar';
var tabData={};
var loaderEl=null;

function fmt(v,d){return v!=null?v.toFixed(d||1):'--'}
function ts(t){return new Date(t*1000).toLocaleString()}
function ago(t){var s=Math.floor(Date.now()/1000-t);if(s<60)return s+'s';if(s<3600)return Math.floor(s/60)+'m';if(s<86400)return Math.floor(s/3600)+'h';return Math.floor(s/86400)+'d';}
function flag(f){return f?f+' ':''}
function cardDir(b){var d=['N','NE','E','SE','S','SW','W','NW'];return d[Math.round(((b||0)+360)%360/45)%8]}
function makeRow(cols){var t='';for(var i=0;i<cols.length;i++)t+='<td>'+cols[i]+'</td>';return'<tr>'+t+'</tr>'}

function loadStats(){
 fetch('/api/stats').then(function(r){return r.json()}).then(function(s){
  document.getElementById('stTotal').textContent=(s.total_aircraft||0).toLocaleString();
  document.getElementById('stSightings').textContent=(s.total_sightings||0).toLocaleString();
  document.getElementById('stMil').textContent=(s.military_aircraft||0).toLocaleString();
  document.getElementById('stFurthest').textContent=fmt(s.furthest_nm,1)+'nm';
  document.getElementById('stFurthestExt').textContent=s.furthest_icao||'';
  document.getElementById('stClosest').textContent=fmt(s.closest_nm,1)+'nm';
  document.getElementById('stClosestExt').textContent=s.closest_icao||'';
  document.getElementById('stHighest').textContent=fmt(s.highest_ft,0)+'ft';
  document.getElementById('stHighestExt').textContent=s.highest_icao||'';
  document.getElementById('stLowest').textContent=fmt(s.lowest_ft,0)+'ft';
  document.getElementById('stLowestExt').textContent=s.lowest_icao||'';
  document.getElementById('stFastest').textContent=fmt(s.fastest_kt,1)+'kt';
  document.getElementById('stFastestExt').textContent=s.fastest_icao||'';
  document.getElementById('stSlowest').textContent=fmt(s.slowest_kt,1)+'kt';
  document.getElementById('stSlowestExt').textContent=s.slowest_icao||'';
  document.getElementById('footerTime').textContent=new Date().toLocaleTimeString();
 }).catch(function(){});
}

function loadNearest(){
 var card=document.getElementById('nearestCard');
 fetch('/api/nearest?seen=180').then(function(r){return r.json()}).then(function(n){
  if(!n){card.innerHTML='<div class="empty-state"><div class="big">&#9992;</div><p>No low-altitude aircraft in range</p></div>';return}
  var cs=n.callsign||n.icao24||'--';
  var reg=n.reg||'';
  var typ=[n.manufacturer,n.model].filter(Boolean).join(' ')||'';
  var op=n.operator||'';
  card.innerHTML=
   '<div class="atis-card">'+
   '<div class="atis-section">'+
    '<div class="atis-label">CALLSIGN</div>'+
    '<div class="atis-callsign">'+cs+'</div>'+
    (reg?'<div class="atis-reg">'+reg+'</div>':'')+
    (typ?'<div class="atis-sub">'+typ+'</div>':'')+
    (op?'<div class="atis-sub">'+op+'</div>':'')+
    (n.is_military?'<div class="atis-mil">MILITARY</div>':'')+
   '</div>'+
   '<div class="atis-section">'+
    '<div class="atis-label">DISTANCE</div>'+
    '<div class="atis-value">'+fmt(n.dist_nm,1)+' <span class="atis-unit">NM</span></div>'+
    '<div class="atis-label" style="margin-top:12px">ALTITUDE</div>'+
    '<div class="atis-value">'+(n.altitude_ft||0).toLocaleString()+' <span class="atis-unit">FT</span></div>'+
   '</div>'+
   '<div class="atis-section">'+
    '<div class="atis-label">BEARING</div>'+
    '<div class="atis-value">'+n.bearing+'&deg; <span class="atis-unit">'+cardDir(n.bearing)+'</span></div>'+
    '<div class="atis-label" style="margin-top:12px">SPEED / TRACK</div>'+
    '<div class="atis-value">'+fmt(n.speed_kt,0)+' <span class="atis-unit">KT / '+fmt(n.track,0)+'&deg;</span></div>'+
   '</div>'+
   '<div class="atis-section">'+
    '<div class="atis-label">POSITION</div>'+
    '<div class="atis-pos">'+fmt(n.lat,4)+', '+fmt(n.lon,4)+'</div>'+
    '<div class="atis-label" style="margin-top:12px">LAST SEEN</div>'+
    '<div class="atis-pos">'+ago(n.seen_at)+' ago &middot; '+ts(n.seen_at)+'</div>'+
   '</div>'+
   '</div>';
 }).catch(function(){card.innerHTML='<div class="empty-state"><div class="big">!</div><p>Error loading</p></div>'});
}

var leaderboardCols=null;
function buildLeaderboardHTML(data,tab){
 var h='';
 for(var i=0;i<data.length;i++){
  var a=data[i];
  var icoa='<span class="icao" onclick="showAircraft(\''+a.icao24+'\')">'+a.icao24+'</span>'+(a.is_military?' <span class="mil-badge">MIL</span>':'');
  var reg=a.reg?'<span class="reg-cell">'+a.reg+'</span>':'<span class="dim">-</span>';
  var typ=[a.manufacturer,a.model].filter(Boolean).join(' ')||'<span class="dim">-</span>';
  var op=a.operator?'<span class="op-cell">'+a.operator+'</span>':'<span class="dim">-</span>';
  var cty=flag(a.country_flag)+'<span class="country-cell">'+a.country+'</span>';
  var cs=a.last_callsign?'<span class="callsign-cell">'+a.last_callsign+'</span>':'<span class="dim">-</span>';
  var seen=a.total_sightings.toLocaleString();
  var last=ago(a.last_seen);

  if(tab==='leaderboard'){
   h+=makeRow([icoa,reg,typ,op,cty,cs,seen,
    (a.min_dist_nm>0?fmt(a.min_dist_nm,1)+'/'+fmt(a.max_dist_nm,1):fmt(a.max_dist_nm,1))+' nm',
    (a.min_alt_ft||'-')+'/'+(a.max_alt_ft||'-'),
    fmt(a.max_speed_kt,0)+' kt',last]);
  }else if(tab==='military'){
   h+=makeRow([icoa,reg,typ,op,cty,cs,seen,
    (a.min_dist_nm>0?fmt(a.min_dist_nm,1)+'/'+fmt(a.max_dist_nm,1):fmt(a.max_dist_nm,1))+' nm',
    fmt(a.max_speed_kt,0)+' kt',last]);
  }else if(tab==='distance'){
   h+=makeRow([icoa,reg,typ,op,cty,cs,
    fmt(a.max_dist_nm,1)+' nm',fmt(a.min_dist_nm,1)+' nm',seen,last]);
  }else if(tab==='altitude'){
   h+=makeRow([icoa,reg,typ,op,cty,cs,
    (a.max_alt_ft||'-').toLocaleString()+' ft',(a.min_alt_ft||'-').toLocaleString()+' ft',seen,last]);
  }else if(tab==='speed'){
   h+=makeRow([icoa,reg,typ,op,cty,cs,
    fmt(a.max_speed_kt,0)+' kt',(a.max_alt_ft||'-').toLocaleString()+' ft',seen,last]);
  }
 }
 return h;
}

function loadTab(tab){
 var body,url='';
 switch(tab){
  case'nearest':loadNearest();return;
  case'leaderboard':body=document.getElementById('lbBody');url='/api/leaderboard?sort=sightings&limit=100';break;
  case'military':body=document.getElementById('milBody');url='/api/leaderboard?sort=sightings&limit=100&military=1';break;
  case'distance':body=document.getElementById('distBody');url='/api/leaderboard?sort=distance&limit=100';break;
  case'altitude':body=document.getElementById('altBody');url='/api/leaderboard?sort=altitude&limit=100';break;
  case'speed':body=document.getElementById('spdBody');url='/api/leaderboard?sort=speed&limit=100';break;
  case'recent':body=document.getElementById('recBody');url='/api/recent?limit=100';break;
  case'daily':body=document.getElementById('dayBody');url='/api/daily?limit=30';break;
  default:return;
 }
 var key='tab_'+tab;
 fetch(url).then(function(r){return r.json()}).then(function(data){
  if(!data||!data.length){body.innerHTML='<tr><td colspan="12"><div class="empty-state"><p>No data yet</p></div></td></tr>';return}
  var cached=tabData[key];
  if(cached&&JSON.stringify(cached)===JSON.stringify(data))return;
  tabData[key]=data;
  var h='';
  if(tab==='daily'){
   for(var i=0;i<data.length;i++){var d=data[i];
    h+=makeRow([d.date,'<span class="bright">'+d.unique_aircraft.toLocaleString()+'</span>',d.total_sightings.toLocaleString(),d.military_aircraft.toLocaleString(),fmt(d.max_distance_nm,1)+' nm',(d.max_altitude_ft||'-').toLocaleString()+' ft',fmt(d.max_speed_kt,1)+' kt']);
   }
  }else if(tab==='recent'){
   for(var i=0;i<data.length;i++){var s=data[i];
    h+=makeRow(['<span class="icao" onclick="showAircraft(\''+s.icao24+'\')">'+s.icao24+'</span>','<span class="callsign-cell">'+(s.callsign||'-')+'</span>','<span class="cat-badge">'+(s.category||'-')+'</span>',(s.alt_ft||'-').toLocaleString(),fmt(s.speed_kt,0),fmt(s.track,0)+'&deg;',fmt(s.dist_nm,1)+' nm',s.bearing+'&deg;',fmt(s.lat,4)+' '+fmt(s.lon,4),ago(s.seen_at)]);
   }
  }else{
   h=buildLeaderboardHTML(data,tab);
  }
  body.innerHTML=h;
 }).catch(function(){body.innerHTML='<tr><td colspan="12"><div class="empty-state"><div class="big">!</div><p>Error loading</p></div></td></tr>'});
}

function showAircraft(icao){
 var overlay=document.getElementById('modalOverlay');
 var title=document.getElementById('modalTitle');
 var body=document.getElementById('modalBody');
 title.textContent='Loading...';
 body.innerHTML='<div class="empty-state"><p>Loading...</p></div>';
 overlay.classList.add('open');
 fetch('/api/aircraft/'+icao).then(function(r){return r.json()}).then(function(a){
  title.textContent=a.registration||a.last_callsign||a.icao24;
  var typ=[a.manufacturer,a.model].filter(Boolean).join(' ')||'-';
  var links='';
  if(a.registration){links+='<a class="ext-link" target="_blank" href="https://www.planespotters.net/search?q='+encodeURIComponent(a.registration)+'">PLANESPOTTERS</a>';}
  if(a.registration){links+='<a class="ext-link" target="_blank" href="https://www.flightradar24.com/data/aircraft/'+encodeURIComponent(a.registration)+'">FR24</a>';}
  links+='<a class="ext-link" target="_blank" href="https://globe.adsbexchange.com/?icao='+icao+'">ADSBX</a>';
  body.innerHTML=
   '<div class="modal-grid">'+
   '<div>'+
    '<div class="modal-field"><div class="modal-label">ICAO24</div><div class="modal-value">'+a.icao24+'</div></div>'+
    '<div class="modal-field"><div class="modal-label">REGISTRATION</div><div class="modal-value big">'+(a.reg||'-')+'</div></div>'+
    '<div class="modal-field"><div class="modal-label">TYPE</div><div class="modal-value">'+typ+'</div></div>'+
    '<div class="modal-field"><div class="modal-label">OPERATOR</div><div class="modal-value">'+(a.operator||'-')+'</div></div>'+
   '</div>'+
   '<div>'+
    '<div class="modal-field"><div class="modal-label">COUNTRY</div><div class="modal-value">'+flag(a.country_flag)+(a.country||'-')+'</div></div>'+
    '<div class="modal-field"><div class="modal-label">MILITARY</div><div class="modal-value">'+(a.is_military?'<span style="color:var(--mil);font-weight:700">YES</span>':'No')+'</div></div>'+
    '<div class="modal-field"><div class="modal-label">EMITTER</div><div class="modal-value">'+(a.category||'-')+'</div></div>'+
    '<div class="modal-field"><div class="modal-label">LAST CALLSIGN</div><div class="modal-value" style="color:var(--warn);font-weight:600">'+(a.last_callsign||'-')+'</div></div>'+
   '</div>'+
   '</div>'+
   '<div style="display:grid;grid-template-columns:repeat(4,1fr);gap:12px;margin-top:16px;padding-top:16px;border-top:1px solid var(--border)">'+
   '<div><div class="modal-label">TOTAL SEEN</div><div class="modal-value green">'+(a.total_sightings||0).toLocaleString()+'</div></div>'+
   '<div><div class="modal-label">FIRST SEEN</div><div class="modal-value">'+ts(a.first_seen)+'</div></div>'+
   '<div><div class="modal-label">MIN/MAX DIST</div><div class="modal-value">'+fmt(a.min_dist_nm,1)+' / '+fmt(a.max_dist_nm,1)+' nm</div></div>'+
   '<div><div class="modal-label">MIN/MAX ALT</div><div class="modal-value">'+(a.min_alt_ft||'-')+' / '+(a.max_alt_ft||'-')+' ft</div></div>'+
   '<div><div class="modal-label">MAX SPEED</div><div class="modal-value green">'+fmt(a.max_speed_kt,0)+' kt</div></div>'+
   '<div><div class="modal-label">LAST SEEN</div><div class="modal-value">'+ts(a.last_seen)+'</div></div>'+
   '<div><div class="modal-label">LAST POSITION</div><div class="modal-value">'+fmt(a.last_lat,4)+', '+fmt(a.last_lon,4)+'</div></div>'+
   '<div></div>'+
   '</div>'+
   '<div class="modal-links">'+links+'</div>';
 }).catch(function(){body.innerHTML='<div class="empty-state"><p>Failed to load</p></div>'});
}

function closeModal(){document.getElementById('modalOverlay').classList.remove('open')}
document.getElementById('modalOverlay').addEventListener('click',function(e){if(e.target===this)closeModal()});

function switchTab(tab,silent){
 if(!silent)window.location.hash=tab;
 document.querySelectorAll('nav button').forEach(function(b){b.classList.remove('active');if(b.dataset.tab===tab)b.classList.add('active')});
 document.querySelectorAll('.tab-content').forEach(function(x){x.style.display='none'});
 var main=document.querySelector('main');
 var el=document.getElementById('tab-'+tab);
 if(el){
  el.style.display='block';activeTab=tab;
  if(tab==='radar'){main.classList.add('full');el.style.height='100%'}else{main.classList.remove('full');el.style.height='';loadTab(tab)}
 }
}

document.querySelectorAll('nav button').forEach(function(b){
 b.addEventListener('click',function(){switchTab(b.dataset.tab)});
});

var hash=window.location.hash.replace('#','');
if(hash&&document.getElementById('tab-'+hash)){switchTab(hash,true)}
else{switchTab('radar',true)}

window.addEventListener('hashchange',function(){
 var h=window.location.hash.replace('#','');
 if(h&&document.getElementById('tab-'+h))switchTab(h,true);
});

loadStats();
setInterval(loadStats,10000);
setInterval(function(){loadTab(activeTab)},10000);
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
body{
  background:#060d17;color:#dce4f0;font:14px 'SF Mono','Fira Code',monospace;
  display:flex;align-items:center;justify-content:center;min-height:100vh;padding:16px;
  background-image:radial-gradient(ellipse at 50% 0%,rgba(0,180,140,0.06) 0%,transparent 60%);
}
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
 card.innerHTML=
  '<h2>NEAREST AIRCRAFT</h2>'+
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
