# syntax=docker/dockerfile:1

# ===== Stage 1: Build =====
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# СНАЧАЛА копируем весь исходный код проекта
COPY . .

# ЗАСТАВЛЯЕМ Go самостоятельно найти все нужные библиотеки в коде, 
# скачать их свежие версии и обновить go.mod/go.sum
RUN go mod tidy

# Build binary with optimization flags
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o /server \
    ./cmd/server

# ===== Stage 2: Runtime =====
FROM alpine:3.19

WORKDIR /app

# Install CA certificates for HTTPS
RUN apk add --no-cache ca-certificates tzdata

# Create non-privileged user
RUN adduser -D -g '' appuser

# Copy binary from stage 1
COPY --from=builder /server /app/server

# Copy Firebase credentials (if present)
COPY --from=builder /app/firebase-service-account.json /app/firebase-service-account.json 2>/dev/null || true

# Change ownership
RUN chown -R appuser:appuser /app

# Switch to non-privileged user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Start application
CMD ["/app/server"]
