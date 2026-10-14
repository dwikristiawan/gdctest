package infra

import (
	"context"
	"fmt"
	"net/smtp"

	"github.com/kelseyhightower/envconfig"
)

type MailerConfig struct {
	Host string `envconfig:"SMTP_HOST" default:"localhost"`
	Port int    `envconfig:"SMTP_PORT" default:"1025"`
	From string `envconfig:"SMTP_FROM" default:"no-reply@gdctest.local"`
}

type Mailer struct {
	cfg MailerConfig
}

func (m *Mailer) SendEmail(to, subject, bodyHTML string) error {
	addr := fmt.Sprintf("%s:%d", m.cfg.Host, m.cfg.Port)
	msg := []byte(fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n\r\n"+
		"%s\r\n", m.cfg.From, to, subject, bodyHTML))

	return smtp.SendMail(addr, nil, m.cfg.From, []string{to}, msg)
}

func GetMailer(c *Container) (*Mailer, bool) {
	if c == nil {
		return nil, false
	}
	if res, ok := c.Get("mailer"); ok {
		if m, ok := res.(*Mailer); ok {
			return m, true
		}
	}
	return nil, false
}

func init() {
	Register("mailer", func(ctx context.Context, c *Container) error {
		var cfg MailerConfig
		if err := envconfig.Process("", &cfg); err != nil {
			return fmt.Errorf("failed to parse mailer config: %w", err)
		}
		c.Set("mailer", &Mailer{cfg: cfg})
		return nil
	})
}
