package store

import (
	"time"

	"ai-wrap/internal/models"
)

type RequestLog struct {
	ID            string                 `bson:"_id,omitempty"`
	Timestamp     time.Time              `bson:"timestamp"`
	Model         string                 `bson:"model"`
	Request       models.GeminiRequest   `bson:"request"`
	Response      *models.GeminiResponse `bson:"response,omitempty"`
	StatusCode    int                    `bson:"status_code"`
	Success       bool                   `bson:"success"`
	Error         string                 `bson:"error,omitempty"`
	Cost          models.Cost            `bson:"cost"`
	Temperature   float64                `bson:"temperature"`
	KeySource     string                 `bson:"key_source"`
	CacheHit      bool                   `bson:"cache_hit"`
	RequestHash   string                 `bson:"request_hash"`
	DurationMs    int64                  `bson:"duration_ms"`
	PromptTokens  int                    `bson:"prompt_tokens"`
	OutputTokens  int                    `bson:"output_tokens"`
	TotalTokens   int                    `bson:"total_tokens"`
	IsVision      bool                   `bson:"is_vision"`
}
