# caching

## how it works

request → sha256 hash → redis → mongodb fallback → api

cache only enabled when `temperature <= max_temp` (default 0.3)

## cache layers

1. **redis** - primary cache, fast lookup
2. **mongodb** - fallback cache from logged requests, populates redis on hit
3. **api** - cache miss, call gemini api

## implementation

`internal/cache/redis.go` - redis client
`internal/store/mongodb.go` - FindCached() for fallback
`internal/handler/proxy.go` - cache lookup logic

## config

```yaml
cache:
  max_temp: 0.3
```

env vars:
- `REDIS_URI` (default: redis://localhost:6379)
- `REDIS_TTL` (default: 3600 seconds)
