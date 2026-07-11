# ==========================================
# BUILDER STAGE
# ==========================================
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# - CGO_ENABLED=0 disables CGO for a fully static binary
# - -ldflags="-w -s" strips debugging information and symbols to reduce size
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o bin/fintrack-be .

# ==========================================
# RUNNER STAGE
# ==========================================
FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/bin/fintrack-be .

RUN mkdir -p log

ENV ENV=production
ENV PORT=8080
ENV LOG_FILE=/app/log/app.log

EXPOSE 8080

CMD ["./fintrack-be"]
