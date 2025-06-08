#!/bin/bash
set -e # Exit immediately if a command exits with a non-zero status.

# --- Configuration ---
BASE_URL="http://localhost:8080"
VERBOSE=${VERBOSE:-false} # Set to true for detailed debugging
TEST_TIMEOUT=30 # seconds

# --- Helper Functions and Colors ---
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

print_header() {
    echo -e "\n${YELLOW}======================================================================="
    echo -e "===== $1"
    echo -e "=======================================================================${NC}"
}

print_step() {
    echo -e "\n${BLUE}▶ $1${NC}"
}

print_verbose() {
    if [[ "$VERBOSE" == "true" ]]; then
        echo -e "${BLUE}[DEBUG] $1${NC}" >&2
    fi
}

# Improved function to check HTTP status code with more flexible error handling
check_status() {
    local expected_code="$1"
    local response="$2"
    local description="$3"
    local allow_alternative="${4:-}" # Optional fourth parameter for alternative acceptable codes
    
    local status_code=$(echo "$response" | tail -n1)
    local response_body=$(echo "$response" | sed '$d') # Remove last line (status code)
    
    print_verbose "Expected: $expected_code, Got: $status_code"
    print_verbose "Response body: $response_body"
    
    # Check if status matches expected or alternative
    if [[ "$status_code" == "$expected_code" ]] || [[ -n "$allow_alternative" && "$status_code" == "$allow_alternative" ]]; then
        echo -e "${GREEN}✓ PASSED: ${description} (Expected: ${expected_code}, Got: ${status_code})${NC}"
        return 0
    else
        echo -e "${RED}✗ FAILED: ${description} (Expected: ${expected_code}, Got: ${status_code})${NC}"
        echo "Response body:"
        echo "$response_body"
        return 1
    fi
}

# Function to wait for server to be ready
wait_for_server() {
    local attempts=0
    local max_attempts=30
    
    print_step "Waiting for server to be ready..."
    while [[ $attempts -lt $max_attempts ]]; do
        if curl -s -f "$BASE_URL/health" > /dev/null 2>&1; then
            echo -e "${GREEN}✓ Server is ready${NC}"
            return 0
        fi
        ((attempts++))
        echo -n "."
        sleep 1
    done
    
    echo -e "${RED}✗ Server failed to start within ${max_attempts} seconds${NC}"
    return 1
}

# Function to safely extract JSON field using jq with error handling
extract_json() {
    local response="$1"
    local field="$2"
    local response_body=$(echo "$response" | sed '$d')
    
    if ! echo "$response_body" | jq -e ".$field" > /dev/null 2>&1; then
        echo -e "${RED}✗ Failed to extract '$field' from response${NC}" >&2
        echo "Response: $response_body" >&2
        return 1
    fi
    
    echo "$response_body" | jq -r ".$field"
}

# Function to make HTTP request with timeout and error handling
make_request() {
    local method="$1"
    local url="$2"
    local headers="$3"
    local data="$4"
    
    local curl_cmd="curl -s -w \"\\n%{http_code}\" --max-time $TEST_TIMEOUT"
    
    if [[ -n "$headers" ]]; then
        curl_cmd="$curl_cmd $headers"
    fi
    
    if [[ -n "$data" ]]; then
        curl_cmd="$curl_cmd -d '$data'"
    fi
    
    curl_cmd="$curl_cmd -X $method '$url'"
    
    print_verbose "Executing: $curl_cmd"
    eval "$curl_cmd"
}

# --- Pre-flight Check ---
print_header "PRE-FLIGHT CHECKS"

# Check required tools
for tool in curl jq; do
    if ! command -v $tool &> /dev/null; then
        echo -e "${RED}Error: '$tool' is required but not installed${NC}"
        exit 1
    fi
    echo -e "${GREEN}✓ $tool is available${NC}"
done

# Wait for server
if ! wait_for_server; then
    echo -e "${RED}Please ensure the server is running on $BASE_URL${NC}"
    exit 1
