# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY services/api-services/go.mod services/api-services/go.sum ./services/api-services/

# Download dependencies
WORKDIR /app/services/api-services
RUN go mod download

# Copy source code
COPY services/api-services/ ./

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/build/api-service ./cmd

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/build/api-service .

RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup && \
    touch /app/.restart-proc && \
    chown appuser:appgroup /app/.restart-proc

USER appuser

EXPOSE 8082

CMD ["./api-service"]