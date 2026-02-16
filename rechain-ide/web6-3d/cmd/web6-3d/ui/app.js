const canvas = document.getElementById("viewport");
const ctx = canvas.getContext("2d");
const search = document.getElementById("search");
const mode = document.getElementById("mode");
const depth = document.getElementById("depth");
const typeSelect = document.getElementById("type");
const indicator = document.getElementById("mode-indicator");
const taskState = document.getElementById("task-state");
const taskMerge = document.getElementById("task-merge");
const taskParent = document.getElementById("task-parent");
const taskSort = document.getElementById("task-sort");
const replayMode = document.getElementById("replay-mode");
const replayBtn = document.getElementById("replay-btn");
const debugScope = document.getElementById("debug-scope");
const debugRefreshBtn = document.getElementById("debug-refresh");
const debugCopyBtn = document.getElementById("debug-copy");
const proxyRefreshBtn = document.getElementById("proxy-refresh");
const proxyCopyBtn = document.getElementById("proxy-copy");
const proxyHistoryLimit = document.getElementById("proxy-history-limit");
const proxyHistoryFormat = document.getElementById("proxy-history-format");
const proxyHistoryCopyBtn = document.getElementById("proxy-history-copy");
const proxyHistoryResetBtn = document.getElementById("proxy-history-reset");
const dashboardCopyBtn = document.getElementById("dashboard-copy");
const dashboardWeb6CopyBtn = document.getElementById("dashboard-web6-copy");
const dashboardWeb6HistoryLimit = document.getElementById("dashboard-web6-history-limit");
const dashboardWeb6HistoryLevel = document.getElementById("dashboard-web6-history-level");
const dashboardWeb6HistorySource = document.getElementById("dashboard-web6-history-source");
const dashboardWeb6HistoryCopyBtn = document.getElementById("dashboard-web6-history-copy");
const dashboardWeb6HistoryResetBtn = document.getElementById("dashboard-web6-history-reset");
const quickFailedBtn = document.getElementById("quick-failed");
const quickReplayBtn = document.getElementById("quick-replay");
const quickSoftBtn = document.getElementById("quick-soft");
const compareIdInput = document.getElementById("compare-id");
const compareBtn = document.getElementById("compare-btn");

const params = new URLSearchParams(location.search);
const urlMode = params.get("mode");
const urlDepth = params.get("depth");
const urlType = params.get("type");
const urlQ = params.get("q");

if (urlMode) {
  mode.value = urlMode;
} else {
  const storedMode = localStorage.getItem("web6_mode");
  if (storedMode) mode.value = storedMode;
}
if (urlDepth) {
  depth.value = urlDepth;
} else {
  const storedDepth = localStorage.getItem("web6_depth");
  if (storedDepth) depth.value = storedDepth;
}
if (urlType) {
  typeSelect.value = urlType;
} else {
  const storedType = localStorage.getItem("web6_type");
  if (storedType) typeSelect.value = storedType;
}
if (urlQ) {
  search.value = urlQ;
} else {
  const storedQ = localStorage.getItem("web6_q");
  if (storedQ) search.value = storedQ;
}
const storedTaskState = localStorage.getItem("web6_task_state");
if (storedTaskState) taskState.value = storedTaskState;
const storedTaskMerge = localStorage.getItem("web6_task_merge");
if (storedTaskMerge) taskMerge.value = storedTaskMerge;
const storedTaskParent = localStorage.getItem("web6_task_parent");
if (storedTaskParent) taskParent.value = storedTaskParent;
const storedTaskSort = localStorage.getItem("web6_task_sort");
if (storedTaskSort) taskSort.value = storedTaskSort;
const storedReplayMode = localStorage.getItem("web6_replay_mode");
if (storedReplayMode) replayMode.value = storedReplayMode;
const storedDebugScope = localStorage.getItem("web6_debug_scope");
if (storedDebugScope) debugScope.value = storedDebugScope;

function cssVar(name) {
  return getComputedStyle(document.documentElement).getPropertyValue(name).trim() || "";
}

function setVar(name, val) {
  document.documentElement.style.setProperty(name, val);
}

const colors = window.WEB6_COLORS || {};
if (colors.matchDefault) setVar("--match-default", colors.matchDefault);
if (colors.matchDir) setVar("--match-dir", colors.matchDir);
if (colors.matchFile) setVar("--match-file", colors.matchFile);
if (colors.matchPkg) setVar("--match-pkg", colors.matchPkg);
if (colors.matchUnknown) setVar("--match-unknown", colors.matchUnknown);
if (colors.typeUnknown) setVar("--type-unknown", colors.typeUnknown);

const colorDefault = document.getElementById("color-match-default");
const colorDir = document.getElementById("color-match-dir");
const colorFile = document.getElementById("color-match-file");
const colorPkg = document.getElementById("color-match-pkg");
const colorUnknown = document.getElementById("color-match-unknown");

function initColor(el, varName, key) {
  const stored = localStorage.getItem(key);
  const current = stored || cssVar(varName);
  if (current) el.value = current;
  setVar(varName, el.value);
  el.addEventListener("input", () => {
    setVar(varName, el.value);
    localStorage.setItem(key, el.value);
  });
}

initColor(colorDefault, "--match-default", "web6_match_default");
initColor(colorDir, "--match-dir", "web6_match_dir");
initColor(colorFile, "--match-file", "web6_match_file");
initColor(colorPkg, "--match-pkg", "web6_match_pkg");
initColor(colorUnknown, "--match-unknown", "web6_match_unknown");

