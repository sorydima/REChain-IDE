package main

import (
  "encoding/json"
  "log"
  "net/http"
  "os"
  "strconv"
  "strings"
  "sync"
  "time"

  "rechain-ide/shared/logging"
)

const schemaVersion = "0.1.0"

type CompileRequest struct {
  SchemaVersion string        `json:"schema_version"`
  TaskID        string        `json:"task_id"`
  Results       []ModelResult `json:"results"`
  Policy        string        `json:"policy"`
}

type ModelResult struct {
  ModelID string `json:"model_id"`
  Output  string `json:"output"`
  Diff    string `json:"diff"`
}

type CompileResponse struct {
  SchemaVersion string  `json:"schema_version"`
  TaskID        string  `json:"task_id"`
  SelectedModel string  `json:"selected_model"`
  Diff          string  `json:"diff"`
  QualityScore  float64 `json:"quality_score"`
  Rationale     string  `json:"rationale"`
  Tests         []string `json:"tests"`
}

type Metrics struct {
  mu        sync.Mutex
  compile   int
  health    int
  latencyMs []int64
}

func (m *Metrics) IncCompile() { m.mu.Lock(); m.compile++; m.mu.Unlock() }
func (m *Metrics) IncHealth()  { m.mu.Lock(); m.health++; m.mu.Unlock() }
func (m *Metrics) Observe(ms int64) {
  m.mu.Lock()
  m.latencyMs = append(m.latencyMs, ms)
  if len(m.latencyMs) > 100 {
    m.latencyMs = m.latencyMs[len(m.latencyMs)-100:]
  }
  m.mu.Unlock()
}

func (m *Metrics) Snapshot() map[string]int {
  m.mu.Lock()
  defer m.mu.Unlock()
  return map[string]int{
    "compile": m.compile,
    "health":  m.health,
  }
}

func main() {
  mux := http.NewServeMux()
  metrics := &Metrics{}

  mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
    metrics.IncHealth()
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("ok"))
  })

  mux.HandleFunc("/compile", func(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
      http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
      return
    }
    start := time.Now()
    metrics.IncCompile()

    var req CompileRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
      http.Error(w, "invalid json", http.StatusBadRequest)
      return
    }
    if req.SchemaVersion == "" {
      req.SchemaVersion = schemaVersion
    }
    if len(req.Results) == 0 {
      http.Error(w, "no results", http.StatusBadRequest)
      return
    }

  sel, score, rationale := pickBest(req.Results, req.Policy)
  tests := suggestTests(sel.Diff)
  resp := CompileResponse{
      SchemaVersion: req.SchemaVersion,
      TaskID:        req.TaskID,
      SelectedModel: sel.ModelID,
      Diff:          sel.Diff,
      QualityScore:  score,
      Rationale:     rationale,
      Tests:         tests,
    }
    metrics.Observe(time.Since(start).Milliseconds())
    writeJSON(w, resp)
  })

  mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/plain; version=0.0.4")
    snap := metrics.Snapshot()
    lines := []string{
      "# HELP rechain_agent_compile_total Compile requests",
      "# TYPE rechain_agent_compile_total counter",
      "rechain_agent_compile_total " + strconv.Itoa(snap["compile"]),
      "# HELP rechain_agent_health_total Health requests",
      "# TYPE rechain_agent_health_total counter",
      "rechain_agent_health_total " + strconv.Itoa(snap["health"]),
    }
    lines = append(lines, renderLatencyHistogram(metrics)...)
    w.Write([]byte(strings.Join(lines, "\n")))
  })

  addr := ":8086"
  log.Printf("agent-compiler listening on %s", addr)
  if err := http.ListenAndServe(addr, logging.WithRequestID(mux)); err != nil {
    log.Fatal(err)
  }
}

func pickBest(results []ModelResult, policy string) (ModelResult, float64, string) {
  best := results[0]
  bestScore := scoreResult(best, policy)
  for _, r := range results[1:] {
    s := scoreResult(r, policy)
    if s > bestScore {
      best = r
      bestScore = s
    }
  }
  rationale := "selected highest quality score"
  if strings.EqualFold(policy, "min_diff") {
    rationale = "selected smallest diff footprint"
  }
  return best, bestScore, rationale
}

