# Custom Auth Implementation Plan

**Goal**: Implement custom auth migration, sessions, and email verification with scenario-based integration tests.

---

## Prerequisites
- Database migration `000002_custom_auth.up.sql` must be applied to the test database.
- MailerSend API configuration (mocked for tests).

## Add/Modify Files List

### Domain
- `internal/domain/user.go`: Add `IsVerified` boolean field.
- `internal/domain/session.go`: Create from `refresh_token.go`, add `IPAddress` and `UserAgent`.
- `internal/domain/verification.go`: Create new entity for `VerificationCode`.

### Repositories
- `internal/core/auth/session_repo.go`: Refactor from `refresh_token_repo.go` to use `sessions` table.
- `internal/core/auth/verification_repo.go`: Implement new repository for verification codes.

### Services
- `internal/core/auth/auth_service.go`: 
    - Update `Register` to generate and store verification codes.
    - Implement `VerifyEmail` logic.
    - Update `Login` and `RefreshTokens` to include verification status.

### Handlers & Routes
- `internal/core/auth/auth_handler.go`: Add `VerifyEmail` handler method.
- `internal/server/routes.go`: Register the `/api/auth/verify-email` endpoint.

### Testing
- `test/integration/auth_integration_test.go`: Create new scenario tests.

---

## Edge Cases & Common Problems
- **Code Expiration**: Ensuring verification codes are purged after they expire.
- **Race Conditions**: Multiple verification attempts for the same user.
- **Session Security**: Correctly tracking and updating IP/User-Agent on every refresh.
- **Database Consistency**: Ensuring atomic updates when verifying email and deleting codes.

---

## User Stories

### Story 1: Happy Path - Manual Verification
1. User registers successfully (marked as unverified).
2. Developer/Admin manually verifies user in the test using a direct SQL command.
3. User logs in successfully and receives tokens.
4. User logs out, and the session family is revoked.

### Story 2: Happy Path - Email Verification
1. User registers successfully.
2. System generates a 6-digit verification code.
3. User submits the correct code via the API.
4. System marks the user as verified and deletes the code.
5. User can now access protected features.

### Story 3: Error Path - Invalid Verification
1. User submits an incorrect or expired code.
2. System returns a descriptive error (e.g., "Invalid verification code").
3. User status remains unverified.

---

## Test Case Updates

### Unit Tests
- Update `auth_service_test.go` to mock the new `VerificationRepository` and `SessionRepository`.
- Add test cases for the `VerifyEmail` service method.

### Integration Tests
- Implement `TestAuthScenario`: Registration -> SQL Update -> Login -> Logout.
- Implement `TestEmailVerificationScenario`: Registration -> Retrieve Code from DB -> API Verify -> Success Check.
