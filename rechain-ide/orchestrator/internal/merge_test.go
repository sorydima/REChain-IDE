package internal

import "testing"

func TestMergeResults_SelectsLowestLatency(t *testing.T) {
  results := []ModelResult{
    {Diff: "b", Metrics: []Metric{{Name: "latency_ms", Value: 150}}},
    {Diff: "a", Metrics: []Metric{{Name: "latency_ms", Value: 120}}},
  }

  best, ok := MergeResults(results, "latency_ms", false)
  if !ok {
    t.Fatal("expected ok")
  }
  if best.Diff != "a" {
    t.Fatalf("expected diff a, got %s", best.Diff)
  }
}

func TestMergeResults_TieBreakByDiff(t *testing.T) {
  results := []ModelResult{
    {Diff: "b", Metrics: []Metric{{Name: "latency_ms", Value: 120}}},
    {Diff: "a", Metrics: []Metric{{Name: "latency_ms", Value: 120}}},
  }

  best, ok := MergeResults(results, "latency_ms", false)
  if !ok {
    t.Fatal("expected ok")
  }
  if best.Diff != "a" {
    t.Fatalf("expected diff a, got %s", best.Diff)
  }
}

func TestMergeResults_SelectsLowestCost(t *testing.T) {
  results := []ModelResult{
    {Diff: "b", Metrics: []Metric{{Name: "cost_usd", Value: 0.02}}},
    {Diff: "a", Metrics: []Metric{{Name: "cost_usd", Value: 0.01}}},
  }

  best, ok := MergeResults(results, "cost_usd", false)
  if !ok {
    t.Fatal("expected ok")
  }
  if best.Diff != "a" {
    t.Fatalf("expected diff a, got %s", best.Diff)
  }
}

func TestMergeResults_SelectsHighestQuality(t *testing.T) {
  results := []ModelResult{
    {Diff: "b", Metrics: []Metric{{Name: "quality_score", Value: 0.6}}},
    {Diff: "a", Metrics: []Metric{{Name: "quality_score", Value: 0.9}}},
  }

  best, ok := MergeResults(results, "quality_score", true)
  if !ok {
    t.Fatal("expected ok")
  }
  if best.Diff != "a" {
    t.Fatalf("expected diff a, got %s", best.Diff)
  }
}