func scoreResult(r ModelResult, policy string) float64 {
  stats := diffStats(r.Diff)
  errCount := errorTokenCount(r.Output)
  if strings.EqualFold(policy, "min_diff") {
    return 1.0 / float64(1+stats.totalLines)
  }
  sizeScore := 1.0 / float64(1+stats.totalLines)
  churnScore := 1.0 / float64(1+stats.additions+stats.deletions)
  errScore := 1.0 / float64(1+errCount)

  wSize := envFloat("AGENT_SCORE_WEIGHT_SIZE", 0.4)
  wChurn := envFloat("AGENT_SCORE_WEIGHT_CHURN", 0.3)
  wErr := envFloat("AGENT_SCORE_WEIGHT_ERRORS", 0.3)
  total := wSize + wChurn + wErr
  if total <= 0 {
    total = 1.0
  }
  score := (wSize*sizeScore + wChurn*churnScore + wErr*errScore) / total
  return clamp(score, 0, 1)
}

type diffStat struct {
  files      int
  hunks      int
  additions  int
  deletions  int
  totalLines int
}

func diffStats(diff string) diffStat {
  s := diffStat{}
  if diff == "" {
    return s
  }
  lines := strings.Split(diff, "\n")
  s.totalLines = len(lines)
  for _, line := range lines {
    if strings.HasPrefix(line, "diff --git") {
      s.files++
      continue
    }
    if strings.HasPrefix(line, "@@") {
      s.hunks++
      continue
    }
    if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
      s.additions++
      continue
    }
    if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
      s.deletions++
      continue
    }
  }
  return s
}

func errorTokenCount(text string) int {
  lower := strings.ToLower(text)
  tokens := []string{"error", "exception", "failed", "panic"}
  count := 0
  for _, t := range tokens {
    count += strings.Count(lower, t)
  }
  return count
}

func suggestTests(diff string) []string {
  if diff == "" {
    return []string{"go test ./..."}
  }
  tests := []string{"go test ./..."}
  if strings.Contains(diff, "package.json") {
    tests = append(tests, "npm test")
  }
  if strings.Contains(diff, ".csproj") {
    tests = append(tests, "dotnet test")
  }
  return tests
}

func envFloat(key string, fallback float64) float64 {
  v := strings.TrimSpace(os.Getenv(key))
  if v == "" {
    return fallback
  }
  f, err := strconv.ParseFloat(v, 64)
  if err != nil {
    return fallback
  }
  return f
}

func clamp(v float64, min float64, max float64) float64 {
  if v < min {
    return min
  }
  if v > max {
    return max
  }
  return v
}

func renderLatencyHistogram(m *Metrics) []string {
  buckets := []int{100, 250, 500, 1000, 2000, 5000}
  counts := make([]int, len(buckets)+1)
  total := 0

  m.mu.Lock()
  samples := append([]int64{}, m.latencyMs...)
  m.mu.Unlock()

  for _, v := range samples {
    total++
    placed := false
    for i, b := range buckets {
      if int(v) <= b {
        counts[i]++
        placed = true
        break
      }
    }
    if !placed {
      counts[len(counts)-1]++
    }
  }

  lines := []string{
    "# HELP rechain_agent_compile_latency_ms Compile latency histogram",
    "# TYPE rechain_agent_compile_latency_ms histogram",
  }
  running := 0
  for i, b := range buckets {
    running += counts[i]
    lines = append(lines, "rechain_agent_compile_latency_ms_bucket{le=\""+strconv.Itoa(b)+"\"} "+strconv.Itoa(running))
  }
  running += counts[len(counts)-1]
  lines = append(lines, "rechain_agent_compile_latency_ms_bucket{le=\"+Inf\"} "+strconv.Itoa(running))
  lines = append(lines, "rechain_agent_compile_latency_ms_count "+strconv.Itoa(total))
  return lines
}

func writeJSON(w http.ResponseWriter, v interface{}) {
  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(v)
}
