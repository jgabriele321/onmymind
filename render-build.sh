#!/usr/bin/env bash
# Exit on error
set -o errexit

# Download dependencies
go mod download

# Build with CGO enabled for SQLite support
CGO_ENABLED=1 go build -o mindbot

# Make the binary executable
chmod +x mindbot 