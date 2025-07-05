#!/bin/bash

# API Test Script for goFE API Example
# This script tests all endpoints of the running API server

# Configuration
API_BASE_URL="http://localhost:8080"
API_TIMEOUT=10

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0

# Function to print colored output
print_status() {
    local status=$1
    local message=$2
    case $status in
        "PASS")
            echo -e "${GREEN}✓ PASS${NC}: $message"
            ((TESTS_PASSED++))
            ;;
        "FAIL")
            echo -e "${RED}✗ FAIL${NC}: $message"
            ((TESTS_FAILED++))
            ;;
        "INFO")
            echo -e "${BLUE}ℹ INFO${NC}: $message"
            ;;
        "WARN")
            echo -e "${YELLOW}⚠ WARN${NC}: $message"
            ;;
    esac
}

# Function to make HTTP requests and check response
test_endpoint() {
    local method=$1
    local endpoint=$2
    local expected_status=$3
    local description=$4
    local data=$5
    
    local url="$API_BASE_URL$endpoint"
    local response_file="/tmp/api_test_response_$$.json"
    
    # Make the request
    if [ -n "$data" ]; then
        response=$(curl -s -w "%{http_code}" -o "$response_file" \
            -X "$method" \
            -H "Content-Type: application/json" \
            -d "$data" \
            --max-time $API_TIMEOUT \
            "$url")
    else
        response=$(curl -s -w "%{http_code}" -o "$response_file" \
            -X "$method" \
            --max-time $API_TIMEOUT \
            "$url")
    fi
    
    # Extract status code (last 3 characters)
    status_code=${response: -3}
    
    # Check if request was successful
    if [ $? -eq 0 ] && [ "$status_code" = "$expected_status" ]; then
        print_status "PASS" "$description (Status: $status_code)"
        if [ -f "$response_file" ]; then
            echo "  Response: $(cat "$response_file" | head -c 200)..."
        fi
    else
        print_status "FAIL" "$description (Expected: $expected_status, Got: $status_code)"
        if [ -f "$response_file" ]; then
            echo "  Error: $(cat "$response_file")"
        fi
    fi
    
    # Clean up
    rm -f "$response_file"
}

# Function to extract ID from response
extract_id() {
    local response_file=$1
    local field=$2
    cat "$response_file" | grep -o "\"$field\":\"[^\"]*\"" | cut -d'"' -f4
}

# Check if API server is running
print_status "INFO" "Testing API server at $API_BASE_URL"

# Test 1: Health check (GET /api/users to check if server is up)
print_status "INFO" "=== Testing API Server Availability ==="
test_endpoint "GET" "/api/users" "200" "Server is running and responding"

# Test 2: Get all users
print_status "INFO" "=== Testing User Endpoints ==="
test_endpoint "GET" "/api/users" "200" "Get all users"

# Test 3: Create a new user
print_status "INFO" "=== Testing User Creation ==="
NEW_USER_DATA='{"email":"test@example.com","name":"Test User"}'
test_endpoint "POST" "/api/users" "201" "Create new user" "$NEW_USER_DATA"

# Test 4: Create another user for testing
ANOTHER_USER_DATA='{"email":"test2@example.com","name":"Test User 2"}'
test_endpoint "POST" "/api/users" "201" "Create another user" "$ANOTHER_USER_DATA"

# Test 5: Try to create duplicate user (should fail)
test_endpoint "POST" "/api/users" "409" "Create duplicate user (should fail)" "$NEW_USER_DATA"

# Test 6: Create user with invalid data (should fail)
INVALID_USER_DATA='{"email":"","name":""}'
test_endpoint "POST" "/api/users" "400" "Create user with invalid data (should fail)" "$INVALID_USER_DATA"

# Test 7: Get all users again (should have more users)
test_endpoint "GET" "/api/users" "200" "Get all users (after creation)"

