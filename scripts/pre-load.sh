#!/bin/bash
# Pre-load hook script for galick

echo "Running pre-load hook script..."
echo "Setting up test environment..."

mkdir -p output

# Get target host from environment variable or use default
TARGET_HOST=${TARGET_HOST:-demo-server:8080}

if curl -s http://${TARGET_HOST}/api/health > /dev/null; then
    echo "Server is running and ready for load testing at ${TARGET_HOST}."
else
    echo "Warning: Server does not seem to be running at ${TARGET_HOST}."
    echo "Make sure the demo-server container is running properly."
    exit 1
fi

echo "Pre-load hook completed successfully."
