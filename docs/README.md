# Newsletter API Documentation

This directory contains the Swagger/OpenAPI documentation for the Newsletter API.

## Accessing the Documentation

Once the server is running, you can access the interactive Swagger documentation at:

- **Swagger UI**: `http://localhost:8080/swagger/index.html`
- **Alternative URL**: `http://localhost:8080/docs/index.html`

For production deployment:
- **Production Swagger UI**: `https://strv-vse-go-newsletter-production.up.railway.app/swagger/index.html`

## API Overview

The Newsletter API provides comprehensive functionality for:

### 🔐 Authentication
- Editor registration and sign-in
- Firebase JWT-based authentication
- Password reset functionality

### 📰 Newsletter Management
- Create, read, update, and delete newsletters
- Paginated listing of newsletters
- Editor ownership validation

### 📝 Post Management
- Create and manage newsletter posts
- Publish posts to subscribers
- Full CRUD operations with ownership checks

### 👥 Subscriber Management
- Newsletter subscription functionality
- One-click unsubscribe with tokens
- Subscriber listing for newsletter owners

## Files

- `openapi.yaml` - OpenAPI 3.0 specification (manually maintained)
- `README.md` - Documentation guide

## Authentication

Most endpoints require Firebase JWT authentication. Include the token in the Authorization header:

```
Authorization: Bearer <your-firebase-jwt-token>
```

## Regenerating Documentation

To regenerate the Swagger documentation after code changes:

```bash
# Install swag if not already installed
go install github.com/swaggo/swag/cmd/swag@latest

# Generate documentation
swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal
```

## API Testing

You can test the API endpoints directly from the Swagger UI interface, or use tools like:
- Postman (import the OpenAPI spec)
- curl commands
- Any HTTP client

## Support

For API support, contact: support@newsletter-api.com 