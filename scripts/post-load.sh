#!/bin/bash
# Post-load hook script for galick

echo "Running post-load hook script..."
echo "Cleaning up test environment..."

# Get test results summary
echo "Test Results Summary:"
if [ -f "output/dev/simple/report.json" ]; then
    echo "Success Rate: $(cat output/dev/simple/report.json | grep success | head -n 1)"
    echo "P95 Latency: $(cat output/dev/simple/report.json | grep "95th" | head -n 1)"
fi

echo "Post-load hook completed successfully."
