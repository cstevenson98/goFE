#!/bin/bash

# Test script for Anthropic Agent API
# Make sure the server is running on port 8081

BASE_URL="http://localhost:8081"

echo "Testing Anthropic Agent API..."
echo "================================"

# Test health endpoint
echo "1. Testing health endpoint..."
curl -s -X GET "$BASE_URL/health" | jq .
echo ""

# Test endpoints documentation
echo "2. Testing endpoints documentation..."
curl -s -X GET "$BASE_URL/endpoints" | jq .
echo ""

# Test chat endpoint (requires ANTHROPIC_API_KEY to be set)
echo "3. Testing chat endpoint..."
curl -s -X POST "$BASE_URL/api/chat" \
  -H "Content-Type: application/json" \
  -d '{"message": "Hello, how are you?"}' | jq .
echo ""

echo "Tests completed!" 