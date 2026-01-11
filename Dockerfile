# Build stage
FROM golang:alpine AS builder

# Install specific build dependencies if needed
WORKDIR /app

# Optimize dependency caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build static binary (removing symbols for smaller size with -ldflags="-s -w")
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o ovh-dyndns ./cmd/ovh-dyndns

# Final stage
FROM alpine:latest

# Certificates required for HTTPS
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/ovh-dyndns .

# Create a non-root user to run the application
RUN adduser -D appuser
USER appuser

CMD ["./ovh-dyndns"]

