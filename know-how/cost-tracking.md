# cost tracking

## calculation

```go
inputCost = promptTokenCount * modelPrice.input / 1_000_000
outputCost = candidatesTokenCount * modelPrice.output / 1_000_000
totalCost = inputCost + outputCost
```

prices are per 1M tokens in USD

## response headers

- `X-Cost-Input: 0.000001`
- `X-Cost-Output: 0.000003`
- `X-Cost-Total: 0.000004`
- `X-Cache-Status: HIT|MISS`
- `X-Key-Source: random|env`

## cost blocking

predicts cost before api call using:
- input: character count / 4 (rough token estimate)
- output: maxOutputTokens (default 8192)

if predicted cost exceeds `max_cost`, returns 402 Payment Required:

```json
{
  "error": "predicted cost $1.015625 exceeds maximum allowed cost $0.010000",
  "predicted_cost": 1.015625,
  "max_cost": 0.01
}
```

## config

```yaml
costs:
  max_cost: 0.01
  models:
    - name: gemini-2.0-flash
      input: 0.10
      output: 0.40
```

## model validation

only models defined in config are allowed
requests for undefined models → 400 error

## location

`internal/handler/proxy.go` - predictCost(), calculateCost()
