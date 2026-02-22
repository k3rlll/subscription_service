FROM golang:1.24-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download


COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o subscription-api ./cmd/app/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/subscription-api /app/subscription-api
COPY --from=builder /app/configs/config.yaml /app/configs/config.yaml
EXPOSE 8080

CMD ["/app/subscription-api"]