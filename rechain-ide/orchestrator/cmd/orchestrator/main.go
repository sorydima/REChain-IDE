package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"rechain-ide/orchestrator/internal"
	"rechain-ide/shared/logging"
)

const schemaVersion = "0.1.0"

type TaskSpec struct {
	SchemaVersion string       `json:"schema_version"`
	ID            string       `json:"id"`
	Type          string       `json:"type"`
	Input         string       `json:"input"`
	Context       []ContextRef `json:"context"`
	Constraints   []Constraint `json:"constraints"`
	Metadata      Metadata     `json:"metadata"`
}

type ContextRef struct {
	Type string `json:"type"`
	Path string `json:"path"`
	Rev  string `json:"rev"`
}

type Constraint struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

type Metadata struct {
	Requester string `json:"requester"`
	Priority  string `json:"priority"`
}

type TaskStatus struct {
	SchemaVersion string  `json:"schema_version"`
	ID            string  `json:"id"`
	State         string  `json:"state"`
	Progress      float64 `json:"progress"`
	StartedAt     string  `json:"started_at"`
	UpdatedAt     string  `json:"updated_at"`
}

type ModelResult struct {
	SchemaVersion string   `json:"schema_version"`
	ModelID       string   `json:"model_id"`
	Output        string   `json:"output"`
	Diff          string   `json:"diff"`
	Metrics       []Metric `json:"metrics"`
}

type Metric struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

type MergeResult struct {
	SchemaVersion string  `json:"schema_version"`
	Diff          string  `json:"diff"`
	Rationale     string  `json:"rationale"`
	Confidence    float64 `json:"confidence"`
	QualityScore  float64 `json:"quality_score"`
}

type TraceModelResult struct {
	ModelID      string  `json:"model_id"`
	DiffLen      int     `json:"diff_len"`
	LatencyMs    float64 `json:"latency_ms"`
	CostUSD      float64 `json:"cost_usd"`
	QualityScore float64 `json:"quality_score"`
}

type TaskTrace struct {
	SchemaVersion string             `json:"schema_version"`
	TaskID        string             `json:"task_id"`
	ParentTaskID  string             `json:"parent_task_id,omitempty"`
	State         string             `json:"state"`
	StartedAt     string             `json:"started_at"`
	FinishedAt    string             `json:"finished_at,omitempty"`
	RoutingPolicy string             `json:"routing_policy"`
	Selected      []string           `json:"selected_models,omitempty"`
	Results       []TraceModelResult `json:"results,omitempty"`
	MergeSource   string             `json:"merge_source,omitempty"`
	Merge         *MergeResult       `json:"merge,omitempty"`
	Error         string             `json:"error,omitempty"`
}

type Artifact struct {
	SchemaVersion string `json:"schema_version"`
	ID            string `json:"id"`
	Type          string `json:"type"`
	Path          string `json:"path"`
	Sha256        string `json:"sha256"`
	CreatedAt     string `json:"created_at"`
}

type TaskStore struct {
	mu        sync.Mutex
	statuses  map[string]TaskStatus
	artifacts map[string][]Artifact
	results   map[string]MergeResult
	traces    map[string]TaskTrace
	specs     map[string]TaskSpec
}

func (s *TaskStore) TraceMetrics() (map[string]int, map[string]int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	byState := map[string]int{}
	byMerge := map[string]int{}
	for _, tr := range s.traces {
		state := strings.TrimSpace(tr.State)
		if state == "" {
			state = "unknown"
		}
		byState[state]++
		source := strings.TrimSpace(tr.MergeSource)
		if source == "" {
			source = "none"
		}
		byMerge[source]++
	}
	return byState, byMerge
}

func (s *TaskStore) TraceParentLinks() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	total := 0
	for _, tr := range s.traces {
		if strings.TrimSpace(tr.ParentTaskID) != "" {
			total++
		}
	}
	return total
}

func (s *TaskStore) LatestTrace() (TaskTrace, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var best TaskTrace
	bestKey := ""
	for _, tr := range s.traces {
		key := tr.FinishedAt
		if key == "" {
			key = tr.StartedAt
		}
		if key > bestKey {
			best = tr
			bestKey = key
		}
	}
	if bestKey == "" {
		return TaskTrace{}, false
	}
	return best, true
}

type Metrics struct {
	mu             sync.Mutex
	submitted      int
	replayed       int
	forcedFallback int
	replayMode     map[string]int
	completed      int
	failed         int
	canceled       int
	hfErrors       int
	taskLatencyMs  []int64
	routingCounts  map[string]int
	routingByModel map[string]map[string]int
	mergeChoice    map[string]int
	modelLatencyMs map[string][]int64
	retries        int
	queueDelayMs   []int64
}

func (m *Metrics) IncSubmitted() {
	m.mu.Lock()
	m.submitted++
	m.mu.Unlock()
}

func (m *Metrics) IncReplayed() {
	m.mu.Lock()
	m.replayed++
	m.mu.Unlock()
}

func (m *Metrics) IncReplayMode(mode string) {
	mode = strings.ToLower(strings.TrimSpace(mode))
	if mode == "" {
		mode = "default"
	}
	m.mu.Lock()
	if m.replayMode == nil {
		m.replayMode = map[string]int{}
	}
	m.replayMode[mode]++
	m.mu.Unlock()
}

func (m *Metrics) IncForcedFallback() {
	m.mu.Lock()
	m.forcedFallback++
	m.mu.Unlock()
}

func (m *Metrics) IncCompleted() {
	m.mu.Lock()
	m.completed++
	m.mu.Unlock()
}

func (m *Metrics) IncFailed() {
	m.mu.Lock()
	m.failed++
	m.mu.Unlock()
}

func (m *Metrics) IncCanceled() {
	m.mu.Lock()
	m.canceled++
	m.mu.Unlock()
}

func (m *Metrics) IncHFError() {
	m.mu.Lock()
	m.hfErrors++
	m.mu.Unlock()
}

func (m *Metrics) ObserveLatency(ms int64) {
	m.mu.Lock()
	m.taskLatencyMs = append(m.taskLatencyMs, ms)
	if len(m.taskLatencyMs) > 100 {
		m.taskLatencyMs = m.taskLatencyMs[len(m.taskLatencyMs)-100:]
	}
	m.mu.Unlock()
}

func (m *Metrics) IncRetry() {
	m.mu.Lock()
	m.retries++
	m.mu.Unlock()
}

func (m *Metrics) ObserveQueueDelay(ms int64) {
	m.mu.Lock()
	m.queueDelayMs = append(m.queueDelayMs, ms)
	if len(m.queueDelayMs) > 100 {
		m.queueDelayMs = m.queueDelayMs[len(m.queueDelayMs)-100:]
	}
	m.mu.Unlock()
}

func (m *Metrics) ObserveModelLatency(model string, ms int64) {
	if model == "" || ms <= 0 {
		return
	}
	m.mu.Lock()
	if m.modelLatencyMs == nil {
		m.modelLatencyMs = map[string][]int64{}
	}
	m.modelLatencyMs[model] = append(m.modelLatencyMs[model], ms)
	if len(m.modelLatencyMs[model]) > 100 {
		m.modelLatencyMs[model] = m.modelLatencyMs[model][len(m.modelLatencyMs[model])-100:]
	}
	m.mu.Unlock()
}

func (m *Metrics) Snapshot() map[string]int {
	m.mu.Lock()
	defer m.mu.Unlock()
	avg := 0
	if len(m.taskLatencyMs) > 0 {
		sum := int64(0)
		for _, v := range m.taskLatencyMs {
			sum += v
		}
		avg = int(sum / int64(len(m.taskLatencyMs)))
	}
	qavg := 0
	if len(m.queueDelayMs) > 0 {
		sum := int64(0)
		for _, v := range m.queueDelayMs {
			sum += v
		}
		qavg = int(sum / int64(len(m.queueDelayMs)))
	}
	return map[string]int{
		"submitted":          m.submitted,
		"replayed":           m.replayed,
		"forced_fallback":    m.forcedFallback,
		"completed":          m.completed,
		"failed":             m.failed,
		"canceled":           m.canceled,
		"hf_errors":          m.hfErrors,
		"retries":            m.retries,
		"latency_avg_ms":     avg,
		"queue_delay_avg_ms": qavg,
	}
}

func (m *Metrics) IncMergeChoice(source string) {
	if strings.TrimSpace(source) == "" {
		source = "unknown"
	}
	m.mu.Lock()
	if m.mergeChoice == nil {
		m.mergeChoice = map[string]int{}
	}
	m.mergeChoice[source]++
	m.mu.Unlock()
}

func (m *Metrics) ReplayModeSnapshot() map[string]int {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := map[string]int{}
	for k, v := range m.replayMode {
		out[k] = v
	}
	return out
}

func (m *Metrics) MergeChoiceSnapshot() map[string]int {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := map[string]int{}
	for k, v := range m.mergeChoice {
		out[k] = v
	}
	return out
}

func (m *Metrics) IncRouting(policy string) {
	m.mu.Lock()
	if m.routingCounts == nil {
		m.routingCounts = map[string]int{}
	}
	if policy == "" {
		policy = "latency"
	}
	m.routingCounts[policy]++
	m.mu.Unlock()
}

func (m *Metrics) IncRoutingModel(policy string, model string) {
	m.mu.Lock()
	if m.routingByModel == nil {
		m.routingByModel = map[string]map[string]int{}
	}
	if m.routingByModel[model] == nil {
		m.routingByModel[model] = map[string]int{}
	}
	if policy == "" {
		policy = "latency"
	}
	m.routingByModel[model][policy]++
	m.mu.Unlock()
}

func (m *Metrics) RoutingSnapshot() map[string]int {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := map[string]int{}
	for k, v := range m.routingCounts {
		out[k] = v
	}
	return out
}

func (m *Metrics) RoutingByModelSnapshot() map[string]map[string]int {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := map[string]map[string]int{}
	for model, policies := range m.routingByModel {
		out[model] = map[string]int{}
		for p, v := range policies {
			out[model][p] = v
		}
	}
	return out
}

func NewTaskStore() *TaskStore {
	return &TaskStore{
		statuses:  make(map[string]TaskStatus),
		artifacts: make(map[string][]Artifact),
		results:   make(map[string]MergeResult),
		traces:    make(map[string]TaskTrace),
		specs:     make(map[string]TaskSpec),
	}
}

type TaskSummary struct {
	ID           string  `json:"id"`
	ParentTaskID string  `json:"parent_task_id,omitempty"`
	State        string  `json:"state"`
	UpdatedAt    string  `json:"updated_at"`
	MergeSource  string  `json:"merge_source,omitempty"`
	QualityScore float64 `json:"quality_score,omitempty"`
}

type ReplayChain struct {
	SchemaVersion string        `json:"schema_version"`
	TaskID        string        `json:"task_id"`
	Lineage       []TaskSummary `json:"lineage"`
	Descendants   []TaskSummary `json:"descendants"`
}

func (s *TaskStore) RecentTasks(limit int, stateFilter string, mergeSourceFilter string, hasParent string, sortBy string) []TaskSummary {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]TaskSummary, 0, len(s.statuses))
	for id, st := range s.statuses {
		tr := s.traces[id]
		item := TaskSummary{
			ID:           id,
			ParentTaskID: tr.ParentTaskID,
			State:        st.State,
			UpdatedAt:    st.UpdatedAt,
			MergeSource:  tr.MergeSource,
		}
		if tr.Merge != nil {
			item.QualityScore = tr.Merge.QualityScore
		}
		if stateFilter != "" && stateFilter != "all" && !strings.EqualFold(item.State, stateFilter) {
			continue
		}
		if mergeSourceFilter != "" && mergeSourceFilter != "all" && !strings.EqualFold(item.MergeSource, mergeSourceFilter) {
			continue
		}
		if hasParent == "yes" && strings.TrimSpace(item.ParentTaskID) == "" {
			continue
		}
		if hasParent == "no" && strings.TrimSpace(item.ParentTaskID) != "" {
			continue
		}
		out = append(out, item)
	}
	switch strings.ToLower(strings.TrimSpace(sortBy)) {
	case "quality_desc":
		sort.Slice(out, func(i, j int) bool { return out[i].QualityScore > out[j].QualityScore })
	case "quality_asc":
		sort.Slice(out, func(i, j int) bool { return out[i].QualityScore < out[j].QualityScore })
	case "updated_asc":
		sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt < out[j].UpdatedAt })
	default:
		sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt > out[j].UpdatedAt })
	}
	if len(out) > limit {
		out = out[:limit]
	}
	return out
}

