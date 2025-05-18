# Testing Strategy: Go Newsletter Platform

## Table of Contents
1. [Introduction](#introduction)
2. [Testing Levels](#testing-levels)
   - [Unit Testing](#unit-testing)
   - [Integration Testing](#integration-testing)
   - [API Testing](#api-testing)
   - [End-to-End Testing](#end-to-end-testing)
3. [Test Automation](#test-automation)
4. [Test Coverage](#test-coverage)
5. [Test Environments](#test-environments)
6. [Special Testing Considerations](#special-testing-considerations)
7. [Tools and Technologies](#tools-and-technologies)
8. [Test Documentation](#test-documentation)
9. [Continuous Integration](#continuous-integration)

## Introduction

This document outlines the testing strategy for the Go Newsletter Platform. The strategy ensures that all components of the system are thoroughly tested to deliver a high-quality, reliable, and robust application as required by NFR-01 (Production-Ready Quality).

## Testing Levels

### Unit Testing

**Purpose**: Verify individual components work as expected in isolation.

**Scope**:
- Repository layer: Database operations
- Service layer: Business logic
- Handler/Controller layer: Request validation and error handling
- Utility functions and helpers

**Implementation**:
- Use Go's built-in testing package (`testing`)
- Mock external dependencies using interfaces and mock implementations
- Focus on edge cases, error conditions, and input validation
- Keep tests fast and focused on single units of functionality

**Example**:
```go
func TestCreateNewsletter_ValidInput(t *testing.T) {
    // Setup mock repository
    mockRepo := &MockNewsletterRepository{...}
    
    // Create service with mock
    service := NewNewsletterService(mockRepo)
    
    // Test with valid input
    newsletter, err := service.CreateNewsletter(validInput)
    
    // Assertions
    assert.NoError(t, err)
    assert.NotNil(t, newsletter)
    assert.Equal(t, validInput.Name, newsletter.Name)
}
```

### Integration Testing

**Purpose**: Verify that components work together correctly, especially across service boundaries.

**Scope**:
- Repository + Database: Real database interactions
- Service + Repository: Business logic with real data persistence
- Firebase integration: Subscriber storage and authentication
- Email service integration: Message delivery

**Implementation**:
- Use Docker containers for dependencies (PostgreSQL, Firebase emulator)
- Create isolated test databases for each test run
- Test actual API endpoints with HTTP calls
- Verify database state after operations

**Example**:
```go
func TestNewsletterCreationIntegration(t *testing.T) {
    // Setup test database
    db := setupTestDatabase()
    defer cleanupTestDatabase(db)
    
    // Create actual repository with test DB
    repo := postgres.NewNewsletterRepository(db)
    
    // Create service with real repo
    service := newsletter.NewService(repo)
    
    // Test creating a newsletter
    result, err := service.CreateNewsletter(testInput)
    
    // Verify in database
    savedNewsletter, err := repo.GetByID(result.ID)
    assert.NoError(t, err)
    assert.Equal(t, testInput.Name, savedNewsletter.Name)
}
```

### API Testing

**Purpose**: Verify that API endpoints behave correctly according to specifications.

**Scope**:
- All REST endpoints
- Authentication and authorization
- Request validation
- Response formats and status codes
- Error handling

**Implementation**:
- HTTP-based testing using the Go HTTP test client or specialized API testing tools
- Test cases for success and failure scenarios
- Validation of response schemas
- Authentication token handling

**Example**:
```go
func TestCreateNewsletterEndpoint(t *testing.T) {
    // Setup test server
    router := setupTestRouter()
    server := httptest.NewServer(router)
    defer server.Close()
    
    // Create valid request with auth token
    req := createAuthenticatedRequest(
        "POST", 
        server.URL+"/newsletters", 
        validNewsletterPayload
    )
    
    // Send request
    client := &http.Client{}
    resp, err := client.Do(req)
    
    // Assertions
    assert.NoError(t, err)
    assert.Equal(t, http.StatusCreated, resp.StatusCode)
    
    // Verify response body
    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)
    assert.Equal(t, validNewsletterPayload["name"], result["name"])
}
```

### End-to-End Testing

**Purpose**: Validate complete user journeys and system behavior as a whole.

**Scope**:
- Editor registration and login flow
- Newsletter creation, update, deletion flow
- Subscription and unsubscription flow
- Publishing and email delivery flow

**Implementation**:
- Focus on critical user journeys from PRD
- Test application with all real dependencies (limited mocking)
- Use scripted API calls to simulate user behavior
- Verify expected outcomes in all systems (database, email delivery)

**Example**:
```go
func TestCompleteNewsletterJourney(t *testing.T) {
    // 1. Register editor
    editorID := registerTestEditor(t)
    
    // 2. Create newsletter
    newsletterID := createTestNewsletter(t, editorID)
    
    // 3. Subscribe test user
    subscribeTestUser(t, newsletterID, "test@example.com")
    
    // 4. Publish post
    publishTestPost(t, editorID, newsletterID, "Test Post")
    
    // 5. Verify email was sent (using test email server)
    emails := getTestEmails("test@example.com")
    assert.Len(t, emails, 1)
    assert.Contains(t, emails[0].Subject, "Test Post")
    
    // 6. Unsubscribe
    unsubscribeTestUser(t, newsletterID, "test@example.com")
    
    // 7. Verify user is unsubscribed
    assert.False(t, isUserSubscribed(newsletterID, "test@example.com"))
}
```

## Test Automation

**Strategy**:
- All tests should be automated and runnable with a single command
- Critical tests are run on every pull request
- Complete test suite runs nightly
- Test failures block merges to main branch

**Implementation**:
- Use GitHub Actions for continuous integration
- Script test setup and teardown
- Parallelize tests where possible for speed
- Develop helpers for common test operations

## Test Coverage

**Requirements**:
- Unit tests: 80%+ code coverage for all business logic
- Integration tests: Cover all critical database operations
- API tests: 100% coverage of all API endpoints
- E2E tests: Cover all critical user journeys

**Measurement**:
- Use Go's built-in coverage tools (`go test -cover`)
- Generate coverage reports in CI pipeline
- Review coverage on each pull request

## Test Environments

1. **Local Development**:
   - Developers run tests on their machines
   - Uses Docker for dependencies
   - Focuses on unit and some integration tests

2. **CI Environment**:
   - Runs on every pull request
   - Isolated, ephemeral environment
   - Runs all test levels

3. **Staging**:
   - Closely mirrors production
   - Used for final validation before deployment
   - Runs end-to-end tests against real services

## Special Testing Considerations

### Firebase Testing
- Use Firebase Local Emulator Suite for testing Firebase interactions
- Create isolated test projects for integration tests
- Mock Firebase in unit tests

### Email Delivery Testing
- Use a test SMTP server (like MailHog) for local and CI testing
- Verify email content, headers, and delivery
- Test unsubscribe links functionality

### Security Testing
- Test JWT authentication thoroughly
- Verify authorization on all protected endpoints
- Test for common security issues (injection, XSS if applicable)
- Validate input sanitization

## Tools and Technologies

1. **Testing Framework**:
   - Go's standard `testing` package
   - `testify` for enhanced assertions

2. **Mocking**:
   - Interface-based mocking
   - `mockery` or hand-written mocks

3. **Database Testing**:
   - Docker containers for PostgreSQL
   - Test-specific schemas/databases
   - Database migrations for test setup

4. **API Testing**:
   - Standard HTTP client or specialized tools like `httpexpect`

5. **Coverage Analysis**:
   - Go's built-in coverage tools
   - Coverage visualization in CI

## Test Documentation

**Requirements**:
- Each test file should have a clear purpose described in comments
- Test cases should have descriptive names reflecting what they're testing
- Complex test setups should be documented
- Testing approach for major components should be documented

## Continuous Integration

**Pipeline**:
1. **Build**: Compile and build the application
2. **Unit Tests**: Run all unit tests
3. **Linting**: Ensure code quality standards
4. **Integration Tests**: Run integration tests with dependencies
5. **API Tests**: Test all API endpoints
6. **E2E Tests**: Run critical user journey tests
7. **Coverage Report**: Generate and check test coverage
8. **Security Scan**: Run security scanning tools

**Automation**:
- Configure GitHub Actions workflow for automatic testing
- Enforce status checks before merging to main branch
- Generate and publish test reports 