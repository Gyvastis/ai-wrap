package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"ai-wrap/internal/models"

	"github.com/bradleyjkemp/cupaloy/v2"
)

const baseURL = "http://localhost:8089"

type apiClient struct {
	client  *http.Client
	baseURL string
}

func newAPIClient() *apiClient {
	return &apiClient{
		client:  &http.Client{Timeout: 30 * time.Second},
		baseURL: baseURL,
	}
}

func (c *apiClient) health() (map[string]interface{}, error) {
	resp, err := c.client.Get(c.baseURL + "/health")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *apiClient) generateContent(model string, req models.GeminiRequest) (*http.Response, []byte, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, nil, err
	}

	url := c.baseURL + "/v1beta/models/" + model + ":generateContent"
	httpResp, err := c.client.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, nil, err
	}
	defer httpResp.Body.Close()

	bodyBytes, _ := io.ReadAll(httpResp.Body)

	return httpResp, bodyBytes, nil
}

func TestHealth(t *testing.T) {
	client := newAPIClient()

	health, err := client.health()
	if err != nil {
		t.Fatalf("health check failed: %v", err)
	}

	status, ok := health["status"].(string)
	if !ok || status != "ok" {
		t.Errorf("expected status 'ok', got %v", health["status"])
	}

	models, ok := health["models"].([]interface{})
	if !ok {
		t.Fatalf("expected models array, got %v", health["models"])
	}

	if len(models) == 0 {
		t.Error("expected at least one model, got 0")
	}

	t.Logf("✓ health check passed: %d models loaded", len(models))
}

func TestInvalidModel(t *testing.T) {
	client := newAPIClient()

	req := models.GeminiRequest{
		Contents: []models.Content{
			{Parts: []models.Part{{Text: "test"}}},
		},
	}

	httpResp, bodyBytes, err := client.generateContent("invalid-model", req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if httpResp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", httpResp.StatusCode)
	}

	var bodyMap map[string]interface{}
	json.Unmarshal(bodyBytes, &bodyMap)
	if bodyMap["error"] == nil {
		t.Error("expected error in response")
	}

	t.Log("✓ invalid model correctly rejected")
}

func TestValidRequest(t *testing.T) {
	client := newAPIClient()

	temp := 0.1
	prompt := "what is 3+3? answer in one word"
	req := models.GeminiRequest{
		Contents: []models.Content{
			{Parts: []models.Part{{Text: prompt}}},
		},
		GenerationConfig: models.GenerationConfig{
			Temperature: &temp,
		},
	}

	httpResp, bodyBytes, err := client.generateContent("gemini-2.0-flash", req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", httpResp.StatusCode)
	}

	var resp models.GeminiResponse
	json.Unmarshal(bodyBytes, &resp)

	if len(resp.Candidates) == 0 {
		t.Fatal("expected at least one candidate in response")
	}

	text := resp.Candidates[0].Content.Parts[0].Text
	if text == "" {
		t.Error("expected non-empty response text")
	}

	costInput := httpResp.Header.Get("X-Cost-Input")
	costOutput := httpResp.Header.Get("X-Cost-Output")
	costTotal := httpResp.Header.Get("X-Cost-Total")
	cacheStatus := httpResp.Header.Get("X-Cache-Status")
	keySource := httpResp.Header.Get("X-Key-Source")

	if costInput == "" || costOutput == "" || costTotal == "" {
		t.Error("expected cost headers in response")
	}

	if cacheStatus == "" {
		t.Error("expected cache status header")
	}

	if keySource == "" {
		t.Error("expected key source header")
	}

	t.Logf("✓ valid request succeeded")
	t.Logf("  response: %s", text)
	t.Logf("  cost: input=%s output=%s total=%s", costInput, costOutput, costTotal)
	t.Logf("  cache: %s, key: %s", cacheStatus, keySource)
}