fi

# --- Test Data Generation ---
UNIQUE_ID=$(date +%s)
EDITOR_1_EMAIL="editor1-${UNIQUE_ID}@example.com"
EDITOR_1_PASSWORD="SecurePassword123!"
EDITOR_1_TOKEN=""
EDITOR_1_NL_ID=""
EDITOR_1_POST_ID=""
EDITOR_2_EMAIL="editor2-${UNIQUE_ID}@example.com"
EDITOR_2_PASSWORD="SecurePassword123!"
EDITOR_2_TOKEN=""
EDITOR_2_NL_ID=""

print_header "NEWSLETTER SERVICE E2E TEST SUITE"
echo "Test ID: $UNIQUE_ID"
echo "Base URL: $BASE_URL"

# Track test results
TESTS_PASSED=0
TESTS_FAILED=0

# Function to run test with error handling
run_test() {
    local test_name="$1"
    shift
    
    set +e  # Temporarily disable exit on error
    "$@"
    local result=$?
    set -e  # Re-enable exit on error
    
    if [[ $result -eq 0 ]]; then
        ((TESTS_PASSED++))
        return 0
    else
        ((TESTS_FAILED++))
        echo -e "${RED}✗ Test failed: $test_name${NC}" >&2
        if [[ "${CONTINUE_ON_FAILURE:-true}" != "true" ]]; then
            print_header "TEST SUITE FAILED"
            echo -e "${RED}Failed on: $test_name${NC}"
            echo "Passed: $TESTS_PASSED, Failed: $TESTS_FAILED"
            exit 1
        fi
        return 1
    fi
}

# --- 1. SETUP: CREATE TWO EDITORS ---
print_header "1. SETUP: CREATING TWO EDITORS AND THEIR RESOURCES"

# Editor 1 - Sign up
print_step "Signing up Editor 1..."
SIGNUP_1_RAW=$(make_request "POST" "${BASE_URL}/api/editor/signup" \
    "-H 'Content-Type: application/json'" \
    "{\"email\": \"${EDITOR_1_EMAIL}\", \"password\": \"${EDITOR_1_PASSWORD}\"}")

if run_test "Editor 1 Sign Up" check_status "201" "$SIGNUP_1_RAW" "Editor 1 Sign Up"; then
    echo -e "${GREEN}✓ Editor 1 signup successful${NC}"
else
    echo -e "${RED}✗ Editor 1 signup failed, continuing...${NC}"
fi

# Editor 1 - Sign in
print_step "Signing in Editor 1..."
SIGNIN_1_RAW=$(make_request "POST" "${BASE_URL}/api/editor/signin" \
    "-H 'Content-Type: application/json'" \
    "{\"email\": \"${EDITOR_1_EMAIL}\", \"password\": \"${EDITOR_1_PASSWORD}\"}")

if run_test "Editor 1 Sign In" check_status "200" "$SIGNIN_1_RAW" "Editor 1 Sign In"; then
    # Extract token
    EDITOR_1_TOKEN=$(extract_json "$SIGNIN_1_RAW" "token" 2>/dev/null || echo "")
    if [[ -n "$EDITOR_1_TOKEN" ]]; then
        echo -e "${GREEN}✓ Editor 1 signin successful, token extracted${NC}"
    else
        echo -e "${RED}✗ Failed to extract token for Editor 1${NC}"
        EDITOR_1_TOKEN=""
    fi
else
    echo -e "${RED}✗ Editor 1 signin failed, continuing...${NC}"
    EDITOR_1_TOKEN=""
fi

