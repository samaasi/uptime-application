package email

import (
	"context"
	"fmt"

	"github.com/samaasi/uptime-application/services/api-services/internal/utils"
	"github.com/wneessen/go-mail"
	//"log"
)

// SMTPEmailProvider implements EmailProvider for SMTP
type SMTPEmailProvider struct {
	host     string
	port     int
	username string
	password string
	from     string
	client   *mail.Client
}

// NewSMTPEmailProvider creates a new SMTPEmailProvider.
func NewSMTPEmailProvider(host string, port int, username, password, from string) *SMTPEmailProvider {
	return &SMTPEmailProvider{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}
}

// SendEmail builds and dispatches an email using go-mail.
func (p *SMTPEmailProvider) SendEmail(ctx context.Context, from, to, subject, body string) error {
	msg := mail.NewMsg()
	if err := msg.From(from); err != nil {
		return fmt.Errorf("smtp provider: invalid From address: %w", err)
	}
	if err := msg.To(to); err != nil {
		return fmt.Errorf("smtp provider: invalid To address: %w", err)
	}
	msg.Subject(subject)
	msg.SetBodyString(mail.TypeTextPlain, body)

	opts := []mail.Option{
		mail.WithPort(p.port),
		mail.WithTLSPolicy(mail.TLSMandatory), // Production
		//mail.WithTLSPolicy(mail.TLSOpportunistic), // Development
	}

	if p.username != "" || p.password != "" {
		opts = append(opts, mail.WithUsername(p.username))
		opts = append(opts, mail.WithPassword(p.password))
		opts = append(opts, mail.WithSMTPAuth(mail.SMTPAuthLogin))
	}

	client, err := mail.NewClient(p.host, opts...)
	if err != nil {
		return fmt.Errorf("smtp provider: could not create client: %w", err)
	}

	if err := client.DialAndSend(msg); err != nil {
		if ctx.Err() != nil {
			// log.Printf("WARN: SMTP send operation cancelled or timed out for email to %s: %v", to, ctx.Err())
			return fmt.Errorf("smtp provider: send cancelled or timed out: %w", ctx.Err())
		}
		// log.Printf("ERROR: Failed to send email via SMTP to %s: %v", to, err)
		return fmt.Errorf("smtp provider: failed to send: %w", err)
	}

	// log.Printf("INFO: [SMTP] Sent email from %s to %s with subject \"%s\"", from, to, subject)
	// log.Printf("DEBUG: [SMTP] Body: %s", body)
	return nil
}

// Name returns the provider name
func (p *SMTPEmailProvider) Name() string {
	return "smtp"
}

// HealthCheck verifies SMTP connection
func (p *SMTPEmailProvider) HealthCheck(ctx context.Context) error {
	opts := []mail.Option{
		mail.WithPort(p.port),
		mail.WithTLSPolicy(mail.TLSMandatory),
	}

	if p.username != "" || p.password != "" {
		opts = append(opts, mail.WithUsername(p.username))
		opts = append(opts, mail.WithPassword(p.password))
		opts = append(opts, mail.WithSMTPAuth(mail.SMTPAuthLogin))
	}

	client, err := mail.NewClient(p.host, opts...)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer utils.CheckError(client.Close())

	if err := client.DialWithContext(ctx); err != nil {
		return fmt.Errorf("SMTP connection failed: %w", err)
	}

	return nil
}

// GetFromAddress returns the configured from email for SMTP.
func (p *SMTPEmailProvider) GetFromAddress() string {
	return p.from
}
