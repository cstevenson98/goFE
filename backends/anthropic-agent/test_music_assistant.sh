#!/bin/bash

# Test script for Music Composition Assistant API

BASE_URL="http://localhost:8081"

echo "🎵 Testing Music Composition Assistant API"
echo "=========================================="

# Test health check
echo -e "\n1. Testing health check..."
curl -s "$BASE_URL/health" | jq '.'

# Test document creation
echo -e "\n2. Testing document creation..."
DOC_RESPONSE=$(curl -s -X POST "$BASE_URL/api/documents" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Test Music Score",
    "content": "\\documentclass{article}\\begin{document}Test music content\\end{document}"
  }')

echo "$DOC_RESPONSE" | jq '.'
DOC_ID=$(echo "$DOC_RESPONSE" | jq -r '.data.id')

# Test document retrieval
echo -e "\n3. Testing document retrieval..."
curl -s "$BASE_URL/api/documents/$DOC_ID" | jq '.'

# Test LilyPond document creation with valid LilyPond
echo -e "\n4. Testing LilyPond document creation..."
LILYPOND_RESPONSE=$(curl -s -X POST "$BASE_URL/api/lilypond" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Test LilyPond Music Score",
    "content": "\\version \"2.24.0\"\\score {\\new Staff {c d e f g a b c}\\layout {}}"
  }')

echo "$LILYPOND_RESPONSE" | jq '.'
LILYPOND_ID=$(echo "$LILYPOND_RESPONSE" | jq -r '.data.id')

# Test LilyPond source retrieval
echo -e "\n5. Testing LilyPond source retrieval..."
curl -s "$BASE_URL/api/lilypond/$LILYPOND_ID/source" | jq '.'

# Test LilyPond file path retrieval
echo -e "\n6. Testing LilyPond file path retrieval..."
curl -s "$BASE_URL/api/lilypond/$LILYPOND_ID/file-path" | jq '.'

# Test LilyPond temporary directory
echo -e "\n7. Testing LilyPond temporary directory..."
curl -s "$BASE_URL/api/lilypond/temp-dir" | jq '.'

# Test LilyPond document update
echo -e "\n8. Testing LilyPond document update..."
curl -s -X PUT "$BASE_URL/api/lilypond/$LILYPOND_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Updated LilyPond Music Score",
    "content": "\\version \"2.24.0\"\\score {\\new Staff {c d e f g a b c d}\\layout {}}"
  }' | jq '.'

# Test LilyPond compilation
echo -e "\n9. Testing LilyPond compilation..."
COMPILE_RESPONSE=$(curl -s -X POST "$BASE_URL/api/lilypond/$LILYPOND_ID/compile")
echo "$COMPILE_RESPONSE" | jq '.'

# Test PDF path retrieval
echo -e "\n10. Testing PDF path retrieval..."
curl -s "$BASE_URL/api/lilypond/$LILYPOND_ID/pdf-path" | jq '.'

# Test PDF file download
echo -e "\n11. Testing PDF file download..."
PDF_RESPONSE=$(curl -s -I "$BASE_URL/api/lilypond/$LILYPOND_ID/pdf")
echo "$PDF_RESPONSE" | head -5

# Test chat with music composition assistant
echo -e "\n12. Testing chat with music composition assistant..."
curl -s -X POST "$BASE_URL/api/chat" \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Create a simple C major scale in LilyPond notation"
  }' | jq '.'

# Test document listing
echo -e "\n13. Testing document listing..."
curl -s "$BASE_URL/api/documents" | jq '.'

# Test LilyPond document listing
echo -e "\n14. Testing LilyPond document listing..."
curl -s "$BASE_URL/api/lilypond" | jq '.'

# Test LilyPond document deletion
echo -e "\n15. Testing LilyPond document deletion..."
curl -s -X DELETE "$BASE_URL/api/lilypond/$LILYPOND_ID" | jq '.'

# Test endpoints documentation
echo -e "\n16. Testing endpoints documentation..."
curl -s "$BASE_URL/endpoints" | jq '.data.endpoints | length'

echo -e "\n✅ Music Composition Assistant API tests completed!"
echo -e "\n📁 LilyPond files are stored in the temporary directory shown above."
echo -e "\n📄 PDF files are generated in the same directory when compilation succeeds." 