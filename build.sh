#!/bin/bash
# Exit immediately if a command exits with a non-zero status
set -e

echo "Building Employee Desktop Agent binaries..."

# 1. Build for Linux (Ubuntu/Debian)
echo "Building Linux amd64 binary..."
GOOS=linux GOARCH=amd64 go build -o employee-agent .
echo "Linux binary compiled successfully -> employee-agent"

# 2. Build for Windows
echo "Building Windows amd64 binary..."
GOOS=windows GOARCH=amd64 go build -ldflags "-H=windowsgui" -o employee-agent.exe .
echo "Windows binary compiled successfully -> employee-agent.exe"

echo "Build complete."
