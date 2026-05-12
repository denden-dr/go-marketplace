# Implementation Plan: Google OAuth Integration

**Goal**: Add Google Authentication with automatic account linking and robust integration testing using mocked Google APIs.

## 1. Prerequisites
- [ ] Create Google Cloud Project and obtain `Client ID` and `Client Secret`.
- [ ] Configure Authorized Redirect URIs in Google Console (e.g., `http://localhost:3000/api/auth/google/callback`).
- [ ] Install `golang.org/x/oauth2` package (`go get golang.org/x/oauth2`).
- [ ] Install `google.golang.org/api/oauth2/v2` package for UserInfo retrieval.

## 2. Add/Modify Files List

### New Files
- `internal/core/auth/google_client.go` — `GoogleClient` interface and `googleClient` implementation wrapping `golang.org/x/oauth2`.

### Modified Files (Core Logic)
- `internal/config/config.go` — Add `GoogleOAuthConfig` struct with `ClientID`, `ClientSecret`, `RedirectURL`, `LoginRedirectURL`. Load from env vars.
- `internal/core/auth/auth_service.go` — Add `HandleGoogleLogin` method to `AuthService` interface and `authService` implementation.
- `internal/core/auth/auth_handler.go` — Add `GoogleLogin` (redirect) and `GoogleCallback` (code exchange + session creation) handlers. Add `GoogleClient` dependency to `AuthHandler`.
- `internal/domain/errors.go` — Add `ErrInvalidOAuthState` error.
- `internal/server/routes.go` — Register `GET /auth/google/login` and `GET /auth/google/callback` as public routes.
- `cmd/api/main.go` — Wire `GoogleClient`, inject into `AuthService` and `AuthHandler`. Add env var validation for Google OAuth config.

