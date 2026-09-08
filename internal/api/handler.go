package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/olamitunde/marketpulse-ai/internal/cache"
	"github.com/olamitunde/marketpulse-ai/internal/client"
	"github.com/olamitunde/marketpulse-ai/internal/model"
)

type Handler struct {
	ml          *client.MLClient
	cache       *cache.Cache
	requests    atomic.Uint64
	cacheHits   atomic.Uint64
	inferenceMs atomic.Uint64
}

func New(ml *client.MLClient, cache *cache.Cache) *Handler {
	return &Handler{ml: ml, cache: cache}
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("GET /metrics", h.metrics)
	mux.HandleFunc("POST /v1/analyze", h.analyze)
	return h.withMiddleware(mux)
}

func (h *Handler) analyze(w http.ResponseWriter, r *http.Request) {
	h.requests.Add(1)
	var req model.AnalyzeRequest
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request: " + err.Error()})
		return
	}
	if strings.TrimSpace(req.Article.Headline) == "" || strings.TrimSpace(req.Article.Body) == "" {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "headline and body are required"})
		return
	}
	key := contentKey(req.Article)
	if result, ok := h.cache.Get(key); ok {
		h.cacheHits.Add(1)
		w.Header().Set("X-Cache", "HIT")
		writeJSON(w, http.StatusOK, result)
		return
	}
	start := time.Now()
	result, err := h.ml.Analyze(r.Context(), req.Article)
	h.inferenceMs.Store(uint64(time.Since(start).Milliseconds()))
	if err != nil {
		slog.Error("inference failed", "error", err, "article_id", req.Article.ID)
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "inference temporarily unavailable"})
		return
	}
	h.cache.Set(key, result)
	w.Header().Set("X-Cache", "MISS")
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "gateway"})
}

func (h *Handler) metrics(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	_, _ = fmt.Fprintf(w, "marketpulse_requests_total %d\nmarketpulse_cache_hits_total %d\nmarketpulse_last_inference_ms %d\n", h.requests.Load(), h.cacheHits.Load(), h.inferenceMs.Load())
}

func (h *Handler) withMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
		slog.Info("request", "method", r.Method, "path", r.URL.Path, "duration_ms", time.Since(start).Milliseconds())
	})
}

func contentKey(a model.Article) string {
	sum := sha256.Sum256([]byte(a.Ticker + "\x00" + a.Headline + "\x00" + a.Body))
	return hex.EncodeToString(sum[:])
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
