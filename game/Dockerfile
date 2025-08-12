# Multi-stage build for Go application
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make protobuf protobuf-dev

# Install Buf
RUN go install github.com/bufbuild/buf/cmd/buf@latest

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy proto files and generate code
COPY buf.yaml buf.gen.yaml ./
COPY proto/ ./proto/
RUN buf generate

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o merchant-tails ./cmd/gdextension

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 -S merchant && \
    adduser -u 1000 -S merchant -G merchant

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/merchant-tails .

# Copy any necessary config files
COPY --from=builder /app/proto/ ./proto/

# Change ownership
RUN chown -R merchant:merchant /app

USER merchant

# Expose ports
EXPOSE 8080 50051

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
ENTRYPOINT ["./merchant-tails"]
CMD ["--port=8080"]
