# RFC-004: Subscriber Management

## Summary
Implement subscribe, confirmation, and unsubscribe flows for newsletter subscribers. Store subscriber data in Firebase.

## Included Features
- FE-SB-01: Subscribe to Newsletter
- FE-SB-02: Subscription Confirmation
- FE-SB-03: Unsubscribe from Newsletter

## Dependencies
- RFC-003

## Technical Approach
- Endpoints for subscribe, confirm (email), and unsubscribe
- Store subscriber info in Firebase
- Generate unique subscription and unsubscribe links
- Integrate with email service for confirmation/unsubscribe
- Input validation and security for links

## Acceptance Criteria
- Users can subscribe, receive confirmation, and unsubscribe
- Subscriber data is accurate in Firebase
- Links are secure and functional
- Tests for all flows

## APIs
- POST /newsletters/{id}/subscribe
- POST /newsletters/{id}/unsubscribe

## Data Models
- subscribers (in Firebase: email, newsletter_id, subscribed_at, ...)

## State
- Subscriber management is functional and secure

## Error Handling
- Standardized error responses (JSON)
- Secure handling of links and tokens

## Testing
- Unit and integration tests for all endpoints
- Mock email and Firebase in tests

## Notes
- Double opt-in is not required unless specified 