func TestCaching(t *testing.T) {
	client := newAPIClient()

	temp := 0.1
	prompt := "what is 5+5? answer in one word"
	req := models.GeminiRequest{
		Contents: []models.Content{
			{Parts: []models.Part{{Text: prompt}}},
		},
		GenerationConfig: models.GenerationConfig{
			Temperature: &temp,
		},
	}

	httpResp1, _, err := client.generateContent("gemini-2.0-flash", req)
	if err != nil {
		t.Fatalf("first request failed: %v", err)
	}

	cacheStatus1 := httpResp1.Header.Get("X-Cache-Status")
	if cacheStatus1 != "MISS" {
		t.Logf("warning: first request cache status was %s, expected MISS", cacheStatus1)
	}

	time.Sleep(100 * time.Millisecond)

	httpResp2, _, err := client.generateContent("gemini-2.0-flash", req)
	if err != nil {
		t.Fatalf("second request failed: %v", err)
	}

	cacheStatus2 := httpResp2.Header.Get("X-Cache-Status")
	if cacheStatus2 != "HIT" {
		t.Errorf("expected cache HIT on second request, got %s", cacheStatus2)
	} else {
		t.Log("✓ cache hit on second identical request")
	}
}

func TestHighTemperatureNoCache(t *testing.T) {
	client := newAPIClient()

	temp := 0.9
	req := models.GeminiRequest{
		Contents: []models.Content{
			{Parts: []models.Part{{Text: "tell me a random fact"}}},
		},
		GenerationConfig: models.GenerationConfig{
			Temperature: &temp,
		},
	}

	httpResp, _, err := client.generateContent("gemini-2.0-flash", req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	cacheStatus := httpResp.Header.Get("X-Cache-Status")
	if cacheStatus != "MISS" {
		t.Errorf("expected cache MISS for high temp (0.9), got %s", cacheStatus)
	}

	t.Log("✓ high temperature request not cached")
}

func TestCostBlocking(t *testing.T) {
	client := newAPIClient()

	maxTokens := 100000
	longText := make([]byte, 50000)
	for i := range longText {
		longText[i] = 'a'
	}

	req := models.GeminiRequest{
		Contents: []models.Content{
			{Parts: []models.Part{{Text: string(longText)}}},
		},
		GenerationConfig: models.GenerationConfig{
			MaxOutputTokens: &maxTokens,
		},
	}

	httpResp, bodyBytes, err := client.generateContent("gemini-2.5-pro", req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if httpResp.StatusCode != http.StatusPaymentRequired {
		t.Errorf("expected status 402 Payment Required, got %d", httpResp.StatusCode)
	}

	var bodyMap map[string]interface{}
	json.Unmarshal(bodyBytes, &bodyMap)

	if bodyMap["error"] == nil {
		t.Error("expected error message in response")
	}

	predictedCost, ok := bodyMap["predicted_cost"].(float64)
	if !ok {
		t.Error("expected predicted_cost in response")
	}

	maxCost, ok := bodyMap["max_cost"].(float64)
	if !ok {
		t.Error("expected max_cost in response")
	}

	t.Log("✓ cost blocking works correctly")
	t.Logf("  predicted: $%.6f, max: $%.6f", predictedCost, maxCost)
}

func TestVisionRequest(t *testing.T) {
	client := newAPIClient()
	optimizer := NewImageOptimizer()

	base64Image, err := optimizer.OptimizeAndEncode("example_pricing_table.png")
	if err != nil {
		t.Fatalf("failed to optimize image: %v", err)
	}

	temp := 0.1
	maxTokens := 8192
	req := models.GeminiRequest{
		Contents: []models.Content{
			{
				Parts: []models.Part{
					{Text: "extract all data from this pricing table. return as markdown table with exact values. include any header/footer text found."},
					{InlineData: &models.InlineData{
						MimeType: "image/jpeg",
						Data:     base64Image,
					}},
				},
			},
		},
		GenerationConfig: models.GenerationConfig{
			Temperature:     &temp,
			MaxOutputTokens: &maxTokens,
		},
	}

	httpResp, bodyBytes, err := client.generateContent("gemini-2.0-flash", req)
	if err != nil {
		t.Fatalf("vision request failed: %v", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d. response: %s", httpResp.StatusCode, string(bodyBytes))
	}

	var resp models.GeminiResponse
	if err := json.Unmarshal(bodyBytes, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(resp.Candidates) == 0 {
		t.Fatal("expected at least one candidate in response")
	}

	text := resp.Candidates[0].Content.Parts[0].Text
	if text == "" {
		t.Error("expected non-empty response text")
	}

	cupaloy.SnapshotT(t, text)

	costTotal := httpResp.Header.Get("X-Cost-Total")
	cacheStatus := httpResp.Header.Get("X-Cache-Status")

	t.Log("✓ vision request succeeded")
	t.Logf("  cost: %s", costTotal)
	t.Logf("  cache: %s", cacheStatus)
}
