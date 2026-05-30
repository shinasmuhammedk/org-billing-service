FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o billing-service ./cmd/server

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/billing-service .

RUN mkdir -p logs

EXPOSE 50052
EXPOSE 8081

CMD ["./billing-service"]