### Modified Files (Config & Documentation)
- `.env.example` — Remove stale Supabase entries (`SUPABASE_URL`, `SUPABASE_JWT_SECRET`). Add `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `GOOGLE_REDIRECT_URL`, `GOOGLE_LOGIN_REDIRECT_URL`. Add `MAILERSEND_API_KEY`, `MAILERSEND_FROM_EMAIL`.
- `GEMINI.md` — Update Auth tech stack description (remove Supabase references, add Google OAuth).
- `CLAUDE.md` — Update Auth tech stack description (remove Supabase references, add Google OAuth).

### Auto-generated Files (via `make mock`)
- `internal/core/auth/mock_GoogleClient.go` — Generated mock for `GoogleClient` interface.
- `internal/core/auth/mock_AuthService.go` — Regenerated to include new `HandleGoogleLogin` method.

### Test Files
- `internal/core/auth/auth_service_test.go` — Add unit tests for `HandleGoogleLogin` with mocked `GoogleClient`.
- `test/integration/api/auth_api_test.go` — Add integration tests using `httptest.NewServer` to simulate Google API.

### Post-implementation
- Run `make mock` to regenerate all mocks.
- Run `make swagger` to regenerate Swagger documentation.

## 3. Edge Cases & Common Problems

### OAuth Flow
- **State Mismatch (CSRF)**: The `state` parameter is stored in a short-lived cookie (`oauth_state`, 10 min TTL, `SameSite=Lax`). On callback, it must match the query `state`. Use `SameSite=Lax` (not `Strict`) because Google redirects back cross-origin.
- **Code Exchange Failure**: Google may return an error if the code is expired or already used. Handle gracefully with `domain.ErrInvalidSocialToken`.
- **Google API Downtime**: Wrap HTTP calls with proper context timeout. Return a clear error if Google is unreachable.

### Account Identity
- **Email already exists (any provider)**: Automatic linking — allow Google login for existing accounts (whether local or already Google). The `auth_provider` field is **not** changed on the existing user.
- **Google user tries email/password registration**: Blocked. `Register` already returns `domain.ErrUserAlreadyExists`.
- **Google-only user tries email/password login**: Already handled. `Login` checks `user.Password == nil` and returns `domain.ErrInvalidCredentials`. User should use "Forgot Password" to set a password.
- **Username collision**: When creating a Google user, auto-generate username from email prefix (e.g., `denden@gmail.com` → `denden`). If taken, append a random suffix.

### Testing
- **Mock Server Lifecycle**: Start `httptest.NewServer` in test setup, close in teardown. Pass configurable base URL to `GoogleClient`.
- **Token Format**: Mock server should return realistic-looking (but fake) `access_token` and `id_token` JSON responses.
- **Missing fields**: Handle cases where Google UserInfo returns incomplete data (e.g., no name).

## 4. Stories & Flows

### Story 1: New User Registration via Google (Happy Path)
1. User clicks "Login with Google" on frontend.
2. Frontend calls `GET /api/auth/google/login`.
3. Handler generates random `state`, sets `oauth_state` cookie, redirects to Google.
4. User approves app on Google consent screen.
5. Google redirects to `GET /api/auth/google/callback?code=abc&state=xyz`.
6. Handler reads `oauth_state` cookie, verifies it matches `state` query param.
7. Handler clears `oauth_state` cookie.
8. Service calls `GoogleClient.ExchangeCode("abc")` → gets `oauth2.Token`.
9. Service calls `GoogleClient.GetUserInfo(token)` → gets `{email: "new@gmail.com", name: "New User", sub: "12345"}`.
10. Service calls `userRepo.GetUserByEmail("new@gmail.com")` → returns `nil` (user doesn't exist).
11. Service creates new user: `auth_provider: "google"`, `provider_id: "12345"`, `is_verified: true`, `password: nil`.
12. Service generates JWT access + refresh tokens via `generateAuthResponse`.
13. Handler sets `access_token` and `refresh_token` cookies.
14. Handler redirects (302) to `GOOGLE_LOGIN_REDIRECT_URL`.

### Story 2: Account Linking — Existing Local User Logs in via Google (Happy Path)
1. User previously registered with `denden@gmail.com` + password.
2. User clicks "Login with Google".
3. Flow proceeds through steps 2–9 (same as Story 1), returning `email: "denden@gmail.com"`.
4. Service calls `userRepo.GetUserByEmail("denden@gmail.com")` → returns existing user with `auth_provider: "local"`.
5. Service allows the login (automatic linking).
6. Service generates JWT access + refresh tokens via `generateAuthResponse`.
7. Handler sets cookies and redirects to success URL.
8. **User's `auth_provider` remains "local"** — both login methods continue to work.

### Story 3: Returning Google User Logs in Again (Happy Path)
1. User previously created account via Google OAuth.
2. User clicks "Login with Google" again.
3. Flow proceeds to identity retrieval.
4. Service calls `userRepo.GetUserByEmail` → returns existing user with `auth_provider: "google"`.
5. Standard login: generate JWT tokens, set cookies, redirect.

### Story 4: Google User Tries Email/Password Registration (Error Path — Blocked)
1. User previously registered via Google with `denden@gmail.com`.
2. User navigates to standard registration form and submits `denden@gmail.com` + a password.
3. `AuthService.Register` calls `userRepo.GetUserByEmail("denden@gmail.com")` → user exists.
4. Service returns `domain.ErrUserAlreadyExists`.
5. Handler returns `409 Conflict: "user already exists"`.
6. **No code change needed** — this is already handled by the existing `Register` logic.

### Story 5: Google-only User Tries Email/Password Login (Error Path — No Password)
1. User only has a Google account (no password set).
2. User tries `POST /api/auth/login` with email + password.
3. `AuthService.Login` finds the user, checks `user.Password == nil`.
4. Returns `domain.ErrInvalidCredentials`.
5. Handler returns `401 Unauthorized`.
6. **No code change needed** — already handled. User should use "Forgot Password" to set a password first.

### Story 6: Invalid OAuth State (Error Path — CSRF Attack)
1. Attacker crafts a callback URL with a manipulated `state` parameter.
2. Handler reads `oauth_state` cookie, compares with query `state`.
3. Mismatch detected → returns `domain.ErrInvalidOAuthState`.
4. Handler returns `401 Unauthorized: "invalid OAuth state"`.
5. No code exchange occurs.

### Story 7: Google Returns Error During Code Exchange (Error Path)
1. User's authorization code is expired or already used.
2. `GoogleClient.ExchangeCode` fails → returns error.
3. Service wraps the error as `domain.ErrInvalidSocialToken`.
4. Handler returns `401 Unauthorized: "invalid or expired social authentication token"`.

## 5. Test Case Updates

### Unit Tests (`internal/core/auth/auth_service_test.go`)
- **Test `HandleGoogleLogin` — New User**: Mock `GoogleClient` to return valid user info. Mock `UserRepo.GetUserByEmail` to return nil. Assert `CreateUser` is called with `auth_provider: "google"`, `is_verified: true`, `password: nil`.
- **Test `HandleGoogleLogin` — Existing Local User (Linking)**: Mock `GoogleClient` to return user info. Mock `UserRepo.GetUserByEmail` to return existing local user. Assert no `CreateUser` call. Assert `generateAuthResponse` is called with existing user.
- **Test `HandleGoogleLogin` — Returning Google User**: Mock `UserRepo.GetUserByEmail` to return existing Google user. Assert standard login flow.
- **Test `HandleGoogleLogin` — Google API Failure**: Mock `GoogleClient.ExchangeCode` to return error. Assert `domain.ErrInvalidSocialToken` is returned.
- **Test `HandleGoogleLogin` — Invalid State**: Pass mismatched state values. Assert `domain.ErrInvalidOAuthState` is returned.
- **Test `Register` — Existing Google User Blocked** (may already exist): Mock `UserRepo.GetUserByEmail` to return Google user. Assert `domain.ErrUserAlreadyExists`.

### Integration Tests (`test/integration/api/auth_api_test.go`)
- **Setup**: Create a `MockGoogleServer` using `httptest.NewServer` with handlers for `/token` (returns access token JSON) and `/userinfo` (returns user profile JSON).
- **Test Google Login Redirect**: `GET /api/auth/google/login` → assert `302 redirect` and `oauth_state` cookie is set.
- **Test Google Callback — New User**: Hit `/api/auth/google/callback` with valid code and matching state cookie → assert `302 redirect`, `access_token` cookie set, user created in DB.
- **Test Google Callback — Existing Local User**: Pre-insert a local user with same email → hit callback → assert login success, user not duplicated.
- **Test Google Callback — Invalid State**: Hit callback with wrong state → assert `401` error.
- **Test Email/Password Registration Blocked for Google User**: Pre-insert a Google user → try `POST /api/auth/register` with same email → assert `409 Conflict`.

### Mock Regeneration
- After adding `GoogleClient` interface and updating `AuthService` interface, run `make mock` to regenerate:
  - `internal/core/auth/mock_GoogleClient.go`
  - `internal/core/auth/mock_AuthService.go`
