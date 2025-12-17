# key rotation

## how it works

1. load keys from `data/keys.csv`
2. filter only `active=true` keys
3. sort by model priority (best models first)
4. random selection among keys with best priority
5. on failure → rotate through available keys by priority
6. if all keys fail → return actual gemini api error (status + body)

## model priority

defined in `internal/keymanager/keymanager.go`:
```go
gemini-3-pro-preview
gemini-2.5-pro
gemini-flash-latest
gemini-flash-lite-latest
gemini-2.5-flash
gemini-2.5-flash-lite
gemini-2.0-flash
gemini-2.0-flash-lite
```

## csv format

```csv
key,provider,active,working_models,checked_at
AIza...,gemini,true,models/gemini-2.5-flash|models/gemini-2.0-flash,2025-12-17T11:25:25Z
```

- `working_models`: pipe-separated list of supported models
- uses `github.com/gocarina/gocsv` for parsing

## fallback

if csv fails or empty → uses `GEMINI_API_KEY` env var

## retry logic

implemented in `internal/client/gemini.go`:
- tries all keys sequentially
- returns actual api response when exhausted
- preserves gemini error details (code, message, status)

## location

`internal/keymanager/keymanager.go` (key management)
`internal/client/gemini.go` (retry logic)

## response header

`X-Key-Source: random` (from csv) or `env` (from env var)
