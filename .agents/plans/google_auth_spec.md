# Technical Specification: Google OAuth Integration

## 1. Goal
Implement a robust Google OAuth2 authentication flow for the Go Marketplace backend, supporting both new user registration and automatic account linking for existing users.

## 2. Architecture

### 2.1 Component Overview
- **`GoogleClient` (Interface)**: Abstraction for Google OAuth2 operations (exchanging codes, fetching user info). Lives in `internal/core/auth/google_client.go`.
- **`AuthHandler`**: Handles HTTP redirects and callbacks. Modified in `internal/core/auth/auth_handler.go`.
- **`AuthService`**: Business logic for identity verification, user provisioning, and account linking. Modified in `internal/core/auth/auth_service.go`.
- **`UserRepository`**: Already has `GetUserByEmail`, `GetUserByProviderID`, and `CreateUser` — no changes needed.

### 2.2 New Interface: `GoogleClient`
```
type GoogleUserInfo struct {
    Sub           string  // Google's unique user ID
    Email         string  // Verified email address
    EmailVerified bool
    Name          string  // Full name
    Picture       string  // Profile picture URL (optional)
}

type GoogleClient interface {
    GetAuthURL(state string) string
    ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error)
    GetUserInfo(ctx context.Context, token *oauth2.Token) (*GoogleUserInfo, error)
}
```

### 2.3 AuthService Interface Changes
Add a new method to the existing `AuthService` interface:
```
HandleGoogleLogin(ctx context.Context, code, state, expectedState, ipAddress, userAgent string) (*AuthResponse, error)
```

### 2.4 Sequence Flow
1. **Initiate**: `GET /api/auth/google/login` → Handler generates a random `state`, stores it in a short-lived cookie (`oauth_state`, 10 min TTL, HttpOnly, Secure, SameSite=Lax), then redirects (302) to Google's OAuth2 consent URL.
2. **Consent**: User approves app on Google.
3. **Callback**: `GET /api/auth/google/callback?code=xxx&state=yyy` → Handler reads `oauth_state` cookie, compares with query `state`.
4. **Exchange**: Backend exchanges `code` for Google Tokens via `GoogleClient.ExchangeCode`.
5. **Fetch**: Backend fetches UserInfo via `GoogleClient.GetUserInfo`.
6. **Orchestrate** (in `AuthService.HandleGoogleLogin`):
   - Call `userRepo.GetUserByEmail(email)`.
   - **Existing user (any provider)**: Allow login. Issue local JWT + Refresh Token using existing `generateAuthResponse`.
   - **New user**: Create user with `auth_provider: "google"`, `provider_id: sub`, `is_verified: true`, `password: nil`. Auto-generate a username from the email prefix (e.g., `denden` from `denden@gmail.com`). Issue local JWT + Refresh Token.
7. **Session**: Generate local JWT + Refresh Token using existing session family logic.
8. **Finish**: Set JWT cookies (`access_token`, `refresh_token`) and redirect to frontend success URL (configurable via `GOOGLE_LOGIN_REDIRECT_URL` env var).

### 2.5 State Storage Strategy
- The `state` parameter is stored in a **short-lived cookie** (`oauth_state`) set during the `/google/login` redirect.
- On callback, the handler reads the cookie and compares it to the `state` query parameter.
- The cookie is cleared immediately after verification.
- Cookie settings: `HttpOnly: true`, `Secure: true`, `SameSite: Lax`, `MaxAge: 600` (10 minutes).
- `SameSite=Lax` is required (not `Strict`) because Google redirects back cross-origin.

## 3. Data Model Changes
- **No schema migration needed**: The `users` table already has `auth_provider` (text) and `provider_id` (text, nullable) fields. The `password` field is already nullable (`*string` in Go, nullable in DB).
- We will use `auth_provider = "google"` and `provider_id = <Google sub>`.

## 4. Account Linking & Registration Rules

### 4.1 Google Login → Existing Local User (Automatic Linking)
- User registered with email/password, then clicks "Login with Google" with the same email.
- We trust Google's verified email. Allow login to the existing account.
- The user's `auth_provider` field is **not** changed (stays "local"). They can still log in with their password.
- Future: optionally update to a multi-provider model.

### 4.2 Google Login → Existing Google User (Returning User)
- Standard login. Find user by email, issue new session.