const defaults = {
  def: cssVar("--match-default"),
  dir: cssVar("--match-dir"),
  file: cssVar("--match-file"),
  pkg: cssVar("--match-pkg"),
  unknown: cssVar("--match-unknown"),
};

function resetColors() {
  setVar("--match-default", defaults.def);
  setVar("--match-dir", defaults.dir);
  setVar("--match-file", defaults.file);
  setVar("--match-pkg", defaults.pkg);
  setVar("--match-unknown", defaults.unknown);
  localStorage.removeItem("web6_match_default");
  localStorage.removeItem("web6_match_dir");
  localStorage.removeItem("web6_match_file");
  localStorage.removeItem("web6_match_pkg");
  localStorage.removeItem("web6_match_unknown");
  colorDefault.value = defaults.def;
  colorDir.value = defaults.dir;
  colorFile.value = defaults.file;
  colorPkg.value = defaults.pkg;
  colorUnknown.value = defaults.unknown;
}

document.getElementById("reset-colors").addEventListener("click", () => resetColors());
document.getElementById("reset-all").addEventListener("click", () => {
  localStorage.removeItem("web6_mode");
  localStorage.removeItem("web6_depth");
  localStorage.removeItem("web6_type");
  localStorage.removeItem("web6_q");
  localStorage.removeItem("web6_task_state");
  localStorage.removeItem("web6_task_merge");
  localStorage.removeItem("web6_task_parent");
  localStorage.removeItem("web6_task_sort");
  localStorage.removeItem("web6_replay_mode");
  localStorage.removeItem("web6_debug_scope");
  resetColors();
  mode.value = "tree";
  depth.value = "1";
  typeSelect.value = "all";
  search.value = "";
  taskState.value = "all";
  taskMerge.value = "all";
  taskParent.value = "all";
  taskSort.value = "updated_desc";
  replayMode.value = "force-policy";
  debugScope.value = "all";
  query = "";
  history.replaceState(null, "", location.pathname);
  loadGraph();
});

const refreshBtn = document.getElementById("refresh");
const clearBtn = document.getElementById("clear");
const settingsBtn = document.getElementById("settings");
const drawer = document.getElementById("drawer");
const shareBtn = document.getElementById("share");

settingsBtn.addEventListener("click", () => drawer.classList.toggle("open"));
shareBtn.addEventListener("click", async () => {
  try {
    await navigator.clipboard.writeText(location.href);
    shareBtn.textContent = "copied";
  } catch (_e) {
    shareBtn.textContent = "failed";
  }
  setTimeout(() => { shareBtn.textContent = "copy share URL"; }, 1200);
});

let nodes = [];
let edges = [];
let t = 0;
let zoom = 1;
let offsetX = 0;
let offsetY = 0;
let dragging = false;
let lastX = 0;
let lastY = 0;
let query = "";
let selectedTaskId = "";
let lastProxyProm = "";
let lastProxyHistoryProm = "";
let lastDashboardProm = "";
let lastDashboardWeb6SummaryProm = "";
let lastDashboardWeb6HistoryProm = "";

const storedWeb6HistoryLimit = localStorage.getItem("web6_dashboard_web6_history_limit");
if (storedWeb6HistoryLimit) dashboardWeb6HistoryLimit.value = storedWeb6HistoryLimit;
const storedWeb6HistoryLevel = localStorage.getItem("web6_dashboard_web6_history_level");
if (storedWeb6HistoryLevel) dashboardWeb6HistoryLevel.value = storedWeb6HistoryLevel;
const storedWeb6HistorySource = localStorage.getItem("web6_dashboard_web6_history_source");
if (storedWeb6HistorySource) dashboardWeb6HistorySource.value = storedWeb6HistorySource;
const storedProxyHistoryLimit = localStorage.getItem("web6_proxy_history_limit");
if (storedProxyHistoryLimit) proxyHistoryLimit.value = storedProxyHistoryLimit;
const storedProxyHistoryFormat = localStorage.getItem("web6_proxy_history_format");
if (storedProxyHistoryFormat) proxyHistoryFormat.value = storedProxyHistoryFormat;

const typeColors = { dir: "#0aa", file: "#555", pkg: "#7a3", unknown: "var(--type-unknown)" };
const matchColors = { dir: "--match-dir", file: "--match-file", pkg: "--match-pkg", unknown: "--match-unknown", default: "--match-default" };

canvas.addEventListener("wheel", (e) => {
  e.preventDefault();
  const delta = Math.sign(e.deltaY) * 0.1;
  zoom = Math.max(0.4, Math.min(2.5, zoom - delta));
});
canvas.addEventListener("mousedown", (e) => {
  dragging = true;
  lastX = e.clientX;
  lastY = e.clientY;
});
window.addEventListener("mouseup", () => { dragging = false; });
window.addEventListener("mousemove", (e) => {
  if (!dragging) return;
  offsetX += e.clientX - lastX;
  offsetY += e.clientY - lastY;
  lastX = e.clientX;
  lastY = e.clientY;
});

function proj(x, y, z) {
  const scale = 350 / (350 + z);
  return {
    x: 450 + (x * scale * zoom) + offsetX,
    y: 240 + (y * scale * zoom) + offsetY,
    scale: scale * zoom,
  };
}