func (s *TaskStore) ReplayChain(taskID string) (ReplayChain, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.statuses[taskID]; !ok {
		return ReplayChain{}, false
	}
	summaryOf := func(id string) TaskSummary {
		st := s.statuses[id]
		tr := s.traces[id]
		item := TaskSummary{
			ID:           id,
			ParentTaskID: tr.ParentTaskID,
			State:        st.State,
			UpdatedAt:    st.UpdatedAt,
			MergeSource:  tr.MergeSource,
		}
		if tr.Merge != nil {
			item.QualityScore = tr.Merge.QualityScore
		}
		return item
	}

	lineageReversed := []TaskSummary{}
	seen := map[string]bool{}
	cur := taskID
	for cur != "" && !seen[cur] {
		seen[cur] = true
		lineageReversed = append(lineageReversed, summaryOf(cur))
		parent := strings.TrimSpace(s.traces[cur].ParentTaskID)
		cur = parent
	}
	lineage := make([]TaskSummary, 0, len(lineageReversed))
	for i := len(lineageReversed) - 1; i >= 0; i-- {
		lineage = append(lineage, lineageReversed[i])
	}

	children := map[string][]string{}
	for id, tr := range s.traces {
		p := strings.TrimSpace(tr.ParentTaskID)
		if p != "" {
			children[p] = append(children[p], id)
		}
	}
	descendants := []TaskSummary{}
	queue := append([]string{}, children[taskID]...)
	seenDesc := map[string]bool{}
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		if seenDesc[id] {
			continue
		}
		seenDesc[id] = true
		if _, ok := s.statuses[id]; !ok {
			continue
		}
		descendants = append(descendants, summaryOf(id))
		queue = append(queue, children[id]...)
	}
	sort.Slice(descendants, func(i, j int) bool {
		return descendants[i].UpdatedAt > descendants[j].UpdatedAt
	})

	return ReplayChain{
		SchemaVersion: schemaVersion,
		TaskID:        taskID,
		Lineage:       lineage,
		Descendants:   descendants,
	}, true
}

type queuedTask struct {
	id       string
	spec     TaskSpec
	enqueued time.Time
}

type TaskQueue struct {
	high   chan queuedTask
	normal chan queuedTask
	low    chan queuedTask
}

func NewTaskQueue(size int) *TaskQueue {
	if size <= 0 {
		size = 200
	}
	return &TaskQueue{
		high:   make(chan queuedTask, size),
		normal: make(chan queuedTask, size),
		low:    make(chan queuedTask, size),
	}
}

func (q *TaskQueue) Enqueue(t queuedTask) error {
	switch strings.ToLower(strings.TrimSpace(t.spec.Metadata.Priority)) {
	case "high":
		q.high <- t
	case "low":
		q.low <- t
	default:
		q.normal <- t
	}
	return nil
}

func (q *TaskQueue) Dequeue(ctx context.Context) (queuedTask, bool) {
	for {
		select {
		case t := <-q.high:
			return t, true
		default:
		}
		select {
		case t := <-q.high:
			return t, true
		case t := <-q.normal:
			return t, true
		case t := <-q.low:
			return t, true
		case <-ctx.Done():
			return queuedTask{}, false
		}
	}
}

func (q *TaskQueue) Depth() int {
	return len(q.high) + len(q.normal) + len(q.low)
}

func enqueueReplayTask(store *TaskStore, queue *TaskQueue, metrics *Metrics, parentID string, mode string) (string, TaskStatus, error) {
	parentID = strings.TrimSpace(parentID)
	if parentID == "" {
		return "", TaskStatus{}, errors.New("missing parent task id")
	}
	store.mu.Lock()
	parentSpec, ok := store.specs[parentID]
	store.mu.Unlock()
	if !ok {
		return "", TaskStatus{}, errors.New("parent task not found")
	}

	replaySpec := parentSpec
	replaySpec.ID = "task_" + randString(8)
	mode = strings.ToLower(strings.TrimSpace(mode))
	switch mode {
	case "force-agent":
		replaySpec.Constraints = upsertConstraint(replaySpec.Constraints, "force_merge_source", "agent_compiler")
	case "force-agent-soft":
		replaySpec.Constraints = upsertConstraint(replaySpec.Constraints, "force_merge_source", "agent_compiler_soft")
	case "force-policy":
		replaySpec.Constraints = upsertConstraint(replaySpec.Constraints, "force_merge_source", "policy_merge")
	default:
		mode = "default"
	}

	now := time.Now().UTC().Format(time.RFC3339)
	replayStatus := TaskStatus{
		SchemaVersion: schemaVersion,
		ID:            replaySpec.ID,
		State:         "queued",
		Progress:      0.0,
		StartedAt:     now,
		UpdatedAt:     now,
	}
	replayTrace := TaskTrace{
		SchemaVersion: schemaVersion,
		TaskID:        replaySpec.ID,
		ParentTaskID:  parentID,
		State:         "queued",
		StartedAt:     now,
		RoutingPolicy: constraintString(replaySpec.Constraints, "routing"),
	}
	store.mu.Lock()
	store.statuses[replaySpec.ID] = replayStatus
	store.specs[replaySpec.ID] = replaySpec
	store.traces[replaySpec.ID] = replayTrace
	store.mu.Unlock()
	if metrics != nil {
		metrics.IncSubmitted()
		metrics.IncReplayed()
		metrics.IncReplayMode(mode)
	}
	_ = queue.Enqueue(queuedTask{id: replaySpec.ID, spec: replaySpec, enqueued: time.Now()})
	return replaySpec.ID, replayStatus, nil
}

type Driver interface {
	ID() string
	Run(ctx context.Context, spec TaskSpec) (ModelResult, error)
}

type DriverRegistry struct {
	mu      sync.Mutex
	drivers []Driver
	meta    map[string]DriverMeta
}

func NewDriverRegistry() *DriverRegistry {
	return &DriverRegistry{drivers: []Driver{}, meta: map[string]DriverMeta{}}
}

func (r *DriverRegistry) Register(d Driver, meta DriverMeta) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.drivers = append(r.drivers, d)
	r.meta[d.ID()] = meta
}

func (r *DriverRegistry) List() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	ids := make([]string, 0, len(r.drivers))
	for _, d := range r.drivers {
		ids = append(ids, d.ID())
	}
	return ids
}

func (r *DriverRegistry) Drivers() []Driver {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]Driver, 0, len(r.drivers))
	out = append(out, r.drivers...)
	return out
}

func (r *DriverRegistry) Meta() map[string]DriverMeta {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := map[string]DriverMeta{}
	for k, v := range r.meta {
		out[k] = v
	}
	return out
}

type DriverMeta struct {
	ID           string   `json:"id"`
	Kind         string   `json:"kind"`
	CostUSD      float64  `json:"cost_usd"`
	Capabilities []string `json:"capabilities"`
	Description  string   `json:"description"`
}

type ModelRegistryEntry struct {
	ID           string   `json:"id"`
	DriverID     string   `json:"driver_id"`
	Kind         string   `json:"kind"`
	Source       string   `json:"source"`
	CostUSD      float64  `json:"cost_usd"`
	Capabilities []string `json:"capabilities"`
	Description  string   `json:"description"`
}

func (r *DriverRegistry) ModelEntries() []ModelRegistryEntry {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := []ModelRegistryEntry{}
	seen := map[string]bool{}
	for _, d := range r.drivers {
		id := d.ID()
		meta := r.meta[id]
		switch typed := d.(type) {
		case HuggingFaceDriver:
			ids := []string{typed.modelID}
			ids = append(ids, typed.fallback...)
			for i, mid := range ids {
				src := "fallback"
				if i == 0 {
					src = "primary"
				}
				key := id + "|" + mid + "|" + src
				if seen[key] {
					continue
				}
				seen[key] = true
				out = append(out, ModelRegistryEntry{
					ID:           mid,
					DriverID:     id,
					Kind:         meta.Kind,
					Source:       src,
					CostUSD:      meta.CostUSD,
					Capabilities: append([]string{}, meta.Capabilities...),
					Description:  meta.Description,
				})
			}
		default:
			key := id + "|" + id + "|driver"
			if seen[key] {
				continue
			}
			seen[key] = true
			out = append(out, ModelRegistryEntry{
				ID:           id,
				DriverID:     id,
				Kind:         meta.Kind,
				Source:       "driver",
				CostUSD:      meta.CostUSD,
				Capabilities: append([]string{}, meta.Capabilities...),
				Description:  meta.Description,
			})
		}
	}
	return out
}

func (r *DriverRegistry) HFModelIDs(extra []string) []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	set := map[string]bool{}
	for _, d := range r.drivers {
		if h, ok := d.(HuggingFaceDriver); ok {
			if h.modelID != "" {
				set[h.modelID] = true
			}
			for _, m := range h.fallback {
				if strings.TrimSpace(m) != "" {
					set[strings.TrimSpace(m)] = true
				}
			}
		}
	}
	for _, m := range extra {
		m = strings.TrimSpace(m)
		if m != "" {
			set[m] = true
		}
	}
	out := make([]string, 0, len(set))
	for m := range set {
		out = append(out, m)
	}
	sort.Strings(out)
	return out
}

type StubDriver struct {
	id      string
	latency time.Duration
	diff    string
	costUSD float64
	quality float64
}

func (d StubDriver) ID() string { return d.id }

func (d StubDriver) Run(ctx context.Context, spec TaskSpec) (ModelResult, error) {
	select {
	case <-time.After(d.latency):
	case <-ctx.Done():
		return ModelResult{}, ctx.Err()
	}

	_ = spec
	return ModelResult{
		SchemaVersion: schemaVersion,
		ModelID:       d.id,
		Output:        "stub result from " + d.id,
		Diff:          d.diff,
		Metrics: []Metric{
			{Name: "latency_ms", Value: float64(d.latency.Milliseconds())},
			{Name: "cost_usd", Value: d.costUSD},
			{Name: "quality_score", Value: d.quality},
		},
	}, nil
}

type HuggingFaceDriver struct {
	id          string
	modelID     string
	apiURL      string
	apiToken    string
	latency     time.Duration
	timeout     time.Duration
	fallback    []string
	pingTimeout time.Duration
}

var pingSvcGlobal *PingService
var metricsGlobal *Metrics

func (d HuggingFaceDriver) ID() string { return d.id }

func (d HuggingFaceDriver) Run(ctx context.Context, spec TaskSpec) (ModelResult, error) {
	start := time.Now()
	models := []string{d.modelID}
	models = append(models, d.fallback...)

	for i, modelID := range models {
		if pingSvcGlobal != nil {
			if !pingSvcGlobal.IsAvailable(modelID, d) {
				if i == len(models)-1 {
					return ModelResult{}, errors.New("hf ping failed for all models")
				}
				continue
			}
		} else if !d.pingAvailable(modelID) {
			if i == len(models)-1 {
				return ModelResult{}, errors.New("hf ping failed for all models")
			}
			continue
		}
		generated, err := d.callHF(ctx, modelID, spec)
		if err != nil {
			if metricsGlobal != nil {
				metricsGlobal.IncHFError()
			}
			if i == len(models)-1 {
				return ModelResult{}, err
			}
			continue
		}

		latencyMs := float64(time.Since(start).Milliseconds())
		quality := estimateQuality(generated, "diff --git a/file b/file\n+stub change HF\n")
		return ModelResult{
			SchemaVersion: schemaVersion,
			ModelID:       d.id,
			Output:        generated,
			Diff:          "diff --git a/file b/file\n+stub change HF\n",
			Metrics: []Metric{
				{Name: "latency_ms", Value: latencyMs},
				{Name: "cost_usd", Value: 0.05},
				{Name: "quality_score", Value: quality},
			},
		}, nil
	}

	return ModelResult{}, errors.New("hf error: no model succeeded")
}