### 4.3 Google Login → New User
- Create a new user record with `auth_provider: "google"`, `provider_id`, `is_verified: true`, `password: nil`.

### 4.4 Email/Password Registration → Existing Google User (BLOCKED)
- The existing `Register` method already calls `userRepo.GetUserByEmail` and returns `domain.ErrUserAlreadyExists` if a user exists.
- **No code change needed** — this is already handled.

### 4.5 Email/Password Login → Google-only User (Already Handled)
- The existing `Login` method checks `if user.Password == nil` and returns `domain.ErrInvalidCredentials`.
- Google-only users (no password set) cannot log in via email/password. They must use Google or "Forgot Password" to set a password.

## 5. Error Handling
- **New domain error**: `ErrInvalidOAuthState` — returned when callback `state` doesn't match cookie.
- **Existing errors to reuse**: `ErrUserAlreadyExists`, `ErrInvalidCredentials`.
- **Existing unused errors to leverage**: `ErrInvalidSocialToken` (for when Google returns an error during code exchange or userinfo fetch).

## 6. Configuration

### 6.1 New Environment Variables
| Variable | Required | Description |
|----------|----------|-------------|
| `GOOGLE_CLIENT_ID` | Yes | OAuth2 Client ID from Google Cloud Console |
| `GOOGLE_CLIENT_SECRET` | Yes | OAuth2 Client Secret from Google Cloud Console |
| `GOOGLE_REDIRECT_URL` | Yes | Callback URL registered in Google Console (e.g., `http://localhost:3000/api/auth/google/callback`) |
| `GOOGLE_LOGIN_REDIRECT_URL` | Yes | Frontend URL to redirect after successful login (e.g., `http://localhost:5173/login-success`) |

### 6.2 Config Struct Changes
Add `GoogleOAuth` section to `internal/config/config.go`:
```
type GoogleOAuthConfig struct {
    ClientID     string
    ClientSecret string
    RedirectURL  string
    LoginRedirectURL string
}
```

### 6.3 `.env.example` Cleanup
- Remove stale Supabase entries (`SUPABASE_URL`, `SUPABASE_JWT_SECRET`).
- Add Google OAuth entries.
- Add MailerSend entries (currently missing from `.env.example`).

## 7. Files to Modify

### New Files
| File | Purpose |
|------|---------|
| `internal/core/auth/google_client.go` | `GoogleClient` interface + `googleClient` implementation |

### Modified Files
| File | Changes |
|------|---------|
| `internal/config/config.go` | Add `GoogleOAuthConfig` struct and load from env |
| `internal/core/auth/auth_service.go` | Add `HandleGoogleLogin` to interface + implementation |
| `internal/core/auth/auth_handler.go` | Add `GoogleLogin` and `GoogleCallback` handlers |
| `internal/core/auth/auth_dto.go` | Add `GoogleUserInfo` struct (if not in google_client.go) |
| `internal/domain/errors.go` | Add `ErrInvalidOAuthState` |
| `internal/server/routes.go` | Register `GET /auth/google/login` and `GET /auth/google/callback` |
| `cmd/api/main.go` | Wire `GoogleClient` and pass to `AuthService`/`AuthHandler` |
| `.env.example` | Remove Supabase vars, add Google + MailerSend vars |
| `GEMINI.md` | Update auth description (remove Supabase, add Google OAuth) |
| `CLAUDE.md` | Update auth description (remove Supabase, add Google OAuth) |

### Auto-generated (via `make mock`)
| File | Purpose |
|------|---------|
| `internal/core/auth/mock_GoogleClient.go` | Mock for `GoogleClient` interface |
| `internal/core/auth/mock_AuthService.go` | Regenerated to include `HandleGoogleLogin` |

## 8. Mocking Strategy for Integration Tests
- **Approach**: Mock HTTP Server using `httptest.NewServer`.
- **Implementation**:
  - The `googleClient` implementation accepts configurable `AuthURL` and `TokenURL` (via `oauth2.Config.Endpoint`).
  - Integration tests start a local `httptest.NewServer` that mimics Google's `/token` and `/userinfo` responses.
  - Test helper creates a `MockGoogleServer` struct with methods to configure responses (success, error, different user profiles).
  - This allows testing the full handler → service → repository flow without external network calls.

## 9. Swagger Documentation
- Add Swagger comments to `GoogleLogin` and `GoogleCallback` handlers.
- Run `make swagger` to regenerate docs after implementation.