# Editor 1 - Create Newsletter (only if we have a token)
if [[ -n "$EDITOR_1_TOKEN" ]]; then
    print_step "Creating Newsletter for Editor 1..."
    NL_1_RAW=$(make_request "POST" "${BASE_URL}/api/newsletters" \
        "-H 'Authorization: Bearer ${EDITOR_1_TOKEN}' -H 'Content-Type: application/json'" \
        '{"name": "Editor 1 Newsletter", "description": "Belongs to Editor 1"}')
    
    if run_test "Editor 1 Newsletter Creation" check_status "201" "$NL_1_RAW" "Editor 1 Newsletter Creation"; then
        EDITOR_1_NL_ID=$(extract_json "$NL_1_RAW" "id" 2>/dev/null || echo "")
        echo -e "${GREEN}✓ Editor 1 newsletter created${NC}"
    else
        echo -e "${RED}✗ Editor 1 newsletter creation failed${NC}"
        EDITOR_1_NL_ID=""
    fi
else
    echo -e "${YELLOW}⚠ Skipping Editor 1 newsletter creation (no token)${NC}"
fi

# Editor 2 - Sign up
print_step "Signing up Editor 2..."
SIGNUP_2_RAW=$(make_request "POST" "${BASE_URL}/api/editor/signup" \
    "-H 'Content-Type: application/json'" \
    "{\"email\": \"${EDITOR_2_EMAIL}\", \"password\": \"${EDITOR_2_PASSWORD}\"}")

if run_test "Editor 2 Sign Up" check_status "201" "$SIGNUP_2_RAW" "Editor 2 Sign Up"; then
    echo -e "${GREEN}✓ Editor 2 signup successful${NC}"
else
    echo -e "${RED}✗ Editor 2 signup failed, continuing...${NC}"
fi

# Editor 2 - Sign in
print_step "Signing in Editor 2..."
SIGNIN_2_RAW=$(make_request "POST" "${BASE_URL}/api/editor/signin" \
    "-H 'Content-Type: application/json'" \
    "{\"email\": \"${EDITOR_2_EMAIL}\", \"password\": \"${EDITOR_2_PASSWORD}\"}")

if run_test "Editor 2 Sign In" check_status "200" "$SIGNIN_2_RAW" "Editor 2 Sign In"; then
    EDITOR_2_TOKEN=$(extract_json "$SIGNIN_2_RAW" "token" 2>/dev/null || echo "")
    if [[ -n "$EDITOR_2_TOKEN" ]]; then
        echo -e "${GREEN}✓ Editor 2 signin successful, token extracted${NC}"
    else
        echo -e "${RED}✗ Failed to extract token for Editor 2${NC}"
        EDITOR_2_TOKEN=""
    fi
else
    echo -e "${RED}✗ Editor 2 signin failed, continuing...${NC}"
    EDITOR_2_TOKEN=""
fi

echo -e "${GREEN}✓ Setup phase completed${NC}"

# --- 2. CONFLICT TESTS ---
print_header "2. TESTING CONFLICTS AND IDEMPOTENCY"

print_step "Attempting to sign up Editor 1 again with same email..."
SIGNUP_CONFLICT_RAW=$(make_request "POST" "${BASE_URL}/api/editor/signup" \
    "-H 'Content-Type: application/json'" \
    "{\"email\": \"${EDITOR_1_EMAIL}\", \"password\": \"${EDITOR_1_PASSWORD}\"}")

run_test "Duplicate Email Sign Up" check_status "409" "$SIGNUP_CONFLICT_RAW" "Duplicate Email Sign Up"

if [[ -n "$EDITOR_1_TOKEN" ]]; then
    print_step "Attempting to create newsletter with duplicate name..."
    NL_CONFLICT_RAW=$(make_request "POST" "${BASE_URL}/api/newsletters" \
        "-H 'Authorization: Bearer ${EDITOR_1_TOKEN}' -H 'Content-Type: application/json'" \
        '{"name": "Editor 1 Newsletter", "description": "This should fail"}')
    
    run_test "Duplicate Newsletter Name" check_status "409" "$NL_CONFLICT_RAW" "Duplicate Newsletter Name"
else
    echo -e "${YELLOW}⚠ Skipping duplicate newsletter test (no Editor 1 token)${NC}"
