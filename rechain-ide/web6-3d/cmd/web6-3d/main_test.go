package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRenderIndexHTMLContainsCriticalUIBlocks(t *testing.T) {
	html := renderIndexHTML()
	required := []string{
		`id="drawer"`,
		`id="debug-compare-prom"`,
		`id="dashboard-prom"`,
		`id="proxy-metrics"`,
		`id="proxy-prom"`,
		`id="proxy-history"`,
		`id="proxy-history-limit"`,
		`id="proxy-history-format"`,
		`id="proxy-history-copy"`,
		`id="proxy-history-reset"`,
		`id="proxy-health"`,
		`id="proxy-alerts"`,
		`id="dashboard-web6"`,
		`id="dashboard-web6-prom"`,
		`id="dashboard-web6-alerts"`,
		`id="dashboard-web6-summary"`,
		`id="dashboard-web6-history"`,
		`id="dashboard-web6-history-limit"`,
		`id="dashboard-web6-history-level"`,
		`id="dashboard-web6-history-source"`,
		`id="dashboard-web6-history-copy"`,
		`id="dashboard-web6-history-reset"`,
		`id="color-match-default"`,
		`id="share"`,
	}
	for _, token := range required {
		if !strings.Contains(html, token) {
			t.Fatalf("rendered index does not include required token: %s", token)
		}
	}
}

func TestComputeProxyAlert(t *testing.T) {
	level, score, _ := computeProxyAlert(5, 7, 30, 90)
	if level != "ok" || score != 0 {
		t.Fatalf("expected ok/0, got %s/%d", level, score)
	}
	level, score, _ = computeProxyAlert(31, 7, 30, 90)
	if level != "warn" || score != 1 {
		t.Fatalf("expected warn/1, got %s/%d", level, score)
	}
	level, score, _ = computeProxyAlert(91, 7, 30, 90)
	if level != "critical" || score != 2 {
		t.Fatalf("expected critical/2, got %s/%d", level, score)
	}
	level, score, _ = computeProxyAlert(-1, 7, 30, 90)
	if level != "critical" || score != 2 {
		t.Fatalf("expected critical/2 for missing, got %s/%d", level, score)
	}
}

func TestBuildDebugComparePromIncludesRequiredMetrics(t *testing.T) {
	compare := map[string]interface{}{
		"id1": "task_a",
		"id2": "task_b",
		"metrics": map[string]interface{}{
			"quality_delta":    0.125,
			"confidence_delta": -0.25,
			"merge_source_1":   "policy_merge",
			"merge_source_2":   "agent_compiler",
		},
	}
	prom := buildDebugCompareProm(compare)
	required := []string{
		"rechain_web6_debug_compare_quality_delta",
		"rechain_web6_debug_compare_confidence_delta",
		"rechain_web6_debug_compare_total",
		`id1="task_a"`,
		`id2="task_b"`,
		`merge_source_1="policy_merge"`,
		`merge_source_2="agent_compiler"`,
	}
	for _, token := range required {
		if !strings.Contains(prom, token) {
			t.Fatalf("prom output missing token: %s\n%s", token, prom)
		}
	}
}

func TestFetchDashboardSummaryProm(t *testing.T) {
	body := "# TYPE rechain_dashboard_tasks_total gauge\nrechain_dashboard_tasks_total{state=\"completed\"} 3\n"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/dashboard/summary" {
			http.NotFound(w, r)
			return
		}
		if got := r.URL.Query().Get("format"); got != "prom" {
			t.Fatalf("expected format=prom, got %q", got)
		}
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	text, err := fetchDashboardSummaryProm(srv.URL)
	if err != nil {
		t.Fatalf("fetchDashboardSummaryProm returned error: %v", err)
	}
	if text != body {
		t.Fatalf("unexpected body: %q", text)
	}
}

