FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o add-indexes ./cmd/add-indexes

FROM alpine:latest

RUN apk --no-cache add ca-certificates wget

WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/add-indexes .
COPY --from=builder /app/config.yaml .
COPY --from=builder /app/data ./data

EXPOSE 8089

CMD ["./main"]
