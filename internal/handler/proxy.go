package handler

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"ai-wrap/internal/cache"
	"ai-wrap/internal/client"
	"ai-wrap/internal/config"
	"ai-wrap/internal/keymanager"
	"ai-wrap/internal/models"
	"ai-wrap/internal/store"

	"github.com/gin-gonic/gin"
)

type ProxyHandler struct {
	cfg    *config.Config
	cache  *cache.RedisCache
	store  *store.MongoStore
	client *client.GeminiClient
	km     *keymanager.KeyManager
}

func NewProxyHandler(cfg *config.Config, redisCache *cache.RedisCache, mongoStore *store.MongoStore, geminiClient *client.GeminiClient, km *keymanager.KeyManager) *ProxyHandler {
	return &ProxyHandler{
		cfg:    cfg,
		cache:  redisCache,
		store:  mongoStore,
		client: geminiClient,
		km:     km,
	}
}

func (h *ProxyHandler) logAsync(log *store.RequestLog) {
	go func() {
		if err := h.store.LogRequest(log); err != nil {
			fmt.Printf("failed to log request: %v\n", err)
		}
	}()
}

func (h *ProxyHandler) Handle(c *gin.Context) {
	path := c.Param("path")
	userAPIKey := c.Query("key")

	parts := strings.Split(strings.TrimPrefix(path, "/"), ":")
	if len(parts) != 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid path format, expected /model:action"})
		return
	}

	model := parts[0]
	action := parts[1]

	if action != "generateContent" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only generateContent action is supported"})
		return
	}

	modelCost, exists := h.cfg.GetModelCost(model)
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("model '%s' not allowed. only models defined in config are permitted", model),
		})
		return
	}

	var req models.GeminiRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	predictedCost := h.predictCost(req, modelCost)
	if h.cfg.Costs.MaxCost > 0 && predictedCost > h.cfg.Costs.MaxCost {
		c.JSON(http.StatusPaymentRequired, gin.H{
			"error":          fmt.Sprintf("predicted cost $%.6f exceeds maximum allowed cost $%.6f", predictedCost, h.cfg.Costs.MaxCost),
			"predicted_cost": predictedCost,
			"max_cost":       h.cfg.Costs.MaxCost,
		})
		return
	}

	temp := h.getTemperature(req)
	requestHash := store.HashRequest(req)
	startTime := time.Now()
	ctx := c.Request.Context()

	cacheEnabled := temp <= h.cfg.Cache.MaxTemp
	var cached *models.GeminiResponse
	var cachedCost models.Cost
	var cacheSource string

	if cacheEnabled {
		cached, _ = h.cache.Get(ctx, requestHash)
		if cached != nil {
			cacheSource = "redis"
		} else {
			dbLog, _ := h.store.FindCached(requestHash)
			if dbLog != nil && dbLog.Response != nil {
				cached = dbLog.Response
				cacheSource = "mongodb"
				if err := h.cache.Set(ctx, requestHash, cached); err != nil {
					log.Printf("failed to populate redis from mongodb: %v", err)
				}
			}
		}

		if cached != nil {
			cachedCost = h.calculateCost(cached.UsageMetadata, modelCost)
			log.Printf("%s cache hit for model %s (saved $%.6f)", cacheSource, model, cachedCost.Total)
			h.addCostHeaders(c, cachedCost, true, userAPIKey)
			c.JSON(http.StatusOK, cached)

			h.logAsync(&store.RequestLog{
				Timestamp:    time.Now(),
				Model:        model,
				Request:      req,
				Response:     cached,
				StatusCode:   http.StatusOK,
				Success:      true,
				Cost:         cachedCost,
				Temperature:  temp,
				KeySource:    h.getKeySource(userAPIKey),
				CacheHit:     true,
				RequestHash:  requestHash,
				DurationMs:   0,
				PromptTokens: cached.UsageMetadata.PromptTokenCount,
				OutputTokens: cached.UsageMetadata.CandidatesTokenCount,
				TotalTokens:  cached.UsageMetadata.TotalTokenCount,
				IsVision:     h.isVisionRequest(req),
			})

			return
		}
	}

	resp, statusCode, err := h.client.GenerateContent(model, req, userAPIKey)
	duration := time.Since(startTime)

	success := err == nil && statusCode == http.StatusOK
	var cost models.Cost
	var errorMsg string

	if success {
		cost = h.calculateCost(resp.UsageMetadata, modelCost)

		if cacheEnabled {
			if err := h.cache.Set(ctx, requestHash, &resp); err != nil {
				log.Printf("failed to cache response: %v", err)
			}
		}
	} else {
		errorMsg = err.Error()
		log.Printf("gemini api error: %v", err)
	}

	requestLog := &store.RequestLog{
		Timestamp:   time.Now(),
		Model:       model,
		Request:     req,
		StatusCode:  statusCode,
		Success:     success,
		Error:       errorMsg,
		Cost:        cost,
		Temperature: temp,
		KeySource:   h.getKeySource(userAPIKey),
		CacheHit:    false,
		RequestHash: requestHash,
		DurationMs:  duration.Milliseconds(),
		IsVision:    h.isVisionRequest(req),
	}

	if success {
		requestLog.Response = &resp
		requestLog.PromptTokens = resp.UsageMetadata.PromptTokenCount
		requestLog.OutputTokens = resp.UsageMetadata.CandidatesTokenCount
		requestLog.TotalTokens = resp.UsageMetadata.TotalTokenCount
	}

	h.logAsync(requestLog)

	if err != nil {
		if resp.Error != nil && resp.Error.Code != 0 {
			c.JSON(statusCode, resp)
		} else {
			c.JSON(statusCode, gin.H{"error": err.Error()})
		}
		return
	}

	h.addCostHeaders(c, cost, false, userAPIKey)
	c.JSON(http.StatusOK, resp)
}

