package mailer

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
)

// Sender sends transactional emails via SMTP.
type Sender struct {
	host     string
	port     string
	username string
	password string
	from     string
}

// NewFromEnv creates a Sender from environment variables.
// Returns nil if SMTP is not configured (development mode).
func NewFromEnv() *Sender {
	host := os.Getenv("SMTP_HOST")
	if host == "" {
		return nil
	}
	return &Sender{
		host:     host,
		port:     envOrDefault("SMTP_PORT", "587"),
		username: os.Getenv("SMTP_USERNAME"),
		password: os.Getenv("SMTP_PASSWORD"),
		from:     envOrDefault("SMTP_FROM", "noreply@quotetrack.com"),
	}
}

// SendForgotPassword sends a password-reset email.
func (s *Sender) SendForgotPassword(toEmail, resetURL string) error {
	if s == nil {
		log.Printf("[mailer] SMTP not configured — skipping email to %s", toEmail)
		log.Printf("[mailer] Reset URL: %s", resetURL)
		return nil
	}

	subject := "Reset your QuoteTrack password"
	body := fmt.Sprintf(
		"Hello,\n\nYou requested a password reset for your QuoteTrack account.\n\n"+
			"Click the link below to reset your password:\n\n%s\n\n"+
			"This link expires in 1 hour.\n\n"+
			"If you didn't request this, you can safely ignore this email.\n\n"+
			"QuoteTrack",
		resetURL,
	)

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		s.from, toEmail, subject, body)

	addr := s.host + ":" + s.port
	auth := smtp.PlainAuth("", s.username, s.password, s.host)

	err := smtp.SendMail(addr, auth, s.from, []string{toEmail}, []byte(msg))
	if err != nil {
		log.Printf("[mailer] failed to send email to %s: %v", toEmail, err)
		return fmt.Errorf("send email: %w", err)
	}
	log.Printf("[mailer] password reset email sent to %s", toEmail)
	return nil
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
