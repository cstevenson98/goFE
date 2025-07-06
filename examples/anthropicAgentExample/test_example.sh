#!/bin/bash

# Test script for the Anthropic Agent Example
# This script tests the frontend example with the backend API

echo "Testing Anthropic Agent Example"
echo "================================"

# Base URL
BASE_URL="http://localhost:8081"

# Test 1: Health check
echo -e "\n1. Testing health check..."
HEALTH_RESPONSE=$(curl -s "$BASE_URL/health")
echo "Health Response:"
echo "$HEALTH_RESPONSE" | jq '.'

# Test 2: Test regular chat API
echo -e "\n2. Testing regular chat API..."
CHAT_RESPONSE=$(curl -s -X POST "$BASE_URL/api/chat" \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Hello, can you help me with music composition?"
  }')

echo "Chat Response:"
echo "$CHAT_RESPONSE" | jq '.'

# Test 2b: Test regular chat API with LilyPond content
echo -e "\n2b. Testing regular chat API with LilyPond content..."
CHAT_WITH_CONTENT_RESPONSE=$(curl -s -X POST "$BASE_URL/api/chat" \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Add a bass line to this piece",
    "content": "\\version \"2.24.0\"\\score{\\new Staff{\\clef treble c\\'4 d\\'4 e\\'4 f\\'4}\\layout{}}"
  }')

echo "Chat with Content Response:"
echo "$CHAT_WITH_CONTENT_RESPONSE" | jq '.'

# Test 3: Test music assistant API
echo -e "\n3. Testing music assistant API..."
MUSIC_RESPONSE=$(curl -s -X POST "$BASE_URL/api/music-assistant/chat" \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Create a simple C major scale in 4/4 time",
    "context": {
      "style": "classical",
      "tempo": "moderate"
    }
  }')

echo "Music Assistant Response:"
echo "$MUSIC_RESPONSE" | jq '.'

# Test 4: Test stream setup for regular chat
echo -e "\n4. Testing stream setup for regular chat..."
STREAM_SETUP_RESPONSE=$(curl -s -X POST "$BASE_URL/api/chat/stream" \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Explain LilyPond notation"
  }')

echo "Stream Setup Response:"
echo "$STREAM_SETUP_RESPONSE" | jq '.'

# Test 5: Test stream setup for music assistant
echo -e "\n5. Testing stream setup for music assistant..."
MUSIC_STREAM_SETUP_RESPONSE=$(curl -s -X POST "$BASE_URL/api/music-assistant/stream" \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Create a simple melody in G major",
    "context": {
      "style": "folk",
      "tempo": "fast"
    }
  }')

echo "Music Stream Setup Response:"
echo "$MUSIC_STREAM_SETUP_RESPONSE" | jq '.'

# Test 6: Test LilyPond document creation
echo -e "\n6. Testing LilyPond document creation..."
DOC_CREATE_RESPONSE=$(curl -s -X POST "$BASE_URL/api/lilypond" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Test Document",
    "content": "\\version \"2.24.0\"\\score{\\new Staff{c\\'4 d\\'4 e\\'4 f\\'4}\\layout{}}"
  }')

echo "Document Creation Response:"
echo "$DOC_CREATE_RESPONSE" | jq '.'

# Test 7: Test endpoints documentation
echo -e "\n7. Testing endpoints documentation..."
ENDPOINTS_RESPONSE=$(curl -s "$BASE_URL/endpoints")

echo "Endpoints Response:"
echo "$ENDPOINTS_RESPONSE" | jq '.'

echo -e "\n================================="
echo "Anthropic Agent Example Test Complete"
echo ""
echo "To run the frontend example:"
echo "1. Make sure the backend is running on localhost:8081"
echo "2. Navigate to examples/anthropicAgentExample"
echo "3. Run: go run main.go"
echo "4. Open your browser to the displayed URL"
echo ""
echo "The example now supports:"
echo "- Regular chat mode (general AI conversation)"
echo "- Music assistant mode (specialized music composition)"
echo "- Toggle between modes using the UI button"
echo "- Streaming responses for both modes"
echo "- LilyPond document management and compilation" 