func (d HuggingFaceDriver) pingAvailable(modelID string) bool {
	endpoint := strings.TrimRight(d.apiURL, "/") + "/" + url.PathEscape(modelID)
	ctx, cancel := context.WithTimeout(context.Background(), d.pingTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return false
	}
	if d.apiToken != "" {
		req.Header.Set("Authorization", "Bearer "+d.apiToken)
	}
	client := &http.Client{Timeout: d.pingTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode < 500
}

func (d HuggingFaceDriver) callHF(ctx context.Context, modelID string, spec TaskSpec) (string, error) {
	endpoint := strings.TrimRight(d.apiURL, "/") + "/" + url.PathEscape(modelID)
	payload := map[string]interface{}{
		"inputs": spec.Input,
		"parameters": map[string]interface{}{
			"max_new_tokens": constraintInt(spec.Constraints, "max_new_tokens", 256),
		},
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if d.apiToken != "" {
		req.Header.Set("Authorization", "Bearer "+d.apiToken)
	}
	if strings.EqualFold(os.Getenv("HF_WAIT_FOR_MODEL"), "true") {
		req.Header.Set("x-wait-for-model", "true")
	}
	if strings.EqualFold(os.Getenv("HF_USE_CACHE"), "false") {
		req.Header.Set("x-use-cache", "false")
	}

	client := &http.Client{Timeout: d.timeout}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		data, _ := io.ReadAll(resp.Body)
		return "", errors.New("hf error: " + string(data))
	}

	data, _ := io.ReadAll(resp.Body)
	generated := parseHFGeneratedText(data)
	return generated, nil
}

type SearchResult struct {
	SchemaVersion string   `json:"schema_version"`
	Query         string   `json:"query"`
	Matches       []string `json:"matches"`
}

func main() {
	rand.Seed(time.Now().UnixNano())
	store := NewTaskStore()
	registry := NewDriverRegistry()
	metrics := &Metrics{}
	metricsGlobal = metrics
	queue := NewTaskQueue(envInt("ORCH_QUEUE_SIZE", 200))
	workers := envInt("ORCH_WORKERS", 4)

	registry.Register(StubDriver{id: "model_a", latency: 120 * time.Millisecond, diff: "diff --git a/file b/file\n+stub change A\n", costUSD: 0.01, quality: 0.7}, DriverMeta{
		ID:           "model_a",
		Kind:         "stub",
		CostUSD:      0.01,
		Capabilities: []string{"patch", "review"},
		Description:  "local stub driver A",
	})
	registry.Register(StubDriver{id: "model_b", latency: 140 * time.Millisecond, diff: "diff --git a/file b/file\n+stub change B\n", costUSD: 0.02, quality: 0.6}, DriverMeta{
		ID:           "model_b",
		Kind:         "stub",
		CostUSD:      0.02,
		Capabilities: []string{"patch", "testgen"},
		Description:  "local stub driver B",
	})
	registry.Register(HuggingFaceDriver{
		id:          "hf_gigachat3_702b_preview",
		modelID:     envOr("HF_MODEL_ID", "ai-sage/GigaChat3-702B-A36B-preview"),
		apiURL:      envOr("HF_API_URL", "https://router.huggingface.co/hf-inference/models"),
		apiToken:    os.Getenv("HF_TOKEN"),
		latency:     180 * time.Millisecond,
		timeout:     time.Duration(envInt("HF_TIMEOUT_MS", 8000)) * time.Millisecond,
		fallback:    splitCSV(os.Getenv("HF_FALLBACK_MODELS")),
		pingTimeout: time.Duration(envInt("HF_PING_TIMEOUT_MS", 1500)) * time.Millisecond,
	}, DriverMeta{
		ID:           "hf_gigachat3_702b_preview",
		Kind:         "huggingface",
		CostUSD:      0.05,
		Capabilities: []string{"patch", "review", "analysis"},
		Description:  "HuggingFace Inference API driver (stubbed diff)",
	})

	ragURL := strings.TrimRight(envOr("RAG_URL", "http://localhost:8083"), "/")
	kernelURL := strings.TrimRight(envOr("KERNEL_URL", "http://localhost:8082"), "/")
	web6URL := strings.TrimRight(envOr("WEB6_URL", "http://localhost:8084"), "/")
	quantumURL := strings.TrimRight(envOr("QUANTUM_URL", "http://localhost:8085"), "/")
	agentCompilerURL := strings.TrimRight(envOr("AGENT_COMPILER_URL", "http://localhost:8086"), "/")

	pingSvc := NewPingService(
		time.Duration(envInt("HF_PING_TTL_MS", 15000))*time.Millisecond,
		time.Duration(envInt("HF_PING_BACKOFF_MS", 1000))*time.Millisecond,
		time.Duration(envInt("HF_PING_BACKOFF_MAX_MS", 10000))*time.Millisecond,
	)
	pingSvcGlobal = pingSvc
	cacheMetricsURL := strings.TrimRight(envOr("RAG_CACHE_METRICS_URL", ragURL), "/")
	startHFPingLoop(
		splitCSV(os.Getenv("HF_PING_MODELS")),
		registry,
		pingSvc,
		time.Duration(envInt("HF_PING_INTERVAL_MS", 60000))*time.Millisecond,
	)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	mux.HandleFunc("/drivers", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]interface{}{
			"drivers": registry.List(),
			"details": registry.Meta(),
		})
	})

	mux.HandleFunc("/models", func(w http.ResponseWriter, r *http.Request) {
		entries := registry.ModelEntries()
		sort.Slice(entries, func(i, j int) bool {
			if entries[i].Kind == entries[j].Kind {
				if entries[i].DriverID == entries[j].DriverID {
					return entries[i].ID < entries[j].ID
				}
				return entries[i].DriverID < entries[j].DriverID
			}
			return entries[i].Kind < entries[j].Kind
		})
		writeJSON(w, map[string]interface{}{
			"schema_version": schemaVersion,
			"count":          len(entries),
			"models":         entries,
		})
	})

	mux.HandleFunc("/models/health", func(w http.ResponseWriter, r *http.Request) {
		entries := registry.ModelEntries()
		modelIDs := make([]string, 0, len(entries))
		set := map[string]bool{}
		for _, e := range entries {
			if !set[e.ID] {
				set[e.ID] = true
				modelIDs = append(modelIDs, e.ID)
			}
		}
		healthMap := pingSvc.HealthMap(modelIDs)
		out := []map[string]interface{}{}
		okCount := 0
		failCount := 0
		staleCount := 0
		unknownCount := 0
		for _, e := range entries {
			h := healthMap[e.ID]
			if h.Status == "" {
				h = PingHealth{ModelID: e.ID, Status: "unknown"}
			}
			switch h.Status {
			case "ok":
				okCount++
			case "fail":
				failCount++
			case "stale":
				staleCount++
			default:
				unknownCount++
			}
			out = append(out, map[string]interface{}{
				"id":              e.ID,
				"driver_id":       e.DriverID,
				"kind":            e.Kind,
				"source":          e.Source,
				"status":          h.Status,
				"available":       h.Available,
				"cached":          h.Cached,
				"ok_until_unix":   h.OkUntilUnix,
				"fail_until_unix": h.FailUntilUnix,
				"backoff_ms":      h.BackoffMs,
			})
		}
		writeJSON(w, map[string]interface{}{
			"schema_version": schemaVersion,
			"summary": map[string]int{
				"ok":      okCount,
				"fail":    failCount,
				"stale":   staleCount,
				"unknown": unknownCount,
			},
			"models": out,
		})
	})

	mux.HandleFunc("/models/cost-profile", func(w http.ResponseWriter, r *http.Request) {
		entries := registry.ModelEntries()
		budget := 0.0
		if b := strings.TrimSpace(r.URL.Query().Get("budget_usd")); b != "" {
			if v, err := strconv.ParseFloat(b, 64); err == nil && v >= 0 {
				budget = v
			}
		}
		sort.Slice(entries, func(i, j int) bool { return entries[i].CostUSD < entries[j].CostUSD })
		selected := []ModelRegistryEntry{}
		running := 0.0
		for _, e := range entries {
			if budget > 0 && running+e.CostUSD > budget && len(selected) > 0 {
				continue
			}
			selected = append(selected, e)
			running += e.CostUSD
		}
		writeJSON(w, map[string]interface{}{
			"schema_version": schemaVersion,
			"budget_usd":     budget,
			"total_cost_usd": running,
			"models":         selected,
		})
	})

	mux.HandleFunc("/ping-metrics", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, pingSvc.Snapshot())
	})

	mux.HandleFunc("/dashboard/summary", func(w http.ResponseWriter, r *http.Request) {
		taskSnap := metrics.Snapshot()
		replayModeSnap := metrics.ReplayModeSnapshot()
		parentLinks := store.TraceParentLinks()
		traceStateSnap, traceMergeSnap := store.TraceMetrics()
		mergeChoiceSnap := metrics.MergeChoiceSnapshot()
		modelIDs := registry.HFModelIDs(nil)
		healthMap := pingSvc.HealthMap(modelIDs)
		healthSummary := map[string]int{
			"ok":      0,
			"fail":    0,
			"stale":   0,
			"unknown": 0,
		}
		for _, h := range healthMap {
			switch h.Status {
			case "ok":
				healthSummary["ok"]++
			case "fail":
				healthSummary["fail"]++
			case "stale":
				healthSummary["stale"]++
			default:
				healthSummary["unknown"]++
			}
		}

		web6Alerts := fetchJSONMap(web6URL, "/proxy-counters/alerts")
		web6Health := fetchJSONMap(web6URL, "/proxy-counters/health")

		downstream := map[string]map[string]interface{}{
			"kernel": {
				"base_url":     kernelURL,
				"up":           fetchHealth(kernelURL),
				"runs_total":   fetchPromMetric(kernelURL, "rechain_kernel_runs_total"),
				"errors_total": fetchPromMetric(kernelURL, "rechain_kernel_errors_total"),
			},
			"rag": {
				"base_url":        ragURL,
				"up":              fetchHealth(ragURL),
				"cache_hits":      fetchPromMetric(ragURL, "rechain_rag_cache_hits_total"),
				"cache_misses":    fetchPromMetric(ragURL, "rechain_rag_cache_misses_total"),
				"weight_lexical":  fetchPromMetric(ragURL, "rechain_rag_weight_lexical"),
				"weight_semantic": fetchPromMetric(ragURL, "rechain_rag_weight_semantic"),
			},
			"quantum": {
				"base_url":       quantumURL,
				"up":             fetchHealth(quantumURL),
				"optimize_total": fetchPromMetric(quantumURL, "rechain_quantum_optimize_total"),
			},
			"agent_compiler": {
				"base_url":      agentCompilerURL,
				"up":            fetchHealth(agentCompilerURL),
				"compile_total": fetchPromMetric(agentCompilerURL, "rechain_agent_compile_total"),
			},
			"web6": {
				"base_url":                 web6URL,
				"up":                       fetchHealth(web6URL),
				"proxy_alert_level":        toFloat64(web6Alerts["level_score"]),
				"proxy_json_stale":         toBool(web6Health["proxy_json_stale"]),
				"proxy_prom_stale":         toBool(web6Health["proxy_prom_stale"]),
				"proxy_json_age_sec":       toFloat64(web6Health["proxy_json_age_sec"]),
				"proxy_prom_age_sec":       toFloat64(web6Health["proxy_prom_age_sec"]),
				"proxy_stale_threshold":    toFloat64(web6Health["stale_threshold_sec"]),
				"proxy_critical_threshold": toFloat64(web6Alerts["critical_threshold_sec"]),
			},
		}

		if strings.Contains(r.Header.Get("Accept"), "text/plain") || r.URL.Query().Get("format") == "prom" {
			w.Header().Set("Content-Type", "text/plain; version=0.0.4")
			lines := []string{
				"# HELP rechain_dashboard_queue_depth Current orchestrator queue depth",
				"# TYPE rechain_dashboard_queue_depth gauge",
				"rechain_dashboard_queue_depth " + strconv.Itoa(queue.Depth()),
				"# HELP rechain_dashboard_tasks_total Task counters by state",
				"# TYPE rechain_dashboard_tasks_total gauge",
				"rechain_dashboard_tasks_total{state=\"submitted\"} " + strconv.Itoa(taskSnap["submitted"]),
				"rechain_dashboard_tasks_total{state=\"replayed\"} " + strconv.Itoa(taskSnap["replayed"]),
				"rechain_dashboard_tasks_total{state=\"completed\"} " + strconv.Itoa(taskSnap["completed"]),
				"rechain_dashboard_tasks_total{state=\"failed\"} " + strconv.Itoa(taskSnap["failed"]),
				"rechain_dashboard_tasks_total{state=\"canceled\"} " + strconv.Itoa(taskSnap["canceled"]),
				"# HELP rechain_dashboard_models_health_total Model health summary from ping cache",
				"# TYPE rechain_dashboard_models_health_total gauge",
				"rechain_dashboard_models_health_total{status=\"ok\"} " + strconv.Itoa(healthSummary["ok"]),
				"rechain_dashboard_models_health_total{status=\"fail\"} " + strconv.Itoa(healthSummary["fail"]),
				"rechain_dashboard_models_health_total{status=\"stale\"} " + strconv.Itoa(healthSummary["stale"]),
				"rechain_dashboard_models_health_total{status=\"unknown\"} " + strconv.Itoa(healthSummary["unknown"]),
				"# HELP rechain_dashboard_task_trace_parent_links_total Task traces with parent links",
				"# TYPE rechain_dashboard_task_trace_parent_links_total gauge",
				"rechain_dashboard_task_trace_parent_links_total " + strconv.Itoa(parentLinks),
				"# HELP rechain_dashboard_forced_agent_fallback_total Forced-agent-soft fallbacks to policy merge",
				"# TYPE rechain_dashboard_forced_agent_fallback_total gauge",
				"rechain_dashboard_forced_agent_fallback_total " + strconv.Itoa(taskSnap["forced_fallback"]),
			}
			for source, v := range mergeChoiceSnap {
				lines = append(lines,
					"# HELP rechain_dashboard_merge_choice_total Merge strategy choices",
					"# TYPE rechain_dashboard_merge_choice_total gauge",
					"rechain_dashboard_merge_choice_total{source=\""+source+"\"} "+strconv.Itoa(v),
				)
			}
			for mode, v := range replayModeSnap {
				lines = append(lines,
					"# HELP rechain_dashboard_task_replay_mode_total Replay mode usage",
					"# TYPE rechain_dashboard_task_replay_mode_total gauge",
					"rechain_dashboard_task_replay_mode_total{mode=\""+promLabelValue(mode)+"\"} "+strconv.Itoa(v),
				)
			}

			if v, ok := downstream["kernel"]["up"].(bool); ok {
				up := 0
				if v {
					up = 1
				}
				lines = append(lines,
					"# HELP rechain_dashboard_downstream_up Downstream service availability",
					"# TYPE rechain_dashboard_downstream_up gauge",
					"rechain_dashboard_downstream_up{service=\"kernel\"} "+strconv.Itoa(up),
				)
			}
			if v, ok := downstream["rag"]["up"].(bool); ok {
				up := 0
				if v {
					up = 1
				}
				lines = append(lines, "rechain_dashboard_downstream_up{service=\"rag\"} "+strconv.Itoa(up))
			}
			if v, ok := downstream["quantum"]["up"].(bool); ok {
				up := 0
				if v {
					up = 1
				}
				lines = append(lines, "rechain_dashboard_downstream_up{service=\"quantum\"} "+strconv.Itoa(up))
			}
			if v, ok := downstream["agent_compiler"]["up"].(bool); ok {
				up := 0
				if v {
					up = 1
				}
				lines = append(lines, "rechain_dashboard_downstream_up{service=\"agent_compiler\"} "+strconv.Itoa(up))
			}
			if v, ok := downstream["web6"]["up"].(bool); ok {
				up := 0
				if v {
					up = 1
				}
				lines = append(lines, "rechain_dashboard_downstream_up{service=\"web6\"} "+strconv.Itoa(up))
			}

			lines = append(lines,
				"# HELP rechain_dashboard_kernel_runs_total Kernel run total",
				"# TYPE rechain_dashboard_kernel_runs_total gauge",
				"rechain_dashboard_kernel_runs_total "+formatDashboardMetric(downstream["kernel"]["runs_total"]),
				"# HELP rechain_dashboard_kernel_errors_total Kernel error total",
				"# TYPE rechain_dashboard_kernel_errors_total gauge",
				"rechain_dashboard_kernel_errors_total "+formatDashboardMetric(downstream["kernel"]["errors_total"]),
				"# HELP rechain_dashboard_rag_cache_hits_total RAG cache hits",
				"# TYPE rechain_dashboard_rag_cache_hits_total gauge",
				"rechain_dashboard_rag_cache_hits_total "+formatDashboardMetric(downstream["rag"]["cache_hits"]),
				"# HELP rechain_dashboard_rag_cache_misses_total RAG cache misses",
				"# TYPE rechain_dashboard_rag_cache_misses_total gauge",
				"rechain_dashboard_rag_cache_misses_total "+formatDashboardMetric(downstream["rag"]["cache_misses"]),
				"# HELP rechain_dashboard_rag_weight_lexical RAG lexical weight",
				"# TYPE rechain_dashboard_rag_weight_lexical gauge",
				"rechain_dashboard_rag_weight_lexical "+formatDashboardMetric(downstream["rag"]["weight_lexical"]),
				"# HELP rechain_dashboard_rag_weight_semantic RAG semantic weight",
				"# TYPE rechain_dashboard_rag_weight_semantic gauge",
				"rechain_dashboard_rag_weight_semantic "+formatDashboardMetric(downstream["rag"]["weight_semantic"]),
				"# HELP rechain_dashboard_quantum_optimize_total Quantum optimize total",
				"# TYPE rechain_dashboard_quantum_optimize_total gauge",
				"rechain_dashboard_quantum_optimize_total "+formatDashboardMetric(downstream["quantum"]["optimize_total"]),
				"# HELP rechain_dashboard_agent_compile_total Agent compiler total",
				"# TYPE rechain_dashboard_agent_compile_total gauge",
				"rechain_dashboard_agent_compile_total "+formatDashboardMetric(downstream["agent_compiler"]["compile_total"]),
				"# HELP rechain_dashboard_web6_proxy_alert_level Web6 proxy alert level (ok=0,warn=1,critical=2)",
				"# TYPE rechain_dashboard_web6_proxy_alert_level gauge",
				"rechain_dashboard_web6_proxy_alert_level "+formatDashboardMetric(downstream["web6"]["proxy_alert_level"]),
				"# HELP rechain_dashboard_web6_proxy_json_stale Web6 proxy JSON stale flag",
				"# TYPE rechain_dashboard_web6_proxy_json_stale gauge",
				"rechain_dashboard_web6_proxy_json_stale "+formatDashboardMetric(boolToGauge(downstream["web6"]["proxy_json_stale"])),
				"# HELP rechain_dashboard_web6_proxy_prom_stale Web6 proxy Prom stale flag",
				"# TYPE rechain_dashboard_web6_proxy_prom_stale gauge",
				"rechain_dashboard_web6_proxy_prom_stale "+formatDashboardMetric(boolToGauge(downstream["web6"]["proxy_prom_stale"])),
				"# HELP rechain_dashboard_web6_proxy_json_age_seconds Web6 proxy JSON age",
				"# TYPE rechain_dashboard_web6_proxy_json_age_seconds gauge",
				"rechain_dashboard_web6_proxy_json_age_seconds "+formatDashboardMetric(downstream["web6"]["proxy_json_age_sec"]),
				"# HELP rechain_dashboard_web6_proxy_prom_age_seconds Web6 proxy Prom age",
				"# TYPE rechain_dashboard_web6_proxy_prom_age_seconds gauge",
				"rechain_dashboard_web6_proxy_prom_age_seconds "+formatDashboardMetric(downstream["web6"]["proxy_prom_age_sec"]),
			)

			w.Write([]byte(strings.Join(lines, "\n")))
			return
		}

		writeJSON(w, map[string]interface{}{
			"schema_version": schemaVersion,
			"orchestrator": map[string]interface{}{
				"queue_depth": queue.Depth(),
				"tasks": map[string]int{
					"submitted":             taskSnap["submitted"],
					"replayed":              taskSnap["replayed"],
					"forced_agent_fallback": taskSnap["forced_fallback"],
					"completed":             taskSnap["completed"],
					"failed":                taskSnap["failed"],
					"canceled":              taskSnap["canceled"],
				},
				"trace_parent_links_total": parentLinks,
				"trace": map[string]interface{}{
					"by_state":        traceStateSnap,
					"by_merge_source": traceMergeSnap,
				},
				"merge_choice": mergeChoiceSnap,
				"replay_modes": replayModeSnap,
			},
			"models_health": healthSummary,
			"downstream":    downstream,
		})
	})

	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		pingSnap := pingSvc.Snapshot()
		taskSnap := metrics.Snapshot()
		replayModeSnap := metrics.ReplayModeSnapshot()
		routingSnap := metrics.RoutingSnapshot()
		mergeChoiceSnap := metrics.MergeChoiceSnapshot()
		traceStateSnap, traceMergeSnap := store.TraceMetrics()
		parentLinks := store.TraceParentLinks()
		cacheSnap := fetchCacheMetrics(cacheMetricsURL)
		queueDepth := queue.Depth()
		lines := []string{
			"# HELP rechain_ping_ok_total Total successful HF pings",
			"# TYPE rechain_ping_ok_total counter",
			"rechain_ping_ok_total " + strconv.Itoa(pingSnap["ok"]),
			"# HELP rechain_ping_fail_total Total failed HF pings",
			"# TYPE rechain_ping_fail_total counter",
			"rechain_ping_fail_total " + strconv.Itoa(pingSnap["fail"]),
			"# HELP rechain_ping_skip_total Total cached skips",
			"# TYPE rechain_ping_skip_total counter",
			"rechain_ping_skip_total " + strconv.Itoa(pingSnap["skip"]),
			"# HELP rechain_tasks_total Total tasks by state",
			"# TYPE rechain_tasks_total counter",
			"rechain_tasks_total{state=\"submitted\"} " + strconv.Itoa(taskSnap["submitted"]),
			"rechain_tasks_total{state=\"replayed\"} " + strconv.Itoa(taskSnap["replayed"]),
			"rechain_tasks_total{state=\"completed\"} " + strconv.Itoa(taskSnap["completed"]),
			"rechain_tasks_total{state=\"failed\"} " + strconv.Itoa(taskSnap["failed"]),
			"rechain_tasks_total{state=\"canceled\"} " + strconv.Itoa(taskSnap["canceled"]),
			"# HELP rechain_queue_depth Current queue depth",
			"# TYPE rechain_queue_depth gauge",
			"rechain_queue_depth " + strconv.Itoa(queueDepth),
			"# HELP rechain_hf_errors_total HF driver errors",
			"# TYPE rechain_hf_errors_total counter",
			"rechain_hf_errors_total " + strconv.Itoa(taskSnap["hf_errors"]),
			"# HELP rechain_task_retries_total Total task retries",
			"# TYPE rechain_task_retries_total counter",
			"rechain_task_retries_total " + strconv.Itoa(taskSnap["retries"]),
			"# HELP rechain_task_replay_total Total replayed tasks",
			"# TYPE rechain_task_replay_total counter",
			"rechain_task_replay_total " + strconv.Itoa(taskSnap["replayed"]),
			"# HELP rechain_forced_agent_fallback_total Forced-agent-soft fallbacks to policy merge",
			"# TYPE rechain_forced_agent_fallback_total counter",
			"rechain_forced_agent_fallback_total " + strconv.Itoa(taskSnap["forced_fallback"]),
			"# HELP rechain_task_trace_parent_links_total Task traces with parent links",
			"# TYPE rechain_task_trace_parent_links_total gauge",
			"rechain_task_trace_parent_links_total " + strconv.Itoa(parentLinks),
			"# HELP rechain_task_latency_avg_ms Average task latency (last 100)",
			"# TYPE rechain_task_latency_avg_ms gauge",
			"rechain_task_latency_avg_ms " + strconv.Itoa(taskSnap["latency_avg_ms"]),
			"# HELP rechain_queue_delay_avg_ms Average queue delay (last 100)",
			"# TYPE rechain_queue_delay_avg_ms gauge",
			"rechain_queue_delay_avg_ms " + strconv.Itoa(taskSnap["queue_delay_avg_ms"]),
		}
		for k, v := range routingSnap {
			lines = append(lines,
				"# HELP rechain_routing_total Routing policy usage",
				"# TYPE rechain_routing_total counter",
				"rechain_routing_total{policy=\""+k+"\"} "+strconv.Itoa(v),
			)
		}
		for model, policies := range metrics.RoutingByModelSnapshot() {
			for policy, v := range policies {
				lines = append(lines,
					"# HELP rechain_routing_by_model_total Routing policy per model",
					"# TYPE rechain_routing_by_model_total counter",
					"rechain_routing_by_model_total{model=\""+model+"\",policy=\""+policy+"\"} "+strconv.Itoa(v),
				)
			}
		}
		for source, v := range mergeChoiceSnap {
			lines = append(lines,
				"# HELP rechain_merge_choice_total Merge strategy choices",
				"# TYPE rechain_merge_choice_total counter",
				"rechain_merge_choice_total{source=\""+source+"\"} "+strconv.Itoa(v),
			)
		}
		for mode, v := range replayModeSnap {
			lines = append(lines,
				"# HELP rechain_task_replay_mode_total Replay mode usage",
				"# TYPE rechain_task_replay_mode_total counter",
				"rechain_task_replay_mode_total{mode=\""+promLabelValue(mode)+"\"} "+strconv.Itoa(v),
			)
		}
		for state, v := range traceStateSnap {
			lines = append(lines,
				"# HELP rechain_task_trace_total Task trace counters by state",
				"# TYPE rechain_task_trace_total gauge",
				"rechain_task_trace_total{state=\""+state+"\"} "+strconv.Itoa(v),
			)
		}
		for source, v := range traceMergeSnap {
			lines = append(lines,
				"# HELP rechain_task_trace_total Task trace counters by merge source",
				"# TYPE rechain_task_trace_total gauge",
				"rechain_task_trace_total{merge_source=\""+source+"\"} "+strconv.Itoa(v),
			)
		}

		lines = append(lines, renderLatencyHistogram(metrics)...)
		lines = append(lines, renderRoutingModelLatencyHistogram(metrics)...)
		if len(cacheSnap) > 0 {
			lines = append(lines,
				"# HELP rechain_cache_hits_total Cache hits",
				"# TYPE rechain_cache_hits_total counter",
				"rechain_cache_hits_total "+strconv.Itoa(cacheSnap["hits"]),
				"# HELP rechain_cache_misses_total Cache misses",
				"# TYPE rechain_cache_misses_total counter",
				"rechain_cache_misses_total "+strconv.Itoa(cacheSnap["misses"]),
				"# HELP rechain_cache_purges_total Cache purges",
				"# TYPE rechain_cache_purges_total counter",
				"rechain_cache_purges_total "+strconv.Itoa(cacheSnap["purges"]),
				"# HELP rechain_cache_evictions_total Cache evictions",
				"# TYPE rechain_cache_evictions_total counter",
				"rechain_cache_evictions_total "+strconv.Itoa(cacheSnap["evictions"]),
				"# HELP rechain_cache_entries Cache entries",
				"# TYPE rechain_cache_entries gauge",
				"rechain_cache_entries "+strconv.Itoa(cacheSnap["entries"]),
				"# HELP rechain_cache_bytes Cache bytes",
				"# TYPE rechain_cache_bytes gauge",
				"rechain_cache_bytes "+strconv.Itoa(cacheSnap["bytes"]),
			)
		}
		w.Write([]byte(strings.Join(lines, "\n")))
	})

	mux.HandleFunc("/queue-depth", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]int{"queue_depth": queue.Depth()})
	})

	mux.HandleFunc("/quality-score", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			Output string `json:"output"`
			Diff   string `json:"diff"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		score := estimateQuality(req.Output, req.Diff)
		stats := diffStats(req.Diff)
		errCount := errorTokenCount(req.Output)
		writeJSON(w, map[string]interface{}{
			"quality_score": score,
			"details": map[string]interface{}{
				"files":       stats.files,
				"hunks":       stats.hunks,
				"additions":   stats.additions,
				"deletions":   stats.deletions,
				"total_lines": stats.totalLines,
				"errors":      errCount,
				"output_len":  len(req.Output),
			},
		})
	})

	mux.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var spec TaskSpec
		if err := json.NewDecoder(r.Body).Decode(&spec); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}

		if spec.SchemaVersion == "" {
			spec.SchemaVersion = schemaVersion
		}
		if spec.ID == "" {
			spec.ID = "task_" + randString(8)
		}

		now := time.Now().UTC().Format(time.RFC3339)
		status := TaskStatus{
			SchemaVersion: schemaVersion,
			ID:            spec.ID,
			State:         "queued",
			Progress:      0.0,
			StartedAt:     now,
			UpdatedAt:     now,
		}

		store.mu.Lock()
		store.statuses[spec.ID] = status
		store.specs[spec.ID] = spec
		store.traces[spec.ID] = TaskTrace{
			SchemaVersion: schemaVersion,
			TaskID:        spec.ID,
			State:         "queued",
			StartedAt:     now,
			RoutingPolicy: constraintString(spec.Constraints, "routing"),
		}
		store.mu.Unlock()
		metrics.IncSubmitted()
		_ = queue.Enqueue(queuedTask{id: spec.ID, spec: spec, enqueued: time.Now()})

		writeJSON(w, status)
	})

	mux.HandleFunc("/tasks/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/tasks/")
		if path == "" {
			http.NotFound(w, r)
			return
		}

		if path == "recent" {
			limit := 10
			if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
				if n, err := strconv.Atoi(raw); err == nil && n > 0 {
					limit = n
				}
			}
			stateFilter := strings.TrimSpace(r.URL.Query().Get("state"))
			mergeSourceFilter := strings.TrimSpace(r.URL.Query().Get("merge_source"))
			hasParent := strings.TrimSpace(r.URL.Query().Get("has_parent"))
			sortBy := strings.TrimSpace(r.URL.Query().Get("sort"))
			writeJSON(w, map[string]interface{}{
				"schema_version": schemaVersion,
				"tasks":          store.RecentTasks(limit, stateFilter, mergeSourceFilter, hasParent, sortBy),
			})
			return
		}

		if path == "latest/trace" {
			trace, ok := store.LatestTrace()
			if !ok {
				http.NotFound(w, r)
				return
			}
			writeJSON(w, trace)
			return
		}

		if strings.HasSuffix(path, "/artifacts") {
			id := strings.TrimSuffix(path, "/artifacts")
			store.mu.Lock()
			artifacts := store.artifacts[id]
			store.mu.Unlock()
			writeJSON(w, artifacts)
			return
		}

		if strings.HasSuffix(path, "/result") {
			id := strings.TrimSuffix(path, "/result")
			store.mu.Lock()
			result, ok := store.results[id]
			store.mu.Unlock()
			if !ok {
				http.NotFound(w, r)
				return
			}
			writeJSON(w, result)
			return
		}

		if strings.HasSuffix(path, "/trace") {
			id := strings.TrimSuffix(path, "/trace")
			store.mu.Lock()
			trace, ok := store.traces[id]
			store.mu.Unlock()
			if !ok {
				http.NotFound(w, r)
				return
			}
			writeJSON(w, trace)
			return
		}

		if strings.HasSuffix(path, "/replay-chain") {
			id := strings.TrimSuffix(path, "/replay-chain")
			id = strings.TrimSuffix(id, "/")
			chain, ok := store.ReplayChain(id)
			if !ok {
				http.NotFound(w, r)
				return
			}
			writeJSON(w, chain)
			return
		}

		if strings.HasSuffix(path, "/debug") {
			id := strings.TrimSuffix(path, "/debug")
			id = strings.TrimSuffix(id, "/")
			store.mu.Lock()
			status, okStatus := store.statuses[id]
			trace, okTrace := store.traces[id]
			result, okResult := store.results[id]
			artifacts := append([]Artifact{}, store.artifacts[id]...)
			store.mu.Unlock()
			if !okStatus {
				http.NotFound(w, r)
				return
			}
			chain, _ := store.ReplayChain(id)
			traceByState, traceBySource := store.TraceMetrics()
			payload := map[string]interface{}{
				"schema_version": schemaVersion,
				"task_id":        id,
				"status":         status,
				"trace":          trace,
				"replay_chain":   chain,
				"artifacts":      artifacts,
				"merge_metrics": map[string]interface{}{
					"global_choice":   metrics.MergeChoiceSnapshot(),
					"trace_by_state":  traceByState,
					"trace_by_source": traceBySource,
				},
			}
			if okResult {
				payload["result"] = result
			}
			if !okTrace {
				payload["trace"] = TaskTrace{}
			}
			if strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("format")), "prom") || strings.Contains(r.Header.Get("Accept"), "text/plain") {
				w.Header().Set("Content-Type", "text/plain; version=0.0.4")
				scope := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("scope")))
				if scope == "" {
					scope = "all"
				}
				taskID := promLabelValue(id)
				state := promLabelValue(status.State)
				parent := strings.TrimSpace(trace.ParentTaskID)
				hasParent := 0
				if parent != "" {
					hasParent = 1
				}
				hasTrace := 0
				if okTrace {
					hasTrace = 1
				}
				hasResult := 0
				if okResult {
					hasResult = 1
				}
				hasError := 0
				if strings.TrimSpace(trace.Error) != "" {
					hasError = 1
				}
				lineageCount := len(chain.Lineage)
				descCount := len(chain.Descendants)
				selectedCount := len(trace.Selected)
				modelResultCount := len(trace.Results)
				mergeSource := promLabelValue(trace.MergeSource)
				if mergeSource == "" {
					mergeSource = "none"
				}
				quality := 0.0
				confidence := 0.0
				if trace.Merge != nil {
					quality = trace.Merge.QualityScore
					confidence = trace.Merge.Confidence
				}
				lines := []string{
					"# HELP rechain_task_debug_state Task state for debug endpoint",
					"# TYPE rechain_task_debug_state gauge",
					"rechain_task_debug_state{task_id=\"" + taskID + "\",state=\"" + state + "\"} 1",
					"# HELP rechain_task_debug_progress Task progress for debug endpoint",
					"# TYPE rechain_task_debug_progress gauge",
					"rechain_task_debug_progress{task_id=\"" + taskID + "\"} " + formatFloat(status.Progress),
					"# HELP rechain_task_debug_has_trace Task has trace payload",
					"# TYPE rechain_task_debug_has_trace gauge",
					"rechain_task_debug_has_trace{task_id=\"" + taskID + "\"} " + strconv.Itoa(hasTrace),
					"# HELP rechain_task_debug_has_result Task has merge result payload",
					"# TYPE rechain_task_debug_has_result gauge",
					"rechain_task_debug_has_result{task_id=\"" + taskID + "\"} " + strconv.Itoa(hasResult),
					"# HELP rechain_task_debug_has_parent Task has parent replay link",
					"# TYPE rechain_task_debug_has_parent gauge",
					"rechain_task_debug_has_parent{task_id=\"" + taskID + "\"} " + strconv.Itoa(hasParent),
					"# HELP rechain_task_debug_has_error Task trace has error",
					"# TYPE rechain_task_debug_has_error gauge",
					"rechain_task_debug_has_error{task_id=\"" + taskID + "\"} " + strconv.Itoa(hasError),
					"# HELP rechain_task_debug_artifacts_count Task artifacts count",
					"# TYPE rechain_task_debug_artifacts_count gauge",
					"rechain_task_debug_artifacts_count{task_id=\"" + taskID + "\"} " + strconv.Itoa(len(artifacts)),
					"# HELP rechain_task_debug_lineage_count Replay chain lineage count",
					"# TYPE rechain_task_debug_lineage_count gauge",
					"rechain_task_debug_lineage_count{task_id=\"" + taskID + "\"} " + strconv.Itoa(lineageCount),
					"# HELP rechain_task_debug_descendants_count Replay chain descendants count",
					"# TYPE rechain_task_debug_descendants_count gauge",
					"rechain_task_debug_descendants_count{task_id=\"" + taskID + "\"} " + strconv.Itoa(descCount),
					"# HELP rechain_task_debug_selected_models_count Selected model count in trace",
					"# TYPE rechain_task_debug_selected_models_count gauge",
					"rechain_task_debug_selected_models_count{task_id=\"" + taskID + "\"} " + strconv.Itoa(selectedCount),
					"# HELP rechain_task_debug_model_results_count Model result count in trace",
					"# TYPE rechain_task_debug_model_results_count gauge",
					"rechain_task_debug_model_results_count{task_id=\"" + taskID + "\"} " + strconv.Itoa(modelResultCount),
					"# HELP rechain_task_debug_merge_quality_score Merge quality score in trace",
					"# TYPE rechain_task_debug_merge_quality_score gauge",
					"rechain_task_debug_merge_quality_score{task_id=\"" + taskID + "\"} " + formatFloat(quality),
					"# HELP rechain_task_debug_merge_confidence Merge confidence in trace",
					"# TYPE rechain_task_debug_merge_confidence gauge",
					"rechain_task_debug_merge_confidence{task_id=\"" + taskID + "\"} " + formatFloat(confidence),
					"# HELP rechain_task_debug_merge_source Task merge source marker",
					"# TYPE rechain_task_debug_merge_source gauge",
					"rechain_task_debug_merge_source{task_id=\"" + taskID + "\",source=\"" + mergeSource + "\"} 1",
				}
				if scope == "all" || scope == "global" {
					for source, v := range metrics.MergeChoiceSnapshot() {
						lines = append(lines,
							"# HELP rechain_task_debug_global_merge_choice_total Global merge strategy counters",
							"# TYPE rechain_task_debug_global_merge_choice_total gauge",
							"rechain_task_debug_global_merge_choice_total{task_id=\""+taskID+"\",source=\""+promLabelValue(source)+"\"} "+strconv.Itoa(v),
						)
					}
				}
				w.Write([]byte(strings.Join(lines, "\n")))
				return
			}
			writeJSON(w, payload)
			return
		}

		if strings.HasSuffix(path, "/cancel") {
			if r.Method != http.MethodPost {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			id := strings.TrimSuffix(path, "/cancel")
			store.mu.Lock()
			status, ok := store.statuses[id]
			if ok {
				status.State = "canceled"
				status.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
				store.statuses[id] = status
				trace := store.traces[id]
				trace.State = "canceled"
				trace.FinishedAt = status.UpdatedAt
				store.traces[id] = trace
				metrics.IncCanceled()
			}
			store.mu.Unlock()
			if !ok {
				http.NotFound(w, r)
				return
			}
			writeJSON(w, status)
			return
		}

		if strings.HasSuffix(path, "/replay/batch") {
			if r.Method != http.MethodPost {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			parentID := strings.TrimSuffix(path, "/replay/batch")
			parentID = strings.TrimSuffix(parentID, "/")
			if parentID == "" {
				http.NotFound(w, r)
				return
			}
			var req struct {
				Modes []string `json:"modes"`
			}
			if r.Body != nil {
				_ = json.NewDecoder(r.Body).Decode(&req)
			}
			modes := req.Modes
			if len(modes) == 0 {
				modes = splitCSV(r.URL.Query().Get("modes"))
			}
			if len(modes) == 0 {
				modes = []string{"force-policy", "force-agent-soft"}
			}
			type replayItem struct {
				Mode         string     `json:"mode"`
				ReplayTaskID string     `json:"replay_task_id,omitempty"`
				Status       TaskStatus `json:"status,omitempty"`
				Error        string     `json:"error,omitempty"`
			}
			items := []replayItem{}
			for _, mode := range modes {
				replayID, replayStatus, err := enqueueReplayTask(store, queue, metrics, parentID, mode)
				item := replayItem{Mode: strings.ToLower(strings.TrimSpace(mode))}
				if err != nil {
					item.Error = err.Error()
				} else {
					item.ReplayTaskID = replayID
					item.Status = replayStatus
				}
				items = append(items, item)
			}
			writeJSON(w, map[string]interface{}{
				"schema_version": schemaVersion,
				"parent_task_id": parentID,
				"count":          len(items),
				"items":          items,
			})
			return
		}

		if strings.HasSuffix(path, "/replay") {
			if r.Method != http.MethodPost {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			parentID := strings.TrimSuffix(path, "/replay")
			parentID = strings.TrimSuffix(parentID, "/")
			if parentID == "" {
				http.NotFound(w, r)
				return
			}
			replayMode := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("mode")))
			replayID, replayStatus, err := enqueueReplayTask(store, queue, metrics, parentID, replayMode)
			if err != nil {
				http.NotFound(w, r)
				return
			}
			writeJSON(w, map[string]interface{}{
				"schema_version": schemaVersion,
				"parent_task_id": parentID,
				"replay_task_id": replayID,
				"mode":           replayMode,
				"status":         replayStatus,
			})
			return
		}

		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		store.mu.Lock()
		status, ok := store.statuses[path]
		store.mu.Unlock()
		if !ok {
			http.NotFound(w, r)
			return
		}
		writeJSON(w, status)
	})

	addr := ":8081"
	log.Printf("orchestrator listening on %s", addr)
	startWorkers(workers, queue, store, registry, ragURL, metrics)
	if err := http.ListenAndServe(addr, logging.WithRequestID(mux)); err != nil {
		log.Fatal(err)
	}
}

func startWorkers(count int, queue *TaskQueue, store *TaskStore, registry *DriverRegistry, ragURL string, metrics *Metrics) {
	if count <= 0 {
		count = 1
	}
	for i := 0; i < count; i++ {
		go func() {
			for {
				task, ok := queue.Dequeue(context.Background())
				if !ok {
					continue
				}
				now := time.Now().UTC().Format(time.RFC3339)
				store.mu.Lock()
				status := store.statuses[task.id]
				status.State = "running"
				status.Progress = 0.1
				status.StartedAt = now
				status.UpdatedAt = now
				store.statuses[task.id] = status
				trace := store.traces[task.id]
				trace.State = "running"
				trace.StartedAt = now
				if trace.RoutingPolicy == "" {
					trace.RoutingPolicy = constraintString(task.spec.Constraints, "routing")
				}
				store.traces[task.id] = trace
				store.mu.Unlock()

				if metrics != nil {
					delay := time.Since(task.enqueued)
					if delay > 0 {
						metrics.ObserveQueueDelay(delay.Milliseconds())
					}
				}
				processTask(store, registry.Drivers(), registry.Meta(), task.id, task.spec, ragURL, metrics)
			}
		}()
	}
}

func processTask(store *TaskStore, drivers []Driver, meta map[string]DriverMeta, id string, spec TaskSpec, ragURL string, metrics *Metrics) {
	start := time.Now()
	trace := TaskTrace{
		SchemaVersion: schemaVersion,
		TaskID:        id,
		State:         "running",
		StartedAt:     start.UTC().Format(time.RFC3339),
		RoutingPolicy: constraintString(spec.Constraints, "routing"),
	}
	store.mu.Lock()
	existingTrace, ok := store.traces[id]
	store.mu.Unlock()
	if ok {
		trace.ParentTaskID = existingTrace.ParentTaskID
		if existingTrace.StartedAt != "" {
			trace.StartedAt = existingTrace.StartedAt
		}
	}
	timeoutMs := constraintInt(spec.Constraints, "budget_ms", 2000)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutMs)*time.Millisecond)
	defer cancel()

	delay := queueDelayForPriority(spec.Metadata.Priority)
	if delay > 0 {
		time.Sleep(delay)
		if metrics != nil {
			metrics.ObserveQueueDelay(delay.Milliseconds())
		}
	}

	if ragURL != "" {
		if ctxs, err := fetchRAGContext(ctx, ragURL, spec.Input); err == nil && len(ctxs) > 0 {
			spec.Context = append(spec.Context, ctxs...)
		}
	}

	selected := selectDrivers(spec, drivers, meta)
	for _, d := range selected {
		trace.Selected = append(trace.Selected, d.ID())
	}
	results := make([]ModelResult, 0, len(selected))
	for _, d := range selected {
		res, err := runWithRetry(ctx, d, spec, metrics)
		if err == nil {
			results = append(results, res)
			if metrics != nil {
				metrics.ObserveModelLatency(res.ModelID, int64(metricValue(res, "latency_ms")))
			}
		}
	}

	if len(results) == 0 {
		fallbackIDs := splitCSV(constraintString(spec.Constraints, "fallback_models"))
		if len(fallbackIDs) > 0 {
			for _, fid := range fallbackIDs {
				d := findDriverByID(drivers, fid)
				if d == nil {
					continue
				}
				res, err := runWithRetry(ctx, d, spec, metrics)
				if err == nil {
					results = append(results, res)
					if metrics != nil {
						metrics.ObserveModelLatency(res.ModelID, int64(metricValue(res, "latency_ms")))
					}
				}
			}
		}
		if len(results) == 0 {
			store.mu.Lock()
			status := store.statuses[id]
			status.State = "failed"
			status.Progress = 1.0
			status.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
			store.statuses[id] = status
			trace.State = "failed"
			trace.FinishedAt = time.Now().UTC().Format(time.RFC3339)
			trace.Error = "no model results"
			store.traces[id] = trace
			store.mu.Unlock()
			metrics.IncFailed()
			metrics.ObserveLatency(time.Since(start).Milliseconds())
			return
		}
	}

	for _, r := range results {
		trace.Results = append(trace.Results, TraceModelResult{
			ModelID:      r.ModelID,
			DiffLen:      len(r.Diff),
			LatencyMs:    metricValue(r, "latency_ms"),
			CostUSD:      metricValue(r, "cost_usd"),
			QualityScore: metricValue(r, "quality_score"),
		})
	}

	policy := constraintString(spec.Constraints, "routing")
	metrics.IncRouting(policy)
	for _, d := range selected {
		metrics.IncRoutingModel(policy, d.ID())
	}

	forceMergeSource := strings.ToLower(strings.TrimSpace(constraintString(spec.Constraints, "force_merge_source")))
	mergeSource := "agent_compiler"
	var merge MergeResult
	var err error
	switch forceMergeSource {
	case "agent_compiler":
		merge, err = tryAgentCompiler(results, policy)
		if err != nil {
			err = errors.New("forced agent_compiler failed: " + err.Error())
		}
	case "agent_compiler_soft":
		merge, err = tryAgentCompiler(results, policy)
		if err != nil {
			mergeSource = "policy_merge"
			merge, err = mergeResults(
				results,
				policy,
				constraintFloat(spec.Constraints, "weight_cost", 0.3),
				constraintFloat(spec.Constraints, "weight_latency", 0.7),
				constraintFloat(spec.Constraints, "weight_quality", 0.0),
			)
			if err == nil && metrics != nil {
				metrics.IncForcedFallback()
			}
		}
	case "policy_merge":
		mergeSource = "policy_merge"
		merge, err = mergeResults(
			results,
			policy,
			constraintFloat(spec.Constraints, "weight_cost", 0.3),
			constraintFloat(spec.Constraints, "weight_latency", 0.7),
			constraintFloat(spec.Constraints, "weight_quality", 0.0),
		)
	default:
		merge, err = tryAgentCompiler(results, policy)
		if err != nil {
			mergeSource = "policy_merge"
			merge, err = mergeResults(
				results,
				policy,
				constraintFloat(spec.Constraints, "weight_cost", 0.3),
				constraintFloat(spec.Constraints, "weight_latency", 0.7),
				constraintFloat(spec.Constraints, "weight_quality", 0.0),
			)
		}
	}
	if err != nil {
		store.mu.Lock()
		status := store.statuses[id]
		status.State = "failed"
		status.Progress = 1.0
		status.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
		store.statuses[id] = status
		trace.State = "failed"
		trace.FinishedAt = time.Now().UTC().Format(time.RFC3339)
		trace.Error = "merge failed: " + err.Error()
		store.traces[id] = trace
		store.mu.Unlock()
		metrics.IncFailed()
		metrics.ObserveLatency(time.Since(start).Milliseconds())
		return
	}

	artifact := Artifact{
		SchemaVersion: schemaVersion,
		ID:            "artifact_" + randString(8),
		Type:          "diff",
		Path:          "artifacts/" + id + "/patch.diff",
		Sha256:        "",
		CreatedAt:     time.Now().UTC().Format(time.RFC3339),
	}

	store.mu.Lock()
	status := store.statuses[id]
	status.State = "completed"
	status.Progress = 1.0
	status.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	store.statuses[id] = status
	store.artifacts[id] = []Artifact{artifact}
	store.results[id] = merge
	trace.State = "completed"
	trace.FinishedAt = time.Now().UTC().Format(time.RFC3339)
	trace.MergeSource = mergeSource
	trace.Merge = &merge
	store.traces[id] = trace
	store.mu.Unlock()
	if metrics != nil {
		metrics.IncMergeChoice(mergeSource)
	}
	metrics.IncCompleted()
	metrics.ObserveLatency(time.Since(start).Milliseconds())
}

func tryAgentCompiler(results []ModelResult, policy string) (MergeResult, error) {
	base := strings.TrimRight(os.Getenv("AGENT_COMPILER_URL"), "/")
	if base == "" {
		return MergeResult{}, errors.New("agent compiler disabled")
	}
	payload := map[string]interface{}{
		"schema_version": schemaVersion,
		"task_id":        "task",
		"policy":         policy,
		"results":        results,
	}
	body, _ := json.Marshal(payload)
	client := &http.Client{Timeout: 1200 * time.Millisecond}
	resp, err := client.Post(base+"/compile", "application/json", bytes.NewReader(body))
	if err != nil {
		return MergeResult{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return MergeResult{}, errors.New("agent compiler error")
	}
	var out struct {
		Diff         string  `json:"diff"`
		QualityScore float64 `json:"quality_score"`
		Rationale    string  `json:"rationale"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return MergeResult{}, err
	}
	if out.Diff == "" {
		return MergeResult{}, errors.New("agent compiler empty diff")
	}
	return MergeResult{
		SchemaVersion: schemaVersion,
		Diff:          out.Diff,
		Rationale:     "agent compiler: " + out.Rationale,
		Confidence:    0.6,
		QualityScore:  out.QualityScore,
	}, nil
}

