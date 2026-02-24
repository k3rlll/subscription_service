FROM golang:1.24-alpine AS builder
WORKDIR /app

RUN go install github.com/swaggo/swag/cmd/swag@v1.16.6

COPY go.mod go.sum ./
RUN go mod download


COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o subscription-api ./cmd/app/main.go

RUN swag init -g cmd/app/main.go --parseInternal

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/subscription-api /app/subscription-api
COPY --from=builder /app/configs/config.yaml /app/configs/config.yaml
EXPOSE 8080

CMD ["/app/subscription-api"]