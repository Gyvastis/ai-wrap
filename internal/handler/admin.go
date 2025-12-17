package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"ai-wrap/internal/models"
	"ai-wrap/internal/store"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

type AdminHandler struct {
	store *store.MongoStore
}

func NewAdminHandler(store *store.MongoStore) *AdminHandler {
	return &AdminHandler{store: store}
}

type Stats struct {
	TotalRequests   int64   `json:"total_requests"`
	SuccessfulReqs  int64   `json:"successful_requests"`
	FailedReqs      int64   `json:"failed_requests"`
	CacheHits       int64   `json:"cache_hits"`
	TotalCost       float64 `json:"total_cost"`
	AvgResponseTime int64   `json:"avg_response_time_ms"`
}

type RequestsResponse struct {
	Requests   []store.RequestLog `json:"requests"`
	Page       int                `json:"page"`
	PerPage    int                `json:"per_page"`
	Total      int64              `json:"total"`
	TotalPages int                `json:"total_pages"`
}

func (h *AdminHandler) GetStats(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	duration := c.DefaultQuery("duration", "24h")
	filter := h.getTimeFilter(duration)

	collection := h.store.GetCollection()

	// single aggregation for all stats
	pipeline := []bson.M{
		{"$match": filter},
		{"$group": bson.M{
			"_id":               nil,
			"total":             bson.M{"$sum": 1},
			"successful":        bson.M{"$sum": bson.M{"$cond": []interface{}{"$success", 1, 0}}},
			"failed":            bson.M{"$sum": bson.M{"$cond": []interface{}{"$success", 0, 1}}},
			"cache_hits":        bson.M{"$sum": bson.M{"$cond": []interface{}{"$cache_hit", 1, 0}}},
			"total_cost":        bson.M{"$sum": "$cost.total"},
			"avg_response_time": bson.M{"$avg": "$duration_ms"},
		}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cursor.Close(ctx)

	var result struct {
		Total           int64   `bson:"total"`
		Successful      int64   `bson:"successful"`
		Failed          int64   `bson:"failed"`
		CacheHits       int64   `bson:"cache_hits"`
		TotalCost       float64 `bson:"total_cost"`
		AvgResponseTime float64 `bson:"avg_response_time"`
	}

	if cursor.Next(ctx) {
		cursor.Decode(&result)
	}

	stats := Stats{
		TotalRequests:   result.Total,
		SuccessfulReqs:  result.Successful,
		FailedReqs:      result.Failed,
		CacheHits:       result.CacheHits,
		TotalCost:       result.TotalCost,
		AvgResponseTime: int64(result.AvgResponseTime),
	}

	c.JSON(http.StatusOK, stats)
}

func (h *AdminHandler) GetRequests(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	collection := h.store.GetCollection()

	// use estimated count (fast, uses collection metadata)
	total, err := collection.EstimatedDocumentCount(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// calculate pagination
	skip := (page - 1) * perPage
	totalPages := int((total + int64(perPage) - 1) / int64(perPage))

	// get paginated requests (projection excludes request/response bodies)
	requests, err := h.store.FindPaginated(ctx, skip, perPage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := RequestsResponse{
		Requests:   requests,
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	}

	c.JSON(http.StatusOK, response)
}

func (h *AdminHandler) GetRequest(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id required"})
		return
	}

	log, err := h.store.FindByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	h.redactImageData(log)
	c.JSON(http.StatusOK, log)
}

func (h *AdminHandler) redactImageData(log *store.RequestLog) {
	for i := range log.Request.Contents {
		for j := range log.Request.Contents[i].Parts {
			part := &log.Request.Contents[i].Parts[j]
			if part.InlineData != nil {
				part.InlineData = &models.InlineData{
					MimeType: part.InlineData.MimeType,
					Data:     "[redacted]",
				}
			}
		}
	}
}

func (h *AdminHandler) getTimeFilter(duration string) bson.M {
	var since time.Time
	switch duration {
	case "7d":
		since = time.Now().Add(-7 * 24 * time.Hour)
	default:
		since = time.Now().Add(-24 * time.Hour)
	}
	return bson.M{"timestamp": bson.M{"$gte": since}}
}

type TimeSeriesData struct {
	Timestamp string `json:"timestamp"`
	Count     int    `json:"count"`
}

func (h *AdminHandler) GetTimeSeries(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	duration := c.DefaultQuery("duration", "24h")
	filter := h.getTimeFilter(duration)

	var groupBy string
	if duration == "7d" {
		groupBy = "%Y-%m-%d"
	} else {
		groupBy = "%Y-%m-%d %H:00"
	}

	collection := h.store.GetCollection()

	pipeline := []bson.M{
		{"$match": filter},
		{"$group": bson.M{
			"_id":   bson.M{"$dateToString": bson.M{"format": groupBy, "date": "$timestamp"}},
			"count": bson.M{"$sum": 1},
		}},
		{"$sort": bson.M{"_id": 1}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cursor.Close(ctx)

	var results []struct {
		ID    string `bson:"_id"`
		Count int    `bson:"count"`
	}

	if err := cursor.All(ctx, &results); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	data := make([]TimeSeriesData, len(results))
	for i, r := range results {
		data[i] = TimeSeriesData{
			Timestamp: r.ID,
			Count:     r.Count,
		}
	}

	c.JSON(http.StatusOK, data)
}
