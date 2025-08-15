# Build stage
FROM golang:1.22.2-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY deploy/go.mod ./

# Download dependencies
RUN go mod download

# Copy source code
COPY deploy/ ./

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o nonax-server .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/nonax-server .

# Create uploads directory
RUN mkdir -p uploads && \
    chown -R appuser:appgroup uploads

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8443

# Run the application
CMD ["./nonax-server"]
