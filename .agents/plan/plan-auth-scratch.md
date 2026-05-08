# Design: Custom Email/Password Authentication with OTP Verification

## Overview
This document outlines the transition from Supabase-based authentication to a custom-built email/password authentication system. The new system includes mandatory email verification via a 6-digit One-Time Password (OTP) sent through MailerSend.

## Architecture

### 1. Data Models

#### Users Table (`users`)
| Column | Type | Description |
| :--- | :--- | :--- |
| `id` | UUID | Primary Key |
| `email` | VARCHAR | Unique, user's email |
| `password` | VARCHAR | Bcrypt hashed password |
| `full_name` | VARCHAR | User's full name |
| `username` | VARCHAR | Unique username |
| `is_verified` | BOOLEAN | Verification status (default: false) |
| `created_at` | TIMESTAMP | Creation time |

#### OTP Verifications Table (`otp_verifications`)
| Column | Type | Description |
| :--- | :--- | :--- |
| `id` | UUID | Primary Key |
| `user_id` | UUID | Foreign Key to `users.id` |
| `code` | VARCHAR(6) | 6-digit numeric OTP |
| `expires_at` | TIMESTAMP | Expiration time (10 mins) |
| `created_at` | TIMESTAMP | Generation time |

### 2. Services

#### `EmailService`
- **Responsibility**: Interface with MailerSend REST API.
- **Methods**:
    - `SendOTP(email, code)`: Sends the 6-digit code using a template or raw email.

#### `AuthService`
- **Registration**: 
    1. Create user record with `is_verified = false`.
    2. Generate 6-digit OTP.
    3. Save OTP to `otp_verifications`.
    4. Call `EmailService.SendOTP`.
- **OTP Verification**:
    1. Find valid, non-expired OTP for the email.
    2. Mark user as `is_verified = true`.
    3. Generate and return JWT access/refresh tokens.
- **Login**:
    1. Validate password.
    2. Check `is_verified`. If false, deny login and suggest verification.

### 3. API Endpoints (Fiber)
- `POST /auth/register`: Signup with email/password.
- `POST /auth/verify`: Verify OTP.
- `POST /auth/login`: Standard login.
- `POST /auth/resend-otp`: Resend a new OTP if the previous one expired.

## Security Considerations
- **Password Hashing**: Use `bcrypt` with a cost factor of at least 10.
- **OTP Rate Limiting**: Limit the number of OTP attempts per user/IP.
- **JWT Secret**: Use a strong environment-managed secret.
- **Database Cleanup**: Periodically delete expired tokens.

## Cleanup
- Remove `internal/core/auth/supabase.go`.
- Remove all Supabase-related environment variables and dependencies.
- Truncate existing user-related tables to start fresh.


# Implementation Plan: Custom Auth with OTP Verification (Feature-Based)

This plan details the steps to replace Supabase-based authentication with a custom email/password system using 6-digit OTP verification via MailerSend, following the existing feature-folder structure.

## Phase 1: Database and Domain

### 1.1 Database Migrations
Create a new migration (e.g., `000012_custom_auth_otp.up.sql`) to:
- Add `is_verified` (BOOLEAN, default false) to the `users` table.
- Create the `otp_verifications` table:
  ```sql
  CREATE TABLE otp_verifications (
      id UUID PRIMARY KEY,
      user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      code VARCHAR(6) NOT NULL,
      expires_at TIMESTAMP NOT NULL,
      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
  );
  ```

### 1.2 Domain Model Updates
- **`internal/domain/user.go`**: Add `IsVerified bool` to the `User` struct.
- **`internal/domain/otp.go`** (New): Define `OTPVerification` struct and `OTPRepository` interface.
- **`internal/domain/errors.go`**: Add new error types: `ErrUserNotVerified`, `ErrInvalidOTP`, `ErrExpiredOTP`.

## Phase 2: Auth Feature Implementation (internal/core/auth)

### 2.1 Email Integration
- **`internal/core/auth/mailersend.go`** (New): 
    - Define `EmailService` interface.
    - Implement `mailersendAuthClient` using MailerSend REST API (following the `supabase.go` pattern).

### 2.2 OTP Repository
- **`internal/core/auth/otp_repo.go`** (New): Implement `OTPRepository` for PostgreSQL using `pgx`.

### 2.3 Auth Service Refactoring
- **`internal/core/auth/auth_service.go`**:
    - Add `emailService EmailService` and `otpRepo OTPRepository` to `AuthService` struct.
    - Update `Register`:
        1. Create user with `is_verified = false`.
        2. Generate 6-digit OTP.
        3. Save OTP to database using `otpRepo`.
        4. Send OTP via `emailService`.
    - Implement `VerifyOTP(email, code)`:
        1. Validate code and expiration via `otpRepo`.
        2. Set `is_verified = true` for the user in `userRepo`.
        3. Generate and return JWT tokens.
    - Update `Login`:
        1. Check `is_verified` after password validation. Return `ErrUserNotVerified` if false.
    - Implement `ResendOTP(email)`:
        1. Generate and send a new OTP.
    - **Cleanup**: Remove `socialAuthClient` and Supabase-related logic.

## Phase 3: API & Handlers

### 3.1 Auth Handler Updates
- **`internal/core/auth/auth_handler.go`**:
    - Add `VerifyOTP` and `ResendOTP` methods.
    - Update `Register` response to status `201 Created` with a message that verification is required (no tokens returned yet).
    - Update `Login` to handle `ErrUserNotVerified` and return an appropriate error message.

### 3.2 Routing
- **`internal/server/routes.go`**:
    - Add `authRoutes.Post("/verify", authHandler.VerifyOTP)`
    - Add `authRoutes.Post("/resend-otp", authHandler.ResendOTP)`

## Phase 4: Environment & Cleanup

### 4.1 Configuration
- Update `.env` and `.env.example`:
    - Add `MAILERSEND_API_KEY`
    - Add `MAILERSEND_SENDER_EMAIL`
    - Add `MAILERSEND_SENDER_NAME`
    - Remove all `SUPABASE_*` variables.

### 4.2 Code Cleanup
- Delete `internal/core/auth/supabase.go` and `internal/core/auth/supabase_test.go`.
- Remove `github.com/supabase-community/gotrue-go` from `go.mod`.

## Verification Plan
- **Unit Tests**:
    - Update `auth_service_test.go` to mock `EmailService` and `OTPRepository`.
    - Test registration flow with OTP generation.
    - Test OTP verification (success, expired, invalid).
- **Manual Testing**:
    - Register a new user and verify OTP delivery.
    - Verify OTP via `/auth/verify`.
    - Attempt login before and after verification.
