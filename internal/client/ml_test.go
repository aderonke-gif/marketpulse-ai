package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/olamitunde/marketpulse-ai/internal/model"
)

func TestAnalyzeSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/analyze" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"article_id":"a-1","sentiment":"positive","model":"test-v1"}`))
	}))
	defer server.Close()

	got, err := NewMLClient(server.URL, time.Second).Analyze(context.Background(), model.Article{ID: "a-1"})
	if err != nil {
		t.Fatalf("Analyze returned error: %v", err)
	}
	if got.ArticleID != "a-1" || got.Sentiment != "positive" || got.Model != "test-v1" {
		t.Fatalf("unexpected analysis: %#v", got)
	}
}

func TestAnalyzeTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	_, err := NewMLClient(server.URL, 5*time.Millisecond).Analyze(context.Background(), model.Article{ID: "a-1"})
	if err == nil || !strings.Contains(err.Error(), "call inference service") {
		t.Fatalf("expected wrapped timeout error, got %v", err)
	}
}

func TestAnalyzeRejectsUpstreamError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "unavailable", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	_, err := NewMLClient(server.URL, time.Second).Analyze(context.Background(), model.Article{})
	if err == nil || !strings.Contains(err.Error(), "503 Service Unavailable") {
		t.Fatalf("expected upstream status error, got %v", err)
	}
}
