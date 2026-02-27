package blog

import (
	"fmt"
	"net/smtp"
	"strings"
)

// sanitizeHeader strips CR/LF to prevent email header injection.
func sanitizeHeader(s string) string {
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\n", "")
	return s
}

func (app *App) sendMail(to, subject, body string) error {
	if !app.cfg.smtpConfigured() {
		return fmt.Errorf("SMTP not configured")
	}

	to = sanitizeHeader(to)
	subject = sanitizeHeader(subject)

	msg := "From: " + app.cfg.FromEmail + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n" +
		"\r\n" + body

	addr := app.cfg.SMTPHost + ":" + app.cfg.SMTPPort
	auth := smtp.PlainAuth("", app.cfg.SMTPUsername, app.cfg.SMTPPassword, app.cfg.SMTPHost)
	return smtp.SendMail(addr, auth, app.cfg.FromEmail, []string{to}, []byte(msg))
}
