#!/bin/bash
# Pre-load hook script for galick

echo "Running pre-load hook script..."
echo "Setting up test environment..."

# Create output directory if it doesn't exist
mkdir -p output

# Verify if server is running
if curl -s http://localhost:8080/api/health > /dev/null; then
    echo "Server is running and ready for load testing."
else
    echo "Warning: Server does not seem to be running on port 8080."
    echo "Make sure to start the demo server with: ./scripts/run-demo.sh"
    exit 1
fi

echo "Pre-load hook completed successfully."