function isMatch(n) {
  if (!query) return false;
  const q = query.toLowerCase();
  return n.label.toLowerCase().includes(q) || n.id.toLowerCase().includes(q);
}

function nodeColor(n, match) {
  if (match) {
    const key = matchColors[n.type] || matchColors.default;
    const val = cssVar(key);
    return val || "#f39c12";
  }
  if (typeColors[n.type]) return typeColors[n.type];
  return "#2c6";
}

function draw() {
  ctx.clearRect(0, 0, 900, 480);
  ctx.fillStyle = "#fafafa";
  ctx.fillRect(0, 0, 900, 480);

  edges.forEach((e) => {
    const a = nodes.find((n) => n.id === e.from);
    const b = nodes.find((n) => n.id === e.to);
    if (!a || !b) return;
    const pa = proj(a.x, a.y, a.z);
    const pb = proj(b.x, b.y, b.z);
    ctx.strokeStyle = "#999";
    ctx.beginPath();
    ctx.moveTo(pa.x, pa.y);
    ctx.lineTo(pb.x, pb.y);
    ctx.stroke();
  });

  nodes.forEach((n) => {
    const p = proj(n.x, n.y, n.z);
    const match = isMatch(n);
    ctx.fillStyle = nodeColor(n, match);
    ctx.beginPath();
    ctx.arc(p.x, p.y, 6 * p.scale, 0, Math.PI * 2);
    ctx.fill();
    ctx.fillStyle = "#222";
    ctx.fillText(n.label, p.x + 8, p.y - 8);
  });
}

function tick() {
  t += 0.01;
  nodes.forEach((n, i) => { n.z = 80 * Math.sin(t + i); });
  draw();
  requestAnimationFrame(tick);
}

function updateCache() {
  fetch("/cache/metrics")
    .then((r) => r.json())
    .then((d) => {
      const type = d.filter_type || "all";
      const n = d.graph_nodes || 0;
      const e = d.graph_edges || 0;
      document.getElementById("cache").textContent = `cache: nodes ${n} edges ${e} type ${type}`;
    })
    .catch(() => { document.getElementById("cache").textContent = "cache metrics: n/a"; });
}

function updateQueue() {
  fetch("/queue-depth")
    .then((r) => r.json())
    .then((d) => { document.getElementById("queue").textContent = "Queue depth: " + d.queue_depth; })
    .catch(() => { document.getElementById("queue").textContent = "Queue depth: n/a"; });
}

function updateTrace() {
  fetch("/last-task-trace")
    .then((r) => {
      if (!r.ok) throw new Error("bad status");
      return r.json();
    })
    .then((t) => {
      const taskId = t.task_id || "n/a";
      const state = t.state || "unknown";
      const source = t.merge_source || "none";
      const selected = (t.selected_models || []).length;
      const results = (t.results || []).length;
      const quality = (t.merge && typeof t.merge.quality_score === "number") ? t.merge.quality_score.toFixed(3) : "n/a";
      document.getElementById("trace").textContent =
        `Last task trace: ${taskId} | state=${state} | merge=${source} | selected=${selected} | results=${results} | quality=${quality}`;
    })
    .catch(() => {
      document.getElementById("trace").textContent = "Last task trace: n/a";
    });
}

function renderTraceDetails(trace) {
  const details = document.getElementById("trace-details");
  const models = trace.results || [];
  if (!models.length) {
    details.textContent = "Task trace details: no per-model results";
    return;
  }
  const lines = models.map((m) => {
    const id = m.model_id || "n/a";
    const lat = (typeof m.latency_ms === "number") ? m.latency_ms.toFixed(1) : "n/a";
    const cost = (typeof m.cost_usd === "number") ? m.cost_usd.toFixed(4) : "n/a";
    const q = (typeof m.quality_score === "number") ? m.quality_score.toFixed(3) : "n/a";
    return `${id}: latency=${lat}ms cost=${cost} quality=${q}`;
  });
  const selected = (trace.selected_models || []).join(", ");
  details.textContent = `Task trace details: ${trace.task_id || "n/a"} | selected=[${selected}]\n${lines.join("\n")}`;
}

function loadTaskTrace(taskId) {
  if (!taskId) return;
  fetch(`/task-trace?id=${encodeURIComponent(taskId)}`)
    .then((r) => {
      if (!r.ok) throw new Error("bad status");
      return r.json();
    })
    .then((trace) => {
      renderTraceDetails(trace);
    })
    .catch(() => {
      document.getElementById("trace-details").textContent = "Task trace details: n/a";
    });
}

function loadReplayChain(taskId) {
  if (!taskId) return;
  fetch(`/replay-chain?id=${encodeURIComponent(taskId)}`)
    .then((r) => {
      if (!r.ok) throw new Error("bad status");
      return r.json();
    })
    .then((chain) => {
      const lineage = chain.lineage || [];
      const descendants = chain.descendants || [];
      const lineageText = lineage.map((t) => t.id || "n/a").join(" -> ");
      const descText = descendants.slice(0, 8).map((t) => t.id || "n/a").join(", ");
      document.getElementById("replay-chain").textContent =
        `Replay chain: lineage=[${lineageText || "n/a"}] descendants(${descendants.length})=[${descText}]`;
    })
    .catch(() => {
      document.getElementById("replay-chain").textContent = "Replay chain: n/a";
    });
}

