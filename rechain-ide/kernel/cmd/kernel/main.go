package main

import (
  "context"
  "encoding/json"
  "errors"
  "log"
  "net/http"
  "os"
  "os/exec"
  "regexp"
  "runtime"
  "strconv"
  "strings"
  "sync"
  "time"

  "rechain-ide/shared/logging"
)

const schemaVersion = "0.1.0"

type ExecSpec struct {
  SchemaVersion string   `json:"schema_version"`
  ID            string   `json:"id"`
  Command       string   `json:"command"`
  Args          []string `json:"args"`
  TimeoutMs     int      `json:"timeout_ms"`
}

type ExecResult struct {
  SchemaVersion string `json:"schema_version"`
  ID            string `json:"id"`
  Allowed       bool   `json:"allowed"`
  Output        string `json:"output"`
  Error         string `json:"error"`
  StartedAt     string `json:"started_at"`
  CompletedAt   string `json:"completed_at"`
  ExitCode      int    `json:"exit_code"`
}

type PolicyDecision struct {
  Allowed bool   `json:"allowed"`
  Reason  string `json:"reason"`
}

type Metrics struct {
  mu        sync.Mutex
  runs      int
  allowed   int
  denied    int
  errors    int
  runLatencyMs []int64
}

func (m *Metrics) IncRuns()    { m.mu.Lock(); m.runs++; m.mu.Unlock() }
func (m *Metrics) IncAllowed() { m.mu.Lock(); m.allowed++; m.mu.Unlock() }
func (m *Metrics) IncDenied()  { m.mu.Lock(); m.denied++; m.mu.Unlock() }
func (m *Metrics) IncErrors()  { m.mu.Lock(); m.errors++; m.mu.Unlock() }

func (m *Metrics) ObserveLatency(ms int64) {
  m.mu.Lock()
  m.runLatencyMs = append(m.runLatencyMs, ms)
  if len(m.runLatencyMs) > 100 {
    m.runLatencyMs = m.runLatencyMs[len(m.runLatencyMs)-100:]
  }
  m.mu.Unlock()
}

func (m *Metrics) Snapshot() map[string]int {
  m.mu.Lock()
  defer m.mu.Unlock()
  return map[string]int{
    "runs":    m.runs,
    "allowed": m.allowed,
    "denied":  m.denied,
    "errors":  m.errors,
  }
}

func main() {
  mux := http.NewServeMux()
  metrics := &Metrics{}
  mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("ok"))
  })

  mux.HandleFunc("/run", func(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
      http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
      return
    }

    metrics.IncRuns()
    start := time.Now()
    var spec ExecSpec
    if err := json.NewDecoder(r.Body).Decode(&spec); err != nil {
      http.Error(w, "invalid json", http.StatusBadRequest)
      return
    }

    if spec.SchemaVersion == "" {
      spec.SchemaVersion = schemaVersion
    }

    decision := enforcePolicy(spec)
    now := time.Now().UTC().Format(time.RFC3339)
    result := ExecResult{
      SchemaVersion: schemaVersion,
      ID:            spec.ID,
      Allowed:       decision.Allowed,
      Output:        "",
      Error:         "",
      StartedAt:     now,
      CompletedAt:   now,
      ExitCode:      0,
    }

    if !decision.Allowed {
      result.Error = decision.Reason
      metrics.IncDenied()
      metrics.ObserveLatency(time.Since(start).Milliseconds())
      writeJSON(w, result)
      return
    }
    metrics.IncAllowed()

    timeout := 2000 * time.Millisecond
    if spec.TimeoutMs > 0 {
      maxMs := maxTimeoutMs()
      if spec.TimeoutMs > maxMs {
        http.Error(w, "timeout exceeds max", http.StatusBadRequest)
        return
      }
      timeout = time.Duration(spec.TimeoutMs) * time.Millisecond
    }

    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()

    out, code, err := runSandboxed(ctx, spec.Command, spec.Args)
    result.CompletedAt = time.Now().UTC().Format(time.RFC3339)
    result.Output = out
    result.ExitCode = code
    if err != nil {
      result.Error = err.Error()
      metrics.IncErrors()
    }

    metrics.ObserveLatency(time.Since(start).Milliseconds())
    writeJSON(w, result)
  })

  mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/plain; version=0.0.4")
    snap := metrics.Snapshot()
    lines := []string{
      "# HELP rechain_kernel_runs_total Total run requests",
      "# TYPE rechain_kernel_runs_total counter",
      "rechain_kernel_runs_total " + strconv.Itoa(snap["runs"]),
      "# HELP rechain_kernel_allowed_total Allowed runs",
      "# TYPE rechain_kernel_allowed_total counter",
      "rechain_kernel_allowed_total " + strconv.Itoa(snap["allowed"]),
      "# HELP rechain_kernel_denied_total Denied runs",
      "# TYPE rechain_kernel_denied_total counter",
      "rechain_kernel_denied_total " + strconv.Itoa(snap["denied"]),
      "# HELP rechain_kernel_errors_total Run errors",
      "# TYPE rechain_kernel_errors_total counter",
      "rechain_kernel_errors_total " + strconv.Itoa(snap["errors"]),
    }
    lines = append(lines, renderKernelLatencyHistogram(metrics)...)
    w.Write([]byte(strings.Join(lines, "\n")))
  })

  addr := ":8082"
  log.Printf("kernel listening on %s", addr)
  if err := http.ListenAndServe(addr, logging.WithRequestID(mux)); err != nil {
    log.Fatal(err)
  }
}

