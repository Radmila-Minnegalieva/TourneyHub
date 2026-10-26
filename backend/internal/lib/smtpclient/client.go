// Package smtpclient доставляет письма, используя переданную типизированную конфигурацию.
package smtpclient

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
	"time"

	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/config"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/model"
)

type Client struct {
	cfg config.SMTP
}

func New(cfg config.SMTP) *Client { return &Client{cfg: cfg} }

// Send отправляет конверт и DATA. Ошибка оставляет повторную доставку вызывающему сервису.
func (c *Client) Send(ctx context.Context, message model.MailMessage) error {
	from, err := c.sender(message)
	if err != nil {
		return err
	}
	client, err := c.connect(ctx)
	if err != nil {
		return err
	}
	return c.send(ctx, client, from, message)
}

func (c *Client) send(ctx context.Context, client *smtp.Client, from string, message model.MailMessage) error {
	stopCancellation := context.AfterFunc(ctx, func() { _ = client.Close() })
	defer stopCancellation()
	defer client.Close()
	if err := client.Mail(from); err != nil {
		return err
	}
	if err := client.Rcpt(message.To); err != nil {
		return err
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	if err := writeMessage(w, from, message); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return client.Quit()
}

func writeMessage(w io.Writer, from string, message model.MailMessage) error {
	_, err := fmt.Fprintf(w, "From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s\r\n", from, message.To, mime.QEncoding.Encode("UTF-8", message.Subject), strings.ReplaceAll(strings.ReplaceAll(message.Text, "\r\n", "\n"), "\n", "\r\n"))
	return err
}

func (c *Client) sender(message model.MailMessage) (string, error) {
	from := c.cfg.From
	if _, err := mail.ParseAddress(from); err != nil {
		return "", fmt.Errorf("SMTP_FROM: %w", err)
	}
	if _, err := mail.ParseAddress(message.To); err != nil {
		return "", fmt.Errorf("invalid recipient: %w", err)
	}
	if strings.ContainsAny(from+message.To+message.Subject, "\r\n") {
		return "", fmt.Errorf("invalid mail header")
	}
	return from, nil
}

func (c *Client) connect(ctx context.Context) (*smtp.Client, error) {
	host, _, err := net.SplitHostPort(c.cfg.Address)
	if err != nil {
		return nil, fmt.Errorf("SMTP_ADDRESS: %w", err)
	}
	dialer := net.Dialer{Timeout: c.cfg.DialTimeout}
	conn, err := dialer.DialContext(ctx, "tcp", c.cfg.Address)
	if err != nil {
		return nil, err
	}
	// Отмена должна прерывать также ожидание приветствия, STARTTLS и AUTH.
	stopCancellation := context.AfterFunc(ctx, func() { _ = conn.Close() })
	defer stopCancellation()
	deadline := time.Now().Add(c.cfg.Timeout)
	if requestDeadline, ok := ctx.Deadline(); ok && requestDeadline.Before(deadline) {
		deadline = requestDeadline
	}
	if err := conn.SetDeadline(deadline); err != nil {
		conn.Close()
		return nil, err
	}
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		conn.Close()
		return nil, err
	}
	if err := c.configure(client, host); err != nil {
		client.Close()
		return nil, err
	}
	return client, nil
}

func (c *Client) configure(client *smtp.Client, host string) error {
	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(&tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}); err != nil {
			return err
		}
	} else if !c.cfg.AllowInsecure {
		return fmt.Errorf("SMTP requires STARTTLS; allow insecure mode only for a local relay")
	}
	if c.cfg.User != "" {
		return client.Auth(smtp.PlainAuth("", c.cfg.User, c.cfg.Password, host))
	}
	return nil
}
