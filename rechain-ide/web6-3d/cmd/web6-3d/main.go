package main

import (
	_ "embed"
	"encoding/json"
	"errors"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"rechain-ide/shared/logging"
)

type Graph struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

type Node struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Type  string `json:"type"`
}

type Edge struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type Metrics struct {
	mu                   sync.Mutex
	graph                int
	root                 int
	health               int
	debugCompare         int
	debugCompareProm     int
	proxyCounters        int
	proxyCountersProm    int
	proxyAlerts          int
	proxyAlertsProm      int
	proxyLastJSONUnix    int64
	proxyLastPromUnix    int64
	dashboardSummary     int
	dashboardSummaryProm int
	filterMatches        int
	filterEdges          int
	depthCounts          map[string]int
	typeCounts           map[string]int
	typeFilter           string
}

type DashboardWeb6AlertEvent struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Score     int    `json:"score"`
	Reason    string `json:"reason"`
	Source    string `json:"source"`
}

type ProxyCountersEvent struct {
	Timestamp            string `json:"timestamp"`
	Format               string `json:"format"`
	DebugCompareTotal    int    `json:"debug_compare_total"`
	DebugCompareProm     int    `json:"debug_compare_prom_total"`
	ProxyCountersTotal   int    `json:"proxy_counters_total"`
	ProxyCountersProm    int    `json:"proxy_counters_prom_total"`
	ProxyAlertsTotal     int    `json:"proxy_alerts_total"`
	ProxyAlertsPromTotal int    `json:"proxy_alerts_prom_total"`
	DashboardSummary     int    `json:"dashboard_summary_total"`
	DashboardSummaryProm int    `json:"dashboard_summary_prom_total"`
	ProxyLastJSONUnix    int64  `json:"proxy_last_json_unix"`
	ProxyLastPromUnix    int64  `json:"proxy_last_prom_unix"`
}

type ProxyCountersHistory struct {
	mu    sync.Mutex
	items []ProxyCountersEvent
	max   int
}

func NewProxyCountersHistory(max int) *ProxyCountersHistory {
	if max <= 0 {
		max = 200
	}
	return &ProxyCountersHistory{max: max}
}

func (h *ProxyCountersHistory) Add(ev ProxyCountersEvent) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.items = append(h.items, ev)
	if len(h.items) > h.max {
		h.items = h.items[len(h.items)-h.max:]
	}
}

func (h *ProxyCountersHistory) List(limit int) []ProxyCountersEvent {
	h.mu.Lock()
	defer h.mu.Unlock()
	if limit <= 0 || limit > len(h.items) {
		limit = len(h.items)
	}
	start := len(h.items) - limit
	out := make([]ProxyCountersEvent, limit)
	copy(out, h.items[start:])
	return out
}

func (h *ProxyCountersHistory) Reset() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.items = nil
}

type DashboardWeb6AlertHistory struct {
	mu    sync.Mutex
	items []DashboardWeb6AlertEvent
	max   int
}

func NewDashboardWeb6AlertHistory(max int) *DashboardWeb6AlertHistory {
	if max <= 0 {
		max = 100
	}
	return &DashboardWeb6AlertHistory{max: max}
}

func (h *DashboardWeb6AlertHistory) Add(ev DashboardWeb6AlertEvent) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.items = append(h.items, ev)
	if len(h.items) > h.max {
		h.items = h.items[len(h.items)-h.max:]
	}
}

func (h *DashboardWeb6AlertHistory) List(limit int) []DashboardWeb6AlertEvent {
	h.mu.Lock()
	defer h.mu.Unlock()
	if limit <= 0 || limit > len(h.items) {
		limit = len(h.items)
	}
	start := len(h.items) - limit
	out := make([]DashboardWeb6AlertEvent, limit)
	copy(out, h.items[start:])
	return out
}

func (h *DashboardWeb6AlertHistory) Reset() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.items = nil
}

func (m *Metrics) IncGraph()            { m.mu.Lock(); m.graph++; m.mu.Unlock() }
func (m *Metrics) IncRoot()             { m.mu.Lock(); m.root++; m.mu.Unlock() }
func (m *Metrics) IncHealth()           { m.mu.Lock(); m.health++; m.mu.Unlock() }
func (m *Metrics) IncDebugCompare()     { m.mu.Lock(); m.debugCompare++; m.mu.Unlock() }
func (m *Metrics) IncDebugCompareProm() { m.mu.Lock(); m.debugCompareProm++; m.mu.Unlock() }
func (m *Metrics) TouchProxyCounters() {
	m.mu.Lock()
	m.proxyCounters++
	m.proxyLastJSONUnix = time.Now().UTC().Unix()
	m.mu.Unlock()
}
func (m *Metrics) TouchProxyCountersProm() {
	m.mu.Lock()
	m.proxyCountersProm++
	m.proxyLastPromUnix = time.Now().UTC().Unix()
	m.mu.Unlock()
}
func (m *Metrics) IncProxyAlerts()          { m.mu.Lock(); m.proxyAlerts++; m.mu.Unlock() }
func (m *Metrics) IncProxyAlertsProm()      { m.mu.Lock(); m.proxyAlertsProm++; m.mu.Unlock() }
func (m *Metrics) IncDashboardSummary()     { m.mu.Lock(); m.dashboardSummary++; m.mu.Unlock() }
func (m *Metrics) IncDashboardSummaryProm() { m.mu.Lock(); m.dashboardSummaryProm++; m.mu.Unlock() }

func (m *Metrics) Snapshot() map[string]int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return map[string]int{
		"graph":                  m.graph,
		"root":                   m.root,
		"health":                 m.health,
		"debug_compare":          m.debugCompare,
		"debug_compare_prom":     m.debugCompareProm,
		"proxy_counters":         m.proxyCounters,
		"proxy_counters_prom":    m.proxyCountersProm,
		"proxy_alerts":           m.proxyAlerts,
		"proxy_alerts_prom":      m.proxyAlertsProm,
		"dashboard_summary":      m.dashboardSummary,
		"dashboard_summary_prom": m.dashboardSummaryProm,
	}
}

func (m *Metrics) ProxyMetaSnapshot() map[string]int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return map[string]int64{
		"proxy_last_json_unix": m.proxyLastJSONUnix,
		"proxy_last_prom_unix": m.proxyLastPromUnix,
	}
}

func (m *Metrics) RecordFilter(matches int, edges int) {
	m.mu.Lock()
	m.filterMatches = matches
	m.filterEdges = edges
	m.mu.Unlock()
}

func (m *Metrics) FilterSnapshot() map[string]int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return map[string]int{
		"matches": m.filterMatches,
		"edges":   m.filterEdges,
	}
}

func (m *Metrics) IncDepth(depth int) {
	key := strconv.Itoa(depth)
	m.mu.Lock()
	if m.depthCounts == nil {
		m.depthCounts = map[string]int{}
	}
	m.depthCounts[key]++
	m.mu.Unlock()
}

func (m *Metrics) DepthSnapshot() map[string]int {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := map[string]int{}
	for k, v := range m.depthCounts {
		out[k] = v
	}
	return out
}

func (m *Metrics) RecordTypes(nodes []Node) {
	counts := countTypes(nodes)
	m.mu.Lock()
	m.typeCounts = counts
	m.mu.Unlock()
}

func (m *Metrics) TypeSnapshot() map[string]int {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := map[string]int{}
	for k, v := range m.typeCounts {
		out[k] = v
	}
	return out
}

func (m *Metrics) SetTypeFilter(t string) {
	m.mu.Lock()
	m.typeFilter = t
	m.mu.Unlock()
}

func (m *Metrics) TypeFilter() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.typeFilter
}

