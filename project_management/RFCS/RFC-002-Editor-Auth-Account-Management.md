# RFC-002: Editor Auth & Account Management

## Summary
Implement editor registration, login, JWT-based stateless authentication, and password reset functionality.

## Included Features
- FE-ED-01: Editor Registration
- FE-ED-02: Editor Login
- FE-ED-03: Stateless Auth (JWT)
- FE-ED-04: Password Reset

## Dependencies
- RFC-001

## Technical Approach
- Implement endpoints for registration, login, password reset request, and password reset
- Use PostgreSQL for editor data
- Integrate Firebase Auth for JWT management
- Secure password storage (hashing)
- Email integration for password reset
- Input validation and error handling

## Acceptance Criteria
- Editors can register, log in, and receive JWT
- Password reset flow works (email, token, update)
- All endpoints secured and validated
- Tests for all flows

## APIs
- POST /editor/signin
- POST /editor/signup
- POST /editor/password-reset-request

## Data Models
- editors (id, firebase_id, email)
- credentials are stored in firebase

## State
- Editor auth flows are functional and secure

## Error Handling
- Standardized error responses (JSON)
- Secure handling of auth and reset tokens

## Testing
- Unit and integration tests for all endpoints
- Mock email and JWT in tests

## Notes
- Social auth is handled in a later (optional) RFC 