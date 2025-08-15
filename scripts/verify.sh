#!/bin/bash

# Verification script for fx-gin-scaffold
# This script verifies that the project is set up correctly

set -e

echo "🔍 Verifying fx-gin-scaffold setup..."

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21 or later."
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
echo "✅ Go version: $GO_VERSION"

# Check if .env file exists
if [ ! -f .env ]; then
    echo "⚠️  .env file not found. Creating from .env.example..."
    cp .env.example .env
fi

# Check if data directory exists
if [ ! -d data ]; then
    echo "📁 Creating data directory..."
    mkdir -p data
fi

# Download dependencies
echo "📦 Downloading dependencies..."
go mod download
go mod tidy

# Try to build the project
echo "🔨 Building project..."
go build -o bin/app ./cmd/server

if [ $? -eq 0 ]; then
    echo "✅ Build successful!"
else
    echo "❌ Build failed!"
    exit 1
fi

# Run tests
echo "🧪 Running tests..."
go test ./internal/repo/... -v

if [ $? -eq 0 ]; then
    echo "✅ Tests passed!"
else
    echo "❌ Tests failed!"
    exit 1
fi

# Check if migrations exist
if [ -d migrations ]; then
    echo "✅ Migration files found"
else
    echo "❌ Migration files not found"
    exit 1
fi

echo ""
echo "🎉 Verification complete! Your fx-gin-scaffold is ready to use."
echo ""
echo "Next steps:"
echo "1. Update .env with your configuration"
echo "2. Run 'make migrate-up' to setup database"
echo "3. Run 'make dev' to start development server"
echo "4. Visit http://localhost:8080/swagger/index.html for API docs"