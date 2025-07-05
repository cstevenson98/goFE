#!/bin/bash

# Start server script for goFE
# This script ensures environment variables are properly passed to docker-compose

echo "Starting goFE services..."

# Check if ANTHROPIC_API_KEY is set
if [ -z "$ANTHROPIC_API_KEY" ]; then
    echo "Warning: ANTHROPIC_API_KEY environment variable is not set."
    echo "The anthropic-agent service will not work without it."
    echo "Please set it with: export ANTHROPIC_API_KEY='your-api-key-here'"
    echo ""
fi

# Build and start services with environment variables
echo "Building and starting services..."
ANTHROPIC_API_KEY="$ANTHROPIC_API_KEY" docker compose up --build -d

echo ""
echo "Services started!"
echo "API Example: http://localhost:8080"
echo "Anthropic Agent: http://localhost:8081"
echo "Frontend: http://localhost:80"
echo ""
echo "To view logs: docker-compose logs -f"
echo "To stop services: docker-compose down" 