func mergeResults(results []ModelResult, policy string, weightCost float64, weightLatency float64, weightQuality float64) (MergeResult, error) {
	metric := "latency_ms"
	if strings.EqualFold(policy, "cost") {
		metric = "cost_usd"
	} else if strings.EqualFold(policy, "quality") {
		metric = "quality_score"
		mapped := make([]internal.ModelResult, 0, len(results))
		for _, r := range results {
			mapped = append(mapped, internal.ModelResult{
				Diff:    r.Diff,
				Metrics: []internal.Metric{{Name: metric, Value: metricValue(r, metric)}},
			})
		}
		best, ok := internal.MergeResults(mapped, metric, true)
		if !ok {
			return MergeResult{}, errors.New("no results")
		}
		return MergeResult{
			SchemaVersion: schemaVersion,
			Diff:          best.Diff,
			Rationale:     "selected highest " + metric + " with deterministic tie-breaker",
			Confidence:    0.6,
			QualityScore:  metricValueInternal(best, "quality_score"),
		}, nil
	} else if strings.EqualFold(policy, "quantum") {
		if best, ok := tryQuantumOptimize(results); ok {
			return MergeResult{
				SchemaVersion: schemaVersion,
				Diff:          best.Diff,
				Rationale:     "selected via quantum optimizer",
				Confidence:    0.6,
				QualityScore:  metricValue(best, "quality_score"),
			}, nil
		}
		return MergeResult{}, errors.New("quantum optimize failed")
	} else if strings.EqualFold(policy, "weighted") {
		best, score := weightedBest(results, weightCost, weightLatency, 0)
		return MergeResult{
			SchemaVersion: schemaVersion,
			Diff:          best.Diff,
			Rationale:     "selected lowest weighted score " + formatFloat(score),
			Confidence:    0.6,
			QualityScore:  metricValue(best, "quality_score"),
		}, nil
	} else if strings.EqualFold(policy, "weighted_quality") {
		best, score := weightedBest(results, weightCost, weightLatency, weightQuality)
		return MergeResult{
			SchemaVersion: schemaVersion,
			Diff:          best.Diff,
			Rationale:     "selected lowest weighted score " + formatFloat(score),
			Confidence:    0.6,
			QualityScore:  metricValue(best, "quality_score"),
		}, nil
	}
	mapped := make([]internal.ModelResult, 0, len(results))
	for _, r := range results {
		mapped = append(mapped, internal.ModelResult{
			Diff:    r.Diff,
			Metrics: []internal.Metric{{Name: metric, Value: metricValue(r, metric)}},
		})
	}

	best, ok := internal.MergeResults(mapped, metric, false)
	if !ok {
		return MergeResult{}, errors.New("no results")
	}

	return MergeResult{
		SchemaVersion: schemaVersion,
		Diff:          best.Diff,
		Rationale:     "selected lowest " + metric + " with deterministic tie-breaker",
		Confidence:    0.6,
		QualityScore:  metricValueInternal(best, "quality_score"),
	}, nil
}

