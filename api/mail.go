package main

import (
	"crypto/tls"
	"fmt"
	"log/slog"
	"mime"
	"net/smtp"
	"strings"
	"time"
)

// Mailer sends the login codes, the only emails Games Night sends. Without
// SMTP_HOST (local dev) it only logs what it would send.
type Mailer struct {
	cfg Config
	log *slog.Logger
}

func NewMailer(cfg Config, log *slog.Logger) *Mailer { return &Mailer{cfg: cfg, log: log} }

func (m *Mailer) SendCode(to, code string) error {
	if m.cfg.SMTPHost == "" {
		m.log.Info("DEV login code (SMTP not configured)", "email", to, "code", code)
		return nil
	}
	subject := "Seu código do Games Night: " + code
	body := fmt.Sprintf("Olá!\r\n\r\nSeu código para entrar no Games Night é:\r\n\r\n    %s\r\n\r\nEle vale por 10 minutos. Se não foi você, ignore este e-mail.\r\n", code)
	return m.Send(to, subject, body)
}

// Send delivers one plain-text email.
func (m *Mailer) Send(to, subject, body string) error {
	if m.cfg.SMTPHost == "" {
		m.log.Info("DEV email (SMTP not configured)", "to", to, "subject", subject)
		return nil
	}
	msg := strings.Join([]string{
		"From: Games Night <" + m.cfg.MailFrom + ">",
		"To: " + to,
		"Subject: " + mime.QEncoding.Encode("utf-8", subject),
		"Date: " + time.Now().Format(time.RFC1123Z),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n")
	auth := smtp.PlainAuth("", m.cfg.SMTPUser, m.cfg.SMTPPass, m.cfg.SMTPHost)
	addr := m.cfg.SMTPHost + ":" + m.cfg.SMTPPort
	if m.cfg.SMTPPort != "465" {
		// Port 587: STARTTLS, which SendMail negotiates.
		return smtp.SendMail(addr, auth, m.cfg.MailFrom, []string{to}, []byte(msg))
	}
	// Port 465: implicit TLS.
	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: m.cfg.SMTPHost})
	if err != nil {
		return err
	}
	c, err := smtp.NewClient(conn, m.cfg.SMTPHost)
	if err != nil {
		return err
	}
	defer c.Close()
	if err := c.Auth(auth); err != nil {
		return err
	}
	if err := c.Mail(m.cfg.MailFrom); err != nil {
		return err
	}
	if err := c.Rcpt(to); err != nil {
		return err
	}
	wc, err := c.Data()
	if err != nil {
		return err
	}
	if _, err := wc.Write([]byte(msg)); err != nil {
		return err
	}
	if err := wc.Close(); err != nil {
		return err
	}
	return c.Quit()
}