func TestMetricsSnapshotIncludesProxyCounters(t *testing.T) {
	m := &Metrics{}
	m.IncDebugCompare()
	m.IncDebugCompareProm()
	m.TouchProxyCounters()
	m.TouchProxyCountersProm()
	m.IncDashboardSummary()
	m.IncDashboardSummaryProm()
	snap := m.Snapshot()
	if snap["debug_compare"] != 1 {
		t.Fatalf("debug_compare expected 1, got %d", snap["debug_compare"])
	}
	if snap["debug_compare_prom"] != 1 {
		t.Fatalf("debug_compare_prom expected 1, got %d", snap["debug_compare_prom"])
	}
	if snap["proxy_counters"] != 1 {
		t.Fatalf("proxy_counters expected 1, got %d", snap["proxy_counters"])
	}
	if snap["proxy_counters_prom"] != 1 {
		t.Fatalf("proxy_counters_prom expected 1, got %d", snap["proxy_counters_prom"])
	}
	if snap["dashboard_summary"] != 1 {
		t.Fatalf("dashboard_summary expected 1, got %d", snap["dashboard_summary"])
	}
	if snap["dashboard_summary_prom"] != 1 {
		t.Fatalf("dashboard_summary_prom expected 1, got %d", snap["dashboard_summary_prom"])
	}
}

func TestBuildProxyCountersProm(t *testing.T) {
	prom := buildProxyCountersProm(map[string]int{
		"debug_compare_total":          2,
		"debug_compare_prom_total":     3,
		"proxy_counters_total":         7,
		"proxy_counters_prom_total":    8,
		"proxy_alerts_total":           9,
		"proxy_alerts_prom_total":      10,
		"dashboard_summary_total":      4,
		"dashboard_summary_prom_total": 5,
	}, map[string]interface{}{
		"proxy_last_json_unix": int64(100),
		"proxy_last_prom_unix": int64(101),
	})
	required := []string{
		"rechain_web6_debug_compare_total 2",
		"rechain_web6_debug_compare_prom_total 3",
		"rechain_web6_proxy_counters_total 7",
		"rechain_web6_proxy_counters_prom_total 8",
		"rechain_web6_proxy_alerts_total 9",
		"rechain_web6_proxy_alerts_prom_total 10",
		"rechain_web6_proxy_counters_last_json_unix 100",
		"rechain_web6_proxy_counters_last_prom_unix 101",
		"rechain_web6_dashboard_summary_total 4",
		"rechain_web6_dashboard_summary_prom_total 5",
	}
	for _, token := range required {
		if !strings.Contains(prom, token) {
			t.Fatalf("proxy counters prom missing token: %s\n%s", token, prom)
		}
	}
}

func TestProxyCountersHistoryLimit(t *testing.T) {
	h := NewProxyCountersHistory(2)
	h.Add(ProxyCountersEvent{Format: "json", ProxyCountersTotal: 1})
	h.Add(ProxyCountersEvent{Format: "prom", ProxyCountersTotal: 2})
	h.Add(ProxyCountersEvent{Format: "json", ProxyCountersTotal: 3})
	items := h.List(10)
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].ProxyCountersTotal != 2 || items[1].ProxyCountersTotal != 3 {
		t.Fatalf("unexpected proxy history order/content: %#v", items)
	}
}

func TestBuildProxyCountersHistoryProm(t *testing.T) {
	items := []ProxyCountersEvent{
		{Format: "json", ProxyCountersTotal: 3, ProxyCountersProm: 2, DashboardSummary: 5, DashboardSummaryProm: 1},
		{Format: "prom", ProxyCountersTotal: 4, ProxyCountersProm: 7, DashboardSummary: 6, DashboardSummaryProm: 2},
	}
	prom := buildProxyCountersHistoryProm(items, map[string]int{"json": 1, "prom": 1}, "prom")
	required := []string{
		"rechain_web6_proxy_history_total 2",
		`rechain_web6_proxy_history_by_format{format="prom"} 1`,
		`rechain_web6_proxy_history_filter_active{kind="format",value="prom"} 1`,
		`rechain_web6_proxy_history_last_counter{name="proxy_prom_total"} 7`,
		`rechain_web6_proxy_history_last_format{format="prom"} 1`,
	}
	for _, token := range required {
		if !strings.Contains(prom, token) {
			t.Fatalf("proxy history prom missing token: %s\n%s", token, prom)
		}
	}
}

