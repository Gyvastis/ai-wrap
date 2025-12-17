# ai-wrap

minimal go proxy for gemini llm with redis caching, mongodb logging, cost blocking, and key rotation



## features

- proxies gemini api (same request/response format)
- cost tracking via headers + mongodb logs
- redis cache with mongodb fallback (temp < 0.3)
- blocks requests exceeding max cost (402)
- api key rotation from csv
- admin ui for monitoring

## quick start

```bash
docker-compose up
# api: http://localhost:8089
# admin: http://localhost:3000
```

## usage

```bash
curl -X POST http://localhost:8089/v1beta/models/gemini-2.0-flash:generateContent \
  -H "Content-Type: application/json" \
  -d '{"contents": [{"parts": [{"text": "hello"}]}]}'
```

response headers: `X-Cost-Total`, `X-Cache-Status`, `X-Key-Source`

## docker images

builds push to docker hub on main:
- [gyvastis/ai-wrap-api](https://hub.docker.com/r/gyvastis/ai-wrap-api)
- [gyvastis/ai-wrap-app](https://hub.docker.com/r/gyvastis/ai-wrap-app)

## config

`config.yaml` - models and costs (per 1M tokens usd)

env vars: `PORT`, `MONGO_URI`, `REDIS_URI`, `GEMINI_TIMEOUT`

## api keys

optional `data/keys.csv` for key rotation:

```csv
key,provider,active,working_models,checked_at
AIza...,gemini,true,models/gemini-2.5-flash|models/gemini-2.0-flash,2025-12-17T11:25:25Z
```

- only `active=true` keys are used
- keys sorted by model priority (newest models first)
- on failure, rotates through all available keys
- if csv missing/empty, uses `GEMINI_API_KEY` env var

see `CLAUDE.md` for dev workflow, `know-how/*.md` for implementation details
