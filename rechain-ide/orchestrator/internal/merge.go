package internal

type Metric struct {
  Name  string
  Value float64
}

type ModelResult struct {
  Diff    string
  Metrics []Metric
}

func MergeResults(results []ModelResult, metric string, preferHigher bool) (ModelResult, bool) {
  if len(results) == 0 {
    return ModelResult{}, false
  }

  best := results[0]
  bestValue := metricOf(best, metric)
  for _, r := range results[1:] {
    v := metricOf(r, metric)
    if preferHigher {
      if v > bestValue || (v == bestValue && r.Diff < best.Diff) {
        best = r
        bestValue = v
      }
      continue
    }
    if v < bestValue || (v == bestValue && r.Diff < best.Diff) {
      best = r
      bestValue = v
    }
  }

  return best, true
}

func metricOf(r ModelResult, name string) float64 {
  for _, m := range r.Metrics {
    if m.Name == name {
      return m.Value
    }
  }
  return 0
}
