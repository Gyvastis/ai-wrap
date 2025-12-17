package main

import (
	"fmt"
	"log"

	"ai-wrap/internal/cache"
	"ai-wrap/internal/client"
	"ai-wrap/internal/config"
	"ai-wrap/internal/handler"
	"ai-wrap/internal/keymanager"
	"ai-wrap/internal/store"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	log.Printf("loaded %d models from config:", len(cfg.Costs.Models))
	for _, model := range cfg.Costs.Models {
		log.Printf("  - %s", model.Name)
	}

	km, err := keymanager.New("data/gemini_capabilities.csv")
	if err != nil {
		log.Printf("warning: failed to load keys from csv: %v", err)
		log.Printf("will accept user-provided api keys only")
	}
	if km != nil && km.ActiveCount() > 0 {
		log.Printf("loaded %d active keys from csv", km.ActiveCount())
	}

	redisCache, err := cache.NewRedisCache(cfg)
	if err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}
	defer redisCache.Close()
	log.Printf("connected to redis")

	mongoStore, err := store.NewMongoStore(cfg)
	if err != nil {
		log.Fatalf("failed to connect to mongodb: %v", err)
	}
	defer mongoStore.Close()
	log.Printf("connected to mongodb")

	geminiClient := client.NewGeminiClient(cfg, km)
	proxyHandler := handler.NewProxyHandler(cfg, redisCache, mongoStore, geminiClient, km)
	adminHandler := handler.NewAdminHandler(mongoStore)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())
	r.Use(cors.Default())

	r.GET("/health", proxyHandler.Health)
	r.POST("/v1beta/models/*path", proxyHandler.Handle)

	admin := r.Group("/admin")
	{
		admin.GET("/stats", adminHandler.GetStats)
		admin.GET("/requests", adminHandler.GetRequests)
		admin.GET("/requests/:id", adminHandler.GetRequest)
		admin.GET("/timeseries", adminHandler.GetTimeSeries)
	}

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("starting server on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