function loadTaskDebugProm(taskId) {
  if (!taskId) return;
  const scope = debugScope.value || "all";
  fetch(`/task-debug-prom?id=${encodeURIComponent(taskId)}&scope=${encodeURIComponent(scope)}`)
    .then((r) => {
      if (!r.ok) throw new Error("bad status");
      return r.text();
    })
    .then((text) => {
      const lines = text.split("\n").filter((l) => l && !l.startsWith("#")).slice(0, 12);
      document.getElementById("debug-prom").textContent =
        `Task debug metrics (prom) [scope=${scope}]` + "\n" + (lines.join("\n") || "n/a");
    })
    .catch(() => {
      document.getElementById("debug-prom").textContent = "Task debug metrics (prom): n/a";
    });
}

function loadDebugCompare(id1, id2) {
  if (!id1 || !id2) return;
  fetch(`/debug-compare?id1=${encodeURIComponent(id1)}&id2=${encodeURIComponent(id2)}`)
    .then((r) => {
      if (!r.ok) throw new Error("bad status");
      return r.json();
    })
    .then((d) => {
      const m = d.metrics || {};
      const qd = typeof m.quality_delta === "number" ? m.quality_delta.toFixed(3) : "n/a";
      const cd = typeof m.confidence_delta === "number" ? m.confidence_delta.toFixed(3) : "n/a";
      document.getElementById("debug-compare").textContent =
        `Debug compare: ${d.id1 || "n/a"} vs ${d.id2 || "n/a"} | quality_delta=${qd} confidence_delta=${cd} merge=(${m.merge_source_1 || "n/a"} vs ${m.merge_source_2 || "n/a"})`;
    })
    .catch(() => {
      document.getElementById("debug-compare").textContent = "Debug compare: n/a";
    });
  fetch(`/debug-compare?format=prom&id1=${encodeURIComponent(id1)}&id2=${encodeURIComponent(id2)}`)
    .then((r) => {
      if (!r.ok) throw new Error("bad status");
      return r.text();
    })
    .then((text) => {
      const lines = text.split("\n").filter((l) => l && !l.startsWith("#")).slice(0, 8);
      document.getElementById("debug-compare-prom").textContent =
        "Debug compare scrape preview (prom)\n" + (lines.join("\n") || "n/a");
    })
    .catch(() => {
      document.getElementById("debug-compare-prom").textContent = "Debug compare scrape preview (prom): n/a";
    });
}