fi

# --- 3. AUTHORIZATION TESTS ---
print_header "3. TESTING AUTHORIZATION (CROSS-TENANT ACCESS)"

if [[ -n "$EDITOR_2_TOKEN" && -n "$EDITOR_1_NL_ID" ]]; then
    print_step "Editor 2 attempts to GET Editor 1's newsletter..."
    AUTH_GET_RAW=$(make_request "GET" "${BASE_URL}/api/newsletters/${EDITOR_1_NL_ID}" \
        "-H 'Authorization: Bearer ${EDITOR_2_TOKEN}'" "")
    
    run_test "Cross-Tenant GET" check_status "403" "$AUTH_GET_RAW" "Cross-Tenant GET"
    
    print_step "Editor 2 attempts to PATCH Editor 1's newsletter..."
    AUTH_PATCH_RAW=$(make_request "PATCH" "${BASE_URL}/api/newsletters/${EDITOR_1_NL_ID}" \
        "-H 'Authorization: Bearer ${EDITOR_2_TOKEN}' -H 'Content-Type: application/json'" \
        '{"name": "Hacked!"}')
    
    run_test "Cross-Tenant PATCH" check_status "403" "$AUTH_PATCH_RAW" "Cross-Tenant PATCH"
    
    print_step "Editor 2 attempts to DELETE Editor 1's newsletter..."
    AUTH_DELETE_RAW=$(make_request "DELETE" "${BASE_URL}/api/newsletters/${EDITOR_1_NL_ID}" \
        "-H 'Authorization: Bearer ${EDITOR_2_TOKEN}'" "")
    
    run_test "Cross-Tenant DELETE" check_status "403" "$AUTH_DELETE_RAW" "Cross-Tenant DELETE"
else
    echo -e "${YELLOW}⚠ Skipping authorization tests (missing tokens or newsletter ID)${NC}"
fi

# --- 4. STATE TRANSITION TESTS ---
print_header "4. TESTING STATE TRANSITIONS"

if [[ -n "$EDITOR_1_TOKEN" && -n "$EDITOR_1_NL_ID" ]]; then
    # Create post
    print_step "Creating a post for state testing..."
    POST_RAW=$(make_request "POST" "${BASE_URL}/api/newsletters/${EDITOR_1_NL_ID}/posts" \
        "-H 'Authorization: Bearer ${EDITOR_1_TOKEN}' -H 'Content-Type: application/json'" \
        '{"title": "State Test Post", "content": "This is test content for state transitions."}')
    
    if run_test "Post Creation" check_status "201" "$POST_RAW" "Post Creation for State Test"; then
        EDITOR_1_POST_ID=$(extract_json "$POST_RAW" "id" 2>/dev/null || echo "")
        
        if [[ -n "$EDITOR_1_POST_ID" ]]; then
            # Publish post
            print_step "Publishing the post..."
            PUBLISH_1_RAW=$(make_request "POST" "${BASE_URL}/api/posts/${EDITOR_1_POST_ID}/publish" \
                "-H 'Authorization: Bearer ${EDITOR_1_TOKEN}'" "")
            
            run_test "First Post Publish" check_status "200" "$PUBLISH_1_RAW" "First Post Publish"
        else
            echo -e "${YELLOW}⚠ Skipping post publish (failed to extract post ID)${NC}"
        fi
    else
        echo -e "${YELLOW}⚠ Skipping post publish (post creation failed)${NC}"
    fi
    
    # Subscribe user
    print_step "Subscribing a user to the newsletter..."
    SUBSCRIBE_1_RAW=$(make_request "POST" "${BASE_URL}/api/newsletters/${EDITOR_1_NL_ID}/subscribe" \
        "-H 'Content-Type: application/json'" \
        '{"email": "subscriber@test.com"}')
    
    run_test "First Subscription" check_status "201" "$SUBSCRIBE_1_RAW" "First Subscription"
    
    # Duplicate subscription
    print_step "Attempting duplicate subscription..."
    SUBSCRIBE_2_RAW=$(make_request "POST" "${BASE_URL}/api/newsletters/${EDITOR_1_NL_ID}/subscribe" \
        "-H 'Content-Type: application/json'" \
        '{"email": "subscriber@test.com"}')
    
    run_test "Duplicate Subscription" check_status "409" "$SUBSCRIBE_2_RAW" "Second (Duplicate) Subscription"
