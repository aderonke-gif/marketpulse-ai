package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/olamitunde/marketpulse-ai/internal/cache"
	"github.com/olamitunde/marketpulse-ai/internal/client"
)

func newTestAPI(t *testing.T, upstream http.Handler) (http.Handler, *atomic.Int32) {
	t.Helper()
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		upstream.ServeHTTP(w, r)
	}))
	t.Cleanup(server.Close)
	return New(client.NewMLClient(server.URL, time.Second), cache.New(time.Minute)).Routes(), &calls
}

func post(t *testing.T, handler http.Handler, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/v1/analyze", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	return recorder
}

func TestAnalyzeRejectsWhitespaceOnlyText(t *testing.T) {
	handler, calls := newTestAPI(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	response := post(t, handler, `{"article":{"headline":"  ","body":"content"}}`)
	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d: %s", response.Code, response.Body.String())
	}
	if calls.Load() != 0 {
		t.Fatalf("invalid input reached inference service %d times", calls.Load())
	}
}

func TestAnalyzeCachesSuccessfulInference(t *testing.T) {
	handler, calls := newTestAPI(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"article_id":"a-1","ticker":"LLY","sentiment":"positive","model":"test-v1"}`)
	}))
	body := `{"article":{"id":"a-1","ticker":"LLY","headline":"Lilly beats estimates","body":"Revenue growth was strong."}}`

	first := post(t, handler, body)
	second := post(t, handler, body)
	if first.Code != http.StatusOK || first.Header().Get("X-Cache") != "MISS" {
		t.Fatalf("expected first request cache miss, got %d %q", first.Code, first.Header().Get("X-Cache"))
	}
	if second.Code != http.StatusOK || second.Header().Get("X-Cache") != "HIT" {
		t.Fatalf("expected second request cache hit, got %d %q", second.Code, second.Header().Get("X-Cache"))
	}
	if calls.Load() != 1 {
		t.Fatalf("expected one inference call, got %d", calls.Load())
	}

	metricsReq := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	metrics := httptest.NewRecorder()
	handler.ServeHTTP(metrics, metricsReq)
	if !strings.Contains(metrics.Body.String(), "marketpulse_requests_total 2") ||
		!strings.Contains(metrics.Body.String(), "marketpulse_cache_hits_total 1") {
		t.Fatalf("unexpected metrics: %s", metrics.Body.String())
	}
}

func TestAnalyzeMapsUpstreamFailureToBadGateway(t *testing.T) {
	handler, _ := newTestAPI(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "unavailable", http.StatusServiceUnavailable)
	}))

	response := post(t, handler, `{"article":{"headline":"Headline","body":"Body"}}`)
	if response.Code != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d: %s", response.Code, response.Body.String())
	}
	if response.Body.String() != "{\"error\":\"inference temporarily unavailable\"}\n" {
		t.Fatalf("expected sanitized error, got: %s", response.Body.String())
	}
}
