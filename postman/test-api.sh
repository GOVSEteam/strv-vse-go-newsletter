#!/bin/bash

# Newsletter API Test Script
# This script demonstrates the basic API functionality using curl commands

BASE_URL="https://strv-vse-go-newsletter-production.up.railway.app"
TEST_EMAIL="demo-$(date +%s)@example.com"  # Unique email with timestamp
TEST_PASSWORD="SecurePassword123!"

echo "🚀 Newsletter API Test Script"
echo "=============================="
echo "Base URL: $BASE_URL"
echo "Test Email: $TEST_EMAIL"
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_step() {
    echo -e "${BLUE}$1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

# Step 1: Health Check
print_step "1. Testing Health Check..."
HEALTH_RESPONSE=$(curl -s -w "%{http_code}" -o /tmp/health_response.json "$BASE_URL/healthz")
if [ "$HEALTH_RESPONSE" = "200" ]; then
    print_success "Health check passed"
else
    print_error "Health check failed (HTTP $HEALTH_RESPONSE)"
    exit 1
fi

# Step 2: Editor Sign Up
print_step "2. Creating editor account..."
SIGNUP_RESPONSE=$(curl -s -w "%{http_code}" -o /tmp/signup_response.json \
    -X POST "$BASE_URL/editor/signup" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\"}")

if [ "$SIGNUP_RESPONSE" = "201" ]; then
    EDITOR_ID=$(cat /tmp/signup_response.json | grep -o '"editor_id":"[^"]*"' | cut -d'"' -f4)
    print_success "Editor account created (ID: $EDITOR_ID)"
else
    print_error "Editor signup failed (HTTP $SIGNUP_RESPONSE)"
    cat /tmp/signup_response.json
    exit 1
fi

# Step 3: Editor Sign In
print_step "3. Signing in editor..."
SIGNIN_RESPONSE=$(curl -s -w "%{http_code}" -o /tmp/signin_response.json \
    -X POST "$BASE_URL/editor/signin" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\"}")

if [ "$SIGNIN_RESPONSE" = "200" ]; then
    AUTH_TOKEN=$(cat /tmp/signin_response.json | grep -o '"idToken":"[^"]*"' | cut -d'"' -f4)
    print_success "Editor signed in successfully"
else
    print_error "Editor signin failed (HTTP $SIGNIN_RESPONSE)"
    cat /tmp/signin_response.json
    exit 1
fi

# Step 4: Create Newsletter
print_step "4. Creating newsletter..."
NEWSLETTER_RESPONSE=$(curl -s -w "%{http_code}" -o /tmp/newsletter_response.json \
    -X POST "$BASE_URL/api/newsletters" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $AUTH_TOKEN" \
    -d '{"name":"Test Newsletter","description":"A test newsletter created by the API test script"}')

if [ "$NEWSLETTER_RESPONSE" = "201" ]; then
    NEWSLETTER_ID=$(cat /tmp/newsletter_response.json | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
    print_success "Newsletter created (ID: $NEWSLETTER_ID)"
else
    print_error "Newsletter creation failed (HTTP $NEWSLETTER_RESPONSE)"
    cat /tmp/newsletter_response.json
    exit 1
fi

# Step 5: List Newsletters
print_step "5. Listing newsletters..."
LIST_RESPONSE=$(curl -s -w "%{http_code}" -o /tmp/list_response.json \
    -X GET "$BASE_URL/api/newsletters?limit=10&offset=0" \
    -H "Authorization: Bearer $AUTH_TOKEN")

if [ "$LIST_RESPONSE" = "200" ]; then
    NEWSLETTER_COUNT=$(cat /tmp/list_response.json | grep -o '"total":[0-9]*' | cut -d':' -f2)
    print_success "Listed newsletters (Total: $NEWSLETTER_COUNT)"
else
    print_error "Newsletter listing failed (HTTP $LIST_RESPONSE)"
    cat /tmp/list_response.json
fi

# Step 6: Create Post
print_step "6. Creating post..."
POST_RESPONSE=$(curl -s -w "%{http_code}" -o /tmp/post_response.json \
    -X POST "$BASE_URL/api/newsletters/$NEWSLETTER_ID/posts" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $AUTH_TOKEN" \
    -d '{"title":"Test Post","content":"This is a test post created by the API test script. It demonstrates the post creation functionality."}')

if [ "$POST_RESPONSE" = "201" ]; then
    POST_ID=$(cat /tmp/post_response.json | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
    print_success "Post created (ID: $POST_ID)"
else
    print_error "Post creation failed (HTTP $POST_RESPONSE)"
    cat /tmp/post_response.json
fi

# Step 7: Subscribe to Newsletter
print_step "7. Subscribing to newsletter..."
SUBSCRIBE_RESPONSE=$(curl -s -w "%{http_code}" -o /tmp/subscribe_response.json \
    -X POST "$BASE_URL/api/newsletters/$NEWSLETTER_ID/subscribe" \
    -H "Content-Type: application/json" \
    -d '{"email":"test-subscriber@example.com"}')

if [ "$SUBSCRIBE_RESPONSE" = "201" ]; then
    print_success "Subscription created"
else
    print_warning "Subscription failed (HTTP $SUBSCRIBE_RESPONSE) - This might be expected if email service is not configured"
    cat /tmp/subscribe_response.json
fi

# Step 8: Get Subscribers
print_step "8. Getting newsletter subscribers..."
SUBSCRIBERS_RESPONSE=$(curl -s -w "%{http_code}" -o /tmp/subscribers_response.json \
    -X GET "$BASE_URL/api/newsletters/$NEWSLETTER_ID/subscribers" \
    -H "Authorization: Bearer $AUTH_TOKEN")

if [ "$SUBSCRIBERS_RESPONSE" = "200" ]; then
    print_success "Retrieved subscribers list"
else
    print_warning "Getting subscribers failed (HTTP $SUBSCRIBERS_RESPONSE)"
    cat /tmp/subscribers_response.json
fi

# Step 9: Update Post
if [ ! -z "$POST_ID" ]; then
    print_step "9. Updating post..."
    UPDATE_RESPONSE=$(curl -s -w "%{http_code}" -o /tmp/update_response.json \
        -X PUT "$BASE_URL/api/posts/$POST_ID" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $AUTH_TOKEN" \
        -d '{"title":"Updated Test Post","content":"This post has been updated by the API test script."}')

    if [ "$UPDATE_RESPONSE" = "200" ]; then
        print_success "Post updated"
    else
        print_error "Post update failed (HTTP $UPDATE_RESPONSE)"
        cat /tmp/update_response.json
    fi
fi

# Step 10: Publish Post
if [ ! -z "$POST_ID" ]; then
    print_step "10. Publishing post..."
    PUBLISH_RESPONSE=$(curl -s -w "%{http_code}" -o /tmp/publish_response.json \
        -X POST "$BASE_URL/api/posts/$POST_ID/publish" \
        -H "Authorization: Bearer $AUTH_TOKEN")

    if [ "$PUBLISH_RESPONSE" = "200" ]; then
        print_success "Post published"
    else
        print_warning "Post publishing failed (HTTP $PUBLISH_RESPONSE) - This might be expected if email service is not configured"
        cat /tmp/publish_response.json
    fi
fi

# Cleanup temporary files
rm -f /tmp/*_response.json

echo ""
print_step "🎉 API Test Complete!"
echo "=============================="
echo "Summary:"
echo "- Editor Email: $TEST_EMAIL"
echo "- Editor ID: $EDITOR_ID"
echo "- Newsletter ID: $NEWSLETTER_ID"
if [ ! -z "$POST_ID" ]; then
    echo "- Post ID: $POST_ID"
fi
echo ""
echo "You can now use these IDs in Postman or other API testing tools."
echo "Note: Some operations (email sending) might fail if external services are not configured." 