func selectDrivers(spec TaskSpec, drivers []Driver, meta map[string]DriverMeta) []Driver {
	preferred := constraintString(spec.Constraints, "models")
	maxModels := constraintInt(spec.Constraints, "max_models", len(drivers))
	minModels := constraintInt(spec.Constraints, "min_models", 0)
	budgetUSD := constraintFloat(spec.Constraints, "budget_usd", 0)

	out := []Driver{}
	if preferred != "" {
		allowed := map[string]bool{}
		for _, m := range strings.Split(preferred, ",") {
			allowed[strings.TrimSpace(m)] = true
		}
		for _, d := range drivers {
			if allowed[d.ID()] {
				out = append(out, d)
			}
		}
	} else {
		if budgetUSD > 0 {
			out = append(out, selectByBudget(drivers, meta, budgetUSD)...)
		} else {
			out = append(out, drivers...)
		}
	}

	if maxModels > 0 && len(out) > maxModels {
		return out[:maxModels]
	}
	if minModels > 0 && len(out) < minModels {
		for _, d := range drivers {
			if len(out) >= minModels {
				break
			}
			exists := false
			for _, s := range out {
				if s.ID() == d.ID() {
					exists = true
					break
				}
			}
			if !exists {
				out = append(out, d)
			}
		}
	}
	return out
}

