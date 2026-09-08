package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/aderonke-gif/marketpulse-ai/internal/model"
)

type MLClient struct {
	baseURL string
	http    *http.Client
}

func NewMLClient(baseURL string, timeout time.Duration) *MLClient {
	return &MLClient{baseURL: baseURL, http: &http.Client{Timeout: timeout}}
}

func (c *MLClient) Analyze(ctx context.Context, article model.Article) (model.Analysis, error) {
	payload, err := json.Marshal(model.AnalyzeRequest{Article: article})
	if err != nil {
		return model.Analysis{}, fmt.Errorf("encode inference request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/analyze", bytes.NewReader(payload))
	if err != nil {
		return model.Analysis{}, fmt.Errorf("create inference request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return model.Analysis{}, fmt.Errorf("call inference service: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return model.Analysis{}, fmt.Errorf("inference service returned %s", resp.Status)
	}
	var result model.Analysis
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return model.Analysis{}, fmt.Errorf("decode inference response: %w", err)
	}
	return result, nil
}
