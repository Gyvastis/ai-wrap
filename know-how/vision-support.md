# vision support

## request format

```json
{
  "contents": [{
    "parts": [
      {"text": "what is in this image?"},
      {
        "inlineData": {
          "mimeType": "image/png",
          "data": "base64_encoded_image..."
        }
      }
    ]
  }],
  "generationConfig": {
    "temperature": 0.1,
    "maxOutputTokens": 8192
  }
}
```

## implementation details

### models (internal/models/gemini.go)

```go
type Part struct {
    Text       string      `json:"text,omitempty"`
    InlineData *InlineData `json:"inlineData,omitempty"`
}

type InlineData struct {
    MimeType string `json:"mimeType"`
    Data     string `json:"data"`
}
```

### image optimization (recommended)

optimize images before sending to reduce costs and payload size:

**gemini limits:**
- max request size: 20mb for inline data
- images auto-scaled to max 3072x3072
- supported: png, jpeg, webp, heic, heif

**optimization strategy:**
- resize to 2048px max dimension (maintains aspect ratio)
- convert to jpeg quality 85%
- reduces payload by 60-80%
- reduces cost (smaller base64 = faster processing)

**helper in tests/image_helper.go:**
```go
optimizer := NewImageOptimizer()
base64Image, err := optimizer.OptimizeAndEncode("image.png")
```

### cost prediction (internal/handler/proxy.go)

- text tokens: char count / 4
- image tokens: +258 per image (fixed estimate)
- prediction doesn't analyze base64 data size

### timeout settings

default timeout increased to 120s:
- text-only: ~2-5s
- vision: ~10-30s
- configurable via `GEMINI_TIMEOUT` env var

### caching

vision requests cached same as text:
- temperature <= max_temp (0.3)
- cache key = sha256(entire request including image data)
- redis → mongodb fallback → api
- same image + same prompt = cache hit (useful for repeated extractions)
- different prompt on same image = cache miss (new hash)

## supported formats

gemini supports: png, jpeg, webp, heic, heif

## structured data extraction

best use case is extracting exact data from images as markdown tables or csv:

```bash
# base64 encode image
base64 -i pricing_table.png -o encoded.txt

# extract as markdown table
curl -X POST http://localhost:8089/v1beta/models/gemini-2.0-flash:generateContent \
  -H "Content-Type: application/json" \
  -d '{
    "contents": [{
      "parts": [
        {"text": "extract all data from this pricing table. return as markdown table with exact values. include any header/footer text found."},
        {"inlineData": {"mimeType": "image/png", "data": "..."}}
      ]
    }],
    "generationConfig": {
      "temperature": 0.1,
      "maxOutputTokens": 8192
    }
  }'

# extract as csv
curl -X POST http://localhost:8089/v1beta/models/gemini-2.0-flash:generateContent \
  -H "Content-Type: application/json" \
  -d '{
    "contents": [{
      "parts": [
        {"text": "extract all data from this table as CSV. include column headers."},
        {"inlineData": {"mimeType": "image/png", "data": "..."}}
      ]
    }],
    "generationConfig": {"temperature": 0.1}
  }'
```

## testing

integration test in `tests/integration_test.go`:
```bash
go test -v -run TestVisionRequest ./tests
```

test uses:
- image optimization (2048px, jpeg 85%)
- cupaloy snapshots to verify consistent extraction
- validates markdown table structure
- cost with optimization: ~$0.0009 vs ~$0.001 unoptimized
