#!/bin/bash
# Script to run the demo server for galick testing

# Function to clean up on exit
cleanup() {
  echo "Cleaning up..."
  if [ ! -z "$SERVER_PID" ]; then
    echo "Stopping demo server (PID: $SERVER_PID)"
    kill $SERVER_PID
  fi
  exit 0
}

# Set up trap for cleanup
trap cleanup EXIT INT TERM

# Make sure the demo server is compiled
echo "Compiling demo server..."
go build -o bin/demo-server scripts/demo-server.go

# Start the demo server in the background
echo "Starting demo server on port 8080..."
./bin/demo-server &
SERVER_PID=$!

# Wait for server to start
echo "Waiting for server to start..."
sleep 2

# Run VHS to create the demo
echo "Running VHS to create demo GIF..."
vhs demo.tape

echo "Demo complete! The demo GIF has been created as galick-demo.gif"
echo "Press Ctrl+C to stop the server and clean up."

# Wait for the server process to finish (or until the script is interrupted)
wait $SERVER_PID
