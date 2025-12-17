# ai-wrap

minimal go proxy for gemini llm with redis caching, mongodb logging, cost blocking, and key rotation

## what it does

- proxies gemini api requests (same request/response format)
- tracks costs per request (headers + mongodb logs)
- redis cache with mongodb fallback (< 0.3 temp)
- blocks requests exceeding max cost (402 Payment Required)
- rotates api keys from csv (14 active keys)
- validates models against config
- production ready (gin release mode, recovery, logging)

## quick start

```bash
# development
docker-compose up

# production
docker-compose -f docker-compose.prod.yml up --build
```

api runs on port 8089 by default

## usage

**health check:**
```bash
curl http://localhost:8089/health
```

**generate content:**
```bash
curl -X POST http://localhost:8089/v1beta/models/gemini-2.0-flash:generateContent \
  -H "Content-Type: application/json" \
  -d '{
    "contents": [{"parts": [{"text": "what is 2+2?"}]}],
    "generationConfig": {"temperature": 0.1}
  }'
```

**response headers:**
- `X-Cost-Input`, `X-Cost-Output`, `X-Cost-Total` (usd)
- `X-Cache-Status` (HIT/MISS)
- `X-Key-Source` (random/env)

## config

`config.yaml` - model costs and cache settings:
```yaml
cache:
  max_temp: 0.3

costs:
  max_cost: 0.01
  models:
    - name: gemini-2.0-flash
      input: 0.10
      output: 0.40
```

env vars (see config.go for full list):
- `PORT` (default: 8089)
- `MONGO_URI` (default: mongodb://localhost:27017)
- `REDIS_URI` (default: redis://localhost:6379)
- `GEMINI_API_URL` (default: https://generativelanguage.googleapis.com/v1beta)

**available models:**
- gemini-2.5-pro ($1.25 / $10.0)
- gemini-2.5-flash ($0.075 / $0.30)
- gemini-2.5-flash-lite ($0.10 / $0.40)
- gemini-2.0-flash ($0.10 / $0.40)
- gemini-2.0-flash-lite ($0.075 / $0.30)
- gemini-flash-latest ($0.075 / $0.30)
- gemini-flash-lite-latest ($0.015 / $0.06)
- gemini-1.5-flash ($0.15 / $0.60)
- gemini-1.5-pro ($1.25 / $5.00)

prices are input/output per 1M tokens in usd. only models in config are allowed

## architecture

- **redis** - primary cache (TTL: 3600s)
- **mongodb** - request logging + cache fallback
- **gin** - http server (release mode, recovery, logger middleware)
- **key rotation** - sequential retry on all keys before failure

## api keys

uses keys from `data/gemini_capabilities.csv` (14 active keys loaded)
retries all keys on failure before returning actual gemini error
falls back to `GEMINI_API_KEY` env var if csv fails

## mongodb indexes

run once to optimize cache lookups:
```bash
go run ./cmd/add-indexes
```

creates indexes on:
- request_hash + success (cache lookup)
- timestamp (analytics)
- model (analytics)
- success (analytics)

## testing

```bash
go test -v ./tests
```

tests cover:
- health check
- invalid model validation
- valid request with cost tracking
- redis caching
- high temperature (no cache)
- cost blocking (402)

## deployment

production setup with docker-compose:
- multi-stage build (golang:alpine → alpine:latest)
- mongodb + redis services
- restart policies
- volume persistence

see `docker-compose.prod.yml` for full config
