package email

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/samaasi/uptime-application/services/api-services/internal/config"
)

// NoopService implements EmailService but does nothing. (non-nil EmailService interface)
type NoopService struct{}

// BasicTemplateRenderer is a simple implementation for demonstration.
type BasicTemplateRenderer struct{}

// Provider defines the interface for email sending providers
type Provider interface {
	SendEmail(ctx context.Context, from, to, subject, body string) error
	Name() string
	GetFromAddress() string
	HealthCheck(ctx context.Context) error
}

// Service manages multiple email providers with failover
type Service interface {
	SendEmail(ctx context.Context, to, subject, body string) error
	SendTemplatedEmail(ctx context.Context, to, templateContent, templateSubject string, templateData map[string]string) error
	HealthCheck(ctx context.Context) error
}

type ServiceImpl struct {
	defaultProviderName string
	providersMap        map[string]Provider
	failoverOrder       []string
	cfg                 *config.EmailConfig
	templateRenderer    TemplateRenderer
}

// TemplateRenderer defines an interface for rendering email templates.
type TemplateRenderer interface {
	Render(templateContent string, data map[string]string) (string, error)
}

// NewEmailService creates a new EmailService with multiple providers based on the application configuration.
func NewEmailService(cfg *config.EmailConfig) (Service, error) {
	if !cfg.Enable {
		log.Printf("INFO: Email service is globally disabled by configuration (EMAIL_ENABLE=false).")
		return nil, nil
	}

	providersMap := make(map[string]Provider)

	if cfg.SMTP.Enable {
		smtpProvider := NewSMTPEmailProvider(
			cfg.SMTP.Host,
			cfg.SMTP.Port,
			cfg.SMTP.Username,
			cfg.SMTP.Password,
			cfg.SMTP.FromAddress,
		)
		providersMap[smtpProvider.Name()] = smtpProvider
		log.Printf("INFO: SMTP Email Provider enabled and initialized.")
	}

	if len(providersMap) == 0 {
		return nil, fmt.Errorf("no email providers enabled in configuration")
	}

	var failoverOrder []string
	if cfg.DefaultProvider != "" {
		if _, ok := providersMap[cfg.DefaultProvider]; ok {
			failoverOrder = append(failoverOrder, cfg.DefaultProvider)
			log.Printf("INFO: Default email provider set to: %s", cfg.DefaultProvider)
		} else {
			log.Printf("WARN: Configured default email provider '%s' is not enabled or does not exist. Ignoring default provider setting.", cfg.DefaultProvider)
		}
	}

	if cfg.ProviderOrder != "" {
		orderedNames := strings.Split(cfg.ProviderOrder, ",")
		for _, name := range orderedNames {
			trimmedName := strings.TrimSpace(name)
			if trimmedName != "" {
				if _, ok := providersMap[trimmedName]; ok {
					isAlreadyDefault := false
					if len(failoverOrder) > 0 && failoverOrder[0] == trimmedName {
						isAlreadyDefault = true
					}
					if !isAlreadyDefault {
						failoverOrder = append(failoverOrder, trimmedName)
					}
				} else {
					log.Printf("WARN: Configured provider '%s' in PROVIDER_ORDER is not enabled or does not exist. Skipping.", trimmedName)
				}
			}
		}
	}

	// Add any remaining enabled providers to the end of the failover order if not already included
	for name := range providersMap {
		found := false
		for _, orderedName := range failoverOrder {
			if name == orderedName {
				found = true
				break
			}
		}
		if !found {
			failoverOrder = append(failoverOrder, name)
		}
	}

	if len(failoverOrder) == 0 {
		return nil, fmt.Errorf("no active email providers available after processing configuration")
	}

	return &ServiceImpl{
		defaultProviderName: cfg.DefaultProvider,
		providersMap:        providersMap,
		failoverOrder:       failoverOrder,
		cfg:                 cfg,
		templateRenderer:    &BasicTemplateRenderer{},
	}, nil
}

func (etr *BasicTemplateRenderer) Render(templateContent string, data map[string]string) (string, error) {
	renderedContent := templateContent
	for key, value := range data {
		renderedContent = strings.ReplaceAll(renderedContent, fmt.Sprintf("{{.%s}}", key), value)
	}
	return renderedContent, nil
}

// SendEmail attempts to send an email using the configured providers with a failover mechanism.
func (s *ServiceImpl) SendEmail(ctx context.Context, to, subject, body string) error {
	fromAddress := s.cfg.DefaultFromAddress

	if s == nil || s.cfg == nil || !s.cfg.Enable {
		log.Printf("WARN: Attempted to send email but email service is globally disabled or not initialized.")
		return fmt.Errorf("email service is disabled")
	}

	for _, providerName := range s.failoverOrder {
		provider, ok := s.providersMap[providerName]
		if !ok {
			log.Printf("ERROR: Attempted to use non-existent provider: %s", providerName)
			continue
		}

		providerFrom := provider.GetFromAddress()
		if providerFrom != "" {
			fromAddress = providerFrom
		}

		log.Printf("INFO: Attempting to send email to %s using %s provider (From: %s).", to, provider.Name(), fromAddress)

		sendCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		err := provider.SendEmail(sendCtx, fromAddress, to, subject, body)
		cancel()

		if err == nil {
			log.Printf("INFO: Email successfully sent to %s using %s provider.", to, provider.Name())
			return nil
		}
		log.Printf("ERROR: Failed to send email via %s: %v", provider.Name(), err)
	}

	return fmt.Errorf("all configured email providers failed to send email to %s", to)
}

// SendTemplatedEmail renders the template and then sends the email.
func (s *ServiceImpl) SendTemplatedEmail(ctx context.Context, to, templateContent, templateSubject string, templateData map[string]string) error {
	if s == nil || s.cfg == nil || !s.cfg.Enable {
		log.Printf("WARN: Attempted to send templated email but email service is globally disabled or not initialized.")
		return fmt.Errorf("email service is disabled")
	}

	log.Printf("INFO: Sending templated email to %s with subject: %s", to, templateSubject)

	renderedBody, err := s.templateRenderer.Render(templateContent, templateData)
	if err != nil {
		return fmt.Errorf("failed to render email template: %w", err)
	}

	return s.SendEmail(ctx, to, templateSubject, renderedBody)
}

func (s *ServiceImpl) HealthCheck(ctx context.Context) error {
	if s == nil || s.cfg == nil || !s.cfg.Enable {
		return fmt.Errorf("email service is disabled")
	}

	var lastErr error
	anyProviderHealthy := false

	for _, providerName := range s.failoverOrder {
		provider, ok := s.providersMap[providerName]
		if !ok {
			continue
		}

		if err := provider.HealthCheck(ctx); err != nil {
			log.Printf("WARN: Email provider %s health check failed: %v", provider.Name(), err)
			lastErr = err
		} else {
			anyProviderHealthy = true
		}
	}

	if !anyProviderHealthy {
		return fmt.Errorf("all email providers failed health check: %w", lastErr)
	}
	return nil
}

func (n *NoopService) HealthCheck(ctx context.Context) error {
	return nil // No-op service is always healthy
}

func (n *NoopService) SendEmail(ctx context.Context, to, subject, body string) error {
	log.Printf("DEBUG: SendEmail (no-op) called for %s", to)
	return nil
}

func (n *NoopService) SendTemplatedEmail(ctx context.Context, to, templateContent, templateSubject string, templateData map[string]string) error {
	log.Printf("DEBUG: SendTemplatedEmail (no-op) called for %s", to)
	return nil
}
