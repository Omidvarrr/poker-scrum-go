# Multi-stage build for Go application
FROM golang:1.23-alpine AS builder

# Install build dependencies for CGO
RUN apk add --no-cache gcc musl-dev

# Install templ CLI for template generation
RUN go install github.com/a-h/templ/cmd/templ@v0.3.943

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Generate templates
RUN templ generate

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage - minimal runtime image
FROM alpine:latest

# Install ca-certificates for HTTPS requests and sqlite
RUN apk --no-cache add ca-certificates sqlite

# Create app directory and user
RUN adduser -D -s /bin/sh appuser
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/main .

# Copy static assets and templates
COPY --from=builder /app/web ./web
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/qoutes.md ./qoutes.md

# Create data directory for uploads and database
RUN mkdir -p data/uploads && chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Run the application
CMD ["./main"]
