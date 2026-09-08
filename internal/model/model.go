package model

import "time"

type Article struct {
	ID          string    `json:"id"`
	Ticker      string    `json:"ticker"`
	Headline    string    `json:"headline"`
	Body        string    `json:"body"`
	PublishedAt time.Time `json:"published_at"`
}

type Analysis struct {
	ArticleID  string   `json:"article_id"`
	Ticker     string   `json:"ticker"`
	Sentiment  string   `json:"sentiment"`
	Score      float64  `json:"score"`
	Events     []string `json:"events"`
	Summary    string   `json:"summary"`
	Confidence float64  `json:"confidence"`
	Model      string   `json:"model"`
}

type AnalyzeRequest struct {
	Article Article `json:"article"`
}

