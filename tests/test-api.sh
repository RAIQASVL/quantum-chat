#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Base URL
BASE_URL="http://localhost:8080"
TEST_USERNAME=""
TEST_PASSWORD="testpass123"

echo -e "${GREEN}Starting API Tests...${NC}"

# Generate unique username
timestamp=$(date +%s)
TEST_USERNAME="testuser_${timestamp}"

echo -e "\n${YELLOW}1. Testing Registration${NC}"
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/api/auth/register" \
    -H "Content-Type: application/json" \
    -d "{
        \"username\": \"$TEST_USERNAME\",
        \"password\": \"$TEST_PASSWORD\",
        \"public_key\": \"$(openssl rand -base64 32)\"
    }")

echo "Registration Response: $REGISTER_RESPONSE"

if [[ $REGISTER_RESPONSE == *"access_token"* ]]; then
    echo -e "${GREEN}✓ Registration successful${NC}"
    ACCESS_TOKEN=$(echo $REGISTER_RESPONSE | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
    REFRESH_TOKEN=$(echo $REGISTER_RESPONSE | grep -o '"refresh_token":"[^"]*"' | cut -d'"' -f4)
    
    echo -e "\n${YELLOW}2. Testing Login${NC}"
    LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/api/auth/login" \
        -H "Content-Type: application/json" \
        -d "{
            \"username\": \"$TEST_USERNAME\",
            \"password\": \"$TEST_PASSWORD\"
        }")
    
    echo "Login Response: $LOGIN_RESPONSE"
    
    if [[ $LOGIN_RESPONSE == *"access_token"* ]]; then
        echo -e "${GREEN}✓ Login successful${NC}"
        ACCESS_TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
        
        echo -e "\n${YELLOW}3. Testing Token Refresh${NC}"
        REFRESH_RESPONSE=$(curl -s -X POST "$BASE_URL/api/auth/refresh" \
            -H "Authorization: Bearer $REFRESH_TOKEN" \
            -H "Content-Type: application/json")
        
        echo "Refresh Response: $REFRESH_RESPONSE"
        
        if [[ $REFRESH_RESPONSE == *"access_token"* ]]; then
            echo -e "${GREEN}✓ Token refresh successful${NC}"
        else
            echo -e "${RED}✗ Token refresh failed${NC}"
        fi
    else
        echo -e "${RED}✗ Login failed${NC}"
    fi
else
    echo -e "${RED}✗ Registration failed${NC}"
fi

echo -e "\n${YELLOW}4. Testing Health Endpoints${NC}"
HEALTH_RESPONSE=$(curl -s "$BASE_URL/health")
if [[ $HEALTH_RESPONSE == "OK" ]]; then
    echo -e "${GREEN}✓ Health check passed${NC}"
else
    echo -e "${RED}✗ Health check failed${NC}"
fi

if [ ! -z "$ACCESS_TOKEN" ]; then
    echo -e "\n${YELLOW}Access Token for WebSocket testing:${NC}"
    echo "$ACCESS_TOKEN"
fi

echo -e "\n${GREEN}Tests completed!${NC}"