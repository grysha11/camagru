package mailer

import (
	"fmt"
	"net/smtp"
)

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	From     string
}

type Mailer struct {
	cfg Config
}

func New(cfg Config) *Mailer {
	return &Mailer{cfg: cfg}
}

func (m *Mailer) Send(to, subject, body string) error {
	addr := m.cfg.Host + ":" + m.cfg.Port

	var auth smtp.Auth
	if m.cfg.User != "" {
		auth = smtp.PlainAuth("", m.cfg.User, m.cfg.Password, m.cfg.Host)
	}

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		m.cfg.From, to, subject, body)

	return smtp.SendMail(addr, auth, m.cfg.From, []string{to}, []byte(msg))
}
