package mail

import (
	"GoShort/config"
	"GoShort/pkg/logger"
	"fmt"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type SendGridService struct {
	Cfg    *config.AppConfig
	Log    *logger.Logger
	Client *sendgrid.Client
}

func (s *SendGridService) SendEmail(to, subject, body string) error {
	message := mail.NewSingleEmail(
		mail.NewEmail("Sender", s.Cfg.SendGrid.SenderEmail),
		subject,
		mail.NewEmail("Recipient", to),
		body,
		body,
	)

	response, err := s.Client.Send(message)
	if err != nil {
		s.Log.Error("Failed to send email: ", err)
		return err
	}

	if response.StatusCode >= 400 {
		s.Log.Error("SendGrid API error: ", response.StatusCode, response.Body)
		return fmt.Errorf("failed to send email, status code: %d", response.StatusCode)
	}

	s.Log.Info("Email sent successfully to: ", to)
	return nil
}

func (s *SendGridService) SendEmailWithTemplate(to, templateID string, dynamicTemplateData map[string]interface{}) error {
	message := mail.NewV3Mail()
	message.SetFrom(mail.NewEmail("Sender", s.Cfg.SendGrid.SenderEmail))
	message.SetTemplateID(templateID)

	personalization := mail.NewPersonalization()
	personalization.AddTos(mail.NewEmail("Recipient", to))

	for key, value := range dynamicTemplateData {
		personalization.SetDynamicTemplateData(key, value)
	}

	message.AddPersonalizations(personalization)

	response, err := s.Client.Send(message)
	if err != nil {
		s.Log.Error("Failed to send email with template: ", err)
		return err
	}

	if response.StatusCode >= 400 {
		s.Log.Error("SendGrid API error: ", response.StatusCode, response.Body)
		return fmt.Errorf("failed to send email, status code: %d", response.StatusCode)
	}

	s.Log.Info("Email with template sent successfully to: ", to)
	return nil
}

type ISendGridService interface {
	SendEmail(to, subject, body string) error
	SendEmailWithTemplate(to, templateID string, dynamicTemplateData map[string]interface{}) error
}

func NewSendGridService(cfg *config.AppConfig, log *logger.Logger) ISendGridService {
	client := sendgrid.NewSendClient(cfg.SendGrid.APIKey)
	return &SendGridService{
		Cfg:    cfg,
		Log:    log,
		Client: client,
	}
}
