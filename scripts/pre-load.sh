#!/bin/bash
# Pre-load hook script for galick

echo "Running pre-load hook script..."
echo "Setting up test environment..."

mkdir -p output

# Get base URL from environment variable or extract from config
BASE_URL=${BASE_URL:-""}
if [ -z "$BASE_URL" ]; then
    # Try to extract from config
    if [ -f "/data/loadtest.yaml" ]; then
        ENV=$(grep "environment:" /data/loadtest.yaml | head -1 | awk '{print $2}')
        ENV=${ENV:-"dev"}
        BASE_URL=$(grep -A 5 "${ENV}:" /data/loadtest.yaml | grep "base_url:" | head -1 | awk '{print $2}')
    fi
    # Default fallback
    BASE_URL=${BASE_URL:-"http://demo-server:8080"}
fi

# Extract host from BASE_URL (remove http:// or https:// prefix)
TARGET_HOST=$(echo $BASE_URL | sed -e 's|^[^/]*//||' -e 's|/.*$||')
HEALTH_URL="${BASE_URL}/api/health"

echo "Checking server health at: ${HEALTH_URL}"

if curl -s "${HEALTH_URL}" > /dev/null; then
    echo "Server is running and ready for load testing at ${TARGET_HOST}."
else
    echo "Warning: Server does not seem to be running at ${TARGET_HOST}."
    echo "Make sure the server is running properly."
    exit 1
fi

echo "Pre-load hook completed successfully."