else
    echo -e "${YELLOW}⚠ Skipping state transition tests (missing Editor 1 token or newsletter ID)${NC}"
fi

# --- 5. VALIDATION TESTS ---
print_header "5. TESTING VALIDATION & EDGE CASES"

if [[ -n "$EDITOR_1_TOKEN" ]]; then
    LONG_STRING=$(printf '%*s' 200 '' | tr ' ' 'a')
    print_step "Testing newsletter name length validation..."
    VALIDATE_LONG_NAME_RAW=$(make_request "POST" "${BASE_URL}/api/newsletters" \
        "-H 'Authorization: Bearer ${EDITOR_1_TOKEN}' -H 'Content-Type: application/json'" \
        "{\"name\": \"${LONG_STRING}\"}")
    
    run_test "Long Name Validation" check_status "400" "$VALIDATE_LONG_NAME_RAW" "Long Name Validation"
else
    echo -e "${YELLOW}⚠ Skipping validation tests (no Editor 1 token)${NC}"
fi

if [[ -n "$EDITOR_1_NL_ID" ]]; then
    print_step "Testing invalid email validation..."
    VALIDATE_EMAIL_RAW=$(make_request "POST" "${BASE_URL}/api/newsletters/${EDITOR_1_NL_ID}/subscribe" \
        "-H 'Content-Type: application/json'" \
        '{"email": "not-an-email"}')
    
    run_test "Invalid Email Validation" check_status "400" "$VALIDATE_EMAIL_RAW" "Invalid Email Validation"
else
    echo -e "${YELLOW}⚠ Skipping email validation test (no newsletter ID)${NC}"
fi

if [[ -n "$EDITOR_1_TOKEN" ]]; then
    print_step "Testing pagination validation..."
    VALIDATE_PAGINATION_RAW=$(make_request "GET" "${BASE_URL}/api/newsletters?limit=invalid" \
        "-H 'Authorization: Bearer ${EDITOR_1_TOKEN}'" "")
    
    run_test "Invalid Pagination Validation" check_status "400" "$VALIDATE_PAGINATION_RAW" "Invalid Pagination Validation"
else
    echo -e "${YELLOW}⚠ Skipping pagination validation test (no Editor 1 token)${NC}"
fi

# --- 6. CLEANUP ---
print_header "6. CLEANUP"
if [[ -n "$EDITOR_1_TOKEN" && -n "$EDITOR_1_NL_ID" ]]; then
    print_step "Deleting test resources..."
    curl -s -o /dev/null -X DELETE "${BASE_URL}/api/newsletters/${EDITOR_1_NL_ID}" \
        -H "Authorization: Bearer ${EDITOR_1_TOKEN}" 2>/dev/null || true
    echo -e "${GREEN}✓ Test resources cleaned up${NC}"
else
    echo -e "${YELLOW}⚠ Skipping cleanup (missing token or newsletter ID)${NC}"
fi

# --- FINAL RESULTS ---
print_header "TEST SUITE COMPLETED"
echo -e "Results:"
echo -e "  ${GREEN}Passed: $TESTS_PASSED${NC}"
echo -e "  ${RED}Failed: $TESTS_FAILED${NC}"

if [[ $TESTS_FAILED -eq 0 ]]; then
    echo -e "\n${GREEN}🎉 ALL TESTS PASSED!${NC}"
    exit 0
else
    echo -e "\n${YELLOW}⚠️  Some tests failed. Check the output above for details.${NC}"
    exit 1
fi