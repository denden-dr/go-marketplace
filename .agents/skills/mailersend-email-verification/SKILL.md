---
name: mailersend-email-verification
description: Use when integrating MailerSend for email verification or transactional emails in Go projects.
---

# MailerSend Email Verification

## Overview
Secure and reliable email verification using MailerSend's official Go SDK. Prioritizes SDK usage, context timeouts, and template-based delivery.

## When to Use
- Implementing signup/registration verification.
- Sending One-Time Passwords (OTP).
- Implementing transactional emails (password reset, notifications).

## Core Pattern

### 1. SDK Integration
Always use `github.com/mailersend/mailersend-go` instead of raw HTTP calls.

### 2. Context & Timeouts
Never use `context.Background()` or `context.TODO()` for API calls. Always use a timeout.

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
```

### 3. Template-Based Sending (Recommended)
Avoid hardcoding HTML in Go code. Use MailerSend templates for better maintenance.

```go
message := ms.Email.NewMessage()
message.SetTemplateID("your-template-id")
message.SetPersonalization([]mailersend.Personalization{
    {
        Email: email,
        Data: map[string]interface{}{
            "account_name": "Go Marketplace",
            "otp_code":     code,
        },
    },
})
```

## Implementation Example

```go
import (
    "context"
    "time"
    "github.com/mailersend/mailersend-go"
)

func SendOTP(apiKey, fromEmail, toEmail, code string) error {
    ms := mailersend.NewMailersend(apiKey)
    
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    message := ms.Email.NewMessage()
    message.SetFrom(mailersend.From{Name: "Auth System", Email: fromEmail})
    message.SetRecipients([]mailersend.Recipient{{Email: toEmail}})
    message.SetSubject("Verify Your Email")
    message.SetText("Your code is: " + code)
    message.SetHTML("<b>Your code is: " + code + "</b>")

    _, err := ms.Email.Send(ctx, message)
    return err
}
```

## Common Mistakes
- **No Timeout**: API hangs can block the whole request.
- **Raw HTTP**: Missing out on SDK features like retries or type-safe payloads.
- **Insecure OTP**: Not using `crypto/rand` for code generation.

## Red Flags
- Using `net/http` directly for MailerSend.
- Missing `context.WithTimeout`.
- Hardcoding complex HTML strings in `.go` files.
