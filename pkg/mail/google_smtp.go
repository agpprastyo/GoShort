package mail

import (
	"GoShort/config"
	"GoShort/pkg/logger"
	"fmt"
	"net/smtp"
)

// GoogleSMTPService holds the necessary configuration for sending emails via Gmail's SMTP server.
type GoogleSMTPService struct {
	Cfg *config.AppConfig
	Log *logger.Logger
	// auth holds the authentication information for the SMTP server.
	auth smtp.Auth
}

// IGoogleSMTPService defines the interface for our Google SMTP email service.
// By using an interface, you can easily swap this implementation with another one (like SendGrid)
// without changing the code in your authentication service.
type IGoogleSMTPService interface {
	SendEmail(to, subject, body string) error
}

// NewGoogleSMTPService creates and initializes a new GoogleSMTPService.
// It sets up the SMTP authentication using the credentials from your app's configuration.
func NewGoogleSMTPService(cfg *config.AppConfig, log *logger.Logger) IGoogleSMTPService {
	// PlainAuth is used to authenticate with the SMTP server.
	// It requires the sender's identity, username (usually the email), password, and the SMTP host.
	auth := smtp.PlainAuth(
		"", // identity, can be left empty
		cfg.GoogleSMTP.SenderEmail,
		cfg.GoogleSMTP.AppPassword, // Use the App Password here, not your regular Gmail password
		cfg.GoogleSMTP.Host,
	)

	return &GoogleSMTPService{
		Cfg:  cfg,
		Log:  log,
		auth: auth,
	}
}

// SendEmail constructs and sends an email using the configured Google SMTP server.
func (s *GoogleSMTPService) SendEmail(to, subject, body string) error {
	// The message must be formatted according to RFC 822.
	// We need to set the From, To, Subject, and MIME-Version headers,
	// and specify that the content is HTML.
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	from := fmt.Sprintf("From: %s", s.Cfg.GoogleSMTP.SenderEmail)
	toHeader := fmt.Sprintf("To: %s", to)
	subjectHeader := fmt.Sprintf("Subject: %s", subject)

	// Combine the headers and the HTML body to form the complete email message.
	msg := []byte(from + "\n" + toHeader + "\n" + subjectHeader + "\n" + mime + "\n" + body)

	// The address for the SMTP server is in the format "host:port".
	addr := fmt.Sprintf("%s:%d", s.Cfg.GoogleSMTP.Host, s.Cfg.GoogleSMTP.Port)

	// smtp.SendMail connects to the server, authenticates, and sends the email.
	err := smtp.SendMail(addr, s.auth, s.Cfg.GoogleSMTP.SenderEmail, []string{to}, msg)
	if err != nil {
		s.Log.Error("Failed to send email via Google SMTP", "error", err)
		return err
	}

	s.Log.Info("Email sent successfully via Google SMTP to", "recipient", to)
	return nil
}
