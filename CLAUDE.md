# ai-wrap

go proxy for gemini with caching, logging, cost control

## dev workflow

```bash
docker-compose up          # api + redis + mongo (hot reload)
make dev-ui                # admin ui at localhost:3000
go test -v ./tests         # integration tests
make add-indexes           # mongodb indexes
```

## config

`config.yaml`:
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

env: `PORT`, `MONGO_URI`, `REDIS_URI`, `GEMINI_TIMEOUT` (default 120s)

## file structure

```
main.go              # entry + routes
config.yaml          # models, costs, cache
data/*.csv           # api keys
internal/
  config/            # yaml + env loading
  handler/           # proxy + admin endpoints
  client/            # gemini api calls
  cache/             # redis
  store/             # mongodb
  keymanager/        # key rotation
app/                 # nextjs admin ui
cmd/add-indexes/     # mongodb index tool
```

## key concepts

- **caching**: redis when temp <= max_temp, sha256 hash key, mongodb fallback
- **cost tracking**: per 1M tokens, headers + mongodb logs, blocks if > max_cost (402)
- **key rotation**: csv keys, retries on 429/5xx, marks inactive on 403
- **vision**: base64 inlineData, +258 tokens per image estimate
- **admin api**: `/admin/stats`, `/admin/requests`, `/admin/timeseries`

## ci/cd

github actions builds on push to main:
- `api.yml` → tests + builds `gyvastis/ai-wrap-api:latest`
- `app.yml` → builds `gyvastis/ai-wrap-app:latest`

deployment handled separately (not in this repo)

see `know-how/*.md` for detailed implementation notes
