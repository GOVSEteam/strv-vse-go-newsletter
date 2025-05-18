# RFC-005: Publishing & Email Delivery

## Summary
Implement post publishing, email delivery to subscribers, and archiving of published posts.

## Included Features
- FE-PB-01: Publish Post to Newsletter
- FE-PB-02: Email Delivery of Posts
- FE-PB-03: Archive Published Posts

## Dependencies
- RFC-004

## Technical Approach
- Endpoint for publishing a post to a newsletter
- Store post in PostgreSQL (archive)
- Fetch subscribers from Firebase
- Integrate with email service for delivery
- Batch/async email sending for scale
- Input validation and error handling

## Acceptance Criteria
- Editors can publish posts to their newsletters
- All subscribers receive the post via email
- Posts are archived in DB
- Delivery failures are logged/handled
- Tests for all flows

## APIs
- POST /newsletters/{id}/posts
- GET /newsletters/{id}/posts

## Data Models
- posts (id, newsletter_id, title, body, created_at, ...)

## State
- Publishing and email delivery are functional and reliable

## Error Handling
- Standardized error responses (JSON)
- Log and handle email delivery failures

## Testing
- Unit and integration tests for all endpoints
- Mock email and Firebase in tests

## Notes
- Consider rate limiting and retries for email delivery 