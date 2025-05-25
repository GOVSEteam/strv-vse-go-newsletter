# Integration Test Suite

This directory contains comprehensive integration tests for the Go Newsletter application. These tests validate the entire API flow from HTTP requests through all layers to the database.

## 📁 Structure

```
tests/integration/
├── setup/                    # Test infrastructure
│   ├── test_server.go       # HTTP server setup
│   ├── test_database.go     # Database setup/cleanup
│   ├── test_auth.go         # Authentication helpers
│   └── test_fixtures.go     # Test data fixtures
├── api/                     # Individual API endpoint tests
│   └── health_test.go       # Health check tests
├── workflows/               # End-to-end workflow tests
│   ├── newsletter_workflow_test.go   # Complete newsletter CRUD
│   ├── subscription_workflow_test.go # Subscription flow
│   └── publishing_workflow_test.go   # Publishing workflow
└── README.md               # This file
```

## 🚀 Quick Start

### Prerequisites

1. **Database**: PostgreSQL instance (local or Docker)
2. **Environment**: Go 1.21+ 
3. **Dependencies**: All Go modules installed (`go mod download`)

### Running Tests

#### Option 1: Using Docker Compose (Recommended)

```bash
# Start test database and run all integration tests
docker-compose -f docker-compose.test.yml up test-runner

# Or run specific test
docker-compose -f docker-compose.test.yml run test-runner go test -v ./tests/integration/workflows/newsletter_workflow_test.go
```

#### Option 2: Local Database

```bash
# Set up test database URL
export TEST_DATABASE_URL="postgres://user:password@localhost:5432/newsletter_test?sslmode=disable"

# Run migrations
go run cmd/migrate/main.go up

# Run all integration tests
go test -v ./tests/integration/... -race -cover

# Run specific test suite
go test -v ./tests/integration/workflows/ -race
```

## 🧪 Test Categories

### 1. Setup Infrastructure (`setup/`)

**Purpose**: Provides reusable test utilities and infrastructure.

- **`test_server.go`**: Creates HTTP test server with real dependencies
- **`test_database.go`**: Database connection, cleanup, and transaction management
- **`test_auth.go`**: Firebase authentication mocking and JWT helpers
- **`test_fixtures.go`**: Test data creation and management

### 2. API Tests (`api/`)

**Purpose**: Tests individual API endpoints in isolation.

- **Health checks**: Basic connectivity and server status
- **Authentication**: JWT validation and security
- **Input validation**: Request parsing and validation
- **Error handling**: Proper HTTP status codes and error responses

### 3. Workflow Tests (`workflows/`)

**Purpose**: End-to-end testing of complete business workflows.

#### Newsletter Workflow (`newsletter_workflow_test.go`)
- ✅ Complete newsletter creation → post creation → publishing flow
- ✅ Newsletter CRUD operations (Create, Read, Update, Delete)
- ✅ Authentication and authorization validation
- ✅ Cross-editor access prevention

#### Subscription Workflow (`subscription_workflow_test.go`)
- ✅ Complete subscription flow: subscribe → confirm → unsubscribe
- ✅ Email validation and error handling
- ✅ Token-based confirmation and unsubscription
- ✅ Double subscription handling
- ✅ Editor viewing subscribers (authenticated)

#### Publishing Workflow (`publishing_workflow_test.go`)
- ✅ Complete publishing flow with multiple subscribers
- ✅ Publishing without subscribers (error cases)
- ✅ Post status transitions (draft → published)
- ✅ Cross-editor publishing prevention
- ✅ Double publishing handling

## 🔧 Configuration

### Environment Variables

```bash
# Database (required)
TEST_DATABASE_URL="postgres://user:pass@host:port/dbname?sslmode=disable"
DATABASE_URL="postgres://user:pass@host:port/dbname?sslmode=disable"  # Fallback

# Email service (optional - uses console service if not set)
RESEND_API_KEY="your_resend_api_key"
EMAIL_FROM="noreply@yourdomain.com"

# Firebase (optional - mocked in tests)
FIREBASE_API_KEY="your_firebase_api_key"
```

### Test Database Setup

