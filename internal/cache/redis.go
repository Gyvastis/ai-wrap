package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"ai-wrap/internal/config"
	"ai-wrap/internal/models"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisCache(cfg *config.Config) (*RedisCache, error) {
	opt, err := redis.ParseURL(cfg.Redis.URI)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis uri: %w", err)
	}

	client := redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return &RedisCache{
		client: client,
		ttl:    time.Duration(cfg.Redis.TTL) * time.Second,
	}, nil
}

func (c *RedisCache) Close() error {
	return c.client.Close()
}

func (c *RedisCache) Get(ctx context.Context, key string) (*models.GeminiResponse, error) {
	data, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var resp models.GeminiResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *RedisCache) Set(ctx context.Context, key string, resp *models.GeminiResponse) error {
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, data, c.ttl).Err()
}