func selectByBudget(drivers []Driver, meta map[string]DriverMeta, budgetUSD float64) []Driver {
	type item struct {
		d    Driver
		cost float64
	}
	items := []item{}
	for _, d := range drivers {
		cost := 0.0
		if m, ok := meta[d.ID()]; ok {
			cost = m.CostUSD
		}
		items = append(items, item{d: d, cost: cost})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].cost < items[j].cost })

	out := []Driver{}
	sum := 0.0
	for _, it := range items {
		if sum+it.cost <= budgetUSD || len(out) == 0 {
			out = append(out, it.d)
			sum += it.cost
		}
	}
	return out
}

func constraintString(constraints []Constraint, key string) string {
	for _, c := range constraints {
		if c.Key == key {
			if v, ok := c.Value.(string); ok {
				return v
			}
		}
	}
	return ""
}

func constraintInt(constraints []Constraint, key string, fallback int) int {
	for _, c := range constraints {
		if c.Key == key {
			switch v := c.Value.(type) {
			case float64:
				return int(v)
			case int:
				return v
			case string:
				if v == "" {
					return fallback
				}
			}
		}
	}
	return fallback
}

func constraintFloat(constraints []Constraint, key string, fallback float64) float64 {
	for _, c := range constraints {
		if c.Key == key {
			switch v := c.Value.(type) {
			case float64:
				return v
			case int:
				return float64(v)
			}
		}
	}
	return fallback
}

