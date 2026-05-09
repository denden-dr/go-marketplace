# Design Spec: Custom Auth & Integration Tests

**Topic**: Authentication Integration Tests & Email Verification flow

## 1. Overview
Transition the project to a local, custom authentication system. This involves replacing the legacy Supabase-style refresh tokens with a dedicated session management system and implementing a mandatory email verification step.

## 2. Goals
- Verify registration, login, and logout via integration tests.
- Implement and verify the email verification flow.
- Update data models to support verification status and session metadata.
- Simulate manual verification in tests to test authenticated flows.

## 3. Scenarios

### Scenario A: Registration and Authenticated Flow
1. **User Registration**: A new user registers via the API.
2. **Status Check**: Verify the user is marked as unverified in the database.
3. **Manual Verification**: Simulate the verification process by directly updating the user's status in the database using SQL.
4. **Login**: The now-verified user logs in and receives access/refresh tokens.
5. **Logout**: The user logs out, and the session is invalidated in the database.

### Scenario B: Email Verification Flow
1. **User Registration**: A new user registers.
2. **Code Generation**: A verification code is automatically generated and stored.
3. **API Verification**: The user submits the code via the verification endpoint.
4. **Validation**: The system validates the code, marks the user as verified, and removes the code from the database.

## 4. Error Handling
- Invalid or expired verification codes.
- Login attempts by unverified users (permitted but flagged).
- Session reuse detection (invalidating compromised token families).
