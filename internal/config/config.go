package config

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server  ServerConfig
	Gemini  GeminiConfig
	MongoDB MongoDBConfig
	Redis   RedisConfig
	Cache   CacheConfig
	Costs   CostsConfig
}

type ServerConfig struct {
	Port int
}

type GeminiConfig struct {
	APIURL  string
	Timeout int
}

type MongoDBConfig struct {
	URI        string
	Database   string
	Collection string
}

type RedisConfig struct {
	URI string
	TTL int
}

type CacheConfig struct {
	MaxTemp float64 `yaml:"max_temp"`
}

type CostsConfig struct {
	MaxCost float64       `yaml:"max_cost"`
	Models  []ModelConfig `yaml:"models"`
}

type ModelConfig struct {
	Name   string  `yaml:"name"`
	Input  float64 `yaml:"input"`
	Output float64 `yaml:"output"`
}

type ModelCost struct {
	Input  float64
	Output float64
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}

func getEnvFloat(key string, defaultVal float64) float64 {
	if val := os.Getenv(key); val != "" {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
	}
	return defaultVal
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var yamlCfg struct {
		Cache CacheConfig `yaml:"cache"`
		Costs CostsConfig `yaml:"costs"`
	}

	if err := yaml.Unmarshal(data, &yamlCfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	cfg := &Config{
		Server: ServerConfig{
			Port: getEnvInt("PORT", 8089),
		},
		Gemini: GeminiConfig{
			APIURL:  "https://generativelanguage.googleapis.com/v1beta",
			Timeout: getEnvInt("GEMINI_TIMEOUT", 120),
		},
		MongoDB: MongoDBConfig{
			URI:        getEnv("MONGO_URI", "mongodb://localhost:27017"),
			Database:   getEnv("MONGO_DATABASE", "aiwrap"),
			Collection: "requests",
		},
		Redis: RedisConfig{
			URI: getEnv("REDIS_URI", "redis://localhost:6379"),
			TTL: getEnvInt("REDIS_TTL", 3600),
		},
		Cache: yamlCfg.Cache,
		Costs: yamlCfg.Costs,
	}

	return cfg, nil
}

func (c *Config) GetModelCost(model string) (ModelCost, bool) {
	for _, m := range c.Costs.Models {
		if m.Name == model {
			return ModelCost{
				Input:  m.Input,
				Output: m.Output,
			}, true
		}
	}
	return ModelCost{}, false
}
