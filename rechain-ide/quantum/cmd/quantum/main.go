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

type OptimizeRequest struct {
  SchemaVersion string     `json:"schema_version"`
  Candidates    []Candidate `json:"candidates"`
  Objective     string     `json:"objective"`
}

type Candidate struct {
  ID        string  `json:"id"`
  CostUSD   float64 `json:"cost_usd"`
  LatencyMs float64 `json:"latency_ms"`
  Quality   float64 `json:"quality"`
}

type OptimizeResponse struct {
  SchemaVersion string   `json:"schema_version"`
  SelectedID    string   `json:"selected_id"`
  Score         float64  `json:"score"`
  Rationale     string   `json:"rationale"`
  Evaluated     int      `json:"evaluated"`
}

type Metrics struct {
  mu          sync.Mutex
  optimize    int
  health      int
  latencyMs   []int64
}

func (m *Metrics) IncOptimize() { m.mu.Lock(); m.optimize++; m.mu.Unlock() }
func (m *Metrics) IncHealth()   { m.mu.Lock(); m.health++; m.mu.Unlock() }
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
    "optimize": m.optimize,
    "health":   m.health,
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

  mux.HandleFunc("/optimize", func(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
      http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
      return
    }
    start := time.Now()
    metrics.IncOptimize()

    var req OptimizeRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
      http.Error(w, "invalid json", http.StatusBadRequest)
      return
    }
    if req.SchemaVersion == "" {
      req.SchemaVersion = schemaVersion
    }
    if len(req.Candidates) == 0 {
      http.Error(w, "no candidates", http.StatusBadRequest)
      return
    }

    selected, score, rationale := selectCandidate(req.Candidates, req.Objective)
    resp := OptimizeResponse{
      SchemaVersion: req.SchemaVersion,
      SelectedID:    selected,
      Score:         score,
      Rationale:     rationale,
      Evaluated:     len(req.Candidates),
    }
    metrics.Observe(time.Since(start).Milliseconds())
    writeJSON(w, resp)
  })

  mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/plain; version=0.0.4")
    snap := metrics.Snapshot()
    lines := []string{
      "# HELP rechain_quantum_optimize_total Optimize requests",
      "# TYPE rechain_quantum_optimize_total counter",
      "rechain_quantum_optimize_total " + strconv.Itoa(snap["optimize"]),
      "# HELP rechain_quantum_health_total Health requests",
      "# TYPE rechain_quantum_health_total counter",
      "rechain_quantum_health_total " + strconv.Itoa(snap["health"]),
    }
    lines = append(lines, renderLatencyHistogram(metrics)...)
    w.Write([]byte(strings.Join(lines, "\n")))
  })

  addr := ":8085"
  log.Printf("quantum listening on %s", addr)
  if err := http.ListenAndServe(addr, logging.WithRequestID(mux)); err != nil {
    log.Fatal(err)
  }
}

func selectCandidate(cands []Candidate, objective string) (string, float64, string) {
  best := cands[0]
  minCost, maxCost := minMax(cands, "cost")
  minLat, maxLat := minMax(cands, "latency")
  minQ, maxQ := minMax(cands, "quality")
  bestScore := scoreCandidate(best, objective, minCost, maxCost, minLat, maxLat, minQ, maxQ)
  for _, c := range cands[1:] {
    s := scoreCandidate(c, objective, minCost, maxCost, minLat, maxLat, minQ, maxQ)
    if s < bestScore {
      best = c
      bestScore = s
    }
  }
  rationale := "minimized " + objective
  if objective == "" {
    rationale = "minimized weighted score"
  }
  return best.ID, bestScore, rationale
}

func scoreCandidate(c Candidate, objective string, minCost float64, maxCost float64, minLat float64, maxLat float64, minQ float64, maxQ float64) float64 {
  switch strings.ToLower(objective) {
  case "cost":
    return c.CostUSD
  case "latency":
    return c.LatencyMs
  case "quality":
    return 1 - c.Quality
  default:
    wCost := envFloat("QUANTUM_WEIGHT_COST", 0.3)
    wLat := envFloat("QUANTUM_WEIGHT_LATENCY", 0.5)
    wQual := envFloat("QUANTUM_WEIGHT_QUALITY", 0.2)
    total := wCost + wLat + wQual
    if total <= 0 {
      total = 1
    }
    nCost := normalize(c.CostUSD, minCost, maxCost)
    nLat := normalize(c.LatencyMs, minLat, maxLat)
    nQual := normalize(c.Quality, minQ, maxQ)
    return (wCost*nCost + wLat*nLat + wQual*(1-nQual)) / total
  }
}

func minMax(cands []Candidate, metric string) (float64, float64) {
  min := 0.0
  max := 0.0
  for i, c := range cands {
    v := 0.0
    switch metric {
    case "cost":
      v = c.CostUSD
    case "latency":
      v = c.LatencyMs
    case "quality":
      v = c.Quality
    }
    if i == 0 || v < min {
      min = v
    }
    if i == 0 || v > max {
      max = v
    }
  }
  return min, max
}

func normalize(v float64, min float64, max float64) float64 {
  if max-min == 0 {
    return 0
  }
  return (v - min) / (max - min)
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
    "# HELP rechain_quantum_optimize_latency_ms Optimize latency histogram",
    "# TYPE rechain_quantum_optimize_latency_ms histogram",
  }
  running := 0
  for i, b := range buckets {
    running += counts[i]
    lines = append(lines, "rechain_quantum_optimize_latency_ms_bucket{le=\""+strconv.Itoa(b)+"\"} "+strconv.Itoa(running))
  }
  running += counts[len(counts)-1]
  lines = append(lines, "rechain_quantum_optimize_latency_ms_bucket{le=\"+Inf\"} "+strconv.Itoa(running))
  lines = append(lines, "rechain_quantum_optimize_latency_ms_count "+strconv.Itoa(total))
  return lines
}

func writeJSON(w http.ResponseWriter, v interface{}) {
  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(v)
}