func main() {
	mux := http.NewServeMux()
	metrics := &Metrics{}
	orchURL := strings.TrimRight(envOr("ORCH_URL", "http://localhost:8081"), "/")
	graphPath := os.Getenv("WEB6_GRAPH_PATH")
	graphRoot := envOr("WEB6_ROOT", ".")
	maxNodes := envInt("WEB6_MAX_NODES", 2000)
	ttl := time.Duration(envInt("WEB6_GRAPH_TTL_SEC", 5)) * time.Second
	proxyStaleSec := envInt("WEB6_PROXY_STALE_SEC", 30)
	proxyCriticalSec := envInt("WEB6_PROXY_CRITICAL_SEC", proxyStaleSec*3)
	proxyHistory := NewProxyCountersHistory(envInt("WEB6_PROXY_HISTORY_MAX", 200))
	alertHistory := NewDashboardWeb6AlertHistory(envInt("WEB6_DASHBOARD_WEB6_HISTORY_MAX", 100))
	defaultDepth := envInt("WEB6_GRAPH_DEPTH_DEFAULT", 1)
	mode := envOr("WEB6_GRAPH_MODE", "tree")
	cache := &GraphCache{}
	goListCache := &GoListCache{}
	goListMetrics := &GoListMetrics{}

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		metrics.IncHealth()
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	mux.HandleFunc("/graph", func(w http.ResponseWriter, r *http.Request) {
		metrics.IncGraph()
		reqMode := strings.TrimSpace(r.URL.Query().Get("mode"))
		if reqMode == "" {
			reqMode = mode
		}
		g := loadGraph(graphPath, graphRoot, maxNodes, ttl, cache, reqMode, goListCache, goListMetrics)
		q := strings.TrimSpace(r.URL.Query().Get("q"))
		depth := constraintIntFromQuery(r, "depth", defaultDepth)
		metrics.IncDepth(depth)
		if q != "" {
			g = filterGraphContext(g, q, depth, metrics)
		}
		typeFilter := strings.TrimSpace(r.URL.Query().Get("type"))
		if typeFilter == "" {
			typeFilter = "all"
		}
		if typeFilter != "all" {
			g = filterGraphByType(g, typeFilter)
		}
		metrics.SetTypeFilter(typeFilter)
		metrics.RecordTypes(g.Nodes)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(g)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ui/style.css" {
			w.Header().Set("Content-Type", "text/css; charset=utf-8")
			w.Write([]byte(styleCSSTemplate))
			return
		}
		if r.URL.Path == "/ui/app.js" {
			w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
			w.Write([]byte(appJSTemplate))
			return
		}
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		metrics.IncRoot()
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(renderIndexHTML()))
	})

	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		snap := metrics.Snapshot()
		snapFilter := metrics.FilterSnapshot()
		depthSnap := metrics.DepthSnapshot()
		typeSnap := metrics.TypeSnapshot()
		typeFilter := metrics.TypeFilter()
		meta := metrics.ProxyMetaSnapshot()
		goSnap := goListMetrics.Snapshot()
		lines := []string{
			"# HELP rechain_web6_graph_total Graph requests",
			"# TYPE rechain_web6_graph_total counter",
			"rechain_web6_graph_total " + strconv.Itoa(snap["graph"]),
			"# HELP rechain_web6_root_total Root page requests",
			"# TYPE rechain_web6_root_total counter",
			"rechain_web6_root_total " + strconv.Itoa(snap["root"]),
			"# HELP rechain_web6_health_total Health requests",
			"# TYPE rechain_web6_health_total counter",
			"rechain_web6_health_total " + strconv.Itoa(snap["health"]),
			"# HELP rechain_web6_debug_compare_total Debug compare JSON requests",
			"# TYPE rechain_web6_debug_compare_total counter",
			"rechain_web6_debug_compare_total " + strconv.Itoa(snap["debug_compare"]),
			"# HELP rechain_web6_debug_compare_prom_total Debug compare Prometheus requests",
			"# TYPE rechain_web6_debug_compare_prom_total counter",
			"rechain_web6_debug_compare_prom_total " + strconv.Itoa(snap["debug_compare_prom"]),
			"# HELP rechain_web6_proxy_counters_total Proxy counters JSON requests",
			"# TYPE rechain_web6_proxy_counters_total counter",
			"rechain_web6_proxy_counters_total " + strconv.Itoa(snap["proxy_counters"]),
			"# HELP rechain_web6_proxy_counters_prom_total Proxy counters Prometheus requests",
			"# TYPE rechain_web6_proxy_counters_prom_total counter",
			"rechain_web6_proxy_counters_prom_total " + strconv.Itoa(snap["proxy_counters_prom"]),
			"# HELP rechain_web6_proxy_alerts_total Proxy alerts JSON requests",
			"# TYPE rechain_web6_proxy_alerts_total counter",
			"rechain_web6_proxy_alerts_total " + strconv.Itoa(snap["proxy_alerts"]),
			"# HELP rechain_web6_proxy_alerts_prom_total Proxy alerts Prometheus requests",
			"# TYPE rechain_web6_proxy_alerts_prom_total counter",
			"rechain_web6_proxy_alerts_prom_total " + strconv.Itoa(snap["proxy_alerts_prom"]),
			"# HELP rechain_web6_proxy_counters_last_json_unix Last proxy-counters JSON request unix timestamp",
			"# TYPE rechain_web6_proxy_counters_last_json_unix gauge",
			"rechain_web6_proxy_counters_last_json_unix " + strconv.FormatInt(meta["proxy_last_json_unix"], 10),
			"# HELP rechain_web6_proxy_counters_last_prom_unix Last proxy-counters Prom request unix timestamp",
			"# TYPE rechain_web6_proxy_counters_last_prom_unix gauge",
			"rechain_web6_proxy_counters_last_prom_unix " + strconv.FormatInt(meta["proxy_last_prom_unix"], 10),
			"# HELP rechain_web6_dashboard_summary_total Dashboard summary JSON requests",
			"# TYPE rechain_web6_dashboard_summary_total counter",
			"rechain_web6_dashboard_summary_total " + strconv.Itoa(snap["dashboard_summary"]),
			"# HELP rechain_web6_dashboard_summary_prom_total Dashboard summary Prometheus requests",
			"# TYPE rechain_web6_dashboard_summary_prom_total counter",
			"rechain_web6_dashboard_summary_prom_total " + strconv.Itoa(snap["dashboard_summary_prom"]),
			"# HELP rechain_web6_filter_matches Filtered match count",
			"# TYPE rechain_web6_filter_matches gauge",
			"rechain_web6_filter_matches " + strconv.Itoa(snapFilter["matches"]),
			"# HELP rechain_web6_filter_edges Filtered edge count",
			"# TYPE rechain_web6_filter_edges gauge",
			"rechain_web6_filter_edges " + strconv.Itoa(snapFilter["edges"]),
			"# HELP rechain_web6_go_list_cache_hits_total Go list cache hits",
			"# TYPE rechain_web6_go_list_cache_hits_total counter",
			"rechain_web6_go_list_cache_hits_total " + strconv.Itoa(goSnap["hits"]),
			"# HELP rechain_web6_go_list_cache_misses_total Go list cache misses",
			"# TYPE rechain_web6_go_list_cache_misses_total counter",
			"rechain_web6_go_list_cache_misses_total " + strconv.Itoa(goSnap["misses"]),
		}
		for depth, v := range depthSnap {
			lines = append(lines,
				"# HELP rechain_web6_filter_depth_total Filter depth usage",
				"# TYPE rechain_web6_filter_depth_total counter",
				"rechain_web6_filter_depth_total{depth=\""+depth+"\"} "+strconv.Itoa(v),
			)
		}
		for t, v := range typeSnap {
			lines = append(lines,
				"# HELP rechain_web6_nodes_total Node counts by type",
				"# TYPE rechain_web6_nodes_total gauge",
				"rechain_web6_nodes_total{type=\""+t+"\"} "+strconv.Itoa(v),
			)
		}
		if typeFilter == "" {
			typeFilter = "all"
		}
		for _, t := range []string{"all", "dir", "file", "pkg", "unknown"} {
			v := 0
			if t == typeFilter {
				v = 1
			}
			lines = append(lines,
				"# HELP rechain_web6_filter_type_active Active type filter",
				"# TYPE rechain_web6_filter_type_active gauge",
				"rechain_web6_filter_type_active{type=\""+t+"\"} "+strconv.Itoa(v),
			)
		}
		w.Write([]byte(strings.Join(lines, "\n")))
	})

	mux.HandleFunc("/proxy-counters", func(w http.ResponseWriter, r *http.Request) {
		isProm := strings.Contains(r.Header.Get("Accept"), "text/plain") || strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("format")), "prom")
		if isProm {
			metrics.TouchProxyCountersProm()
		} else {
			metrics.TouchProxyCounters()
		}
		snap := metrics.Snapshot()
		meta := metrics.ProxyMetaSnapshot()
		payload := map[string]int{
			"debug_compare_total":          snap["debug_compare"],
			"debug_compare_prom_total":     snap["debug_compare_prom"],
			"proxy_counters_total":         snap["proxy_counters"],
			"proxy_counters_prom_total":    snap["proxy_counters_prom"],
			"proxy_alerts_total":           snap["proxy_alerts"],
			"proxy_alerts_prom_total":      snap["proxy_alerts_prom"],
			"dashboard_summary_total":      snap["dashboard_summary"],
			"dashboard_summary_prom_total": snap["dashboard_summary_prom"],
		}
		payloadMeta := map[string]interface{}{
			"proxy_last_json_unix": meta["proxy_last_json_unix"],
			"proxy_last_prom_unix": meta["proxy_last_prom_unix"],
		}
		proxyHistory.Add(ProxyCountersEvent{
			Timestamp:            time.Now().UTC().Format(time.RFC3339),
			Format:               map[bool]string{true: "prom", false: "json"}[isProm],
			DebugCompareTotal:    payload["debug_compare_total"],
			DebugCompareProm:     payload["debug_compare_prom_total"],
			ProxyCountersTotal:   payload["proxy_counters_total"],
			ProxyCountersProm:    payload["proxy_counters_prom_total"],
			ProxyAlertsTotal:     payload["proxy_alerts_total"],
			ProxyAlertsPromTotal: payload["proxy_alerts_prom_total"],
			DashboardSummary:     payload["dashboard_summary_total"],
			DashboardSummaryProm: payload["dashboard_summary_prom_total"],
			ProxyLastJSONUnix:    meta["proxy_last_json_unix"],
			ProxyLastPromUnix:    meta["proxy_last_prom_unix"],
		})
		if isProm {
			w.Header().Set("Content-Type", "text/plain; version=0.0.4")
			w.Write([]byte(buildProxyCountersProm(payload, payloadMeta)))
			return
		}
		out := map[string]interface{}{}
		for k, v := range payload {
			out[k] = v
		}
		for k, v := range payloadMeta {
			out[k] = v
		}
		writeJSON(w, out)
	})

	mux.HandleFunc("/proxy-counters/history", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			proxyHistory.Reset()
			writeJSON(w, map[string]interface{}{"status": "reset"})
			return
		}
		limit := constraintIntFromQuery(r, "limit", 20)
		items := proxyHistory.List(limit)
		formatFilter := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("format_filter")))
		filtered := make([]ProxyCountersEvent, 0, len(items))
		formatCounts := map[string]int{"json": 0, "prom": 0}
		for _, it := range items {
			if formatFilter != "" && formatFilter != "all" && formatFilter != strings.ToLower(strings.TrimSpace(it.Format)) {
				continue
			}
			filtered = append(filtered, it)
			f := strings.ToLower(strings.TrimSpace(it.Format))
			if f == "" {
				f = "json"
			}
			formatCounts[f]++
		}
		if strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("format")), "prom") {
			w.Header().Set("Content-Type", "text/plain; version=0.0.4")
			w.Write([]byte(buildProxyCountersHistoryProm(filtered, formatCounts, formatFilter)))
			return
		}
		last := map[string]interface{}{}
		if len(filtered) > 0 {
			last = map[string]interface{}{
				"timestamp":                    filtered[len(filtered)-1].Timestamp,
				"format":                       filtered[len(filtered)-1].Format,
				"proxy_counters_total":         filtered[len(filtered)-1].ProxyCountersTotal,
				"proxy_counters_prom_total":    filtered[len(filtered)-1].ProxyCountersProm,
				"dashboard_summary_total":      filtered[len(filtered)-1].DashboardSummary,
				"dashboard_summary_prom_total": filtered[len(filtered)-1].DashboardSummaryProm,
			}
		}
		writeJSON(w, map[string]interface{}{
			"count":         len(filtered),
			"format_counts": formatCounts,
			"format_filter": dashboardFilterOrAll(formatFilter),
			"items":         filtered,
			"last":          last,
		})
	})

	mux.HandleFunc("/proxy-counters/health", func(w http.ResponseWriter, r *http.Request) {
		meta := metrics.ProxyMetaSnapshot()
		now := time.Now().UTC().Unix()
		jsonAge := int64(-1)
		promAge := int64(-1)
		if meta["proxy_last_json_unix"] > 0 {
			jsonAge = now - meta["proxy_last_json_unix"]
		}
		if meta["proxy_last_prom_unix"] > 0 {
			promAge = now - meta["proxy_last_prom_unix"]
		}
		jsonStale := jsonAge < 0 || jsonAge > int64(proxyStaleSec)
		promStale := promAge < 0 || promAge > int64(proxyStaleSec)
		payload := map[string]interface{}{
			"proxy_last_json_unix": meta["proxy_last_json_unix"],
			"proxy_last_prom_unix": meta["proxy_last_prom_unix"],
			"proxy_json_age_sec":   jsonAge,
			"proxy_prom_age_sec":   promAge,
			"proxy_json_stale":     jsonStale,
			"proxy_prom_stale":     promStale,
			"stale_threshold_sec":  proxyStaleSec,
		}
		if strings.Contains(r.Header.Get("Accept"), "text/plain") || strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("format")), "prom") {
			w.Header().Set("Content-Type", "text/plain; version=0.0.4")
			lines := []string{
				"# HELP rechain_web6_proxy_json_age_seconds Age in seconds since last proxy-counters JSON request",
				"# TYPE rechain_web6_proxy_json_age_seconds gauge",
				"rechain_web6_proxy_json_age_seconds " + strconv.FormatInt(jsonAge, 10),
				"# HELP rechain_web6_proxy_prom_age_seconds Age in seconds since last proxy-counters Prom request",
				"# TYPE rechain_web6_proxy_prom_age_seconds gauge",
				"rechain_web6_proxy_prom_age_seconds " + strconv.FormatInt(promAge, 10),
				"# HELP rechain_web6_proxy_json_stale Proxy JSON freshness status (1=stale)",
				"# TYPE rechain_web6_proxy_json_stale gauge",
				"rechain_web6_proxy_json_stale " + strconv.Itoa(boolToInt(jsonStale)),
				"# HELP rechain_web6_proxy_prom_stale Proxy Prom freshness status (1=stale)",
				"# TYPE rechain_web6_proxy_prom_stale gauge",
				"rechain_web6_proxy_prom_stale " + strconv.Itoa(boolToInt(promStale)),
				"# HELP rechain_web6_proxy_stale_threshold_seconds Proxy freshness threshold in seconds",
				"# TYPE rechain_web6_proxy_stale_threshold_seconds gauge",
				"rechain_web6_proxy_stale_threshold_seconds " + strconv.Itoa(proxyStaleSec),
			}
			w.Write([]byte(strings.Join(lines, "\n") + "\n"))
			return
		}
		writeJSON(w, payload)
	})

	mux.HandleFunc("/proxy-counters/alerts", func(w http.ResponseWriter, r *http.Request) {
		meta := metrics.ProxyMetaSnapshot()
		now := time.Now().UTC().Unix()
		jsonAge := int64(-1)
		promAge := int64(-1)
		if meta["proxy_last_json_unix"] > 0 {
			jsonAge = now - meta["proxy_last_json_unix"]
		}
		if meta["proxy_last_prom_unix"] > 0 {
			promAge = now - meta["proxy_last_prom_unix"]
		}
		level, score, reason := computeProxyAlert(jsonAge, promAge, int64(proxyStaleSec), int64(proxyCriticalSec))
		if strings.Contains(r.Header.Get("Accept"), "text/plain") || strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("format")), "prom") {
			metrics.IncProxyAlertsProm()
			w.Header().Set("Content-Type", "text/plain; version=0.0.4")
			lines := []string{
				"# HELP rechain_web6_proxy_alert_level Proxy alert level (ok=0,warn=1,critical=2)",
				"# TYPE rechain_web6_proxy_alert_level gauge",
				"rechain_web6_proxy_alert_level " + strconv.Itoa(score),
				"# HELP rechain_web6_proxy_alert_state Proxy alert state flags by level",
				"# TYPE rechain_web6_proxy_alert_state gauge",
				"rechain_web6_proxy_alert_state{level=\"ok\"} " + strconv.Itoa(boolToInt(level == "ok")),
				"rechain_web6_proxy_alert_state{level=\"warn\"} " + strconv.Itoa(boolToInt(level == "warn")),
				"rechain_web6_proxy_alert_state{level=\"critical\"} " + strconv.Itoa(boolToInt(level == "critical")),
			}
			w.Write([]byte(strings.Join(lines, "\n") + "\n"))
			return
		}
		metrics.IncProxyAlerts()
		writeJSON(w, map[string]interface{}{
			"level":                  level,
			"level_score":            score,
			"reason":                 reason,
			"proxy_json_age_sec":     jsonAge,
			"proxy_prom_age_sec":     promAge,
			"stale_threshold_sec":    proxyStaleSec,
			"critical_threshold_sec": proxyCriticalSec,
		})
	})

	mux.HandleFunc("/cache/metrics", func(w http.ResponseWriter, r *http.Request) {
		snap := metrics.Snapshot()
		meta := metrics.ProxyMetaSnapshot()
		now := time.Now().UTC().Unix()
		jsonAge := int64(-1)
		promAge := int64(-1)
		if meta["proxy_last_json_unix"] > 0 {
			jsonAge = now - meta["proxy_last_json_unix"]
		}
		if meta["proxy_last_prom_unix"] > 0 {
			promAge = now - meta["proxy_last_prom_unix"]
		}
		jsonStale := jsonAge < 0 || jsonAge > int64(proxyStaleSec)
		promStale := promAge < 0 || promAge > int64(proxyStaleSec)
		typeSnap := metrics.TypeSnapshot()
		typeFilter := metrics.TypeFilter()
		if typeFilter == "" {
			typeFilter = "all"
		}
		filterMap := map[string]int{
			"all":     0,
			"dir":     0,
			"file":    0,
			"pkg":     0,
			"unknown": 0,
		}
		if _, ok := filterMap[typeFilter]; ok {
			filterMap[typeFilter] = 1
		}
		cacheSnap := map[string]interface{}{
			"graph_ttl_sec":      int(ttl.Seconds()),
			"graph_nodes":        len(cache.graph.Nodes),
			"graph_edges":        len(cache.graph.Edges),
			"types":              typeSnap,
			"filter_type":        typeFilter,
			"filter_type_active": filterMap,
			"proxy_health": map[string]interface{}{
				"proxy_json_age_sec":  jsonAge,
				"proxy_prom_age_sec":  promAge,
				"proxy_json_stale":    jsonStale,
				"proxy_prom_stale":    promStale,
				"stale_threshold_sec": proxyStaleSec,
			},
			"proxy_counters": map[string]interface{}{
				"debug_compare_total":          snap["debug_compare"],
				"debug_compare_prom_total":     snap["debug_compare_prom"],
				"proxy_counters_total":         snap["proxy_counters"],
				"proxy_counters_prom_total":    snap["proxy_counters_prom"],
				"proxy_alerts_total":           snap["proxy_alerts"],
				"proxy_alerts_prom_total":      snap["proxy_alerts_prom"],
				"proxy_last_json_unix":         meta["proxy_last_json_unix"],
				"proxy_last_prom_unix":         meta["proxy_last_prom_unix"],
				"dashboard_summary_total":      snap["dashboard_summary"],
				"dashboard_summary_prom_total": snap["dashboard_summary_prom"],
			},
		}
		goSnap := goListMetrics.Snapshot()
		goCacheSnap := goListCache.Snapshot()
		cacheSnap["go_list_hits"] = goSnap["hits"]
		cacheSnap["go_list_misses"] = goSnap["misses"]
		cacheSnap["go_list_cached"] = goCacheSnap["cached"]
		cacheSnap["go_list_age_sec"] = goCacheSnap["age_sec"]

		if strings.Contains(r.Header.Get("Accept"), "text/plain") || r.URL.Query().Get("format") == "prom" {
			w.Header().Set("Content-Type", "text/plain; version=0.0.4")
			lines := []string{
				"# HELP rechain_web6_graph_ttl_seconds Graph cache TTL seconds",
				"# TYPE rechain_web6_graph_ttl_seconds gauge",
				"rechain_web6_graph_ttl_seconds " + strconv.Itoa(int(ttl.Seconds())),
				"# HELP rechain_web6_graph_nodes Graph node count",
				"# TYPE rechain_web6_graph_nodes gauge",
				"rechain_web6_graph_nodes " + strconv.Itoa(len(cache.graph.Nodes)),
				"# HELP rechain_web6_graph_edges Graph edge count",
				"# TYPE rechain_web6_graph_edges gauge",
				"rechain_web6_graph_edges " + strconv.Itoa(len(cache.graph.Edges)),
				"# HELP rechain_web6_go_list_cache_hits_total Go list cache hits",
				"# TYPE rechain_web6_go_list_cache_hits_total counter",
				"rechain_web6_go_list_cache_hits_total " + strconv.Itoa(goSnap["hits"]),
				"# HELP rechain_web6_go_list_cache_misses_total Go list cache misses",
				"# TYPE rechain_web6_go_list_cache_misses_total counter",
				"rechain_web6_go_list_cache_misses_total " + strconv.Itoa(goSnap["misses"]),
				"# HELP rechain_web6_go_list_cached Go list cache present",
				"# TYPE rechain_web6_go_list_cached gauge",
				"rechain_web6_go_list_cached " + strconv.Itoa(goCacheSnap["cached"]),
				"# HELP rechain_web6_go_list_cache_age_seconds Go list cache age seconds",
				"# TYPE rechain_web6_go_list_cache_age_seconds gauge",
				"rechain_web6_go_list_cache_age_seconds " + strconv.Itoa(goCacheSnap["age_sec"]),
				"# HELP rechain_web6_debug_compare_total Debug compare JSON requests",
				"# TYPE rechain_web6_debug_compare_total counter",
				"rechain_web6_debug_compare_total " + strconv.Itoa(snap["debug_compare"]),
				"# HELP rechain_web6_debug_compare_prom_total Debug compare Prometheus requests",
				"# TYPE rechain_web6_debug_compare_prom_total counter",
				"rechain_web6_debug_compare_prom_total " + strconv.Itoa(snap["debug_compare_prom"]),
				"# HELP rechain_web6_proxy_counters_total Proxy counters JSON requests",
				"# TYPE rechain_web6_proxy_counters_total counter",
				"rechain_web6_proxy_counters_total " + strconv.Itoa(snap["proxy_counters"]),
				"# HELP rechain_web6_proxy_counters_prom_total Proxy counters Prometheus requests",
				"# TYPE rechain_web6_proxy_counters_prom_total counter",
				"rechain_web6_proxy_counters_prom_total " + strconv.Itoa(snap["proxy_counters_prom"]),
				"# HELP rechain_web6_proxy_alerts_total Proxy alerts JSON requests",
				"# TYPE rechain_web6_proxy_alerts_total counter",
				"rechain_web6_proxy_alerts_total " + strconv.Itoa(snap["proxy_alerts"]),
				"# HELP rechain_web6_proxy_alerts_prom_total Proxy alerts Prometheus requests",
				"# TYPE rechain_web6_proxy_alerts_prom_total counter",
				"rechain_web6_proxy_alerts_prom_total " + strconv.Itoa(snap["proxy_alerts_prom"]),
				"# HELP rechain_web6_proxy_counters_last_json_unix Last proxy-counters JSON request unix timestamp",
				"# TYPE rechain_web6_proxy_counters_last_json_unix gauge",
				"rechain_web6_proxy_counters_last_json_unix " + strconv.FormatInt(meta["proxy_last_json_unix"], 10),
				"# HELP rechain_web6_proxy_counters_last_prom_unix Last proxy-counters Prom request unix timestamp",
				"# TYPE rechain_web6_proxy_counters_last_prom_unix gauge",
				"rechain_web6_proxy_counters_last_prom_unix " + strconv.FormatInt(meta["proxy_last_prom_unix"], 10),
				"# HELP rechain_web6_proxy_json_age_seconds Age in seconds since last proxy-counters JSON request",
				"# TYPE rechain_web6_proxy_json_age_seconds gauge",
				"rechain_web6_proxy_json_age_seconds " + strconv.FormatInt(jsonAge, 10),
				"# HELP rechain_web6_proxy_prom_age_seconds Age in seconds since last proxy-counters Prom request",
				"# TYPE rechain_web6_proxy_prom_age_seconds gauge",
				"rechain_web6_proxy_prom_age_seconds " + strconv.FormatInt(promAge, 10),
				"# HELP rechain_web6_proxy_json_stale Proxy JSON freshness status (1=stale)",
				"# TYPE rechain_web6_proxy_json_stale gauge",
				"rechain_web6_proxy_json_stale " + strconv.Itoa(boolToInt(jsonStale)),
				"# HELP rechain_web6_proxy_prom_stale Proxy Prom freshness status (1=stale)",
				"# TYPE rechain_web6_proxy_prom_stale gauge",
				"rechain_web6_proxy_prom_stale " + strconv.Itoa(boolToInt(promStale)),
				"# HELP rechain_web6_proxy_stale_threshold_seconds Proxy freshness threshold in seconds",
				"# TYPE rechain_web6_proxy_stale_threshold_seconds gauge",
				"rechain_web6_proxy_stale_threshold_seconds " + strconv.Itoa(proxyStaleSec),
				"# HELP rechain_web6_dashboard_summary_total Dashboard summary JSON requests",
				"# TYPE rechain_web6_dashboard_summary_total counter",
				"rechain_web6_dashboard_summary_total " + strconv.Itoa(snap["dashboard_summary"]),
				"# HELP rechain_web6_dashboard_summary_prom_total Dashboard summary Prometheus requests",
				"# TYPE rechain_web6_dashboard_summary_prom_total counter",
				"rechain_web6_dashboard_summary_prom_total " + strconv.Itoa(snap["dashboard_summary_prom"]),
			}
			for t, v := range typeSnap {
				lines = append(lines,
					"# HELP rechain_web6_nodes_total Node counts by type",
					"# TYPE rechain_web6_nodes_total gauge",
					"rechain_web6_nodes_total{type=\""+t+"\"} "+strconv.Itoa(v),
				)
			}
			if typeFilter == "" {
				typeFilter = "all"
			}
			for _, t := range []string{"all", "dir", "file", "pkg", "unknown"} {
				v := 0
				if t == typeFilter {
					v = 1
				}
				lines = append(lines,
					"# HELP rechain_web6_filter_type_active Active type filter",
					"# TYPE rechain_web6_filter_type_active gauge",
					"rechain_web6_filter_type_active{type=\""+t+"\"} "+strconv.Itoa(v),
				)
			}
			w.Write([]byte(strings.Join(lines, "\n")))
			return
		}

		writeJSON(w, cacheSnap)
	})

	mux.HandleFunc("/cache/clear", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		cache.mu.Lock()
		cache.graph = Graph{}
		cache.updated = time.Time{}
		cache.mu.Unlock()
		goListCache.Clear()
		writeJSON(w, map[string]string{"status": "cleared"})
	})

	mux.HandleFunc("/queue-depth", func(w http.ResponseWriter, r *http.Request) {
		depth, err := fetchQueueDepth(orchURL)
		if err != nil {
			http.Error(w, "queue depth unavailable", http.StatusBadGateway)
			return
		}
		writeJSON(w, map[string]int{"queue_depth": depth})
	})

	mux.HandleFunc("/last-task-trace", func(w http.ResponseWriter, r *http.Request) {
		trace, err := fetchLatestTaskTrace(orchURL)
		if err != nil {
			http.Error(w, "last trace unavailable", http.StatusBadGateway)
			return
		}
		writeJSON(w, trace)
	})

	mux.HandleFunc("/tasks/recent", func(w http.ResponseWriter, r *http.Request) {
		limit := constraintIntFromQuery(r, "limit", 8)
		stateFilter := strings.TrimSpace(r.URL.Query().Get("state"))
		mergeSource := strings.TrimSpace(r.URL.Query().Get("merge_source"))
		hasParent := strings.TrimSpace(r.URL.Query().Get("has_parent"))
		sortBy := strings.TrimSpace(r.URL.Query().Get("sort"))
		tasks, err := fetchRecentTasks(orchURL, limit, stateFilter, mergeSource, hasParent, sortBy)
		if err != nil {
			http.Error(w, "recent tasks unavailable", http.StatusBadGateway)
			return
		}
		writeJSON(w, tasks)
	})

	mux.HandleFunc("/task-trace", func(w http.ResponseWriter, r *http.Request) {
		taskID := strings.TrimSpace(r.URL.Query().Get("id"))
		if taskID == "" {
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}
		trace, err := fetchTaskTrace(orchURL, taskID)
		if err != nil {
			http.Error(w, "trace unavailable", http.StatusBadGateway)
			return
		}
		writeJSON(w, trace)
	})

	mux.HandleFunc("/task-debug-prom", func(w http.ResponseWriter, r *http.Request) {
		taskID := strings.TrimSpace(r.URL.Query().Get("id"))
		scope := strings.TrimSpace(r.URL.Query().Get("scope"))
		if taskID == "" {
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}
		body, err := fetchTaskDebugProm(orchURL, taskID, scope)
		if err != nil {
			http.Error(w, "task debug prom unavailable", http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		w.Write([]byte(body))
	})

	mux.HandleFunc("/debug-compare", func(w http.ResponseWriter, r *http.Request) {
		id1 := strings.TrimSpace(r.URL.Query().Get("id1"))
		id2 := strings.TrimSpace(r.URL.Query().Get("id2"))
		if id1 == "" || id2 == "" {
			http.Error(w, "missing id1 or id2", http.StatusBadRequest)
			return
		}
		d1, err := fetchTaskDebugJSON(orchURL, id1)
		if err != nil {
			http.Error(w, "debug compare unavailable", http.StatusBadGateway)
			return
		}
		d2, err := fetchTaskDebugJSON(orchURL, id2)
		if err != nil {
			http.Error(w, "debug compare unavailable", http.StatusBadGateway)
			return
		}
		compare := buildDebugCompare(d1, d2)
		if strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("format")), "prom") {
			metrics.IncDebugCompareProm()
			w.Header().Set("Content-Type", "text/plain; version=0.0.4")
			w.Write([]byte(buildDebugCompareProm(compare)))
			return
		}
		metrics.IncDebugCompare()
		writeJSON(w, compare)
	})

	mux.HandleFunc("/replay-chain", func(w http.ResponseWriter, r *http.Request) {
		taskID := strings.TrimSpace(r.URL.Query().Get("id"))
		if taskID == "" {
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}
		chain, err := fetchReplayChain(orchURL, taskID)
		if err != nil {
			http.Error(w, "replay chain unavailable", http.StatusBadGateway)
			return
		}
		writeJSON(w, chain)
	})

	mux.HandleFunc("/task-replay", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		taskID := strings.TrimSpace(r.URL.Query().Get("id"))
		mode := strings.TrimSpace(r.URL.Query().Get("mode"))
		if taskID == "" {
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}
		if mode == "" {
			mode = "force-policy"
		}
		res, err := triggerTaskReplay(orchURL, taskID, mode)
		if err != nil {
			http.Error(w, "replay unavailable", http.StatusBadGateway)
			return
		}
		writeJSON(w, res)
	})

	mux.HandleFunc("/dashboard-summary", func(w http.ResponseWriter, r *http.Request) {
		if strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("format")), "prom") {
			metrics.IncDashboardSummaryProm()
			text, err := fetchDashboardSummaryProm(orchURL)
			if err != nil {
				http.Error(w, "dashboard unavailable", http.StatusBadGateway)
				return
			}
			w.Header().Set("Content-Type", "text/plain; version=0.0.4")
			w.Write([]byte(text))
			return
		}
		metrics.IncDashboardSummary()
		summary, err := fetchDashboardSummary(orchURL)
		if err != nil {
			http.Error(w, "dashboard unavailable", http.StatusBadGateway)
			return
		}
		writeJSON(w, summary)
	})

	mux.HandleFunc("/dashboard-web6", func(w http.ResponseWriter, r *http.Request) {
		if strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("format")), "prom") {
			text, err := fetchDashboardSummaryProm(orchURL)
			if err != nil {
				http.Error(w, "dashboard unavailable", http.StatusBadGateway)
				return
			}
			w.Header().Set("Content-Type", "text/plain; version=0.0.4")
			w.Write([]byte(extractDashboardWeb6Prom(text)))
			return
		}
		summary, err := fetchDashboardSummary(orchURL)
		if err != nil {
			http.Error(w, "dashboard unavailable", http.StatusBadGateway)
			return
		}
		downstream, _ := summary["downstream"].(map[string]interface{})
		web6, _ := downstream["web6"].(map[string]interface{})
		if web6 == nil {
			web6 = map[string]interface{}{}
		}
		writeJSON(w, map[string]interface{}{
			"service": "web6",
			"metrics": web6,
		})
	})

	mux.HandleFunc("/dashboard-web6/alerts", func(w http.ResponseWriter, r *http.Request) {
		summary, err := fetchDashboardSummary(orchURL)
		if err != nil {
			http.Error(w, "dashboard unavailable", http.StatusBadGateway)
			return
		}
		web6 := extractDashboardWeb6Metrics(summary)
		level, score, reason := computeDashboardWeb6Alert(web6)
		alertHistory.Add(DashboardWeb6AlertEvent{
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Level:     level,
			Score:     score,
			Reason:    reason,
			Source:    "alerts",
		})
		if strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("format")), "prom") {
			w.Header().Set("Content-Type", "text/plain; version=0.0.4")
			lines := []string{
				"# HELP rechain_web6_dashboard_web6_alert_level Dashboard web6 alert level (ok=0,warn=1,critical=2)",
				"# TYPE rechain_web6_dashboard_web6_alert_level gauge",
				"rechain_web6_dashboard_web6_alert_level " + strconv.Itoa(score),
				"# HELP rechain_web6_dashboard_web6_alert_state Dashboard web6 alert state flags by level",
				"# TYPE rechain_web6_dashboard_web6_alert_state gauge",
				"rechain_web6_dashboard_web6_alert_state{level=\"ok\"} " + strconv.Itoa(boolToInt(level == "ok")),
				"rechain_web6_dashboard_web6_alert_state{level=\"warn\"} " + strconv.Itoa(boolToInt(level == "warn")),
				"rechain_web6_dashboard_web6_alert_state{level=\"critical\"} " + strconv.Itoa(boolToInt(level == "critical")),
			}
			w.Write([]byte(strings.Join(lines, "\n") + "\n"))
			return
		}
		writeJSON(w, map[string]interface{}{
			"service": "web6",
			"level":   level,
			"score":   score,
			"reason":  reason,
			"metrics": web6,
		})
	})

	mux.HandleFunc("/dashboard-web6/summary", func(w http.ResponseWriter, r *http.Request) {
		summary, err := fetchDashboardSummary(orchURL)
		if err != nil {
			http.Error(w, "dashboard unavailable", http.StatusBadGateway)
			return
		}
		web6 := extractDashboardWeb6Metrics(summary)
		level, score, reason := computeDashboardWeb6Alert(web6)
		alertHistory.Add(DashboardWeb6AlertEvent{
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Level:     level,
			Score:     score,
			Reason:    reason,
			Source:    "summary",
		})
		if strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("format")), "prom") {
			text, err := fetchDashboardSummaryProm(orchURL)
			if err != nil {
				http.Error(w, "dashboard unavailable", http.StatusBadGateway)
				return
			}
			filtered := extractDashboardWeb6Prom(text)
			lines := []string{}
			if filtered != "" {
				lines = append(lines, strings.TrimSpace(filtered))
			}
			lines = append(lines,
				"rechain_web6_dashboard_web6_alert_level "+strconv.Itoa(score),
				"rechain_web6_dashboard_web6_alert_state{level=\"ok\"} "+strconv.Itoa(boolToInt(level == "ok")),
				"rechain_web6_dashboard_web6_alert_state{level=\"warn\"} "+strconv.Itoa(boolToInt(level == "warn")),
				"rechain_web6_dashboard_web6_alert_state{level=\"critical\"} "+strconv.Itoa(boolToInt(level == "critical")),
			)
			w.Header().Set("Content-Type", "text/plain; version=0.0.4")
			w.Write([]byte(strings.Join(lines, "\n") + "\n"))
			return
		}
		writeJSON(w, map[string]interface{}{
			"service": "web6",
			"alert": map[string]interface{}{
				"level":  level,
				"score":  score,
				"reason": reason,
			},
			"metrics": web6,
		})
	})

	mux.HandleFunc("/dashboard-web6/history", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			alertHistory.Reset()
			writeJSON(w, map[string]interface{}{
				"status": "reset",
			})
			return
		}
		limit := constraintIntFromQuery(r, "limit", 20)
		levelFilter := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("level")))
		sourceFilter := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("source")))
		items := alertHistory.List(limit)
		filtered := make([]DashboardWeb6AlertEvent, 0, len(items))
		for _, it := range items {
			if !matchDashboardWeb6HistoryFilter(it, levelFilter, sourceFilter) {
				continue
			}
			filtered = append(filtered, it)
		}
		levelCounts := map[string]int{"ok": 0, "warn": 0, "critical": 0}
		sourceCounts := map[string]int{}
		for _, it := range filtered {
			level := strings.TrimSpace(it.Level)
			if level == "" {
				level = "unknown"
			}
			levelCounts[level]++
			src := strings.TrimSpace(it.Source)
			if src == "" {
				src = "unknown"
			}
			sourceCounts[src]++
		}
		if strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("format")), "prom") {
			w.Header().Set("Content-Type", "text/plain; version=0.0.4")
			w.Write([]byte(buildDashboardWeb6HistoryProm(filtered, levelCounts, sourceCounts, levelFilter, sourceFilter)))
			return
		}
		writeJSON(w, map[string]interface{}{
			"count":         len(filtered),
			"level_counts":  levelCounts,
			"source_counts": sourceCounts,
			"level_filter":  dashboardFilterOrAll(levelFilter),
			"source_filter": dashboardFilterOrAll(sourceFilter),
			"items":         filtered,
		})
	})

	addr := ":8084"
	log.Printf("web6-3d listening on %s", addr)
	if err := http.ListenAndServe(addr, logging.WithRequestID(mux)); err != nil {
		log.Fatal(err)
	}
}

