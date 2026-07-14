#!/bin/bash
# ──────────────────────────────────────────────────────────────────────────────
# HRM Employee Desktop Agent — Cross-Platform Build Script
# Produces production binaries for Windows, Linux, and macOS (Intel + Apple Silicon)
# ──────────────────────────────────────────────────────────────────────────────
set -e

VERSION="2.0.0"
BUILD_DATE=$(date -u '+%Y-%m-%dT%H:%M:%SZ')
LDFLAGS="-s -w -X main.agentVersion=${VERSION}"

echo "╔══════════════════════════════════════════════════════════════╗"
echo "║  Building HRM Employee Desktop Agent v${VERSION}               ║"
echo "║  Build Date: ${BUILD_DATE}                         ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""

# ── Step 1: Generate Windows version info resource ──
echo "┌─ Step 1: Generating Windows version info resource..."
if command -v go &> /dev/null; then
    go run github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest \
        -o resource_windows.syso 2>/dev/null || {
        echo "│  ⚠ goversioninfo not available, skipping version info embedding."
        echo "│  Windows binary will work but may trigger antivirus heuristics."
    }
    echo "└─ Done."
else
    echo "└─ ⚠ Go not found, skipping version info generation."
fi
echo ""

# ── Step 2: Build Linux amd64 ──
echo "┌─ Step 2: Building Linux amd64..."
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build \
    -trimpath \
    -ldflags "${LDFLAGS}" \
    -o employee-agent .
echo "└─ ✓ employee-agent ($(du -h employee-agent | cut -f1))"
echo ""

# ── Step 3: Build Windows amd64 ──
echo "┌─ Step 3: Building Windows amd64..."
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build \
    -trimpath \
    -ldflags "${LDFLAGS} -H=windowsgui" \
    -o employee-agent.exe .
echo "└─ ✓ employee-agent.exe ($(du -h employee-agent.exe | cut -f1))"
echo ""

# ── Step 4: Build macOS Intel (amd64) ──
echo "┌─ Step 4: Building macOS Intel (amd64)..."
GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build \
    -trimpath \
    -ldflags "${LDFLAGS}" \
    -o employee-agent-darwin-amd64 .
echo "└─ ✓ employee-agent-darwin-amd64 ($(du -h employee-agent-darwin-amd64 | cut -f1))"
echo ""

# ── Step 5: Build macOS Apple Silicon (arm64) ──
echo "┌─ Step 5: Building macOS Apple Silicon (arm64)..."
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build \
    -trimpath \
    -ldflags "${LDFLAGS}" \
    -o employee-agent-darwin-arm64 .
echo "└─ ✓ employee-agent-darwin-arm64 ($(du -h employee-agent-darwin-arm64 | cut -f1))"
echo ""

# ── Summary ──
echo "╔══════════════════════════════════════════════════════════════╗"
echo "║  Build Complete! Artifacts:                                  ║"
echo "║                                                              ║"
echo "║  • employee-agent              (Linux amd64)                 ║"
echo "║  • employee-agent.exe          (Windows amd64)               ║"
echo "║  • employee-agent-darwin-amd64 (macOS Intel)                 ║"
echo "║  • employee-agent-darwin-arm64 (macOS Apple Silicon)         ║"
echo "╚══════════════════════════════════════════════════════════════╝"
