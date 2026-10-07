package email

import (
	"fmt"
	"strings"

	"github.com/resend/resend-go/v2"
)

type Config struct {
	APIKey    string
	FromEmail string
	FromName  string
}

type Service struct {
	client    *resend.Client
	fromEmail string
	fromName  string
}

func NewService(config Config) *Service {
	return &Service{
		client:    resend.NewClient(config.APIKey),
		fromEmail: config.FromEmail,
		fromName:  config.FromName,
	}
}

func (s *Service) Send(
	to string,
	subject string,
	body string,
) error {
	to = strings.TrimSpace(to)

	if to == "" {
		return fmt.Errorf("recipient email is required")
	}

	if s.fromEmail == "" {
		return fmt.Errorf("email from address is not configured")
	}

	if s.client == nil {
		return fmt.Errorf("Resend email client is not configured")
	}

	from := s.fromEmail

	if strings.TrimSpace(s.fromName) != "" {
		from = fmt.Sprintf("%s <%s>", s.fromName, s.fromEmail)
	}

	params := &resend.SendEmailRequest{
		From:    from,
		To:      []string{to},
		Subject: subject,
		Text:    body,
	}

	_, err := s.client.Emails.Send(params)
	if err != nil {
		return fmt.Errorf("failed to send email through Resend: %w", err)
	}

	return nil
}
