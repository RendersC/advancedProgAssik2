FROM golang:1.22-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o payment-service ./cmd/main.go

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/payment-service .
COPY .env .
EXPOSE 50051
ENTRYPOINT ["./payment-service"]
