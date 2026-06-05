// Package mailer ships verification codes by email. The interface keeps the
// auth handler free of SMTP wiring; pick the implementation in main.go based
// on whether SMTP_HOST is configured.
package mailer

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/smtp"
	"strings"
	"time"
)

// Mailer abstracts delivery of a verification code to an email address, so the
// auth handler depends on this interface rather than a concrete SMTP client.
type Mailer interface {
	SendCode(toEmail, code string) error
}

// ─── LogMailer ──────────────────────────────────────────────────────────────
// Used when SMTP creds are not configured. Writes the code to stdout so the
// flow keeps working in dev without external services.

// LogMailer is the dev/no-SMTP implementation of Mailer.
type LogMailer struct{}

// NewLogMailer returns a Mailer that only logs codes (no real delivery).
func NewLogMailer() *LogMailer { return &LogMailer{} }

// SendCode prints the verification code to the log instead of emailing it,
// letting the sign-up flow be exercised without SMTP credentials.
func (m *LogMailer) SendCode(toEmail, code string) error {
	log.Printf("[mailer:dev] verification code=%s for %s", code, toEmail)
	return nil
}

// ─── SMTPMailer ─────────────────────────────────────────────────────────────
// Plain-text RU email over SMTP. Supports both implicit-TLS port 465
// (Yandex.SMTP) and STARTTLS on port 587 (Gmail). Auth via app-password.

// SMTPMailer is the production Mailer that delivers codes over real SMTP.
type SMTPMailer struct {
	Host string
	Port string
	User string
	Pass string
	From string // e.g. "ClothesStore <store@yourdomain.ru>"
}

// NewSMTPMailer builds an SMTP-backed Mailer from connection/auth settings.
func NewSMTPMailer(host, port, user, pass, from string) *SMTPMailer {
	return &SMTPMailer{Host: host, Port: port, User: user, Pass: pass, From: from}
}

// SendCode composes the RU-language verification email and dispatches it,
// choosing the transport based on the configured port (see below).
func (m *SMTPMailer) SendCode(toEmail, code string) error {
	subject := "Код подтверждения ClothesStore"
	body := fmt.Sprintf(
		"Привет!\r\n\r\n"+
			"Твой код подтверждения: %s\r\n"+
			"Код действует 10 минут. Если ты не регистрировался — просто проигнорируй это письмо.\r\n\r\n"+
			"— ClothesStore",
		code,
	)
	msg := buildMessage(m.From, toEmail, subject, body)
	auth := smtp.PlainAuth("", m.User, m.Pass, m.Host)

	// Port selects the TLS strategy: 465 (and the empty default) means the
	// connection is TLS-encrypted from the start (implicit TLS); any other port
	// (e.g. 587) starts plaintext and upgrades via STARTTLS.
	switch m.Port {
	case "465", "":
		return m.sendImplicitTLS(auth, toEmail, msg)
	default:
		return m.sendStartTLS(auth, toEmail, msg)
	}
}

// sendImplicitTLS handles port-465 delivery: dial a TLS socket up front, then
// drive the SMTP conversation (AUTH, MAIL, RCPT, DATA) by hand because
// net/smtp's SendMail assumes a plaintext-then-STARTTLS connection.
func (m *SMTPMailer) sendImplicitTLS(auth smtp.Auth, to string, msg []byte) error {
	addr := net.JoinHostPort(m.Host, "465")
	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: m.Host})
	if err != nil {
		return fmt.Errorf("tls dial: %w", err)
	}
	c, err := smtp.NewClient(conn, m.Host)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer c.Close()
	if err := c.Auth(auth); err != nil {
		return fmt.Errorf("auth: %w", err)
	}
	from := senderAddr(m.From, m.User)
	if err := c.Mail(from); err != nil {
		return fmt.Errorf("mail: %w", err)
	}
	if err := c.Rcpt(to); err != nil {
		return fmt.Errorf("rcpt: %w", err)
	}
	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("data: %w", err)
	}
	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("write: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("close data: %w", err)
	}
	return c.Quit()
}

// sendStartTLS handles the non-465 case (e.g. port 587): net/smtp.SendMail
// connects in plaintext and upgrades to TLS via STARTTLS before authenticating.
func (m *SMTPMailer) sendStartTLS(auth smtp.Auth, to string, msg []byte) error {
	addr := net.JoinHostPort(m.Host, m.Port)
	from := senderAddr(m.From, m.User)
	return smtp.SendMail(addr, auth, from, []string{to}, msg)
}

// buildMessage assembles the raw RFC-822 message. The subject is RFC 2047
// encoded-word base64 so non-ASCII (Cyrillic) renders correctly in clients.
func buildMessage(from, to, subject, body string) []byte {
	headers := []string{
		"From: " + from,
		"To: " + to,
		"Subject: =?utf-8?B?" + b64UTF8(subject) + "?=",
		"MIME-Version: 1.0",
		`Content-Type: text/plain; charset="utf-8"`,
		"Content-Transfer-Encoding: 8bit",
		"Date: " + time.Now().UTC().Format(time.RFC1123Z),
	}
	return []byte(strings.Join(headers, "\r\n") + "\r\n\r\n" + body)
}

// senderAddr extracts the bare email used for SMTP MAIL FROM when the From
// header is "Name <addr>"; falls back to the configured username.
func senderAddr(from, fallback string) string {
	if i := strings.LastIndex(from, "<"); i >= 0 {
		if j := strings.Index(from[i:], ">"); j > 0 {
			return from[i+1 : i+j]
		}
	}
	if strings.Contains(from, "@") {
		return from
	}
	return fallback
}

// b64UTF8 returns the standard-base64 encoding of s. It's a hand-rolled
// implementation of the textbook 3-bytes-to-4-characters transform, used only
// to encode the email subject; see the inline note on why it avoids the stdlib.
func b64UTF8(s string) string {
	// avoid importing encoding/base64 just for one tiny helper
	const tab = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	b := []byte(s)
	out := make([]byte, 0, ((len(b)+2)/3)*4)
	// Process the input three bytes at a time; n is the count in this group
	// (1, 2 or 3) so the tail can be padded with '=' as base64 requires.
	for i := 0; i < len(b); i += 3 {
		n := len(b) - i
		if n > 3 {
			n = 3
		}
		var c1, c2, c3 byte
		c1 = b[i]
		if n > 1 {
			c2 = b[i+1]
		}
		if n > 2 {
			c3 = b[i+2]
		}
		// Re-pack the 24 input bits into four 6-bit indices into tab. The 3rd
		// and 4th characters become '=' padding when fewer than 3 bytes remain.
		out = append(out,
			tab[c1>>2],
			tab[((c1&0x03)<<4)|(c2>>4)],
		)
		if n > 1 {
			out = append(out, tab[((c2&0x0f)<<2)|(c3>>6)])
		} else {
			out = append(out, '=')
		}
		if n > 2 {
			out = append(out, tab[c3&0x3f])
		} else {
			out = append(out, '=')
		}
	}
	return string(out)
}
