package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"hash/crc32"
	"io"
	"io/fs"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"rechain-ide/shared/logging"

	"go.etcd.io/bbolt"
)

const schemaVersion = "0.1.0"

type IndexRequest struct {
	SchemaVersion string   `json:"schema_version"`
	Repo          string   `json:"repo"`
	Files         []string `json:"files"`
	Root          string   `json:"root"`
	Extensions    []string `json:"extensions"`
	MaxFiles      int      `json:"max_files"`
}

type SearchResult struct {
	SchemaVersion string   `json:"schema_version"`
	Query         string   `json:"query"`
	Matches       []string `json:"matches"`
	Scores        []Score  `json:"scores"`
}

type Score struct {
	Path   string  `json:"path"`
	Score  float64 `json:"score"`
	Source string  `json:"source,omitempty"`
}

type EmbedRequest struct {
	SchemaVersion string `json:"schema_version"`
	Text          string `json:"text"`
}

type EmbedResponse struct {
	SchemaVersion string    `json:"schema_version"`
	Vector        []float64 `json:"vector"`
}

type Store struct {
	mu      sync.Mutex
	files   []string
	chunks  []Chunk
	vectors map[int][]float64
}

type Chunk struct {
	Path  string `json:"path"`
	Start int    `json:"start"`
	Text  string `json:"text"`
}

type CacheMetrics struct {
	mu             sync.Mutex
	hits           int
	misses         int
	purges         int
	evictions      int
	entries        int
	bytes          int
	embedLatencyMs []int64
	tuneUpdates    int
	tuneImports    int
	tuneExports    int
}

var cacheMetrics *CacheMetrics

func (m *CacheMetrics) IncHit() {
	m.mu.Lock()
	m.hits++
	m.mu.Unlock()
}

func (m *CacheMetrics) IncMiss() {
	m.mu.Lock()
	m.misses++
	m.mu.Unlock()
}

func (m *CacheMetrics) IncPurge() {
	m.mu.Lock()
	m.purges++
	m.mu.Unlock()
}

func (m *CacheMetrics) IncEvict(n int, bytes int, entries int) {
	m.mu.Lock()
	m.evictions += n
	m.bytes = bytes
	m.entries = entries
	m.mu.Unlock()
}

func (m *CacheMetrics) ObserveEmbedLatency(ms int64) {
	m.mu.Lock()
	m.embedLatencyMs = append(m.embedLatencyMs, ms)
	if len(m.embedLatencyMs) > 100 {
		m.embedLatencyMs = m.embedLatencyMs[len(m.embedLatencyMs)-100:]
	}
	m.mu.Unlock()
}

func (m *CacheMetrics) Snapshot() map[string]int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return map[string]int{
		"hits":         m.hits,
		"misses":       m.misses,
		"purges":       m.purges,
		"evictions":    m.evictions,
		"entries":      m.entries,
		"bytes":        m.bytes,
		"tune_updates": m.tuneUpdates,
		"tune_imports": m.tuneImports,
		"tune_exports": m.tuneExports,
	}
}

func (m *CacheMetrics) IncTuneUpdate() {
	m.mu.Lock()
	m.tuneUpdates++
	m.mu.Unlock()
}

func (m *CacheMetrics) IncTuneImport() {
	m.mu.Lock()
	m.tuneImports++
	m.mu.Unlock()
}

func (m *CacheMetrics) IncTuneExport() {
	m.mu.Lock()
	m.tuneExports++
	m.mu.Unlock()
}

type SearchConfig struct {
	mu          sync.Mutex
	Lexical     float64 `json:"lexical_weight"`
	Semantic    float64 `json:"semantic_weight"`
	Temperature float64 `json:"temperature"`
	Version     int     `json:"version"`
	UpdatedAt   string  `json:"updated_at"`
}

type persistedHybridConfig struct {
	LexicalWeight  float64 `json:"lexical_weight"`
	SemanticWeight float64 `json:"semantic_weight"`
	Temperature    float64 `json:"temperature"`
	Version        int     `json:"version"`
	UpdatedAt      string  `json:"updated_at"`
}

func (c *SearchConfig) Snapshot() (float64, float64, float64, int, string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Lexical, c.Semantic, c.Temperature, c.Version, c.UpdatedAt
}

func (c *SearchConfig) Set(lexical float64, semantic float64, temperature float64) {
	c.mu.Lock()
	c.Lexical = lexical
	c.Semantic = semantic
	c.Temperature = temperature
	c.Version++
	c.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	c.mu.Unlock()
}