# Test 8: Get specific user (we'll need to extract an ID first)
print_status "INFO" "=== Testing Individual User Operations ==="
# Get users and extract first user ID
response_file="/tmp/users_response_$$.json"
curl -s "$API_BASE_URL/api/users" > "$response_file"
if [ -f "$response_file" ]; then
    USER_ID=$(cat "$response_file" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    if [ -n "$USER_ID" ]; then
        test_endpoint "GET" "/api/users/$USER_ID" "200" "Get specific user by ID"
        
        # Test 9: Update user
        UPDATE_USER_DATA="{\"name\":\"Updated Test User\"}"
        test_endpoint "PUT" "/api/users/$USER_ID" "200" "Update user" "$UPDATE_USER_DATA"
        
        # Test 10: Get updated user
        test_endpoint "GET" "/api/users/$USER_ID" "200" "Get updated user"
        
        # Test 11: Try to get non-existent user
        test_endpoint "GET" "/api/users/nonexistent-id" "404" "Get non-existent user"
    else
        print_status "WARN" "Could not extract user ID for individual user tests"
    fi
    rm -f "$response_file"
fi

# Test 12: Get all messages
print_status "INFO" "=== Testing Message Endpoints ==="
test_endpoint "GET" "/api/messages" "200" "Get all messages"

# Test 13: Create a new message
print_status "INFO" "=== Testing Message Creation ==="
# Get a user ID for message creation
response_file="/tmp/users_for_message_$$.json"
curl -s "$API_BASE_URL/api/users" > "$response_file"
if [ -f "$response_file" ]; then
    USER_ID_FOR_MESSAGE=$(cat "$response_file" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    if [ -n "$USER_ID_FOR_MESSAGE" ]; then
        NEW_MESSAGE_DATA="{\"content\":\"This is a test message\",\"user_id\":\"$USER_ID_FOR_MESSAGE\"}"
        test_endpoint "POST" "/api/messages" "201" "Create new message" "$NEW_MESSAGE_DATA"
        
        # Test 14: Create another message
        ANOTHER_MESSAGE_DATA="{\"content\":\"Another test message\",\"user_id\":\"$USER_ID_FOR_MESSAGE\"}"
        test_endpoint "POST" "/api/messages" "201" "Create another message" "$ANOTHER_MESSAGE_DATA"
        
        # Test 15: Create message with invalid data (should fail)
        INVALID_MESSAGE_DATA='{"content":"","user_id":""}'
        test_endpoint "POST" "/api/messages" "400" "Create message with invalid data (should fail)" "$INVALID_MESSAGE_DATA"
        
        # Test 16: Get all messages again
        test_endpoint "GET" "/api/messages" "200" "Get all messages (after creation)"
        
        # Test 17: Get specific message
        print_status "INFO" "=== Testing Individual Message Operations ==="
        response_file="/tmp/messages_response_$$.json"
        curl -s "$API_BASE_URL/api/messages" > "$response_file"
        if [ -f "$response_file" ]; then
            MESSAGE_ID=$(cat "$response_file" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
            if [ -n "$MESSAGE_ID" ]; then
                test_endpoint "GET" "/api/messages/$MESSAGE_ID" "200" "Get specific message by ID"
                
                # Test 18: Try to get non-existent message
                test_endpoint "GET" "/api/messages/nonexistent-id" "404" "Get non-existent message"
            else
                print_status "WARN" "Could not extract message ID for individual message tests"
            fi
            rm -f "$response_file"
        fi
    else
        print_status "WARN" "Could not extract user ID for message creation tests"
    fi
    rm -f "$response_file"
fi

# Test 19: Test CORS headers
print_status "INFO" "=== Testing CORS Headers ==="
response_file="/tmp/cors_test_$$.json"
curl -s -I "$API_BASE_URL/api/users" > "$response_file"
if grep -q "Access-Control-Allow-Origin" "$response_file"; then
    print_status "PASS" "CORS headers are present"
else
    print_status "FAIL" "CORS headers are missing"
fi
rm -f "$response_file"

# Test 20: Test OPTIONS request (CORS preflight)
test_endpoint "OPTIONS" "/api/users" "200" "CORS preflight request"

# Test 21: Test invalid endpoint
test_endpoint "GET" "/api/invalid" "404" "Invalid endpoint"

# Test 22: Test invalid method
test_endpoint "PATCH" "/api/users" "405" "Invalid HTTP method"

# Summary
print_status "INFO" "=== Test Summary ==="
echo "Tests passed: $TESTS_PASSED"
echo "Tests failed: $TESTS_FAILED"
echo "Total tests: $((TESTS_PASSED + TESTS_FAILED))"

if [ $TESTS_FAILED -eq 0 ]; then
    print_status "PASS" "All tests passed! 🎉"
    exit 0
else
    print_status "FAIL" "Some tests failed. Please check the API server."
    exit 1
fi 