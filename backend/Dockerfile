FROM golang:1.25-bookworm AS builder

# Install build dependencies
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        git \
        ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o server ./cmd/server/main.go

# Final stage - Use Google's distroless image (more secure, minimal vulnerabilities)
FROM gcr.io/distroless/static-debian12:nonroot AS final

# Copy ca-certificates from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy binary from builder
COPY --from=builder /app/server /app/server

# Copy necessary files
COPY --from=builder /app/configs /app/configs
COPY --from=builder /app/pkg/database /app/database

# Distroless runs as nonroot user by default (UID 65532)
# No need to create user or change ownership

# Set working directory
WORKDIR /app
# Expose port
EXPOSE 8080

# Note: distroless doesn't have shell, so healthcheck must be done externally
# or use a different approach in docker-compose/kubernetes

# Run the application
ENTRYPOINT ["/app/server"]
