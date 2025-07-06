#!/bin/bash

# Test script for the Music Assistant integration
# This script tests the connection between LilyPond and prompt engine

echo "Testing Music Assistant Integration"
echo "=================================="

# Base URL
BASE_URL="http://localhost:8081"

# Test 1: Health check
echo -e "\n1. Testing health check..."
curl -s "$BASE_URL/health" | jq '.'

# Test 2: Chat with music assistant - compose a piece
echo -e "\n2. Testing music assistant chat - composition..."
COMPOSE_RESPONSE=$(curl -s -X POST "$BASE_URL/api/music-assistant/chat" \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Create a simple C major scale in 4/4 time",
    "context": {
      "style": "classical",
      "tempo": "moderate"
    }
  }')

echo "Compose Response:"
echo "$COMPOSE_RESPONSE" | jq '.'

# Extract the new content if successful
NEW_CONTENT=$(echo "$COMPOSE_RESPONSE" | jq -r '.data.newContent // empty')
if [ -n "$NEW_CONTENT" ]; then
    echo -e "\nGenerated LilyPond content:"
    echo "$NEW_CONTENT"
    
    # Test 3: Chat with music assistant - analyze the piece
    echo -e "\n3. Testing music assistant chat - analysis..."
    ANALYSIS_RESPONSE=$(curl -s -X POST "$BASE_URL/api/music-assistant/chat" \
      -H "Content-Type: application/json" \
      -d "{
        \"message\": \"Analyze this musical piece: $NEW_CONTENT\",
        \"context\": {
          \"focus\": \"harmonic analysis\"
        }
      }")
    
    echo "Analysis Response:"
    echo "$ANALYSIS_RESPONSE" | jq '.'
    
    # Test 4: Chat with music assistant - get suggestions
    echo -e "\n4. Testing music assistant chat - suggestions..."
    SUGGEST_RESPONSE=$(curl -s -X POST "$BASE_URL/api/music-assistant/chat" \
      -H "Content-Type: application/json" \
      -d "{
        \"message\": \"Suggest improvements for this piece: $NEW_CONTENT\"
      }")
    
    echo "Suggestions Response:"
    echo "$SUGGEST_RESPONSE" | jq '.'
    
    # Test 5: Chat with music assistant - modify the piece
    echo -e "\n5. Testing music assistant chat - modification..."
    MODIFY_RESPONSE=$(curl -s -X POST "$BASE_URL/api/music-assistant/chat" \
      -H "Content-Type: application/json" \
      -d "{
        \"message\": \"Modify this piece to add dynamics and articulation: $NEW_CONTENT\",
        \"context\": {
          \"style\": \"romantic\"
        }
      }")
    
    echo "Modify Response:"
    echo "$MODIFY_RESPONSE" | jq '.'
    
    # Test 6: Validate and compile
    echo -e "\n6. Testing validation and compilation..."
    VALIDATE_RESPONSE=$(curl -s -X POST "$BASE_URL/api/music-assistant/validate-compile" \
      -H "Content-Type: application/json" \
      -d "{
        \"content\": $(echo "$NEW_CONTENT" | jq -R -s .)
      }")
    
    echo "Validate Response:"
    echo "$VALIDATE_RESPONSE" | jq '.'
else
    echo -e "\nComposition failed, skipping subsequent tests"
fi

# Test 7: Stream setup
echo -e "\n7. Testing stream setup..."
STREAM_SETUP_RESPONSE=$(curl -s -X POST "$BASE_URL/api/music-assistant/stream" \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Create a simple melody in G major",
    "context": {
      "style": "folk",
      "tempo": "fast"
    }
  }')

echo "Stream Setup Response:"
echo "$STREAM_SETUP_RESPONSE" | jq '.'

# Test 8: General chat with different request
echo -e "\n8. Testing general chat with different request..."
GENERAL_RESPONSE=$(curl -s -X POST "$BASE_URL/api/music-assistant/chat" \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Explain the difference between major and minor scales",
    "context": {
      "level": "beginner"
    }
  }')

echo "General Chat Response:"
echo "$GENERAL_RESPONSE" | jq '.'

echo -e "\n=================================="
echo "Music Assistant Integration Test Complete" 