func renderKernelLatencyHistogram(m *Metrics) []string {
  buckets := []int{10, 50, 100, 250, 500, 1000, 2000, 5000}
  counts := make([]int, len(buckets)+1)
  total := 0

  m.mu.Lock()
  samples := append([]int64{}, m.runLatencyMs...)
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
    "# HELP rechain_kernel_run_latency_ms Kernel run latency histogram",
    "# TYPE rechain_kernel_run_latency_ms histogram",
  }
  running := 0
  for i, b := range buckets {
    running += counts[i]
    lines = append(lines, "rechain_kernel_run_latency_ms_bucket{le=\""+strconv.Itoa(b)+"\"} "+strconv.Itoa(running))
  }
  running += counts[len(counts)-1]
  lines = append(lines, "rechain_kernel_run_latency_ms_bucket{le=\"+Inf\"} "+strconv.Itoa(running))
  lines = append(lines, "rechain_kernel_run_latency_ms_count "+strconv.Itoa(total))
  return lines
}

func enforcePolicy(spec ExecSpec) PolicyDecision {
  if spec.Command == "" {
    return PolicyDecision{Allowed: false, Reason: "empty command"}
  }
  if isDeniedCommand(spec.Command) {
    return PolicyDecision{Allowed: false, Reason: "command denied"}
  }
  if !isAllowedCommand(spec.Command) {
    return PolicyDecision{Allowed: false, Reason: "command not allowed"}
  }
  if !argsSafe(spec.Args) {
    return PolicyDecision{Allowed: false, Reason: "unsafe args"}
  }
  return PolicyDecision{Allowed: true, Reason: ""}
}

func isAllowedCommand(cmd string) bool {
  allowed := map[string]bool{}
  list := strings.TrimSpace(os.Getenv("KERNEL_ALLOWLIST"))
  if list == "" {
    list = "echo"
  }
  for _, c := range strings.Split(list, ",") {
    allowed[strings.ToLower(strings.TrimSpace(c))] = true
  }
  return allowed[strings.ToLower(cmd)]
}

func isDeniedCommand(cmd string) bool {
  denied := map[string]bool{}
  list := strings.TrimSpace(os.Getenv("KERNEL_DENYLIST"))
  for _, c := range strings.Split(list, ",") {
    if strings.TrimSpace(c) != "" {
      denied[strings.ToLower(strings.TrimSpace(c))] = true
    }
  }
  return denied[strings.ToLower(cmd)]
}

func argsSafe(args []string) bool {
  re := regexp.MustCompile(`^[a-zA-Z0-9_\-\. ]*$`)
  for _, a := range args {
    if !re.MatchString(a) {
      return false
    }
  }
  return true
}

func runSandboxed(ctx context.Context, cmd string, args []string) (string, int, error) {
  if runtime.GOOS == "windows" {
    cmdline := cmd
    if len(args) > 0 {
      cmdline = cmdline + " " + strings.Join(args, " ")
    }
    c := exec.CommandContext(ctx, "cmd", "/C", cmdline)
    out, err := c.CombinedOutput()
    return string(out), exitCode(c, err), err
  }

  c := exec.CommandContext(ctx, cmd, args...)
  out, err := c.CombinedOutput()
  return string(out), exitCode(c, err), err
}

func exitCode(c *exec.Cmd, err error) int {
  if err == nil {
    return 0
  }
  var ee *exec.ExitError
  if errors.As(err, &ee) {
    return ee.ExitCode()
  }
  return 1
}

func writeJSON(w http.ResponseWriter, v interface{}) {
  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(v)
}

func maxTimeoutMs() int {
  v := strings.TrimSpace(os.Getenv("KERNEL_MAX_TIMEOUT_MS"))
  if v == "" {
    return 5000
  }
  n, err := strconv.Atoi(v)
  if err != nil || n <= 0 {
    return 5000
  }
  return n
}
