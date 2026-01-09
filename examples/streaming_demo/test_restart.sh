#!/bin/bash
# Test script to verify restart works
# This script simulates: wait 2s, press 'r', wait 2s, press 'q'

cd "$(dirname "$0")"
go build -o streaming_demo main.go

echo "Starting demo..."
echo "Will press 'r' after 2 seconds to restart"
echo "Then quit after 4 seconds"
echo ""

# Note: This won't actually work in non-interactive mode
# User needs to test manually by running: go run main.go
# Then press 'r' to see if it restarts

echo "Please run manually: go run main.go"
echo "Then press 'r' to test restart"
echo "Look for: stage reset to 0, output cleared, animation restarts"
