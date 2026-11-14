# ==========================================
# BUILDER STAGE
# ==========================================
FROM golang:alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# - CGO_ENABLED=0 disables CGO for a fully static binary
# - -ldflags="-w -s" strips debugging information and symbols to reduce size
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o bin/netbase-be .

# ==========================================
# RUNNER STAGE
# ==========================================
FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S appgroup \
    && adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /app/bin/netbase-be .

RUN mkdir -p log && chown -R appuser:appgroup /app

USER appuser

ENV ENV=production
ENV PORT=8080
ENV LOG_FILE=/app/log/app.log
ENV TZ=Asia/Jakarta

EXPOSE 8080

CMD ["./netbase-be"]
