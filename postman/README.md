# Newsletter API - Postman Collection

This directory contains a comprehensive Postman collection that showcases all the features and endpoints of the Go Newsletter API deployed on Railway.

## 🚀 Quick Start

### 1. Import the Collection

1. Open Postman
2. Click "Import" in the top left
3. Select the `Newsletter_API_Collection.json` file
4. The collection will be imported with all requests and environment variables

### 2. Collection Structure

The collection is organized into logical groups:

- **🔐 Authentication** - User registration, login, and password reset
- **📰 Newsletter Management** - CRUD operations for newsletters
- **📝 Post Management** - Creating, updating, and publishing posts
- **👥 Subscription Management** - Newsletter subscriptions and unsubscriptions

## 📋 Usage Guide

### Step 1: Health Check
Start by running the **Health Check** request to verify the API is running:
```
GET {{baseUrl}}/healthz
```

### Step 2: Create an Account
Run **Editor Sign Up** to create a new account:
```json
{
    "email": "demo@example.com",
    "password": "SecurePassword123!"
}
```
*Note: The collection automatically saves the `editor_id` from the response.*

### Step 3: Sign In
Run **Editor Sign In** with the same credentials:
```json
{
    "email": "demo@example.com", 
    "password": "SecurePassword123!"
}
```
*Note: The collection automatically saves the `authToken` for subsequent authenticated requests.*

### Step 4: Create a Newsletter
Run **Create Newsletter** to create your first newsletter:
```json
{
    "name": "Tech Weekly Digest",
    "description": "A weekly newsletter covering the latest in technology, programming, and software development."
}
```
*Note: The collection automatically saves the `newsletterId` from the response.*

### Step 5: Create and Manage Posts
1. **Create Post** - Add content to your newsletter
2. **List Posts for Newsletter** - View all posts in the newsletter
3. **Update Post** - Modify existing content
4. **Publish Post** - Send the post to all subscribers

### Step 6: Manage Subscriptions
1. **Subscribe to Newsletter** - Add subscribers (no authentication required)
2. **Get Newsletter Subscribers** - View all subscribers (requires authentication)
3. **Unsubscribe by Token** - Remove subscriptions using email tokens

## 🔧 Environment Variables

The collection uses the following variables that are automatically managed:

| Variable | Description | Auto-populated |
|----------|-------------|----------------|
| `baseUrl` | API base URL | ✅ Pre-configured |
| `authToken` | JWT authentication token | ✅ From sign-in |
| `editorId` | Editor's unique ID | ✅ From sign-up |
| `newsletterId` | Newsletter's unique ID | ✅ From newsletter creation |
| `postId` | Post's unique ID | ✅ From post creation |
| `unsubscribeToken` | Token for unsubscribing | ❌ Manual (from email) |

## 🔐 Authentication

Most endpoints require authentication using Firebase JWT tokens. The collection handles this automatically:

1. Sign in using the **Editor Sign In** request
2. The `authToken` is automatically saved and used for subsequent requests
3. The collection is configured to use Bearer token authentication

## 📊 API Endpoints Overview

### Authentication Endpoints
- `GET /healthz` - Health check
- `POST /editor/signup` - Create new editor account
- `POST /editor/signin` - Sign in editor
- `POST /editor/password-reset-request` - Request password reset

### Newsletter Management (Authenticated)
- `GET /api/newsletters` - List editor's newsletters (with pagination)
- `POST /api/newsletters` - Create new newsletter
- `PATCH /api/newsletters/{id}` - Update newsletter
- `DELETE /api/newsletters/{id}` - Delete newsletter

### Post Management (Authenticated)
- `POST /api/newsletters/{newsletterID}/posts` - Create post
- `GET /api/newsletters/{newsletterID}/posts` - List posts (with pagination)
- `GET /api/posts/{postID}` - Get specific post
- `PUT /api/posts/{postID}` - Update post
- `DELETE /api/posts/{postID}` - Delete post
- `POST /api/posts/{postID}/publish` - Publish post to subscribers

### Subscription Management
- `POST /api/newsletters/{newsletterID}/subscribe` - Subscribe to newsletter (public)
- `GET /api/newsletters/{newsletterID}/subscribers` - List subscribers (authenticated)
- `GET /api/subscriptions/unsubscribe?token={token}` - Unsubscribe via token (public)

## 🎯 Testing Scenarios

### Complete Workflow Test
1. **Health Check** → Verify API is running
2. **Editor Sign Up** → Create account
3. **Editor Sign In** → Get authentication token
4. **Create Newsletter** → Set up newsletter
5. **Create Post** → Add content
6. **Subscribe to Newsletter** → Add a subscriber
7. **Publish Post** → Send to subscribers
8. **Get Newsletter Subscribers** → Verify subscription
9. **Update Post** → Modify content
10. **Delete Post** → Clean up

### Error Testing
- Try accessing protected endpoints without authentication
- Attempt to create newsletters with duplicate names
- Test with invalid UUIDs
- Test pagination with various limits and offsets

## 🔍 Response Examples

### Successful Newsletter Creation
```json
{
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "editor_id": "987fcdeb-51a2-43d1-9f12-123456789abc",
    "name": "Tech Weekly Digest",
    "description": "A weekly newsletter covering the latest in technology...",
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
}
```

### Paginated Posts Response
```json
{
    "data": [
        {
            "id": "456e7890-e89b-12d3-a456-426614174111",
            "newsletter_id": "123e4567-e89b-12d3-a456-426614174000",
            "title": "Introduction to Go Programming",
            "content": "Go, also known as Golang...",
            "published_at": null,
            "created_at": "2024-01-15T11:00:00Z",
            "updated_at": "2024-01-15T11:00:00Z"
        }
    ],
    "total": 1,
    "limit": 10,
    "offset": 0
}
```

## 🚨 Important Notes

1. **Firebase Authentication**: The API uses Firebase for authentication. Make sure to use valid email formats and strong passwords.

2. **UUIDs**: All IDs in the system are UUIDs. The collection automatically captures and reuses them.

3. **Pagination**: List endpoints support `limit` and `offset` parameters for pagination.

4. **Email Functionality**: The subscription system sends real emails. Use test email addresses during development.

5. **Rate Limiting**: Be mindful of rate limits when testing extensively.

## 🛠️ Troubleshooting

### Common Issues

**Authentication Errors (401)**
- Ensure you've run the "Editor Sign In" request first
- Check that the `authToken` variable is populated
- Verify your email/password combination

**Not Found Errors (404)**
- Verify that the required IDs are set in collection variables
- Ensure you've created the necessary resources (newsletter before posts, etc.)

**Validation Errors (400)**
- Check request body format and required fields
- Ensure email addresses are valid
- Verify that titles and content are not empty

### Debug Tips
1. Check the Postman Console for detailed request/response logs
2. Verify collection variables are properly set
3. Use the "Tests" tab results to see auto-population status
4. Check the Railway deployment logs if needed

## 🌐 Deployment Information

- **Production URL**: `https://strv-vse-go-newsletter-production.up.railway.app`
- **Platform**: Railway
- **Database**: PostgreSQL (for newsletters, posts, editors)
- **Storage**: Firestore (for subscribers)
- **Authentication**: Firebase Auth

## 📝 Additional Resources

- [API Documentation](../docs/) - Detailed API documentation
- [Project README](../README.md) - Main project documentation
- [Railway Dashboard](https://railway.app) - Deployment management

---

**Happy Testing! 🎉**

This collection provides a complete showcase of the Newsletter API functionality. Run through the requests in order for the best experience, and don't forget to check the automatic variable population in the Tests tab of each request. 