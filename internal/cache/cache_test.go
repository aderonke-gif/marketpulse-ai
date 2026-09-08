package cache

import (
	"testing"
	"time"

	"github.com/olamitunde/marketpulse-ai/internal/model"
)

func TestCacheHitAndExpiry(t *testing.T) {
	c := New(10 * time.Millisecond)
	want := model.Analysis{ArticleID: "a-1", Sentiment: "positive"}
	c.Set("key", want)
	if got, ok := c.Get("key"); !ok || got.ArticleID != want.ArticleID {
		t.Fatalf("expected cache hit, got %#v, %v", got, ok)
	}
	time.Sleep(15 * time.Millisecond)
	if _, ok := c.Get("key"); ok {
		t.Fatal("expected expired item to be removed")
	}
}
