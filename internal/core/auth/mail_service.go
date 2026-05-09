package auth

import (
	"context"
	"time"

	"github.com/mailersend/mailersend-go"
)

type MailService interface {
	SendVerificationCode(ctx context.Context, email, code string) error
}

type mailService struct {
	apiKey    string
	fromEmail string
}

func NewMailService(apiKey, fromEmail string) MailService {
	return &mailService{
		apiKey:    apiKey,
		fromEmail: fromEmail,
	}
}

func (s *mailService) SendVerificationCode(ctx context.Context, email, code string) error {
	ms := mailersend.NewMailersend(s.apiKey)

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	message := ms.Email.NewMessage()
	message.SetFrom(mailersend.From{Name: "Go Marketplace", Email: s.fromEmail})
	message.SetRecipients([]mailersend.Recipient{{Email: email}})
	message.SetSubject("Verify Your Email")
	message.SetText("Your verification code is: " + code)
	message.SetHTML("<b>Your verification code is: " + code + "</b><p>This code will expire in 15 minutes.</p>")

	_, err := ms.Email.Send(ctx, message)
	return err
}