function updateDashboard() {
  fetch("/dashboard-summary")
    .then((r) => {
      if (!r.ok) throw new Error("bad status");
      return r.json();
    })
    .then((d) => {
      const orch = d.orchestrator || {};
      const tasks = orch.tasks || {};
      const trace = orch.trace || {};
      const byState = trace.by_state || {};
      const byMerge = trace.by_merge_source || {};
      const queue = orch.queue_depth ?? "n/a";
      const replayed = tasks.replayed ?? 0;
      const parentLinks = orch.trace_parent_links_total ?? 0;
      const completed = byState.completed ?? 0;
      const policyMerge = byMerge.policy_merge ?? 0;
      const agentMerge = byMerge.agent_compiler ?? 0;
      document.getElementById("dashboard").textContent =
        `Dashboard summary: queue=${queue} replayed=${replayed} parent_links=${parentLinks} trace_completed=${completed} merge(policy=${policyMerge}, agent=${agentMerge})`;
    })
    .catch(() => {
      document.getElementById("dashboard").textContent = "Dashboard summary: n/a";
    });
  fetch("/dashboard-web6")
    .then((r) => {
      if (!r.ok) throw new Error("bad status");
      return r.json();
    })
    .then((d) => {
      const web6 = d.metrics || {};
      const up = web6.up ? "up" : "down";
      const alertLevel = (typeof web6.proxy_alert_level === "number") ? web6.proxy_alert_level : "n/a";
      const jsonStale = web6.proxy_json_stale ? "stale" : "fresh";
      const promStale = web6.proxy_prom_stale ? "stale" : "fresh";
      const jsonAge = (typeof web6.proxy_json_age_sec === "number") ? `${web6.proxy_json_age_sec}s` : "n/a";
      const promAge = (typeof web6.proxy_prom_age_sec === "number") ? `${web6.proxy_prom_age_sec}s` : "n/a";
      document.getElementById("dashboard-web6").textContent =
        `Orchestrator downstream web6: ${up} alert=${alertLevel} json=${jsonStale}(${jsonAge}) prom=${promStale}(${promAge})`;
    })
    .catch(() => {
      document.getElementById("dashboard-web6").textContent = "Orchestrator downstream web6: n/a";
    });
  fetch("/dashboard-summary?format=prom")
    .then((r) => {
      if (!r.ok) throw new Error("bad status");
      return r.text();
    })
    .then((text) => {
      lastDashboardProm = text || "";
      const lines = text.split("\n").filter((l) => l && !l.startsWith("#")).slice(0, 10);
      document.getElementById("dashboard-prom").textContent =
        "Dashboard scrape preview (prom)\n" + (lines.join("\n") || "n/a");
    })
    .catch(() => {
      lastDashboardProm = "";
      document.getElementById("dashboard-prom").textContent = "Dashboard scrape preview (prom): n/a";
    });
  fetch("/dashboard-web6?format=prom")
    .then((r) => {
      if (!r.ok) throw new Error("bad status");
      return r.text();
    })
    .then((text) => {
      const lines = text.split("\n").filter((l) => l).slice(0, 10);
      document.getElementById("dashboard-web6-prom").textContent =
        "Orchestrator web6 metrics (prom)\n" + (lines.join("\n") || "n/a");
    })
    .catch(() => {
      document.getElementById("dashboard-web6-prom").textContent = "Orchestrator web6 metrics (prom): n/a";
    });
  fetch("/dashboard-web6/alerts")
    .then((r) => {
      if (!r.ok) throw new Error("bad status");
      return r.json();
    })
    .then((d) => {
      const level = d.level || "unknown";
      const score = (typeof d.score === "number") ? d.score : "n/a";
      const reason = d.reason || "n/a";
      document.getElementById("dashboard-web6-alerts").textContent =
        `Orchestrator web6 alerts: level=${level} score=${score} reason=${reason}`;
    })
    .catch(() => {
      document.getElementById("dashboard-web6-alerts").textContent = "Orchestrator web6 alerts: n/a";
    });
  fetch("/dashboard-web6/summary")
    .then((r) => {
      if (!r.ok) throw new Error("bad status");
      return r.json();
    })
    .then((d) => {
      const alert = d.alert || {};
      const metrics = d.metrics || {};
      const level = alert.level || "unknown";
      const score = (typeof alert.score === "number") ? alert.score : "n/a";
      const up = metrics.up ? "up" : "down";
      const jsonAge = (typeof metrics.proxy_json_age_sec === "number") ? `${metrics.proxy_json_age_sec}s` : "n/a";
      const promAge = (typeof metrics.proxy_prom_age_sec === "number") ? `${metrics.proxy_prom_age_sec}s` : "n/a";
      document.getElementById("dashboard-web6-summary").textContent =
        `Orchestrator web6 summary: ${up} level=${level} score=${score} json_age=${jsonAge} prom_age=${promAge}`;
    })
    .catch(() => {
      document.getElementById("dashboard-web6-summary").textContent = "Orchestrator web6 summary: n/a";
    });
  fetch("/dashboard-web6/summary?format=prom")
    .then((r) => {
      if (!r.ok) throw new Error("bad status");
      return r.text();
    })
    .then((text) => {
      lastDashboardWeb6SummaryProm = text || "";
    })
    .catch(() => {
      lastDashboardWeb6SummaryProm = "";
    });
  const historyLimit = dashboardWeb6HistoryLimit.value || "8";
  const historyLevel = dashboardWeb6HistoryLevel.value || "all";
  const historySource = dashboardWeb6HistorySource.value || "all";
  const historyQS = `limit=${encodeURIComponent(historyLimit)}&level=${encodeURIComponent(historyLevel)}&source=${encodeURIComponent(historySource)}`;
  fetch(`/dashboard-web6/history?${historyQS}`)
    .then((r) => {
      if (!r.ok) throw new Error("bad status");
      return r.json();
    })
    .then((d) => {
      const levelCounts = d.level_counts || {};
      const sourceCounts = d.source_counts || {};
      const count = d.count || 0;
      const ok = levelCounts.ok || 0;
      const warn = levelCounts.warn || 0;
      const critical = levelCounts.critical || 0;
      const summary = `count=${count} levels(ok=${ok},warn=${warn},critical=${critical}) filters(level=${historyLevel},source=${historySource})`;
      const sources = `sources(alerts=${sourceCounts.alerts || 0},summary=${sourceCounts.summary || 0})`;
      const items = d.items || [];
      const last = items.length ? items[items.length - 1] : null;
      const lastText = last ? `last(level=${last.level || "n/a"},source=${last.source || "n/a"})` : "last(n/a)";
      document.getElementById("dashboard-web6-history").textContent =
        `Orchestrator web6 alert history: ${summary} ${sources} ${lastText}`;
    })
    .catch(() => {
      document.getElementById("dashboard-web6-history").textContent = "Orchestrator web6 alert history: n/a";
    });
  fetch(`/dashboard-web6/history?${historyQS}&format=prom`)
    .then((r) => {
      if (!r.ok) throw new Error("bad status");
      return r.text();
    })
    .then((text) => {
      lastDashboardWeb6HistoryProm = text || "";
    })
    .catch(() => {
      lastDashboardWeb6HistoryProm = "";
    });
}