func upsertConstraint(constraints []Constraint, key string, value interface{}) []Constraint {
	out := append([]Constraint{}, constraints...)
	for i := range out {
		if strings.EqualFold(out[i].Key, key) {
			out[i].Value = value
			return out
		}
	}
	out = append(out, Constraint{Key: key, Value: value})
	return out
}

func findDriverByID(drivers []Driver, id string) Driver {
	for _, d := range drivers {
		if d.ID() == id {
			return d
		}
	}
	return nil
}

func splitCSV(value string) []string {
	parts := []string{}
	for _, p := range strings.Split(value, ",") {
		v := strings.TrimSpace(p)
		if v != "" {
			parts = append(parts, v)
		}
	}
	return parts
}

func weightedBest(results []ModelResult, weightCost float64, weightLatency float64, weightQuality float64) (ModelResult, float64) {
	minCost, maxCost := minMaxMetric(results, "cost_usd")
	minLat, maxLat := minMaxMetric(results, "latency_ms")
	minQ, maxQ := minMaxMetric(results, "quality_score")

	best := results[0]
	bestScore := scoreResult(results[0], weightCost, weightLatency, weightQuality, minCost, maxCost, minLat, maxLat, minQ, maxQ)
	for _, r := range results[1:] {
		s := scoreResult(r, weightCost, weightLatency, weightQuality, minCost, maxCost, minLat, maxLat, minQ, maxQ)
		if s < bestScore || (s == bestScore && r.Diff < best.Diff) {
			best = r
			bestScore = s
		}
	}
	return best, bestScore
}

func minMaxMetric(results []ModelResult, name string) (float64, float64) {
	min := 0.0
	max := 0.0
	for i, r := range results {
		v := metricValue(r, name)
		if i == 0 || v < min {
			min = v
		}
		if i == 0 || v > max {
			max = v
		}
	}
	return min, max
}

func scoreResult(r ModelResult, weightCost float64, weightLatency float64, weightQuality float64, minCost float64, maxCost float64, minLat float64, maxLat float64, minQ float64, maxQ float64) float64 {
	c := normalize(metricValue(r, "cost_usd"), minCost, maxCost)
	l := normalize(metricValue(r, "latency_ms"), minLat, maxLat)
	q := normalize(metricValue(r, "quality_score"), minQ, maxQ)
	return (weightCost * c) + (weightLatency * l) + (weightQuality * (1 - q))
}

func normalize(v float64, min float64, max float64) float64 {
	if max-min == 0 {
		return 0
	}
	return (v - min) / (max - min)
}

func formatFloat(v float64) string {
	return strconv.FormatFloat(v, 'f', 4, 64)
}

func renderLatencyHistogram(m *Metrics) []string {
	buckets := []int{100, 250, 500, 1000, 2000, 5000}
	counts := make([]int, len(buckets)+1)
	total := 0

	m.mu.Lock()
	samples := append([]int64{}, m.taskLatencyMs...)
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
		"# HELP rechain_task_latency_ms Task latency histogram",
		"# TYPE rechain_task_latency_ms histogram",
	}
	running := 0
	for i, b := range buckets {
		running += counts[i]
		lines = append(lines, "rechain_task_latency_ms_bucket{le=\""+strconv.Itoa(b)+"\"} "+strconv.Itoa(running))
	}
	running += counts[len(counts)-1]
	lines = append(lines, "rechain_task_latency_ms_bucket{le=\"+Inf\"} "+strconv.Itoa(running))
	lines = append(lines, "rechain_task_latency_ms_count "+strconv.Itoa(total))
	return lines
}

func renderRoutingModelLatencyHistogram(m *Metrics) []string {
	buckets := []int{100, 250, 500, 1000, 2000, 5000}
	lines := []string{
		"# HELP rechain_routing_model_latency_ms Routing latency by model",
		"# TYPE rechain_routing_model_latency_ms histogram",
	}

	m.mu.Lock()
	snapshot := map[string][]int64{}
	for model, samples := range m.modelLatencyMs {
		snapshot[model] = append([]int64{}, samples...)
	}
	m.mu.Unlock()

	for model, samples := range snapshot {
		counts := make([]int, len(buckets)+1)
		total := 0
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
		running := 0
		for i, b := range buckets {
			running += counts[i]
			lines = append(lines, "rechain_routing_model_latency_ms_bucket{model=\""+model+"\",le=\""+strconv.Itoa(b)+"\"} "+strconv.Itoa(running))
		}
		running += counts[len(counts)-1]
		lines = append(lines, "rechain_routing_model_latency_ms_bucket{model=\""+model+"\",le=\"+Inf\"} "+strconv.Itoa(running))
		lines = append(lines, "rechain_routing_model_latency_ms_count{model=\""+model+"\"} "+strconv.Itoa(total))
	}
	return lines
}

