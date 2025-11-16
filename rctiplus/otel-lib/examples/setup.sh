#!/bin/bash

# Script untuk setup dan run examples

EXAMPLE_DIR=$1

if [ -z "$EXAMPLE_DIR" ]; then
    echo "Usage: ./setup.sh <example-name>"
    echo "Available examples:"
    echo "  - basic"
    echo "  - advanced"
    echo "  - custom-metrics"
    echo "  - microservice"
    exit 1
fi

if [ ! -d "$EXAMPLE_DIR" ]; then
    echo "Error: Example directory '$EXAMPLE_DIR' not found"
    exit 1
fi

cd "$EXAMPLE_DIR"

echo "Setting up $EXAMPLE_DIR example..."

# Initialize go module if not exists
if [ ! -f "go.mod" ]; then
    go mod init "example-$EXAMPLE_DIR"
fi

# Add replace directive for local library
go mod edit -replace github.com/rctiplus/otel-lib=../../..

# Tidy dependencies
go mod tidy

echo "Setup complete! Run with:"
echo "  cd $EXAMPLE_DIR"
echo "  OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317 go run main.go"