function updateProxyMetrics() {
  fetch("/proxy-counters")
    .then((r) => {
      if (!r.ok) throw new Error("bad status");
      return r.json();
    })
    .then((d) => {
      const c1 = d.debug_compare_total || 0;
      const c2 = d.debug_compare_prom_total || 0;
      const c3 = d.proxy_counters_total || 0;
      const c4 = d.proxy_counters_prom_total || 0;
      const c5 = d.proxy_alerts_total || 0;
      const c6 = d.proxy_alerts_prom_total || 0;
      const c7 = d.dashboard_summary_total || 0;
      const c8 = d.dashboard_summary_prom_total || 0;
      const t1 = d.proxy_last_json_unix || 0;
      const t2 = d.proxy_last_prom_unix || 0;
      const now = Math.floor(Date.now() / 1000);
      const a1 = t1 > 0 ? `${Math.max(0, now - t1)}s` : "n/a";
      const a2 = t2 > 0 ? `${Math.max(0, now - t2)}s` : "n/a";
      document.getElementById("proxy-metrics").textContent =
        `Proxy counters: debug_compare=${c1} debug_compare_prom=${c2} proxy_json=${c3} proxy_prom=${c4} proxy_alerts=${c5} proxy_alerts_prom=${c6} dashboard_summary=${c7} dashboard_summary_prom=${c8} age(json=${a1},prom=${a2})`;
    })
    .catch(() => {
      document.getElementById("proxy-metrics").textContent = "Proxy counters: n/a";
    });
  fetch("/proxy-counters?format=prom")
    .then((r) => {
      if (!r.ok) throw new Error("bad status");
      return r.text();
    })
    .then((text) => {
      lastProxyProm = text || "";
      const lines = text.split("\n").filter((l) => l && !l.startsWith("#")).slice(0, 8);
      document.getElementById("proxy-prom").textContent =
        "Proxy counters scrape preview (prom)\n" + (lines.join("\n") || "n/a");
    })
    .catch(() => {
      lastProxyProm = "";
      document.getElementById("proxy-prom").textContent = "Proxy counters scrape preview (prom): n/a";
    });
  const proxyLimit = proxyHistoryLimit.value || "20";
  const proxyFormat = proxyHistoryFormat.value || "all";
  const proxyQS = `limit=${encodeURIComponent(proxyLimit)}&format_filter=${encodeURIComponent(proxyFormat)}`;
  fetch(`/proxy-counters/history?${proxyQS}`)
    .then((r) => {
      if (!r.ok) throw new Error("bad status");
      return r.json();
    })
    .then((d) => {
      const count = d.count || 0;
      const fc = d.format_counts || {};
      const last = d.last || {};
      const lastFmt = last.format || "n/a";
      const pj = (typeof last.proxy_counters_total === "number") ? last.proxy_counters_total : "n/a";
      const pp = (typeof last.proxy_counters_prom_total === "number") ? last.proxy_counters_prom_total : "n/a";
      document.getElementById("proxy-history").textContent =
        `Proxy history: count=${count} formats(json=${fc.json || 0},prom=${fc.prom || 0}) filter=${proxyFormat} last(format=${lastFmt},proxy_json=${pj},proxy_prom=${pp})`;
    })
    .catch(() => {
      document.getElementById("proxy-history").textContent = "Proxy history: n/a";
    });
  fetch(`/proxy-counters/history?${proxyQS}&format=prom`)
    .then((r) => {
      if (!r.ok) throw new Error("bad status");
      return r.text();
    })
    .then((text) => {
      lastProxyHistoryProm = text || "";
    })
    .catch(() => {
      lastProxyHistoryProm = "";
    });
}

function updateProxyHealth() {
  fetch("/proxy-counters/health")
    .then((r) => {
      if (!r.ok) throw new Error("bad status");
      return r.json();
    })
    .then((d) => {
      const jAge = d.proxy_json_age_sec;
      const pAge = d.proxy_prom_age_sec;
      const jStale = d.proxy_json_stale ? "stale" : "fresh";
      const pStale = d.proxy_prom_stale ? "stale" : "fresh";
      const thr = d.stale_threshold_sec || 0;
      document.getElementById("proxy-health").textContent =
        `Proxy health: json=${jStale}(${jAge}s) prom=${pStale}(${pAge}s) threshold=${thr}s`;
    })
    .catch(() => {
      document.getElementById("proxy-health").textContent = "Proxy health: n/a";
    });
}

function updateProxyAlerts() {
  fetch("/proxy-counters/alerts")
    .then((r) => {
      if (!r.ok) throw new Error("bad status");
      return r.json();
    })
    .then((d) => {
      const lvl = d.level || "unknown";
      const score = (typeof d.level_score === "number") ? d.level_score : "n/a";
      const reason = d.reason || "n/a";
      document.getElementById("proxy-alerts").textContent =
        `Proxy alerts: level=${lvl} score=${score} reason=${reason}`;
    })
    .catch(() => {
      document.getElementById("proxy-alerts").textContent = "Proxy alerts: n/a";
    });
}