func (h *ProxyHandler) calculateCost(usage models.UsageMetadata, modelCost config.ModelCost) models.Cost {
	inputCost := float64(usage.PromptTokenCount) * modelCost.Input / 1_000_000
	outputCost := float64(usage.CandidatesTokenCount) * modelCost.Output / 1_000_000

	return models.Cost{
		Input:  inputCost,
		Output: outputCost,
		Total:  inputCost + outputCost,
	}
}

func (h *ProxyHandler) getTemperature(req models.GeminiRequest) float64 {
	if req.GenerationConfig.Temperature != nil {
		return *req.GenerationConfig.Temperature
	}
	return 1.0
}

func (h *ProxyHandler) getKeySource(userAPIKey string) string {
	if userAPIKey != "" {
		return "user"
	}
	return "pool"
}

func (h *ProxyHandler) predictCost(req models.GeminiRequest, modelCost config.ModelCost) float64 {
	var totalChars int
	hasImage := false

	for _, content := range req.Contents {
		for _, part := range content.Parts {
			totalChars += len(part.Text)
			if part.InlineData != nil {
				hasImage = true
			}
		}
	}

	estimatedPromptTokens := totalChars / 4

	if hasImage {
		estimatedPromptTokens += 258
	}

	maxOutputTokens := 8192
	if req.GenerationConfig.MaxOutputTokens != nil {
		maxOutputTokens = *req.GenerationConfig.MaxOutputTokens
	}

	inputCost := float64(estimatedPromptTokens) * modelCost.Input / 1_000_000
	outputCost := float64(maxOutputTokens) * modelCost.Output / 1_000_000

	return inputCost + outputCost
}

func (h *ProxyHandler) addCostHeaders(c *gin.Context, cost models.Cost, cached bool, userAPIKey string) {
	c.Header("X-Cost-Input", fmt.Sprintf("%.6f", cost.Input))
	c.Header("X-Cost-Output", fmt.Sprintf("%.6f", cost.Output))
	c.Header("X-Cost-Total", fmt.Sprintf("%.6f", cost.Total))

	cacheStatus := "MISS"
	if cached {
		cacheStatus = "HIT"
	}
	c.Header("X-Cache-Status", cacheStatus)
	c.Header("X-Key-Source", h.getKeySource(userAPIKey))
}

func (h *ProxyHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"models": h.getModelList(),
	})
}

func (h *ProxyHandler) getModelList() []string {
	models := make([]string, 0, len(h.cfg.Costs.Models))
	for _, model := range h.cfg.Costs.Models {
		models = append(models, model.Name)
	}
	return models
}

func (h *ProxyHandler) isVisionRequest(req models.GeminiRequest) bool {
	for _, content := range req.Contents {
		for _, part := range content.Parts {
			if part.InlineData != nil {
				return true
			}
		}
	}
	return false
}

func (h *ProxyHandler) ExtractModel(path string) string {
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if part == "models" && i+1 < len(parts) {
			modelAction := parts[i+1]
			return strings.Split(modelAction, ":")[0]
		}
	}
	return ""
}