func main() {
	store := &Store{files: []string{}}
	configPath := envOr("RAG_CONFIG_PATH", ".rag-cache/rag-config.db")
	cfg := &SearchConfig{
		Lexical:     envFloat("RAG_WEIGHT_LEXICAL", 0.6),
		Semantic:    envFloat("RAG_WEIGHT_SEMANTIC", 0.4),
		Temperature: envFloat("RAG_TEMPERATURE", 1.0),
		Version:     1,
		UpdatedAt:   time.Now().UTC().Format(time.RFC3339),
	}
	if loaded, err := loadHybridConfig(configPath); err == nil && loaded != nil {
		cfg.Lexical = loaded.LexicalWeight
		cfg.Semantic = loaded.SemanticWeight
		cfg.Temperature = loaded.Temperature
		if loaded.Version > 0 {
			cfg.Version = loaded.Version
		}
		if strings.TrimSpace(loaded.UpdatedAt) != "" {
			cfg.UpdatedAt = loaded.UpdatedAt
		}
	}
	mux := http.NewServeMux()

	cachePath := envOr("RAG_CACHE_PATH", ".rag-cache/embeddings.db")
	cacheMetrics = &CacheMetrics{}
	purgeInterval := time.Duration(envInt("RAG_CACHE_PURGE_INTERVAL_SEC", 300)) * time.Second
	if purgeInterval > 0 {
		go startCachePurge(cachePath, purgeInterval, cacheMetrics)
	}

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	mux.HandleFunc("/cache-metrics", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, cacheMetrics.Snapshot())
	})

	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		snap := cacheMetrics.Snapshot()
		lexWeight, semWeight, temp, version, updatedAt := cfg.Snapshot()
		updatedUnix := toUnix(updatedAt)
		lines := []string{
			"# HELP rechain_rag_cache_hits_total Cache hits",
			"# TYPE rechain_rag_cache_hits_total counter",
			"rechain_rag_cache_hits_total " + strconv.Itoa(snap["hits"]),
			"# HELP rechain_rag_cache_misses_total Cache misses",
			"# TYPE rechain_rag_cache_misses_total counter",
			"rechain_rag_cache_misses_total " + strconv.Itoa(snap["misses"]),
			"# HELP rechain_rag_cache_purges_total Cache purges",
			"# TYPE rechain_rag_cache_purges_total counter",
			"rechain_rag_cache_purges_total " + strconv.Itoa(snap["purges"]),
			"# HELP rechain_rag_cache_evictions_total Cache evictions",
			"# TYPE rechain_rag_cache_evictions_total counter",
			"rechain_rag_cache_evictions_total " + strconv.Itoa(snap["evictions"]),
			"# HELP rechain_rag_cache_entries Current cache entries",
			"# TYPE rechain_rag_cache_entries gauge",
			"rechain_rag_cache_entries " + strconv.Itoa(snap["entries"]),
			"# HELP rechain_rag_cache_bytes Current cache size bytes",
			"# TYPE rechain_rag_cache_bytes gauge",
			"rechain_rag_cache_bytes " + strconv.Itoa(snap["bytes"]),
			"# HELP rechain_rag_weight_lexical Current lexical weight",
			"# TYPE rechain_rag_weight_lexical gauge",
			"rechain_rag_weight_lexical " + formatFloat(lexWeight),
			"# HELP rechain_rag_weight_semantic Current semantic weight",
			"# TYPE rechain_rag_weight_semantic gauge",
			"rechain_rag_weight_semantic " + formatFloat(semWeight),
			"# HELP rechain_rag_temperature Current temperature",
			"# TYPE rechain_rag_temperature gauge",
			"rechain_rag_temperature " + formatFloat(temp),
			"# HELP rechain_rag_hybrid_tune_updates_total Runtime hybrid tune updates",
			"# TYPE rechain_rag_hybrid_tune_updates_total counter",
			"rechain_rag_hybrid_tune_updates_total " + strconv.Itoa(snap["tune_updates"]),
			"# HELP rechain_rag_hybrid_tune_import_total Runtime hybrid tune imports",
			"# TYPE rechain_rag_hybrid_tune_import_total counter",
			"rechain_rag_hybrid_tune_import_total " + strconv.Itoa(snap["tune_imports"]),
			"# HELP rechain_rag_hybrid_tune_export_total Runtime hybrid tune exports",
			"# TYPE rechain_rag_hybrid_tune_export_total counter",
			"rechain_rag_hybrid_tune_export_total " + strconv.Itoa(snap["tune_exports"]),
			"# HELP rechain_rag_hybrid_tune_config_version Runtime hybrid tune config version",
			"# TYPE rechain_rag_hybrid_tune_config_version gauge",
			"rechain_rag_hybrid_tune_config_version " + strconv.Itoa(version),
			"# HELP rechain_rag_hybrid_tune_updated_unix Runtime hybrid tune config updated unix seconds",
			"# TYPE rechain_rag_hybrid_tune_updated_unix gauge",
			"rechain_rag_hybrid_tune_updated_unix " + formatFloat(updatedUnix),
		}
		lines = append(lines, renderEmbedLatencyHistogram(cacheMetrics)...)
		w.Write([]byte(strings.Join(lines, "\n")))
	})

	mux.HandleFunc("/index", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req IndexRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		if req.SchemaVersion == "" {
			req.SchemaVersion = schemaVersion
		}

		files := append([]string{}, req.Files...)
		if len(files) == 0 && req.Root != "" {
			files = scanFiles(req.Root, req.Extensions, req.MaxFiles)
		}
		chunks := buildChunks(files)
		vectors := map[int][]float64{}
		if strings.EqualFold(envOr("RAG_EMBED_INDEX", "false"), "true") {
			vectors = buildChunkVectors(chunks)
		}
		store.mu.Lock()
		store.files = append([]string{}, files...)
		store.chunks = append([]Chunk{}, chunks...)
		store.vectors = vectors
		store.mu.Unlock()

		writeJSON(w, map[string]string{"status": "indexed"})
	})

	mux.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		handleSearchMode(w, r, store, cfg, "")
	})

	mux.HandleFunc("/search/semantic", func(w http.ResponseWriter, r *http.Request) {
		handleSearchMode(w, r, store, cfg, "semantic")
	})

	mux.HandleFunc("/search/lexical", func(w http.ResponseWriter, r *http.Request) {
		handleSearchMode(w, r, store, cfg, "lexical")
	})

	mux.HandleFunc("/search/hybrid", func(w http.ResponseWriter, r *http.Request) {
		handleSearchMode(w, r, store, cfg, "hybrid")
	})

	mux.HandleFunc("/search/hybrid-tune", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			lex, sem, temp, version, updatedAt := cfg.Snapshot()
			writeJSON(w, map[string]interface{}{
				"schema_version":  schemaVersion,
				"lexical_weight":  lex,
				"semantic_weight": sem,
				"temperature":     temp,
				"version":         version,
				"updated_at":      updatedAt,
			})
			return
		}
		if r.Method != http.MethodPost && r.Method != http.MethodPatch {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			LexicalWeight  float64 `json:"lexical_weight"`
			SemanticWeight float64 `json:"semantic_weight"`
			Temperature    float64 `json:"temperature"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		if req.LexicalWeight < 0 || req.SemanticWeight < 0 {
			http.Error(w, "weights must be >= 0", http.StatusBadRequest)
			return
		}
		if req.Temperature <= 0 {
			http.Error(w, "temperature must be > 0", http.StatusBadRequest)
			return
		}
		sum := req.LexicalWeight + req.SemanticWeight
		if sum <= 0 {
			req.LexicalWeight = 0.5
			req.SemanticWeight = 0.5
		} else {
			req.LexicalWeight = req.LexicalWeight / sum
			req.SemanticWeight = req.SemanticWeight / sum
		}
		cfg.Set(req.LexicalWeight, req.SemanticWeight, req.Temperature)
		cacheMetrics.IncTuneUpdate()
		lexical, semantic, temp, version, updatedAt := cfg.Snapshot()
		persisted := persistedHybridConfig{
			LexicalWeight:  lexical,
			SemanticWeight: semantic,
			Temperature:    temp,
			Version:        version,
			UpdatedAt:      updatedAt,
		}
		_ = saveHybridConfig(configPath, persisted)
		_ = appendHybridConfigHistory(configPath, persisted)
		writeJSON(w, map[string]interface{}{
			"schema_version":  schemaVersion,
			"lexical_weight":  req.LexicalWeight,
			"semantic_weight": req.SemanticWeight,
			"temperature":     req.Temperature,
			"version":         version,
			"updated_at":      updatedAt,
		})
	})

	mux.HandleFunc("/search/hybrid-tune/reset", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		lex := envFloat("RAG_WEIGHT_LEXICAL", 0.6)
		sem := envFloat("RAG_WEIGHT_SEMANTIC", 0.4)
		temp := envFloat("RAG_TEMPERATURE", 1.0)
		if temp <= 0 {
			temp = 1.0
		}
		sum := lex + sem
		if sum <= 0 {
			lex = 0.5
			sem = 0.5
		} else {
			lex = lex / sum
			sem = sem / sum
		}
		cfg.Set(lex, sem, temp)
		cacheMetrics.IncTuneUpdate()
		lexical, semantic, temperature, version, updatedAt := cfg.Snapshot()
		persisted := persistedHybridConfig{
			LexicalWeight:  lexical,
			SemanticWeight: semantic,
			Temperature:    temperature,
			Version:        version,
			UpdatedAt:      updatedAt,
		}
		_ = saveHybridConfig(configPath, persisted)
		_ = appendHybridConfigHistory(configPath, persisted)
		writeJSON(w, map[string]interface{}{
			"schema_version":  schemaVersion,
			"lexical_weight":  lexical,
			"semantic_weight": semantic,
			"temperature":     temperature,
			"version":         version,
			"updated_at":      updatedAt,
			"status":          "reset",
		})
	})

	mux.HandleFunc("/search/hybrid-tune/history", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		limit := constraintIntFromQuery(r, "limit", 20)
		if limit <= 0 {
			limit = 20
		}
		if limit > 200 {
			limit = 200
		}
		history, err := listHybridConfigHistory(configPath, limit)
		if err != nil {
			http.Error(w, "history unavailable", http.StatusBadGateway)
			return
		}
		writeJSON(w, map[string]interface{}{
			"schema_version": schemaVersion,
			"count":          len(history),
			"history":        history,
		})
	})

	mux.HandleFunc("/search/hybrid-tune/export", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		lex, sem, temp, version, updatedAt := cfg.Snapshot()
		cacheMetrics.IncTuneExport()
		writeJSON(w, map[string]interface{}{
			"schema_version":  schemaVersion,
			"lexical_weight":  lex,
			"semantic_weight": sem,
			"temperature":     temp,
			"version":         version,
			"updated_at":      updatedAt,
		})
	})

	mux.HandleFunc("/search/hybrid-tune/import", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req persistedHybridConfig
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		if req.LexicalWeight < 0 || req.SemanticWeight < 0 {
			http.Error(w, "weights must be >= 0", http.StatusBadRequest)
			return
		}
		if req.Temperature <= 0 {
			http.Error(w, "temperature must be > 0", http.StatusBadRequest)
			return
		}
		sum := req.LexicalWeight + req.SemanticWeight
		if sum <= 0 {
			req.LexicalWeight = 0.5
			req.SemanticWeight = 0.5
		} else {
			req.LexicalWeight = req.LexicalWeight / sum
			req.SemanticWeight = req.SemanticWeight / sum
		}
		cfg.Set(req.LexicalWeight, req.SemanticWeight, req.Temperature)
		cacheMetrics.IncTuneUpdate()
		cacheMetrics.IncTuneImport()
		lexical, semantic, temperature, version, updatedAt := cfg.Snapshot()
		persisted := persistedHybridConfig{
			LexicalWeight:  lexical,
			SemanticWeight: semantic,
			Temperature:    temperature,
			Version:        version,
			UpdatedAt:      updatedAt,
		}
		_ = saveHybridConfig(configPath, persisted)
		_ = appendHybridConfigHistory(configPath, persisted)
		writeJSON(w, map[string]interface{}{
			"schema_version":  schemaVersion,
			"lexical_weight":  lexical,
			"semantic_weight": semantic,
			"temperature":     temperature,
			"version":         version,
			"updated_at":      updatedAt,
			"status":          "imported",
		})
	})

	mux.HandleFunc("/embed", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req EmbedRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		if req.SchemaVersion == "" {
			req.SchemaVersion = schemaVersion
		}

		vector, err := embedVector(req.Text)
		if err != nil {
			http.Error(w, "embed failed", http.StatusBadRequest)
			return
		}
		resp := EmbedResponse{
			SchemaVersion: schemaVersion,
			Vector:        vector,
		}
		writeJSON(w, resp)
	})

	addr := ":8083"
	log.Printf("rag listening on %s", addr)
	if err := http.ListenAndServe(addr, logging.WithRequestID(mux)); err != nil {
		log.Fatal(err)
	}
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func scoreMatch(q string, path string) float64 {
	if q == "" || path == "" {
		return 0
	}
	h := crc32.ChecksumIEEE([]byte(strings.ToLower(q + "|" + path)))
	return float64(h%1000) / 1000.0
}

func handleSearchMode(w http.ResponseWriter, r *http.Request, store *Store, cfg *SearchConfig, modeOverride string) {
	q := r.URL.Query().Get("q")
	if q == "" {
		http.Error(w, "missing q", http.StatusBadRequest)
		return
	}

	store.mu.Lock()
	files := append([]string{}, store.files...)
	chunks := append([]Chunk{}, store.chunks...)
	vectors := map[int][]float64{}
	for k, v := range store.vectors {
		vectors[k] = v
	}
	store.mu.Unlock()

	mode := strings.ToLower(strings.TrimSpace(modeOverride))
	if mode == "" {
		mode = strings.ToLower(strings.TrimSpace(r.URL.Query().Get("mode")))
	}
	if mode == "" {
		mode = "hybrid"
	}

	matches := []string{}
	scores := []Score{}
	lexWeight, semWeight, temp, _, _ := cfg.Snapshot()
	if temp <= 0 {
		temp = 1.0
	}
	sum := lexWeight + semWeight
	if sum <= 0 {
		lexWeight = 0.5
		semWeight = 0.5
	} else {
		lexWeight = lexWeight / sum
		semWeight = semWeight / sum
	}
	if mode != "semantic" {
		for _, f := range files {
			if strings.Contains(strings.ToLower(f), strings.ToLower(q)) {
				matches = append(matches, f)
				scores = append(scores, Score{Path: f, Score: applyTemp(scoreMatch(q, f), temp) * lexWeight, Source: "lexical"})
			}
		}
		chunkScores := scoreChunks(q, chunks)
		for _, s := range chunkScores {
			s.Score = applyTemp(s.Score, temp) * lexWeight
			scores = append(scores, s)
		}
	}
	if mode != "lexical" {
		embedScores := scoreChunksByEmbedding(q, chunks, vectors)
		for _, s := range embedScores {
			s.Score = applyTemp(s.Score, temp) * semWeight
			scores = append(scores, s)
		}
	}
	scores = rerankScores(scores, q, constraintIntFromQuery(r, "k", 10))

	resp := SearchResult{
		SchemaVersion: schemaVersion,
		Query:         q,
		Matches:       matches,
		Scores:        scores,
	}
	writeJSON(w, resp)
}

func scanFiles(root string, extensions []string, maxFiles int) []string {
	if root == "" {
		return nil
	}
	extSet := map[string]bool{}
	for _, e := range extensions {
		e = strings.ToLower(strings.TrimSpace(e))
		if e != "" && !strings.HasPrefix(e, ".") {
			e = "." + e
		}
		if e != "" {
			extSet[e] = true
		}
	}
	if maxFiles <= 0 {
		maxFiles = 2000
	}
	out := []string{}
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
		if len(extSet) > 0 {
			if !extSet[strings.ToLower(filepath.Ext(path))] {
				return nil
			}
		}
		out = append(out, path)
		if len(out) >= maxFiles {
			return errors.New("limit reached")
		}
		return nil
	})
	return out
}

func buildChunks(files []string) []Chunk {
	maxBytes := envInt("RAG_MAX_FILE_BYTES", 200000)
	linesPerChunk := envInt("RAG_CHUNK_LINES", 40)
	if linesPerChunk <= 0 {
		linesPerChunk = 40
	}
	chunks := []Chunk{}
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		if maxBytes > 0 && len(data) > maxBytes {
			continue
		}
		lines := strings.Split(string(data), "\n")
		for i := 0; i < len(lines); i += linesPerChunk {
			end := i + linesPerChunk
			if end > len(lines) {
				end = len(lines)
			}
			text := strings.Join(lines[i:end], "\n")
			chunks = append(chunks, Chunk{Path: f, Start: i + 1, Text: text})
		}
	}
	return chunks
}

func scoreChunks(q string, chunks []Chunk) []Score {
	ql := strings.ToLower(q)
	out := []Score{}
	for _, c := range chunks {
		if strings.Contains(strings.ToLower(c.Text), ql) {
			out = append(out, Score{Path: c.Path + ":" + strconv.Itoa(c.Start), Score: 0.6, Source: "lexical"})
		}
	}
	return out
}

func buildChunkVectors(chunks []Chunk) map[int][]float64 {
	maxChunks := envInt("RAG_EMBED_MAX_CHUNKS", 500)
	vectors := map[int][]float64{}
	for i, c := range chunks {
		if i >= maxChunks {
			break
		}
		v, err := embedVector(c.Text)
		if err != nil || len(v) == 0 {
			continue
		}
		vectors[i] = v
	}
	return vectors
}

func scoreChunksByEmbedding(q string, chunks []Chunk, vectors map[int][]float64) []Score {
	if !strings.EqualFold(envOr("RAG_EMBED_INDEX", "false"), "true") {
		return nil
	}
	if len(vectors) == 0 || q == "" {
		return nil
	}
	qv, err := embedVector(q)
	if err != nil || len(qv) == 0 {
		return nil
	}
	out := []Score{}
	for i, c := range chunks {
		v, ok := vectors[i]
		if !ok {
			continue
		}
		sim := cosineSimilarity(qv, v)
		if sim <= 0 {
			continue
		}
		out = append(out, Score{Path: c.Path + ":" + strconv.Itoa(c.Start), Score: sim, Source: "semantic"})
	}
	return out
}

func cosineSimilarity(a []float64, b []float64) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	dot := 0.0
	na := 0.0
	nb := 0.0
	for i := 0; i < n; i++ {
		dot += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	if na == 0 || nb == 0 {
		return 0
	}
	return dot / (sqrt(na) * sqrt(nb))
}

func sqrt(v float64) float64 {
	if v <= 0 {
		return 0
	}
	z := v
	for i := 0; i < 8; i++ {
		z -= (z*z - v) / (2 * z)
	}
	return z
}

func formatFloat(v float64) string {
	return strconv.FormatFloat(v, 'f', 4, 64)
}

func toUnix(ts string) float64 {
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return 0
	}
	return float64(t.Unix())
}

func rerankScores(scores []Score, q string, k int) []Score {
	if k <= 0 {
		k = 10
	}
	sort.Slice(scores, func(i, j int) bool {
		if scores[i].Score == scores[j].Score {
			return scores[i].Path < scores[j].Path
		}
		return scores[i].Score > scores[j].Score
	})
	if len(scores) > k {
		scores = scores[:k]
	}
	return scores
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

func applyTemp(score float64, temp float64) float64 {
	if temp <= 0 {
		return score
	}
	if score <= 0 {
		return 0
	}
	if score > 1 {
		score = 1
	}
	v := math.Pow(score, 1.0/temp)
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func embedStub(text string) []float64 {
	if text == "" {
		return []float64{}
	}
	h := crc32.ChecksumIEEE([]byte(text))
	return []float64{
		float64((h>>0)&0xFF) / 255.0,
		float64((h>>8)&0xFF) / 255.0,
		float64((h>>16)&0xFF) / 255.0,
		float64((h>>24)&0xFF) / 255.0,
	}
}

func embedVector(text string) ([]float64, error) {
	if text == "" {
		return []float64{}, nil
	}
	cachePath := envOr("RAG_CACHE_PATH", ".rag-cache/embeddings.db")
	if v, ok := cacheGetBolt(cachePath, text, cacheMetrics); ok {
		return v, nil
	}

	model := strings.TrimSpace(envOr("RAG_EMBEDDING_MODEL", "sentence-transformers/all-MiniLM-L6-v2"))
	if model == "" {
		start := time.Now()
		v := embedStub(text)
		if cacheMetrics != nil {
			cacheMetrics.ObserveEmbedLatency(time.Since(start).Milliseconds())
		}
		_ = cachePutBolt(cachePath, text, v, cacheMetrics)
		return v, nil
	}

	base := strings.TrimRight(envOr("RAG_EMBEDDING_URL", "https://router.huggingface.co/hf-inference/models"), "/")
	endpoint := base + "/" + url.PathEscape(model)

	start := time.Now()
	payload := map[string]interface{}{
		"inputs": text,
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	token := strings.TrimSpace(os.Getenv("RAG_EMBEDDING_TOKEN"))
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: time.Duration(envInt("RAG_EMBEDDING_TIMEOUT_MS", 8000)) * time.Millisecond}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		data, _ := io.ReadAll(resp.Body)
		return nil, errors.New("embed error: " + string(data))
	}

	data, _ := io.ReadAll(resp.Body)
	vector := parseEmbedding(data)
	if cacheMetrics != nil {
		cacheMetrics.ObserveEmbedLatency(time.Since(start).Milliseconds())
	}
	_ = cachePutBolt(cachePath, text, vector, cacheMetrics)
	return vector, nil
}

func renderEmbedLatencyHistogram(m *CacheMetrics) []string {
	buckets := []int{50, 100, 250, 500, 1000, 2000, 5000}
	counts := make([]int, len(buckets)+1)
	total := 0

	m.mu.Lock()
	samples := append([]int64{}, m.embedLatencyMs...)
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
		"# HELP rechain_rag_embed_latency_ms Embedding latency histogram",
		"# TYPE rechain_rag_embed_latency_ms histogram",
	}
	running := 0
	for i, b := range buckets {
		running += counts[i]
		lines = append(lines, "rechain_rag_embed_latency_ms_bucket{le=\""+strconv.Itoa(b)+"\"} "+strconv.Itoa(running))
	}
	running += counts[len(counts)-1]
	lines = append(lines, "rechain_rag_embed_latency_ms_bucket{le=\"+Inf\"} "+strconv.Itoa(running))
	lines = append(lines, "rechain_rag_embed_latency_ms_count "+strconv.Itoa(total))
	return lines
}

func parseEmbedding(data []byte) []float64 {
	var arr []float64
	if err := json.Unmarshal(data, &arr); err == nil && len(arr) > 0 {
		return arr
	}
	var nested [][]float64
	if err := json.Unmarshal(data, &nested); err == nil && len(nested) > 0 {
		return nested[0]
	}
	return []float64{}
}

func loadHybridConfig(path string) (*persistedHybridConfig, error) {
	if strings.TrimSpace(path) == "" {
		return nil, errors.New("empty config path")
	}
	db, err := bbolt.Open(path, 0600, &bbolt.Options{Timeout: 200 * time.Millisecond, ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer db.Close()
	out := &persistedHybridConfig{}
	err = db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("rag_config"))
		if b == nil {
			return errors.New("missing rag_config bucket")
		}
		raw := b.Get([]byte("hybrid_tune"))
		if len(raw) == 0 {
			return errors.New("missing hybrid_tune key")
		}
		return json.Unmarshal(raw, out)
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func saveHybridConfig(path string, cfg persistedHybridConfig) error {
	if strings.TrimSpace(path) == "" {
		return errors.New("empty config path")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	db, err := bbolt.Open(path, 0600, &bbolt.Options{Timeout: 200 * time.Millisecond})
	if err != nil {
		return err
	}
	defer db.Close()
	raw, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	return db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("rag_config"))
		if err != nil {
			return err
		}
		return b.Put([]byte("hybrid_tune"), raw)
	})
}

func appendHybridConfigHistory(path string, cfg persistedHybridConfig) error {
	if strings.TrimSpace(path) == "" {
		return errors.New("empty config path")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	db, err := bbolt.Open(path, 0600, &bbolt.Options{Timeout: 200 * time.Millisecond})
	if err != nil {
		return err
	}
	defer db.Close()
	raw, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	return db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("rag_config_history"))
		if err != nil {
			return err
		}
		key := []byte(strconv.FormatInt(time.Now().UTC().UnixNano(), 10))
		return b.Put(key, raw)
	})
}

func listHybridConfigHistory(path string, limit int) ([]persistedHybridConfig, error) {
	if strings.TrimSpace(path) == "" {
		return nil, errors.New("empty config path")
	}
	db, err := bbolt.Open(path, 0600, &bbolt.Options{Timeout: 200 * time.Millisecond, ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer db.Close()
	out := []persistedHybridConfig{}
	err = db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("rag_config_history"))
		if b == nil {
			return nil
		}
		c := b.Cursor()
		for k, v := c.Last(); k != nil; k, v = c.Prev() {
			var item persistedHybridConfig
			if err := json.Unmarshal(v, &item); err != nil {
				continue
			}
			out = append(out, item)
			if len(out) >= limit {
				break
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
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

type cacheEntry struct {
	Vector []float64 `json:"vector"`
	TS     int64     `json:"ts"`
	Access int64     `json:"access"`
	Size   int       `json:"size"`
}

func cacheGetBolt(path string, text string, metrics *CacheMetrics) ([]float64, bool) {
	if path == "" {
		if metrics != nil {
			metrics.IncMiss()
		}
		return nil, false
	}
	key := []byte(text)
	var entry cacheEntry
	err := withBolt(path, false, func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("embeddings"))
		if b == nil {
			return nil
		}
		v := b.Get(key)
		if v == nil {
			return nil
		}
		return json.Unmarshal(v, &entry)
	})
	if err != nil || len(entry.Vector) == 0 {
		if metrics != nil {
			metrics.IncMiss()
		}
		return nil, false
	}
	ttl := int64(envInt("RAG_CACHE_TTL_SEC", 86400))
	if ttl > 0 && time.Now().Unix()-entry.TS > ttl {
		_ = cacheDeleteBolt(path, text)
		if metrics != nil {
			metrics.IncMiss()
		}
		return nil, false
	}
	_ = cacheTouchBolt(path, text, entry)
	if metrics != nil {
		metrics.IncHit()
	}
	return entry.Vector, true
}

func cachePutBolt(path string, text string, vector []float64, metrics *CacheMetrics) error {
	if path == "" {
		return nil
	}
	entry := cacheEntry{
		Vector: vector,
		TS:     time.Now().Unix(),
		Access: time.Now().Unix(),
		Size:   0,
	}
	data, _ := json.Marshal(entry)
	entry.Size = len(data)
	data, _ = json.Marshal(entry)
	return withBolt(path, true, func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("embeddings"))
		if err != nil {
			return err
		}
		if err := b.Put([]byte(text), data); err != nil {
			return err
		}
		return enforceCacheLimits(tx, metrics)
	})
}

func cacheTouchBolt(path string, text string, entry cacheEntry) error {
	entry.Access = time.Now().Unix()
	data, _ := json.Marshal(entry)
	return withBolt(path, true, func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("embeddings"))
		if b == nil {
			return nil
		}
		return b.Put([]byte(text), data)
	})
}

func cacheDeleteBolt(path string, text string) error {
	if path == "" {
		return nil
	}
	return withBolt(path, true, func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("embeddings"))
		if b == nil {
			return nil
		}
		return b.Delete([]byte(text))
	})
}

func startCachePurge(path string, interval time.Duration, metrics *CacheMetrics) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		_ = purgeExpired(path, metrics)
	}
}

func purgeExpired(path string, metrics *CacheMetrics) error {
	ttl := int64(envInt("RAG_CACHE_TTL_SEC", 86400))
	if ttl <= 0 {
		return nil
	}
	now := time.Now().Unix()
	return withBolt(path, true, func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("embeddings"))
		if b == nil {
			return nil
		}
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var entry cacheEntry
			if err := json.Unmarshal(v, &entry); err != nil {
				_ = c.Delete()
				continue
			}
			if now-entry.TS > ttl {
				_ = c.Delete()
			}
		}
		if metrics != nil {
			metrics.IncPurge()
		}
		return enforceCacheLimits(tx, metrics)
	})
}

func enforceCacheLimits(tx *bbolt.Tx, metrics *CacheMetrics) error {
	maxEntries := envInt("RAG_CACHE_MAX_ENTRIES", 0)
	maxBytes := envInt("RAG_CACHE_MAX_BYTES", 0)
	if maxEntries <= 0 && maxBytes <= 0 {
		return nil
	}
	b := tx.Bucket([]byte("embeddings"))
	if b == nil {
		return nil
	}
	type item struct {
		key    []byte
		access int64
		size   int
	}
	items := []item{}
	totalBytes := 0
	c := b.Cursor()
	for k, v := c.First(); k != nil; k, v = c.Next() {
		var entry cacheEntry
		if err := json.Unmarshal(v, &entry); err != nil {
			_ = c.Delete()
			continue
		}
		size := entry.Size
		if size <= 0 {
			size = len(v)
		}
		items = append(items, item{key: append([]byte{}, k...), access: entry.Access, size: size})
		totalBytes += size
	}
	sort.Slice(items, func(i, j int) bool { return items[i].access < items[j].access })
	evicted := 0
	for (maxEntries > 0 && len(items) > maxEntries) || (maxBytes > 0 && totalBytes > maxBytes) {
		if len(items) == 0 {
			break
		}
		victim := items[0]
		_ = b.Delete(victim.key)
		totalBytes -= victim.size
		evicted++
		items = items[1:]
	}
	if metrics != nil {
		metrics.IncEvict(evicted, totalBytes, len(items))
	}
	return nil
}

func withBolt(path string, write bool, fn func(tx *bbolt.Tx) error) error {
	dir := "."
	if idx := strings.LastIndex(path, "/"); idx != -1 {
		dir = path[:idx]
	}
	if dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	db, err := bbolt.Open(path, 0o644, &bbolt.Options{Timeout: 500 * time.Millisecond})
	if err != nil {
		return err
	}
	defer db.Close()
	if write {
		return db.Update(fn)
	}
	return db.View(fn)
}
