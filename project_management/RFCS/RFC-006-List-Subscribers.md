# RFC-006: List Subscribers

## Summary
Allow editors to view a list of subscribers for their newsletters.

## Included Features
- FE-ED-09: List Subscribers

## Dependencies
- RFC-004

## Technical Approach
- Endpoint for editors to fetch subscribers for a newsletter
- Fetch subscriber data from Firebase
- Auth and ownership checks
- Input validation and error handling

## Acceptance Criteria
- Editors can retrieve subscriber lists for their newsletters
- Data is accurate and up-to-date
- Only owners can access their newsletter's subscribers
- Tests for all flows

## APIs
- GET /newsletters/{id}/subscribers

## Data Models
- subscribers (from Firebase: email, newsletter_id, subscribed_at, ...)

## State
- Subscriber listing is functional and secure

## Error Handling
- Standardized error responses (JSON)
- Ownership and validation errors handled

## Testing
- Unit and integration tests for endpoint
- Mock Firebase in tests

## Notes
- Consider pagination for large subscriber lists 