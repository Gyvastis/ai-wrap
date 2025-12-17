package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"ai-wrap/internal/config"
	"ai-wrap/internal/keymanager"
	"ai-wrap/internal/models"
)

type GeminiClient struct {
	cfg *config.Config
	km  *keymanager.KeyManager
}

func NewGeminiClient(cfg *config.Config, km *keymanager.KeyManager) *GeminiClient {
	return &GeminiClient{
		cfg: cfg,
		km:  km,
	}
}

func (c *GeminiClient) GenerateContent(model string, req models.GeminiRequest, userAPIKey string) (models.GeminiResponse, int, error) {
	if userAPIKey != "" {
		return c.call(model, req, userAPIKey)
	}

	totalKeys := c.km.ActiveCount()
	if totalKeys == 0 {
		return models.GeminiResponse{}, http.StatusUnauthorized, fmt.Errorf("no api key provided and no keys available in pool")
	}

	apiKey := c.km.GetKey()
	triedKeys := 1

	for {
		resp, statusCode, err := c.call(model, req, apiKey)

		if err == nil {
			return resp, statusCode, nil
		}

		if !c.shouldRetry(statusCode) {
			return resp, statusCode, err
		}

		if statusCode == http.StatusForbidden {
			if markErr := c.km.MarkInactive(apiKey); markErr != nil {
				fmt.Printf("failed to mark key as inactive: %v\n", markErr)
			} else {
				fmt.Printf("marked key as inactive due to 403 response\n")
			}
		}

		if triedKeys >= totalKeys {
			return resp, statusCode, err
		}

		apiKey = c.km.RotateKey(apiKey)
		triedKeys++
	}
}

func (c *GeminiClient) shouldRetry(statusCode int) bool {
	switch statusCode {
	case http.StatusBadRequest, http.StatusNotFound:
		return false
	case http.StatusForbidden, http.StatusTooManyRequests,
		http.StatusInternalServerError, http.StatusBadGateway,
		http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true
	default:
		return true
	}
}

func (c *GeminiClient) call(model string, req models.GeminiRequest, apiKey string) (models.GeminiResponse, int, error) {
	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", c.cfg.Gemini.APIURL, model, apiKey)

	body, err := json.Marshal(req)
	if err != nil {
		return models.GeminiResponse{}, http.StatusInternalServerError, err
	}

	client := &http.Client{
		Timeout: time.Duration(c.cfg.Gemini.Timeout) * time.Second,
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return models.GeminiResponse{}, http.StatusInternalServerError, err
	}

	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := client.Do(httpReq)
	if err != nil {
		return models.GeminiResponse{}, http.StatusInternalServerError, err
	}
	defer httpResp.Body.Close()

	bodyBytes, _ := io.ReadAll(httpResp.Body)

	if httpResp.StatusCode != http.StatusOK {
		var errResp models.GeminiResponse
		json.Unmarshal(bodyBytes, &errResp)
		return errResp, httpResp.StatusCode, fmt.Errorf("gemini api returned %d: %s", httpResp.StatusCode, string(bodyBytes))
	}

	var resp models.GeminiResponse
	if err := json.Unmarshal(bodyBytes, &resp); err != nil {
		return models.GeminiResponse{}, http.StatusInternalServerError, err
	}

	return resp, httpResp.StatusCode, nil
}
