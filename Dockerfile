FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o order-service ./cmd/order-service
RUN CGO_ENABLED=0 GOOS=linux go build -o producer ./cmd/producer

FROM alpine:3.14
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/order-service .
COPY --from=builder /app/producer .
COPY templates ./templates
COPY static ./static
EXPOSE 8080
CMD ["./order-service"]