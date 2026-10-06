package email

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"
)

type Config struct {
	Host      string
	Port      int
	Username  string
	Password  string
	FromEmail string
	FromName  string
}

type Service struct {
	config Config
}

func NewService(config Config) *Service {
	return &Service{
		config: config,
	}
}

func (s *Service) Send(
	to string,
	subject string,
	body string,
) error {
	if strings.TrimSpace(to) == "" {
		return fmt.Errorf("recipient email is required")
	}

	if s.config.Host == "" {
		return fmt.Errorf("SMTP host is not configured")
	}

	if s.config.Username == "" {
		return fmt.Errorf("SMTP username is not configured")
	}

	if s.config.Password == "" {
		return fmt.Errorf("SMTP password is not configured")
	}

	if s.config.FromEmail == "" {
		return fmt.Errorf("SMTP from email is not configured")
	}

	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	header := strings.Builder{}
	header.WriteString(fmt.Sprintf("From: %s <%s>\r\n", s.config.FromName, s.config.FromEmail))
	header.WriteString(fmt.Sprintf("To: %s\r\n", to))
	header.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	header.WriteString("MIME-Version: 1.0\r\n")
	header.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	header.WriteString("\r\n")

	message := []byte(header.String() + body)

	dialer := &net.Dialer{
		Timeout: 15 * time.Second,
	}

	conn, err := dialer.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.config.Host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	if ok, _ := client.Extension("STARTTLS"); ok {
		tlsConfig := &tls.Config{
			ServerName: s.config.Host,
			MinVersion: tls.VersionTLS12,
		}

		if err := client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("failed to start TLS: %w", err)
		}
	}

	auth := smtp.PlainAuth(
		"",
		s.config.Username,
		s.config.Password,
		s.config.Host,
	)

	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP authentication failed: %w", err)
	}

	if err := client.Mail(s.config.FromEmail); err != nil {
		return fmt.Errorf("SMTP MAIL FROM failed: %w", err)
	}

	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("SMTP RCPT TO failed: %w", err)
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("SMTP DATA failed: %w", err)
	}

	if _, err := writer.Write(message); err != nil {
		_ = writer.Close()
		return fmt.Errorf("failed to write email: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to finalize email: %w", err)
	}

	if err := client.Quit(); err != nil {
		return fmt.Errorf("failed to close SMTP session: %w", err)
	}

	return nil
}