func fetchQueueDepth(orchURL string) (int, error) {
	if orchURL == "" {
		return 0, errors.New("missing ORCH_URL")
	}
	client := &http.Client{Timeout: 800 * time.Millisecond}
	resp, err := client.Get(orchURL + "/queue-depth")
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, errors.New("bad status")
	}
	var payload struct {
		QueueDepth int `json:"queue_depth"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return 0, err
	}
	return payload.QueueDepth, nil
}

func fetchLatestTaskTrace(orchURL string) (map[string]interface{}, error) {
	if orchURL == "" {
		return nil, errors.New("missing ORCH_URL")
	}
	client := &http.Client{Timeout: 1200 * time.Millisecond}
	resp, err := client.Get(orchURL + "/tasks/latest/trace")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("bad status")
	}
	out := map[string]interface{}{}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}

func fetchRecentTasks(orchURL string, limit int, stateFilter string, mergeSource string, hasParent string, sortBy string) (map[string]interface{}, error) {
	if orchURL == "" {
		return nil, errors.New("missing ORCH_URL")
	}
	if limit <= 0 {
		limit = 8
	}
	q := url.Values{}
	q.Set("limit", strconv.Itoa(limit))
	if stateFilter != "" {
		q.Set("state", stateFilter)
	}
	if mergeSource != "" {
		q.Set("merge_source", mergeSource)
	}
	if hasParent != "" {
		q.Set("has_parent", hasParent)
	}
	if sortBy != "" {
		q.Set("sort", sortBy)
	}
	client := &http.Client{Timeout: 1200 * time.Millisecond}
	resp, err := client.Get(orchURL + "/tasks/recent?" + q.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("bad status")
	}
	out := map[string]interface{}{}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}

func fetchTaskTrace(orchURL string, taskID string) (map[string]interface{}, error) {
	if orchURL == "" {
		return nil, errors.New("missing ORCH_URL")
	}
	client := &http.Client{Timeout: 1200 * time.Millisecond}
	resp, err := client.Get(orchURL + "/tasks/" + url.QueryEscape(taskID) + "/trace")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("bad status")
	}
	out := map[string]interface{}{}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}

func fetchReplayChain(orchURL string, taskID string) (map[string]interface{}, error) {
	if orchURL == "" {
		return nil, errors.New("missing ORCH_URL")
	}
	client := &http.Client{Timeout: 1200 * time.Millisecond}
	resp, err := client.Get(orchURL + "/tasks/" + url.QueryEscape(taskID) + "/replay-chain")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("bad status")
	}
	out := map[string]interface{}{}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}

func triggerTaskReplay(orchURL string, taskID string, mode string) (map[string]interface{}, error) {
	if orchURL == "" {
		return nil, errors.New("missing ORCH_URL")
	}
	if mode == "" {
		mode = "force-policy"
	}
	client := &http.Client{Timeout: 1500 * time.Millisecond}
	req, err := http.NewRequest(http.MethodPost, orchURL+"/tasks/"+url.PathEscape(taskID)+"/replay?mode="+url.QueryEscape(mode), strings.NewReader("{}"))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("bad status")
	}
	out := map[string]interface{}{}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}

func fetchTaskDebugProm(orchURL string, taskID string, scope string) (string, error) {
	if orchURL == "" {
		return "", errors.New("missing ORCH_URL")
	}
	if scope == "" {
		scope = "all"
	}
	client := &http.Client{Timeout: 1500 * time.Millisecond}
	req, err := http.NewRequest(http.MethodGet, orchURL+"/tasks/"+url.PathEscape(taskID)+"/debug?format=prom&scope="+url.QueryEscape(scope), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "text/plain")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", errors.New("bad status")
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func fetchTaskDebugJSON(orchURL string, taskID string) (map[string]interface{}, error) {
	if orchURL == "" {
		return nil, errors.New("missing ORCH_URL")
	}
	client := &http.Client{Timeout: 1500 * time.Millisecond}
	resp, err := client.Get(orchURL + "/tasks/" + url.PathEscape(taskID) + "/debug")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("bad status")
	}
	out := map[string]interface{}{}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}

func buildDebugCompare(a map[string]interface{}, b map[string]interface{}) map[string]interface{} {
	safeTaskID := func(m map[string]interface{}) string {
		if v, ok := m["task_id"].(string); ok {
			return v
		}
		return ""
	}
	getMerge := func(m map[string]interface{}) map[string]interface{} {
		trace, _ := m["trace"].(map[string]interface{})
		merge, _ := trace["merge"].(map[string]interface{})
		return merge
	}
	getNumber := func(m map[string]interface{}, key string) float64 {
		if v, ok := m[key].(float64); ok {
			return v
		}
		return 0
	}
	getMergeSource := func(m map[string]interface{}) string {
		trace, _ := m["trace"].(map[string]interface{})
		if v, ok := trace["merge_source"].(string); ok {
			return v
		}
		return ""
	}
	ma := getMerge(a)
	mb := getMerge(b)
	qa := getNumber(ma, "quality_score")
	qb := getNumber(mb, "quality_score")
	ca := getNumber(ma, "confidence")
	cb := getNumber(mb, "confidence")
	return map[string]interface{}{
		"id1": safeTaskID(a),
		"id2": safeTaskID(b),
		"metrics": map[string]interface{}{
			"quality_score_1":  qa,
			"quality_score_2":  qb,
			"quality_delta":    qa - qb,
			"confidence_1":     ca,
			"confidence_2":     cb,
			"confidence_delta": ca - cb,
			"merge_source_1":   getMergeSource(a),
			"merge_source_2":   getMergeSource(b),
		},
	}
}

func buildDebugCompareProm(compare map[string]interface{}) string {
	id1, _ := compare["id1"].(string)
	id2, _ := compare["id2"].(string)
	metrics, _ := compare["metrics"].(map[string]interface{})
	qualityDelta, _ := metrics["quality_delta"].(float64)
	confidenceDelta, _ := metrics["confidence_delta"].(float64)
	merge1, _ := metrics["merge_source_1"].(string)
	merge2, _ := metrics["merge_source_2"].(string)
	merge1 = promLabelValue(merge1)
	merge2 = promLabelValue(merge2)
	id1 = promLabelValue(id1)
	id2 = promLabelValue(id2)
	lines := []string{
		"# HELP rechain_web6_debug_compare_quality_delta Quality score delta between compared tasks",
		"# TYPE rechain_web6_debug_compare_quality_delta gauge",
		"rechain_web6_debug_compare_quality_delta{id1=\"" + id1 + "\",id2=\"" + id2 + "\"} " + strconv.FormatFloat(qualityDelta, 'f', 6, 64),
		"# HELP rechain_web6_debug_compare_confidence_delta Confidence delta between compared tasks",
		"# TYPE rechain_web6_debug_compare_confidence_delta gauge",
		"rechain_web6_debug_compare_confidence_delta{id1=\"" + id1 + "\",id2=\"" + id2 + "\"} " + strconv.FormatFloat(confidenceDelta, 'f', 6, 64),
		"# HELP rechain_web6_debug_compare_total Compare calls by merge source pair",
		"# TYPE rechain_web6_debug_compare_total gauge",
		"rechain_web6_debug_compare_total{merge_source_1=\"" + merge1 + "\",merge_source_2=\"" + merge2 + "\"} 1",
	}
	return strings.Join(lines, "\n") + "\n"
}

func fetchDashboardSummary(orchURL string) (map[string]interface{}, error) {
	if orchURL == "" {
		return nil, errors.New("missing ORCH_URL")
	}
	client := &http.Client{Timeout: 1200 * time.Millisecond}
	resp, err := client.Get(orchURL + "/dashboard/summary")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("bad status")
	}
	out := map[string]interface{}{}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}

func fetchDashboardSummaryProm(orchURL string) (string, error) {
	if orchURL == "" {
		return "", errors.New("missing ORCH_URL")
	}
	client := &http.Client{Timeout: 1200 * time.Millisecond}
	resp, err := client.Get(orchURL + "/dashboard/summary?format=prom")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", errors.New("bad status")
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func extractDashboardWeb6Prom(promText string) string {
	lines := strings.Split(promText, "\n")
	out := []string{}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "rechain_dashboard_web6_") {
			out = append(out, line)
			continue
		}
		if strings.HasPrefix(line, "rechain_dashboard_downstream_up{") && strings.Contains(line, `service="web6"`) {
			out = append(out, line)
		}
	}
	if len(out) == 0 {
		return ""
	}
	return strings.Join(out, "\n") + "\n"
}

func extractDashboardWeb6Metrics(summary map[string]interface{}) map[string]interface{} {
	downstream, _ := summary["downstream"].(map[string]interface{})
	web6, _ := downstream["web6"].(map[string]interface{})
	if web6 == nil {
		web6 = map[string]interface{}{}
	}
	return web6
}

func computeDashboardWeb6Alert(web6 map[string]interface{}) (string, int, string) {
	up, _ := web6["up"].(bool)
	if !up {
		return "critical", 2, "downstream web6 is down"
	}
	jsonStale, _ := web6["proxy_json_stale"].(bool)
	promStale, _ := web6["proxy_prom_stale"].(bool)
	if jsonStale || promStale {
		if jsonStale && promStale {
			return "critical", 2, "json+prom stale"
		}
		return "warn", 1, "partial stale"
	}
	levelScore, ok := web6["proxy_alert_level"].(float64)
	if ok {
		if levelScore >= 2 {
			return "critical", 2, "proxy alert critical"
		}
		if levelScore >= 1 {
			return "warn", 1, "proxy alert warn"
		}
	}
	return "ok", 0, "healthy"
}

func buildDashboardWeb6HistoryProm(items []DashboardWeb6AlertEvent, levelCounts map[string]int, sourceCounts map[string]int, levelFilter string, sourceFilter string) string {
	levelFilter = dashboardFilterOrAll(levelFilter)
	sourceFilter = dashboardFilterOrAll(sourceFilter)
	lines := []string{
		"# HELP rechain_web6_dashboard_web6_alert_history_total Dashboard web6 alert history size",
		"# TYPE rechain_web6_dashboard_web6_alert_history_total gauge",
		"rechain_web6_dashboard_web6_alert_history_total " + strconv.Itoa(len(items)),
		"# HELP rechain_web6_dashboard_web6_alert_history_by_level Dashboard web6 alert history grouped by level",
		"# TYPE rechain_web6_dashboard_web6_alert_history_by_level gauge",
	}
	for _, level := range []string{"ok", "warn", "critical", "unknown"} {
		if v, ok := levelCounts[level]; ok {
			lines = append(lines, "rechain_web6_dashboard_web6_alert_history_by_level{level=\""+promLabelValue(level)+"\"} "+strconv.Itoa(v))
		}
	}
	lines = append(lines,
		"# HELP rechain_web6_dashboard_web6_alert_history_by_source Dashboard web6 alert history grouped by source",
		"# TYPE rechain_web6_dashboard_web6_alert_history_by_source gauge",
	)
	for src, v := range sourceCounts {
		lines = append(lines, "rechain_web6_dashboard_web6_alert_history_by_source{source=\""+promLabelValue(src)+"\"} "+strconv.Itoa(v))
	}
	lines = append(lines,
		"# HELP rechain_web6_dashboard_web6_alert_history_filter_active Active filters for dashboard web6 history",
		"# TYPE rechain_web6_dashboard_web6_alert_history_filter_active gauge",
		"rechain_web6_dashboard_web6_alert_history_filter_active{kind=\"level\",value=\""+promLabelValue(levelFilter)+"\"} 1",
		"rechain_web6_dashboard_web6_alert_history_filter_active{kind=\"source\",value=\""+promLabelValue(sourceFilter)+"\"} 1",
	)
	return strings.Join(lines, "\n") + "\n"
}

func dashboardFilterOrAll(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	if v == "" {
		return "all"
	}
	return v
}

func matchDashboardWeb6HistoryFilter(it DashboardWeb6AlertEvent, levelFilter string, sourceFilter string) bool {
	levelFilter = dashboardFilterOrAll(levelFilter)
	sourceFilter = dashboardFilterOrAll(sourceFilter)
	level := strings.ToLower(strings.TrimSpace(it.Level))
	if level == "" {
		level = "unknown"
	}
	source := strings.ToLower(strings.TrimSpace(it.Source))
	if source == "" {
		source = "unknown"
	}
	if levelFilter != "all" && level != levelFilter {
		return false
	}
	if sourceFilter != "all" && source != sourceFilter {
		return false
	}
	return true
}

func buildProxyCountersProm(payload map[string]int, meta map[string]interface{}) string {
	lastJSONUnix := int64(0)
	lastPromUnix := int64(0)
	if v, ok := meta["proxy_last_json_unix"].(int64); ok {
		lastJSONUnix = v
	}
	if v, ok := meta["proxy_last_prom_unix"].(int64); ok {
		lastPromUnix = v
	}
	lines := []string{
		"# HELP rechain_web6_debug_compare_total Debug compare JSON requests",
		"# TYPE rechain_web6_debug_compare_total counter",
		"rechain_web6_debug_compare_total " + strconv.Itoa(payload["debug_compare_total"]),
		"# HELP rechain_web6_debug_compare_prom_total Debug compare Prometheus requests",
		"# TYPE rechain_web6_debug_compare_prom_total counter",
		"rechain_web6_debug_compare_prom_total " + strconv.Itoa(payload["debug_compare_prom_total"]),
		"# HELP rechain_web6_proxy_counters_total Proxy counters JSON requests",
		"# TYPE rechain_web6_proxy_counters_total counter",
		"rechain_web6_proxy_counters_total " + strconv.Itoa(payload["proxy_counters_total"]),
		"# HELP rechain_web6_proxy_counters_prom_total Proxy counters Prometheus requests",
		"# TYPE rechain_web6_proxy_counters_prom_total counter",
		"rechain_web6_proxy_counters_prom_total " + strconv.Itoa(payload["proxy_counters_prom_total"]),
		"# HELP rechain_web6_proxy_alerts_total Proxy alerts JSON requests",
		"# TYPE rechain_web6_proxy_alerts_total counter",
		"rechain_web6_proxy_alerts_total " + strconv.Itoa(payload["proxy_alerts_total"]),
		"# HELP rechain_web6_proxy_alerts_prom_total Proxy alerts Prometheus requests",
		"# TYPE rechain_web6_proxy_alerts_prom_total counter",
		"rechain_web6_proxy_alerts_prom_total " + strconv.Itoa(payload["proxy_alerts_prom_total"]),
		"# HELP rechain_web6_proxy_counters_last_json_unix Last proxy-counters JSON request unix timestamp",
		"# TYPE rechain_web6_proxy_counters_last_json_unix gauge",
		"rechain_web6_proxy_counters_last_json_unix " + strconv.FormatInt(lastJSONUnix, 10),
		"# HELP rechain_web6_proxy_counters_last_prom_unix Last proxy-counters Prom request unix timestamp",
		"# TYPE rechain_web6_proxy_counters_last_prom_unix gauge",
		"rechain_web6_proxy_counters_last_prom_unix " + strconv.FormatInt(lastPromUnix, 10),
		"# HELP rechain_web6_dashboard_summary_total Dashboard summary JSON requests",
		"# TYPE rechain_web6_dashboard_summary_total counter",
		"rechain_web6_dashboard_summary_total " + strconv.Itoa(payload["dashboard_summary_total"]),
		"# HELP rechain_web6_dashboard_summary_prom_total Dashboard summary Prometheus requests",
		"# TYPE rechain_web6_dashboard_summary_prom_total counter",
		"rechain_web6_dashboard_summary_prom_total " + strconv.Itoa(payload["dashboard_summary_prom_total"]),
	}
	return strings.Join(lines, "\n") + "\n"
}

func buildProxyCountersHistoryProm(items []ProxyCountersEvent, formatCounts map[string]int, formatFilter string) string {
	lastFormat := "none"
	lastJSON := 0
	lastProm := 0
	lastDashJSON := 0
	lastDashProm := 0
	if len(items) > 0 {
		last := items[len(items)-1]
		if strings.TrimSpace(last.Format) != "" {
			lastFormat = strings.ToLower(strings.TrimSpace(last.Format))
		}
		lastJSON = last.ProxyCountersTotal
		lastProm = last.ProxyCountersProm
		lastDashJSON = last.DashboardSummary
		lastDashProm = last.DashboardSummaryProm
	}
	lines := []string{
		"# HELP rechain_web6_proxy_history_total Proxy counters history points",
		"# TYPE rechain_web6_proxy_history_total gauge",
		"rechain_web6_proxy_history_total " + strconv.Itoa(len(items)),
		"# HELP rechain_web6_proxy_history_by_format History points by response format",
		"# TYPE rechain_web6_proxy_history_by_format gauge",
		"rechain_web6_proxy_history_by_format{format=\"json\"} " + strconv.Itoa(formatCounts["json"]),
		"rechain_web6_proxy_history_by_format{format=\"prom\"} " + strconv.Itoa(formatCounts["prom"]),
		"# HELP rechain_web6_proxy_history_filter_active Active proxy history filter",
		"# TYPE rechain_web6_proxy_history_filter_active gauge",
		"rechain_web6_proxy_history_filter_active{kind=\"format\",value=\"all\"} " + strconv.Itoa(boolToInt(dashboardFilterOrAll(formatFilter) == "all")),
		"rechain_web6_proxy_history_filter_active{kind=\"format\",value=\"json\"} " + strconv.Itoa(boolToInt(dashboardFilterOrAll(formatFilter) == "json")),
		"rechain_web6_proxy_history_filter_active{kind=\"format\",value=\"prom\"} " + strconv.Itoa(boolToInt(dashboardFilterOrAll(formatFilter) == "prom")),
		"# HELP rechain_web6_proxy_history_last_counter Last proxy counter values from history",
		"# TYPE rechain_web6_proxy_history_last_counter gauge",
		"rechain_web6_proxy_history_last_counter{name=\"proxy_json_total\"} " + strconv.Itoa(lastJSON),
		"rechain_web6_proxy_history_last_counter{name=\"proxy_prom_total\"} " + strconv.Itoa(lastProm),
		"rechain_web6_proxy_history_last_counter{name=\"dashboard_json_total\"} " + strconv.Itoa(lastDashJSON),
		"rechain_web6_proxy_history_last_counter{name=\"dashboard_prom_total\"} " + strconv.Itoa(lastDashProm),
		"# HELP rechain_web6_proxy_history_last_format Last history item format flags",
		"# TYPE rechain_web6_proxy_history_last_format gauge",
		"rechain_web6_proxy_history_last_format{format=\"json\"} " + strconv.Itoa(boolToInt(lastFormat == "json")),
		"rechain_web6_proxy_history_last_format{format=\"prom\"} " + strconv.Itoa(boolToInt(lastFormat == "prom")),
	}
	return strings.Join(lines, "\n") + "\n"
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func renderIndexHTML() string {
	unknown := envOr("WEB6_COLOR_UNKNOWN", "#999")
	matchDir := envOr("WEB6_COLOR_MATCH_DIR", "#f5a623")
	matchFile := envOr("WEB6_COLOR_MATCH_FILE", "#f08c6a")
	matchPkg := envOr("WEB6_COLOR_MATCH_PKG", "#f1b54a")
	matchUnknown := envOr("WEB6_COLOR_MATCH_UNKNOWN", "#f6c08a")
	matchDefault := envOr("WEB6_COLOR_MATCH_DEFAULT", "#f39c12")

	r := strings.NewReplacer(
		"__UNKNOWN_COLOR__", unknown,
		"__MATCH_DIR__", matchDir,
		"__MATCH_FILE__", matchFile,
		"__MATCH_PKG__", matchPkg,
		"__MATCH_UNKNOWN__", matchUnknown,
		"__MATCH_DEFAULT__", matchDefault,
		"__DRAWER__", drawerHTML,
	)
	return r.Replace(indexHTMLTemplate)
}

//go:embed ui/index.html
var indexHTMLTemplate string

//go:embed ui/style.css
var styleCSSTemplate string

//go:embed ui/app.js
var appJSTemplate string

//go:embed ui/drawer.html
var drawerHTML string

type GraphCache struct {
	mu      sync.Mutex
	graph   Graph
	updated time.Time
}

func countTypes(nodes []Node) map[string]int {
	out := map[string]int{}
	for _, n := range nodes {
		t := strings.ToLower(strings.TrimSpace(n.Type))
		if t == "" {
			t = "unknown"
		}
		out[t]++
	}
	return out
}

func loadGraph(path string, root string, maxNodes int, ttl time.Duration, cache *GraphCache, mode string, goCache *GoListCache, goMetrics *GoListMetrics) Graph {
	if ttl > 0 && cache != nil {
		cache.mu.Lock()
		if time.Since(cache.updated) < ttl && len(cache.graph.Nodes) > 0 {
			g := cache.graph
			cache.mu.Unlock()
			return g
		}
		cache.mu.Unlock()
	}

	var g Graph
	if path != "" {
		g = readGraphFile(path)
	}
	if len(g.Nodes) == 0 {
		if strings.EqualFold(mode, "imports") {
			g = buildImportGraphFromRepo(root, maxNodes)
		} else if strings.EqualFold(mode, "imports_go_list") || strings.EqualFold(mode, "go_list") {
			g = buildGoListGraph(root, maxNodes, goCache, goMetrics)
		} else {
			g = buildGraphFromRepo(root, maxNodes)
		}
	}
	if len(g.Nodes) == 0 {
		g = Graph{
			Nodes: []Node{{ID: "root", Label: "repo", Type: "dir"}, {ID: "kernel", Label: "kernel", Type: "dir"}, {ID: "orchestrator", Label: "orchestrator", Type: "dir"}, {ID: "rag", Label: "rag", Type: "dir"}, {ID: "web6-3d", Label: "web6-3d", Type: "dir"}},
			Edges: []Edge{{From: "root", To: "kernel"}, {From: "root", To: "orchestrator"}, {From: "root", To: "rag"}, {From: "root", To: "web6-3d"}},
		}
	}

	if cache != nil {
		cache.mu.Lock()
		cache.graph = g
		cache.updated = time.Now()
		cache.mu.Unlock()
	}
	return g
}

func readGraphFile(path string) Graph {
	if data, err := os.ReadFile(path); err == nil {
		var g Graph
		if err := json.Unmarshal(data, &g); err == nil {
			if len(g.Nodes) > 0 {
				return g
			}
		}
	}
	return Graph{}
}

func buildGraphFromRepo(root string, maxNodes int) Graph {
	if root == "" {
		root = "."
	}
	if maxNodes <= 0 {
		maxNodes = 2000
	}
	nodes := []Node{{ID: ".", Label: filepath.Base(root), Type: "dir"}}
	edges := []Edge{}
	count := 1
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if path == root {
			return nil
		}
		if count >= maxNodes {
			return errors.New("limit reached")
		}
		name := d.Name()
		if strings.HasPrefix(name, ".") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		parent := filepath.ToSlash(filepath.Dir(rel))
		if parent == "." || parent == "" {
			parent = "."
		}
		ntype := "file"
		if d.IsDir() {
			ntype = "dir"
		}
		nodes = append(nodes, Node{ID: rel, Label: name, Type: ntype})
		edges = append(edges, Edge{From: parent, To: rel})
		count++
		return nil
	})
	return Graph{Nodes: nodes, Edges: edges}
}

func buildImportGraphFromRepo(root string, maxNodes int) Graph {
	if root == "" {
		root = "."
	}
	if maxNodes <= 0 {
		maxNodes = 2000
	}

	nodes := []Node{}
	edges := []Edge{}
	nodeSet := map[string]bool{}
	addNode := func(id string) {
		if id == "" || nodeSet[id] {
			return
		}
		nodeSet[id] = true
		nodes = append(nodes, Node{ID: id, Label: filepath.Base(id), Type: "file"})
	}
	addEdge := func(from string, to string) {
		if from == "" || to == "" {
			return
		}
		edges = append(edges, Edge{From: from, To: to})
	}

	count := 0
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".go" && ext != ".js" && ext != ".ts" && ext != ".tsx" {
			return nil
		}
		if count >= maxNodes {
			return errors.New("limit reached")
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		addNode(rel)
		count++

		var imports []string
		if ext == ".go" {
			imports = parseGoImports(path)
		} else {
			imports = parseJSImports(path)
		}
		for _, imp := range imports {
			if !strings.HasPrefix(imp, ".") {
				continue
			}
			resolved := resolveLocalImport(root, filepath.Dir(path), imp)
			if resolved == "" {
				continue
			}
			relTo, err := filepath.Rel(root, resolved)
			if err != nil {
				continue
			}
			relTo = filepath.ToSlash(relTo)
			addNode(relTo)
			addEdge(rel, relTo)
		}
		return nil
	})

	return Graph{Nodes: nodes, Edges: edges}
}

type GoListCache struct {
	mu      sync.Mutex
	graph   Graph
	updated time.Time
	modTime time.Time
}

type GoListMetrics struct {
	mu     sync.Mutex
	hits   int
	misses int
}

func (m *GoListMetrics) Hit() {
	m.mu.Lock()
	m.hits++
	m.mu.Unlock()
}

func (m *GoListMetrics) Miss() {
	m.mu.Lock()
	m.misses++
	m.mu.Unlock()
}

func (m *GoListMetrics) Snapshot() map[string]int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return map[string]int{
		"hits":   m.hits,
		"misses": m.misses,
	}
}

func buildGoListGraph(root string, maxNodes int, cache *GoListCache, metrics *GoListMetrics) Graph {
	if root == "" {
		root = "."
	}
	if maxNodes <= 0 {
		maxNodes = 2000
	}
	if cache != nil {
		if g, ok := cache.Get(root, 30*time.Second); ok {
			if metrics != nil {
				metrics.Hit()
			}
			return g
		}
	}
	if metrics != nil {
		metrics.Miss()
	}
	modPath := goListModulePath(root)
	if modPath == "" {
		return Graph{}
	}

	cmd := exec.Command("go", "list", "-deps", "-json", "./...")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return Graph{}
	}

	dec := json.NewDecoder(strings.NewReader(string(out)))
	nodes := []Node{}
	edges := []Edge{}
	nodeSet := map[string]bool{}
	count := 0

	for {
		var pkg struct {
			ImportPath string   `json:"ImportPath"`
			Imports    []string `json:"Imports"`
		}
		if err := dec.Decode(&pkg); err != nil {
			break
		}
		if pkg.ImportPath == "" {
			continue
		}
		if !strings.HasPrefix(pkg.ImportPath, modPath) {
			continue
		}
		if !nodeSet[pkg.ImportPath] {
			nodeSet[pkg.ImportPath] = true
			nodes = append(nodes, Node{ID: pkg.ImportPath, Label: filepath.Base(pkg.ImportPath), Type: "pkg"})
			count++
		}
		if count >= maxNodes {
			break
		}
		for _, imp := range pkg.Imports {
			if !strings.HasPrefix(imp, modPath) {
				continue
			}
			if !nodeSet[imp] {
				nodeSet[imp] = true
				nodes = append(nodes, Node{ID: imp, Label: filepath.Base(imp), Type: "pkg"})
				count++
			}
			edges = append(edges, Edge{From: pkg.ImportPath, To: imp})
			if count >= maxNodes {
				break
			}
		}
		if count >= maxNodes {
			break
		}
	}
	g := Graph{Nodes: nodes, Edges: edges}
	if cache != nil {
		cache.Set(root, g)
	}
	return g
}

func (c *GoListCache) Get(root string, ttl time.Duration) (Graph, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.graph.Nodes) == 0 {
		return Graph{}, false
	}
	if ttl > 0 && time.Since(c.updated) > ttl {
		return Graph{}, false
	}
	if modTime := latestGoModTime(root); !modTime.IsZero() && modTime.After(c.modTime) {
		return Graph{}, false
	}
	return c.graph, true
}

func (c *GoListCache) Set(root string, g Graph) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.graph = g
	c.updated = time.Now()
	c.modTime = latestGoModTime(root)
}

func (c *GoListCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.graph = Graph{}
	c.updated = time.Time{}
	c.modTime = time.Time{}
}

func (c *GoListCache) Snapshot() map[string]int {
	c.mu.Lock()
	defer c.mu.Unlock()
	age := 0
	if !c.updated.IsZero() {
		age = int(time.Since(c.updated).Seconds())
	}
	cached := 0
	if len(c.graph.Nodes) > 0 {
		cached = 1
	}
	return map[string]int{
		"cached":  cached,
		"age_sec": age,
	}
}

func latestGoModTime(root string) time.Time {
	var latest time.Time
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Name() != "go.mod" {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		if info.ModTime().After(latest) {
			latest = info.ModTime()
		}
		return nil
	})
	return latest
}

func goListModulePath(root string) string {
	cmd := exec.Command("go", "list", "-m", "-json")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	var mod struct {
		Path string `json:"Path"`
	}
	if err := json.Unmarshal(out, &mod); err != nil {
		return ""
	}
	return mod.Path
}

func parseGoImports(path string) []string {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
	if err != nil || f == nil {
		return nil
	}
	imports := []string{}
	for _, imp := range f.Imports {
		v := strings.Trim(imp.Path.Value, "\"")
		if v != "" {
			imports = append(imports, v)
		}
	}
	return imports
}

var reJSImport = regexp.MustCompile(`(?m)^\s*import\s+.*?from\s+['"]([^'"]+)['"]|require\(\s*['"]([^'"]+)['"]\s*\)`)

func parseJSImports(path string) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	matches := reJSImport.FindAllStringSubmatch(string(data), -1)
	imports := []string{}
	for _, m := range matches {
		if len(m) >= 2 && m[1] != "" {
			imports = append(imports, m[1])
		} else if len(m) >= 3 && m[2] != "" {
			imports = append(imports, m[2])
		}
	}
	return imports
}

func resolveLocalImport(root string, fromDir string, spec string) string {
	if !strings.HasPrefix(spec, ".") {
		return ""
	}
	base := filepath.Join(fromDir, spec)
	candidates := []string{
		base,
		base + ".go",
		base + ".js",
		base + ".ts",
		base + ".tsx",
		filepath.Join(base, "index.go"),
		filepath.Join(base, "index.js"),
		filepath.Join(base, "index.ts"),
		filepath.Join(base, "index.tsx"),
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	_ = root
	return ""
}

func filterGraph(g Graph, q string) Graph {
	q = strings.ToLower(q)
	keep := map[string]bool{}
	for _, n := range g.Nodes {
		if strings.Contains(strings.ToLower(n.Label), q) || strings.Contains(strings.ToLower(n.ID), q) {
			keep[n.ID] = true
		}
	}
	nodes := []Node{}
	for _, n := range g.Nodes {
		if keep[n.ID] {
			nodes = append(nodes, n)
		}
	}
	edges := []Edge{}
	for _, e := range g.Edges {
		if keep[e.From] && keep[e.To] {
			edges = append(edges, e)
		}
	}
	return Graph{Nodes: nodes, Edges: edges}
}

func filterGraphContext(g Graph, q string, depth int, metrics *Metrics) Graph {
	q = strings.ToLower(q)
	keep := map[string]bool{}
	for _, n := range g.Nodes {
		if strings.Contains(strings.ToLower(n.Label), q) || strings.Contains(strings.ToLower(n.ID), q) {
			keep[n.ID] = true
		}
	}
	if depth <= 0 {
		depth = 1
	}
	// expand N-hop neighbors
	for i := 0; i < depth; i++ {
		for _, e := range g.Edges {
			if keep[e.From] || keep[e.To] {
				keep[e.From] = true
				keep[e.To] = true
			}
		}
	}
	nodes := []Node{}
	for _, n := range g.Nodes {
		if keep[n.ID] {
			nodes = append(nodes, n)
		}
	}
	edges := []Edge{}
	for _, e := range g.Edges {
		if keep[e.From] && keep[e.To] {
			edges = append(edges, e)
		}
	}
	if metrics != nil {
		metrics.RecordFilter(len(nodes), len(edges))
	}
	return Graph{Nodes: nodes, Edges: edges}
}

func filterGraphByType(g Graph, t string) Graph {
	t = strings.ToLower(strings.TrimSpace(t))
	if t == "" {
		return g
	}
	nodes := []Node{}
	keep := map[string]bool{}
	for _, n := range g.Nodes {
		nt := strings.ToLower(strings.TrimSpace(n.Type))
		if nt == "" {
			nt = "unknown"
		}
		if nt == t {
			nodes = append(nodes, n)
			keep[n.ID] = true
		}
	}
	edges := []Edge{}
	for _, e := range g.Edges {
		if keep[e.From] && keep[e.To] {
			edges = append(edges, e)
		}
	}
	return Graph{Nodes: nodes, Edges: edges}
}

func envOr(key string, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

func promLabelValue(v string) string {
	v = strings.ReplaceAll(v, `\`, `\\`)
	v = strings.ReplaceAll(v, `"`, `\"`)
	if strings.TrimSpace(v) == "" {
		return "unknown"
	}
	return v
}

func computeProxyAlert(jsonAge int64, promAge int64, staleThreshold int64, criticalThreshold int64) (string, int, string) {
	if criticalThreshold < staleThreshold {
		criticalThreshold = staleThreshold
	}
	jsonMissing := jsonAge < 0
	promMissing := promAge < 0
	if jsonMissing || promMissing {
		reason := "missing"
		if jsonMissing && promMissing {
			reason = "json+prom missing"
		} else if jsonMissing {
			reason = "json missing"
		} else if promMissing {
			reason = "prom missing"
		}
		return "critical", 2, reason
	}
	if jsonAge > criticalThreshold || promAge > criticalThreshold {
		return "critical", 2, "age above critical threshold"
	}
	if jsonAge > staleThreshold || promAge > staleThreshold {
		return "warn", 1, "age above stale threshold"
	}
	return "ok", 0, "fresh"
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func envInt(key string, fallback int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func constraintIntFromQuery(r *http.Request, key string, fallback int) int {
	v := strings.TrimSpace(r.URL.Query().Get(key))
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