func TestExtractDashboardWeb6Prom(t *testing.T) {
	src := strings.Join([]string{
		`rechain_dashboard_tasks_total{state="completed"} 2`,
		`rechain_dashboard_downstream_up{service="web6"} 1`,
		`rechain_dashboard_downstream_up{service="rag"} 1`,
		`rechain_dashboard_web6_proxy_alert_level 0`,
		`rechain_dashboard_web6_proxy_json_stale 0`,
	}, "\n")
	out := extractDashboardWeb6Prom(src)
	if !strings.Contains(out, `rechain_dashboard_downstream_up{service="web6"} 1`) {
		t.Fatalf("missing web6 downstream_up in filtered output: %s", out)
	}
	if !strings.Contains(out, `rechain_dashboard_web6_proxy_alert_level 0`) {
		t.Fatalf("missing web6 alert level in filtered output: %s", out)
	}
	if strings.Contains(out, `service="rag"`) {
		t.Fatalf("unexpected non-web6 downstream line in filtered output: %s", out)
	}
}

func TestComputeDashboardWeb6Alert(t *testing.T) {
	level, score, _ := computeDashboardWeb6Alert(map[string]interface{}{
		"up":                true,
		"proxy_json_stale":  false,
		"proxy_prom_stale":  false,
		"proxy_alert_level": 0.0,
	})
	if level != "ok" || score != 0 {
		t.Fatalf("expected ok/0, got %s/%d", level, score)
	}
	level, score, _ = computeDashboardWeb6Alert(map[string]interface{}{
		"up":               true,
		"proxy_json_stale": true,
		"proxy_prom_stale": false,
	})
	if level != "warn" || score != 1 {
		t.Fatalf("expected warn/1, got %s/%d", level, score)
	}
	level, score, _ = computeDashboardWeb6Alert(map[string]interface{}{
		"up": false,
	})
	if level != "critical" || score != 2 {
		t.Fatalf("expected critical/2, got %s/%d", level, score)
	}
}

func TestDashboardWeb6AlertHistoryLimit(t *testing.T) {
	h := NewDashboardWeb6AlertHistory(2)
	h.Add(DashboardWeb6AlertEvent{Level: "ok", Source: "alerts"})
	h.Add(DashboardWeb6AlertEvent{Level: "warn", Source: "summary"})
	h.Add(DashboardWeb6AlertEvent{Level: "critical", Source: "alerts"})
	items := h.List(10)
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].Level != "warn" || items[1].Level != "critical" {
		t.Fatalf("unexpected order/content: %#v", items)
	}
}

func TestBuildDashboardWeb6HistoryProm(t *testing.T) {
	items := []DashboardWeb6AlertEvent{
		{Level: "ok", Source: "alerts"},
		{Level: "warn", Source: "summary"},
	}
	prom := buildDashboardWeb6HistoryProm(items, map[string]int{"ok": 1, "warn": 1}, map[string]int{"alerts": 1, "summary": 1}, "warn", "summary")
	required := []string{
		"rechain_web6_dashboard_web6_alert_history_total 2",
		`rechain_web6_dashboard_web6_alert_history_by_level{level="ok"} 1`,
		`rechain_web6_dashboard_web6_alert_history_by_source{source="alerts"} 1`,
		`rechain_web6_dashboard_web6_alert_history_filter_active{kind="level",value="warn"} 1`,
	}
	for _, token := range required {
		if !strings.Contains(prom, token) {
			t.Fatalf("history prom missing token: %s\n%s", token, prom)
		}
	}
}

func TestMatchDashboardWeb6HistoryFilter(t *testing.T) {
	ev := DashboardWeb6AlertEvent{Level: "warn", Source: "summary"}
	if !matchDashboardWeb6HistoryFilter(ev, "warn", "summary") {
		t.Fatalf("expected filter match")
	}
	if matchDashboardWeb6HistoryFilter(ev, "critical", "summary") {
		t.Fatalf("unexpected level match")
	}
	if matchDashboardWeb6HistoryFilter(ev, "warn", "alerts") {
		t.Fatalf("unexpected source match")
	}
	if !matchDashboardWeb6HistoryFilter(ev, "", "") {
		t.Fatalf("expected all/all match")
	}
}
