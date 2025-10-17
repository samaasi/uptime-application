# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY services/api-gateway/go.mod services/api-gateway/go.sum ./services/api-gateway/
COPY shared/ ./shared/

# Download dependencies
WORKDIR /app/services/api-gateway
RUN go mod download

# Copy source code
COPY services/api-gateway/ ./

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/build/api-gateway ./cmd

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/build/api-gateway .
COPY --from=builder /app/shared ./shared

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

USER appuser

EXPOSE 8081

CMD ["./api-gateway"]