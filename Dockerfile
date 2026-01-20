# ---- Build Stage ----
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy dependencies and download them
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /galick cmd/galick/main.go


# ---- Runtime Stage ----
FROM alpine:latest

# Add ca-certificates for HTTPS calls
RUN apk --no-cache add ca-certificates

# Copy the built binary from the builder stage
COPY --from=builder /galick /usr/local/bin/galick

# Set the entrypoint to the galick binary
ENTRYPOINT [ "galick" ]

# Default command can be showing help
CMD [ "--help" ]