func estimateQuality(output string, diff string) float64 {
	if output == "" && diff == "" {
		return 0
	}

	stats := diffStats(diff)
	diffLen := stats.totalLines
	outputLen := len(output)
	errCount := errorTokenCount(output)

	score := 0.9
	if diffLen > 800 {
		score -= 0.2
	} else if diffLen > 300 {
		score -= 0.1
	}

	if outputLen < 20 {
		score -= 0.1
	}

	if errCount > 0 {
		score -= 0.1 * float64(minInt(errCount, 3))
	}

	if stats.files > 10 {
		score -= 0.1
	}
	if stats.hunks > 20 {
		score -= 0.1
	}
	if stats.additions+stats.deletions > 200 {
		score -= 0.1
	}
	if stats.deletions > 0 {
		ratio := float64(stats.additions) / float64(stats.deletions)
		if ratio < 0.5 || ratio > 2.0 {
			score -= 0.05
		}
	}

	// Blend with a stable hash component so equal outputs are deterministic.
	sum := sha256.Sum256([]byte(output + "|" + diff))
	hashComponent := float64(sum[0]) / 255.0
	score = (score * 0.7) + (hashComponent * 0.3)

	if score < 0 {
		return 0
	}
	if score > 1 {
		return 1
	}
	return score
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

func minInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
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

type pingState struct {
	okUntil   time.Time
	failUntil time.Time
	backoff   time.Duration
}

type PingService struct {
	mu         sync.Mutex
	m          map[string]*pingState
	ttl        time.Duration
	backoff    time.Duration
	backoffMax time.Duration
	okCount    int
	failCount  int
	skipCount  int
}

type PingHealth struct {
	ModelID       string `json:"model_id"`
	Status        string `json:"status"`
	Available     bool   `json:"available"`
	Cached        bool   `json:"cached"`
	OkUntilUnix   int64  `json:"ok_until_unix"`
	FailUntilUnix int64  `json:"fail_until_unix"`
	BackoffMs     int64  `json:"backoff_ms"`
}

func NewPingService(ttl time.Duration, backoff time.Duration, backoffMax time.Duration) *PingService {
	return &PingService{
		m:          map[string]*pingState{},
		ttl:        ttl,
		backoff:    backoff,
		backoffMax: backoffMax,
	}
}

func (p *PingService) Snapshot() map[string]int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return map[string]int{
		"ok":    p.okCount,
		"fail":  p.failCount,
		"skip":  p.skipCount,
		"items": len(p.m),
	}
}

func (p *PingService) Health(modelID string) PingHealth {
	now := time.Now()
	p.mu.Lock()
	defer p.mu.Unlock()
	st, ok := p.m[modelID]
	if !ok {
		return PingHealth{ModelID: modelID, Status: "unknown"}
	}
	h := PingHealth{
		ModelID:       modelID,
		OkUntilUnix:   st.okUntil.Unix(),
		FailUntilUnix: st.failUntil.Unix(),
		BackoffMs:     st.backoff.Milliseconds(),
	}
	if st.okUntil.After(now) {
		h.Status = "ok"
		h.Available = true
		h.Cached = true
		return h
	}
	if st.failUntil.After(now) {
		h.Status = "fail"
		h.Available = false
		h.Cached = true
		return h
	}
	h.Status = "stale"
	h.Available = false
	return h
}

func (p *PingService) HealthMap(modelIDs []string) map[string]PingHealth {
	out := map[string]PingHealth{}
	for _, m := range modelIDs {
		out[m] = p.Health(m)
	}
	return out
}

func (p *PingService) IsAvailable(modelID string, d HuggingFaceDriver) bool {
	now := time.Now()
	p.mu.Lock()
	st, ok := p.m[modelID]
	if !ok {
		st = &pingState{}
		p.m[modelID] = st
	}
	if st.okUntil.After(now) {
		p.skipCount++
		p.mu.Unlock()
		return true
	}
	if st.failUntil.After(now) {
		p.skipCount++
		p.mu.Unlock()
		return false
	}
	p.mu.Unlock()

	ok = d.pingAvailable(modelID)

	p.mu.Lock()
	if ok {
		p.okCount++
		st.backoff = 0
		st.okUntil = now.Add(p.ttl)
		st.failUntil = time.Time{}
	} else {
		p.failCount++
		if st.backoff == 0 {
			st.backoff = p.backoff
		} else {
			st.backoff *= 2
			if st.backoff > p.backoffMax {
				st.backoff = p.backoffMax
			}
		}
		st.failUntil = now.Add(st.backoff)
	}
	p.mu.Unlock()
	return ok
}

func startHFPingLoop(models []string, registry *DriverRegistry, pingSvc *PingService, interval time.Duration) {
	if interval <= 0 {
		return
	}
	list := registry.HFModelIDs(models)
	if len(list) == 0 {
		return
	}

	var hf *HuggingFaceDriver
	for _, d := range registry.Drivers() {
		if h, ok := d.(HuggingFaceDriver); ok {
			hf = &h
			break
		}
	}
	if hf == nil {
		return
	}

	runPingSweep := func() {
		for _, modelID := range list {
			_ = pingSvc.IsAvailable(modelID, *hf)
		}
	}

	// Prime health cache on boot so /models/health is useful before first tick.
	runPingSweep()

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			runPingSweep()
		}
	}()
}

func fetchCacheMetrics(url string) map[string]int {
	if url == "" {
		return map[string]int{}
	}
	resp, err := http.Get(url + "/cache-metrics")
	if err != nil {
		return map[string]int{}
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return map[string]int{}
	}
	var m map[string]int
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return map[string]int{}
	}
	return m
}

func fetchHealth(baseURL string) bool {
	baseURL = strings.TrimRight(baseURL, "/")
	if baseURL == "" {
		return false
	}
	client := &http.Client{Timeout: 800 * time.Millisecond}
	resp, err := client.Get(baseURL + "/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func fetchPromMetric(baseURL string, metricName string) float64 {
	baseURL = strings.TrimRight(baseURL, "/")
	if baseURL == "" || metricName == "" {
		return 0
	}
	client := &http.Client{Timeout: 1200 * time.Millisecond}
	resp, err := client.Get(baseURL + "/metrics")
	if err != nil {
		return 0
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if !strings.HasPrefix(line, metricName) {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		v, err := strconv.ParseFloat(parts[len(parts)-1], 64)
		if err == nil {
			return v
		}
	}
	return 0
}

func fetchJSONMap(baseURL string, path string) map[string]interface{} {
	baseURL = strings.TrimRight(baseURL, "/")
	if baseURL == "" {
		return map[string]interface{}{}
	}
	client := &http.Client{Timeout: 1200 * time.Millisecond}
	resp, err := client.Get(baseURL + path)
	if err != nil {
		return map[string]interface{}{}
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return map[string]interface{}{}
	}
	out := map[string]interface{}{}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return map[string]interface{}{}
	}
	return out
}

func toFloat64(v interface{}) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case float32:
		return float64(t)
	case int:
		return float64(t)
	case int64:
		return float64(t)
	case int32:
		return float64(t)
	case json.Number:
		n, _ := t.Float64()
		return n
	default:
		return 0
	}
}

func toBool(v interface{}) bool {
	switch t := v.(type) {
	case bool:
		return t
	case float64:
		return t != 0
	case int:
		return t != 0
	default:
		return false
	}
}

func formatDashboardMetric(v interface{}) string {
	switch t := v.(type) {
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(t), 'f', -1, 64)
	case int:
		return strconv.Itoa(t)
	case int64:
		return strconv.FormatInt(t, 10)
	case int32:
		return strconv.FormatInt(int64(t), 10)
	default:
		return "0"
	}
}

func boolToGauge(v interface{}) int {
	if toBool(v) {
		return 1
	}
	return 0
}

func metricValue(r ModelResult, name string) float64 {
	for _, m := range r.Metrics {
		if m.Name == name {
			return m.Value
		}
	}
	return 0
}

func metricValueInternal(r internal.ModelResult, name string) float64 {
	for _, m := range r.Metrics {
		if m.Name == name {
			return m.Value
		}
	}
	return 0
}

func tryQuantumOptimize(results []ModelResult) (ModelResult, bool) {
	base := strings.TrimRight(os.Getenv("QUANTUM_URL"), "/")
	if base == "" {
		return ModelResult{}, false
	}
	type cand struct {
		ID        string  `json:"id"`
		CostUSD   float64 `json:"cost_usd"`
		LatencyMs float64 `json:"latency_ms"`
		Quality   float64 `json:"quality"`
	}
	candidates := []cand{}
	for _, r := range results {
		candidates = append(candidates, cand{
			ID:        r.ModelID,
			CostUSD:   metricValue(r, "cost_usd"),
			LatencyMs: metricValue(r, "latency_ms"),
			Quality:   metricValue(r, "quality_score"),
		})
	}
	payload := map[string]interface{}{
		"schema_version": schemaVersion,
		"objective":      "weighted",
		"candidates":     candidates,
	}
	body, _ := json.Marshal(payload)
	client := &http.Client{Timeout: 1200 * time.Millisecond}
	resp, err := client.Post(base+"/optimize", "application/json", bytes.NewReader(body))
	if err != nil {
		return ModelResult{}, false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return ModelResult{}, false
	}
	var out struct {
		SelectedID string `json:"selected_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return ModelResult{}, false
	}
	for _, r := range results {
		if r.ModelID == out.SelectedID {
			return r, true
		}
	}
	return ModelResult{}, false
}

func fetchRAGContext(ctx context.Context, ragURL string, query string) ([]ContextRef, error) {
	if query == "" {
		return nil, nil
	}

	u := ragURL + "/search?q=" + url.QueryEscape(query)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 800 * time.Millisecond}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("rag search failed")
	}

	var s SearchResult
	if err := json.NewDecoder(resp.Body).Decode(&s); err != nil {
		return nil, err
	}

	out := make([]ContextRef, 0, len(s.Matches))
	for _, m := range s.Matches {
		out = append(out, ContextRef{Type: "file", Path: m, Rev: ""})
	}
	return out, nil
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func promLabelValue(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}

func randString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func envOr(key string, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
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

func parseHFGeneratedText(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	var arr []map[string]interface{}
	if err := json.Unmarshal(data, &arr); err == nil && len(arr) > 0 {
		if v, ok := arr[0]["generated_text"].(string); ok {
			return v
		}
	}
	var obj map[string]interface{}
	if err := json.Unmarshal(data, &obj); err == nil {
		if v, ok := obj["generated_text"].(string); ok {
			return v
		}
	}
	return string(data)
}

func queueDelayForPriority(priority string) time.Duration {
	switch strings.ToLower(strings.TrimSpace(priority)) {
	case "low":
		return 150 * time.Millisecond
	case "normal":
		return 50 * time.Millisecond
	case "high":
		return 0
	default:
		return 50 * time.Millisecond
	}
}

func runWithRetry(ctx context.Context, d Driver, spec TaskSpec, metrics *Metrics) (ModelResult, error) {
	retries := constraintInt(spec.Constraints, "retries", 0)
	backoff := time.Duration(constraintInt(spec.Constraints, "retry_backoff_ms", 200)) * time.Millisecond
	var lastErr error
	for attempt := 0; attempt <= retries; attempt++ {
		res, err := d.Run(ctx, spec)
		if err == nil {
			return res, nil
		}
		lastErr = err
		if metrics != nil && attempt < retries {
			metrics.IncRetry()
		}
		if attempt < retries && backoff > 0 {
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return ModelResult{}, ctx.Err()
			}
		}
	}
	if lastErr == nil {
		lastErr = errors.New("driver failed")
	}
	return ModelResult{}, lastErr
}
