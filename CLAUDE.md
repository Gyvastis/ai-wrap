# ai-wrap

minimal go proxy for gemini llm with redis caching, mongodb logging, cost blocking, and key rotation

## working with the project

### config structure

yaml config only for costs/cache, everything else from env vars:

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

env vars: `PORT`, `MONGO_URI`, `MONGO_DATABASE`, `REDIS_URI`, `REDIS_TTL`, `GEMINI_TIMEOUT` (default: 120s for vision support)

hardcoded values:
- `GEMINI_API_URL`: `https://generativelanguage.googleapis.com/v1beta`
- `MONGO_COLLECTION`: `requests`

### adding new models

1. add to `config.yaml` costs.models array
2. restart server
3. model auto-loaded and validated

### development workflow

backend + services (with hot reload):
```bash
docker-compose up
```

admin ui (nextjs dev server):
```bash
make dev-ui  # runs on http://localhost:3000
```

build:
```bash
go build -o main .
make build-ui
```

test:
```bash
go test -v ./tests
```

indexes:
```bash
make add-indexes  # local dev with docker-compose mongodb
# production: run manually in api container: docker exec <api-container> ./add-indexes
```

### deployment

production deployed via github actions to docker swarm:

- **workflows**: `.github/workflows/api.yml`, `.github/workflows/app.yml`, `.github/workflows/deploy.yml`
- **pipeline**: test → build → push → deploy (shared deployment workflow)
- **stack**: `docker-stack-compose.prod.yml` deployed as `ai-wrap-prod`
- **domains**:
  - api: `aiwrap.trunkstar.com`
  - admin: `admin.aiwrap.trunkstar.com`
- **security**: vpn ip whitelist on both routes via traefik middleware
- **resources**: conservative limits (api: 0.5cpu/256m, app: 0.25cpu/128m, redis: 0.25cpu/128m)
- **images**: `trunkstar/ai-wrap-api:latest` and `trunkstar/ai-wrap-app:latest`
- **indexes**: run manually via `docker exec <api-container> ./add-indexes`

secrets required: `DOCKER_USERNAME`, `DOCKER_PASSWORD`, `SSH_HOST_SWARM`, `SSH_USER_SWARM`, `SSH_PRIVATE_KEY_DEPLOY`, `MONGO_URI`, `GEMINI_API_KEY`

**app build configuration:**
- `NEXT_PUBLIC_API_URL` must be set at build time (not runtime)
- dockerfile has `ARG NEXT_PUBLIC_API_URL=https://aiwrap.trunkstar.com` as default
- code uses `process.env.NEXT_PUBLIC_API_URL` with localhost fallback for dev
- `HOSTNAME="0.0.0.0"` set in dockerfile for healthcheck access
- healthchecks use `wget` with GET requests (not HEAD)

### file structure

```
main.go                    # entry point, includes /admin/* routes
config.yaml                # config (models, costs, cache)
data/gemini_capabilities.csv  # api keys (active=true)
tests/                     # go integration tests with custom client
Dockerfile                 # production multi-stage build
docker-compose.prod.yml    # production compose with mongodb, redis
cmd/add-indexes/           # mongodb index creation tool
app/                       # nextjs admin ui (minimal black-white vercel-like with lucide icons)
  app/page.tsx             # main dashboard with duration toggle, stats, graph, paginated requests
  components/              # modular components (KISS/SOLID)
    Stats.tsx              # stat cards with icons
    RequestsChart.tsx      # bar chart with 24h/7d time buckets (fills missing data)
    DurationToggle.tsx     # 24h/7d selector
    RequestsList.tsx       # expandable request list
    RequestItem.tsx        # individual request row
    RequestHeader.tsx      # request preview with metadata icons
    RequestDetails.tsx     # expanded view with formatted json
    Pagination.tsx         # page navigation
  lib/utils.ts             # request preview helper (truncates long messages)
  types.ts                 # typescript interfaces
internal/
  config/    # yaml + env var config loading
  models/    # request/response structs (includes error detail)
  keymanager/  # csv key loading, rotation
  client/    # gemini api calls with retry logic
  cache/     # redis caching (sha256 hash)
  store/     # mongodb request logging, pagination
  handler/   # gin routes, validation, cost calc, async logging, admin endpoints
```

### key concepts

- **model validation**: requests only allowed for models in config
- **cost tracking**: calculated per 1M tokens, added to response headers
- **cost blocking**: predicts cost before api call, blocks if exceeds max_cost (returns 402)
- **caching**: redis cache when temp <= max_temp, key = sha256(request)
- **cache fallback**: redis miss → check mongodb for previous successful request → populate redis
- **request logging**: all requests logged to mongodb asynchronously (cache hits and misses)
- **key rotation**: random from csv, retries on transient errors (429, 500-504), marks keys inactive on 403, fails immediately on client errors (400, 404), returns actual gemini error when all exhausted
- **vision support**: handles image data via inlineData (base64 encoded), adds 258 tokens to cost prediction per image, uses 120s timeout (configurable via `GEMINI_TIMEOUT`). recommend optimizing images (resize to 2048px, jpeg 85%) before sending to reduce costs 10-15%
- **admin endpoints**:
  - `/admin/stats?duration=24h|7d` - aggregated metrics (single aggregation query)
  - `/admin/requests?page=1&per_page=20` - paginated request logs (excludes request/response bodies for speed)
  - `/admin/requests/:id` - single request with full details (lazy loaded on expand)
  - `/admin/timeseries?duration=24h|7d` - time-bucketed request counts for graphing
- **production ready**: gin release mode with recovery/logger middleware, multi-stage docker build

see `know-how/*.md` for specific implementation details