function updateTasks() {
  const qs = new URLSearchParams();
  qs.set("limit", "8");
  if (taskState.value && taskState.value !== "all") qs.set("state", taskState.value);
  if (taskMerge.value && taskMerge.value !== "all") qs.set("merge_source", taskMerge.value);
  if (taskParent.value && taskParent.value !== "all") qs.set("has_parent", taskParent.value);
  if (taskSort.value) qs.set("sort", taskSort.value);
  fetch("/tasks/recent?" + qs.toString())
    .then((r) => {
      if (!r.ok) throw new Error("bad status");
      return r.json();
    })
    .then((payload) => {
      const tasks = payload.tasks || [];
      if (!tasks.length) {
        document.getElementById("tasks").textContent = "Task explorer: no tasks";
        return;
      }
      const rows = tasks.map((t) => {
        const id = t.id || "n/a";
        const state = t.state || "unknown";
        const source = t.merge_source || "none";
        const quality = (typeof t.quality_score === "number") ? t.quality_score.toFixed(3) : "n/a";
        const parent = t.parent_task_id ? ` parent=${t.parent_task_id}` : "";
        const cls = selectedTaskId === id ? "task-item active" : "task-item";
        return `<div class="${cls}" data-task-id="${id}"><strong>${id}</strong><span>state=${state}</span><span>merge=${source}</span><span>quality=${quality}</span><span>${parent}</span></div>`;
      }).join("");
      document.getElementById("tasks").innerHTML = `<div><strong>Task explorer (latest 8)</strong></div>${rows}`;
      const taskEls = document.querySelectorAll("#tasks .task-item[data-task-id]");
      taskEls.forEach((el) => {
        el.addEventListener("click", () => {
          selectedTaskId = el.dataset.taskId || "";
          updateTasks();
          loadTaskTrace(selectedTaskId);
          loadReplayChain(selectedTaskId);
          loadTaskDebugProm(selectedTaskId);
        });
      });
      if (!selectedTaskId && tasks.length > 0 && tasks[0].id) {
        selectedTaskId = tasks[0].id;
        loadTaskTrace(selectedTaskId);
        loadReplayChain(selectedTaskId);
        loadTaskDebugProm(selectedTaskId);
      }
    })
    .catch(() => {
      document.getElementById("tasks").textContent = "Task explorer: n/a";
    });
}
taskState.addEventListener("change", () => {
  localStorage.setItem("web6_task_state", taskState.value);
  updateTasks();
});
taskMerge.addEventListener("change", () => {
  localStorage.setItem("web6_task_merge", taskMerge.value);
  updateTasks();
});
taskParent.addEventListener("change", () => {
  localStorage.setItem("web6_task_parent", taskParent.value);
  updateTasks();
});
taskSort.addEventListener("change", () => {
  localStorage.setItem("web6_task_sort", taskSort.value);
  updateTasks();
});
replayMode.addEventListener("change", () => {
  localStorage.setItem("web6_replay_mode", replayMode.value);
});
debugScope.addEventListener("change", () => {
  localStorage.setItem("web6_debug_scope", debugScope.value);
  if (selectedTaskId) loadTaskDebugProm(selectedTaskId);
});
replayBtn.addEventListener("click", () => {
  if (!selectedTaskId) {
    return;
  }
  const mode = replayMode.value || "force-policy";
  fetch(`/task-replay?id=${encodeURIComponent(selectedTaskId)}&mode=${encodeURIComponent(mode)}`, { method: "POST" })
    .then((r) => {
      if (!r.ok) throw new Error("bad status");
      return r.json();
    })
    .then((_res) => {
      updateTasks();
      updateTrace();
      updateDashboard();
    })
    .catch(() => {});
});
quickFailedBtn.addEventListener("click", () => {
  taskState.value = "failed";
  taskMerge.value = "all";
  taskParent.value = "all";
  updateTasks();
});
quickReplayBtn.addEventListener("click", () => {
  taskState.value = "all";
  taskParent.value = "yes";
  updateTasks();
});
quickSoftBtn.addEventListener("click", () => {
  taskState.value = "completed";
  taskMerge.value = "policy_merge";
  taskParent.value = "yes";
  updateTasks();
});
compareBtn.addEventListener("click", () => {
  const id1 = selectedTaskId;
  const id2 = compareIdInput.value.trim();
  if (!id1 || !id2) return;
  loadDebugCompare(id1, id2);
});
debugRefreshBtn.addEventListener("click", () => {
  if (selectedTaskId) loadTaskDebugProm(selectedTaskId);
});
debugCopyBtn.addEventListener("click", async () => {
  const text = document.getElementById("debug-prom").textContent || "";
  try {
    await navigator.clipboard.writeText(text);
    debugCopyBtn.textContent = "copied";
  } catch (_e) {
    debugCopyBtn.textContent = "failed";
  }
  setTimeout(() => { debugCopyBtn.textContent = "copy debug"; }, 1200);
});
proxyRefreshBtn.addEventListener("click", () => {
  updateProxyMetrics();
  updateProxyHealth();
  updateProxyAlerts();
});
proxyCopyBtn.addEventListener("click", async () => {
  const text = lastProxyProm || "";
  try {
    await navigator.clipboard.writeText(text);
    proxyCopyBtn.textContent = "copied";
  } catch (_e) {
    proxyCopyBtn.textContent = "failed";
  }
  setTimeout(() => { proxyCopyBtn.textContent = "copy proxy prom"; }, 1200);
});
proxyHistoryCopyBtn.addEventListener("click", async () => {
  const text = lastProxyHistoryProm || "";
  try {
    await navigator.clipboard.writeText(text);
    proxyHistoryCopyBtn.textContent = "copied";
  } catch (_e) {
    proxyHistoryCopyBtn.textContent = "failed";
  }
  setTimeout(() => { proxyHistoryCopyBtn.textContent = "copy proxy history prom"; }, 1200);
});
proxyHistoryLimit.addEventListener("change", () => {
  localStorage.setItem("web6_proxy_history_limit", proxyHistoryLimit.value);
  updateProxyMetrics();
});
proxyHistoryFormat.addEventListener("change", () => {
  localStorage.setItem("web6_proxy_history_format", proxyHistoryFormat.value);
  updateProxyMetrics();
});
proxyHistoryResetBtn.addEventListener("click", () => {
  fetch("/proxy-counters/history", { method: "POST" })
    .then((r) => {
      if (!r.ok) throw new Error("bad status");
      return r.json();
    })
    .then(() => {
      updateProxyMetrics();
      proxyHistoryResetBtn.textContent = "history reset";
      setTimeout(() => { proxyHistoryResetBtn.textContent = "reset proxy history"; }, 1200);
    })
    .catch(() => {
      proxyHistoryResetBtn.textContent = "reset failed";
      setTimeout(() => { proxyHistoryResetBtn.textContent = "reset proxy history"; }, 1200);
    });
});
dashboardCopyBtn.addEventListener("click", async () => {
  const text = lastDashboardProm || "";
  try {
    await navigator.clipboard.writeText(text);
    dashboardCopyBtn.textContent = "copied";
  } catch (_e) {
    dashboardCopyBtn.textContent = "failed";
  }
  setTimeout(() => { dashboardCopyBtn.textContent = "copy dashboard prom"; }, 1200);
});
dashboardWeb6CopyBtn.addEventListener("click", async () => {
  const text = lastDashboardWeb6SummaryProm || "";
  try {
    await navigator.clipboard.writeText(text);
    dashboardWeb6CopyBtn.textContent = "copied";
  } catch (_e) {
    dashboardWeb6CopyBtn.textContent = "failed";
  }
  setTimeout(() => { dashboardWeb6CopyBtn.textContent = "copy web6 summary prom"; }, 1200);
});
dashboardWeb6HistoryCopyBtn.addEventListener("click", async () => {
  const text = lastDashboardWeb6HistoryProm || "";
  try {
    await navigator.clipboard.writeText(text);
    dashboardWeb6HistoryCopyBtn.textContent = "copied";
  } catch (_e) {
    dashboardWeb6HistoryCopyBtn.textContent = "failed";
  }
  setTimeout(() => { dashboardWeb6HistoryCopyBtn.textContent = "copy web6 history prom"; }, 1200);
});
dashboardWeb6HistoryLimit.addEventListener("change", () => {
  localStorage.setItem("web6_dashboard_web6_history_limit", dashboardWeb6HistoryLimit.value);
  updateDashboard();
});
dashboardWeb6HistoryLevel.addEventListener("change", () => {
  localStorage.setItem("web6_dashboard_web6_history_level", dashboardWeb6HistoryLevel.value);
  updateDashboard();
});
dashboardWeb6HistorySource.addEventListener("change", () => {
  localStorage.setItem("web6_dashboard_web6_history_source", dashboardWeb6HistorySource.value);
  updateDashboard();
});
dashboardWeb6HistoryResetBtn.addEventListener("click", () => {
  fetch("/dashboard-web6/history", { method: "POST" })
    .then((r) => {
      if (!r.ok) throw new Error("bad status");
      return r.json();
    })
    .then(() => {
      updateDashboard();
      dashboardWeb6HistoryResetBtn.textContent = "history reset";
      setTimeout(() => { dashboardWeb6HistoryResetBtn.textContent = "reset web6 history"; }, 1200);
    })
    .catch(() => {
      dashboardWeb6HistoryResetBtn.textContent = "reset failed";
      setTimeout(() => { dashboardWeb6HistoryResetBtn.textContent = "reset web6 history"; }, 1200);
    });
});