The integration tests require a PostgreSQL database. For safety, use a dedicated test database:

```sql
-- Create test database
CREATE DATABASE newsletter_test;
CREATE USER test_user WITH PASSWORD 'test_password';
GRANT ALL PRIVILEGES ON DATABASE newsletter_test TO test_user;
```

## 🛡️ Safety Features

### Database Protection
- ✅ **Test data isolation**: Uses `@integration.test` and `@test.example.com` email patterns
- ✅ **Automatic cleanup**: Removes test data after each test
- ✅ **Database validation**: Warns if not using a test database
- ✅ **Transaction support**: Optional transaction-based isolation

### Authentication Security
- ✅ **JWT mocking**: Safe Firebase authentication mocking
- ✅ **Test tokens**: Uses predictable test tokens for validation
- ✅ **Cleanup**: Restores original auth functions after tests

## 📊 Test Coverage

### Current Coverage
- **Workflow Tests**: 6 comprehensive test functions
- **API Tests**: Health endpoint validation
- **Test Cases**: 100+ individual test scenarios
- **Error Scenarios**: Authentication, validation, cross-editor access, edge cases

### Test Scenarios Covered
- ✅ **Happy Path**: Complete workflows from start to finish
- ✅ **Error Handling**: Invalid inputs, missing data, unauthorized access
- ✅ **Security**: Cross-editor access, JWT validation, ownership checks
- ✅ **Edge Cases**: Double operations, empty data, large datasets
- ✅ **Performance**: Concurrent operations, large subscriber lists

## 🚨 Troubleshooting

### Common Issues

#### Database Connection Errors
```bash
# Check database is running
pg_isready -h localhost -p 5432

# Verify connection string
psql "postgres://user:password@localhost:5432/newsletter_test"
```

#### Migration Errors
```bash
# Run migrations manually
go run cmd/migrate/main.go up

# Check migration status
go run cmd/migrate/main.go status
```

#### Test Failures
```bash
# Run with verbose output
go test -v ./tests/integration/... -race

# Run specific test
go test -v -run TestCompleteNewsletterWorkflow ./tests/integration/workflows/

# Check test logs
go test -v ./tests/integration/... 2>&1 | tee test.log
```

### Debug Mode

Enable detailed logging by setting:
```bash
export DEBUG=true
go test -v ./tests/integration/...
```

## 🔄 CI/CD Integration

### GitHub Actions Example

```yaml
name: Integration Tests
on: [push, pull_request]

jobs:
  integration-tests:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_DB: newsletter_test
          POSTGRES_USER: test_user
          POSTGRES_PASSWORD: test_password
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Run Integration Tests
        env:
          TEST_DATABASE_URL: postgres://test_user:test_password@localhost:5432/newsletter_test?sslmode=disable
        run: |
          go run cmd/migrate/main.go up
          go test -v ./tests/integration/... -race -cover
```

## 📈 Performance Considerations

### Test Execution Time
- **Individual tests**: 1-5 seconds
- **Complete suite**: 30-60 seconds
- **With Docker**: 2-3 minutes (including setup)

### Resource Usage
- **Memory**: ~50MB per test process
- **Database**: ~10MB test data
- **Network**: Local HTTP requests only

## 🔮 Future Enhancements

### Planned Additions
- [ ] **Performance Tests**: Load testing with concurrent users
- [ ] **API Contract Tests**: OpenAPI/Swagger validation
- [ ] **Email Integration**: Real email service testing
- [ ] **Metrics Tests**: Monitoring and observability validation
- [ ] **Security Tests**: Penetration testing scenarios

### Test Infrastructure Improvements
- [ ] **Parallel Execution**: Database-per-test isolation
- [ ] **Test Data Builders**: Fluent test data creation
- [ ] **Custom Assertions**: Domain-specific test assertions
- [ ] **Test Reporting**: HTML test reports with coverage

---

## 📞 Support

For questions or issues with the integration tests:

1. **Check logs**: Run tests with `-v` flag for detailed output
2. **Verify setup**: Ensure database and environment variables are correct
3. **Review documentation**: Check this README and inline code comments
4. **Create issue**: Open a GitHub issue with test output and environment details 