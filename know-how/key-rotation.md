# key rotation

## how it works

1. load keys from `data/gemini_capabilities.csv`
2. filter only `active=true` keys
3. random selection on each request
4. on failure → rotate through all available keys (one by one)
5. if all keys fail → return actual gemini api error (status + body)

## csv format

```csv
key,provider,active,text,image,video,audio,embedding,thinking,pro,total_models,checked_at
AIza...,gemini,true,true,false,false,false,false,false,false,50,2025-11-13T06:56:43Z
```

uses `github.com/gocarina/gocsv` for parsing

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
