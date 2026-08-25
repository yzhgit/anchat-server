package sender

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"mime"
	"mime/quotedprintable"
	"net"
	"net/mail"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	"flamingo/pkg/config"
)

type smtpEmailSender struct {
	conf config.Email
}

// EmailSender abstracts sending an email to a recipient.
type EmailSender interface {
	Send(to, subject, content string) error
}

var _ EmailSender = (*smtpEmailSender)(nil)

func NewEmailSender(conf config.Email) (EmailSender, error) {
	conf.Host = strings.TrimSpace(conf.Host)
	conf.Username = strings.TrimSpace(conf.Username)
	conf.FromName = strings.TrimSpace(conf.FromName)
	conf.FromAddress = strings.TrimSpace(conf.FromAddress)

	if conf.Host == "" {
		return nil, fmt.Errorf("smtp host is required")
	}
	if conf.Port == 0 {
		conf.Port = 587
	}
	if conf.Port < 1 || conf.Port > 65535 {
		return nil, fmt.Errorf("smtp port is invalid: %d", conf.Port)
	}
	if conf.FromAddress == "" {
		return nil, fmt.Errorf("smtp from address is required")
	}
	if _, err := mail.ParseAddress(conf.FromAddress); err != nil {
		return nil, fmt.Errorf("smtp from address is invalid: %w", err)
	}

	return &smtpEmailSender{conf: conf}, nil
}

func (s *smtpEmailSender) Send(to, subject, content string) error {
	to = strings.TrimSpace(to)
	if to == "" {
		return fmt.Errorf("smtp recipient address is required")
	}
	if _, err := mail.ParseAddress(to); err != nil {
		return fmt.Errorf("smtp recipient address is invalid: %w", err)
	}

	message, err := s.buildMessage(to, subject, content)
	if err != nil {
		return err
	}

	addr := net.JoinHostPort(s.conf.Host, strconv.Itoa(s.conf.Port))
	if s.conf.Port == 465 {
		return s.sendWithTLS(addr, to, message)
	}

	if err := smtp.SendMail(addr, s.auth(), s.conf.FromAddress, []string{to}, message); err != nil {
		return fmt.Errorf("smtp send mail failed: %w", err)
	}
	return nil
}

func (s *smtpEmailSender) auth() smtp.Auth {
	if s.conf.Username == "" {
		return nil
	}
	return smtp.PlainAuth("", s.conf.Username, s.conf.Password, s.conf.Host)
}

func (s *smtpEmailSender) buildMessage(to, subject, content string) ([]byte, error) {
	var body bytes.Buffer
	qp := quotedprintable.NewWriter(&body)
	if _, err := qp.Write([]byte(content)); err != nil {
		return nil, fmt.Errorf("encode smtp body failed: %w", err)
	}
	if err := qp.Close(); err != nil {
		return nil, fmt.Errorf("finalize smtp body failed: %w", err)
	}

	fromHeader := s.conf.FromAddress
	if s.conf.FromName != "" {
		fromHeader = (&mail.Address{
			Name:    s.conf.FromName,
			Address: s.conf.FromAddress,
		}).String()
	}

	headers := []string{
		fmt.Sprintf("From: %s", fromHeader),
		fmt.Sprintf("To: %s", to),
		fmt.Sprintf("Subject: %s", mime.QEncoding.Encode("UTF-8", subject)),
		fmt.Sprintf("Date: %s", time.Now().Format(time.RFC1123Z)),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"Content-Transfer-Encoding: quoted-printable",
		"",
		body.String(),
	}

	return []byte(strings.Join(headers, "\r\n")), nil
}

func (s *smtpEmailSender) sendWithTLS(addr, to string, message []byte) error {
	conn, err := tls.Dial("tcp", addr, &tls.Config{
		ServerName: s.conf.Host,
		MinVersion: tls.VersionTLS12,
	})
	if err != nil {
		return fmt.Errorf("smtp tls dial failed: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.conf.Host)
	if err != nil {
		return fmt.Errorf("smtp client init failed: %w", err)
	}
	defer client.Close()

	if auth := s.auth(); auth != nil {
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth failed: %w", err)
		}
	}
	if err := client.Mail(s.conf.FromAddress); err != nil {
		return fmt.Errorf("smtp MAIL FROM failed: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("smtp RCPT TO failed: %w", err)
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp DATA failed: %w", err)
	}
	if _, err := writer.Write(message); err != nil {
		_ = writer.Close()
		return fmt.Errorf("smtp write message failed: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("smtp finalize message failed: %w", err)
	}
	if err := client.Quit(); err != nil {
		return fmt.Errorf("smtp quit failed: %w", err)
	}
	return nil
}