function syncUrl(includeQ) {
  const p = new URLSearchParams();
  const currentQ = new URLSearchParams(location.search).get("q");
  if (includeQ && query) p.set("q", query);
  else if (currentQ) p.set("q", currentQ);
  p.set("mode", mode.value);
  p.set("depth", depth.value);
  p.set("type", typeSelect.value);
  history.replaceState(null, "", "?" + p.toString());
}

function loadGraph() {
  query = search.value.trim();
  const q = encodeURIComponent(query);
  const m = encodeURIComponent(mode.value);
  const d = encodeURIComponent(depth.value);
  const tsel = encodeURIComponent(typeSelect.value);
  indicator.textContent = `mode: ${mode.value} depth: ${depth.value} type: ${typeSelect.value}`;
  const graphURL = q ? `/graph?q=${q}&mode=${m}&depth=${d}&type=${tsel}` : `/graph?mode=${m}&depth=${d}&type=${tsel}`;
  fetch(graphURL)
    .then((r) => r.json())
    .then((g) => {
      nodes = g.nodes.map((n, i) => ({
        id: n.id,
        label: n.label,
        x: (i * 120) - 120,
        y: (i % 2 ? 60 : -60),
        z: 0,
        type: n.type || "unknown",
      }));
      edges = g.edges;
    });
}

search.addEventListener("input", () => {
  localStorage.setItem("web6_q", search.value);
  loadGraph();
});
search.addEventListener("keydown", (e) => {
  if (e.key === "Enter") {
    query = search.value.trim();
    syncUrl(true);
  }
});
mode.addEventListener("change", () => {
  localStorage.setItem("web6_mode", mode.value);
  loadGraph();
  syncUrl(false);
});
depth.addEventListener("change", () => {
  localStorage.setItem("web6_depth", depth.value);
  loadGraph();
  syncUrl(false);
});
typeSelect.addEventListener("change", () => {
  localStorage.setItem("web6_type", typeSelect.value);
  loadGraph();
  syncUrl(false);
});
refreshBtn.addEventListener("click", () => loadGraph());
clearBtn.addEventListener("click", () => {
  fetch("/cache/clear", { method: "POST" }).then(() => loadGraph());
});

loadGraph();
tick();
updateQueue();
updateCache();
updateTrace();
updateTasks();
updateDashboard();
updateProxyMetrics();
updateProxyHealth();
updateProxyAlerts();
setInterval(updateQueue, 3000);
setInterval(updateCache, 5000);
setInterval(updateTrace, 5000);
setInterval(updateTasks, 5000);
setInterval(updateDashboard, 5000);
setInterval(updateProxyMetrics, 5000);
setInterval(updateProxyHealth, 5000);
setInterval(updateProxyAlerts, 5000);
