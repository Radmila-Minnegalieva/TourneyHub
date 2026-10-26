package smtpclient

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"mime"
	"net"
	"net/smtp"
	"strings"
	"testing"
	"time"

	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/config"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/model"
)

func TestMailHeaderInjectionIsRejectedBeforeConnect(t *testing.T) {
	for _, tc := range []struct {
		name    string
		from    string
		to      string
		subject string
	}{
		{"sender", "sender@example.com\r\nBcc: victim@example.com", "player@example.com", "Welcome"},
		{"recipient", "sender@example.com", "player@example.com\r\nBcc: victim@example.com", "Welcome"},
		{"subject", "sender@example.com", "player@example.com", "Welcome\r\nBcc: victim@example.com"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := config.Default().SMTP
			cfg.From = tc.from
			cfg.Address = "invalid-address"
			err := New(cfg).Send(context.Background(), model.MailMessage{To: tc.to, Subject: tc.subject, Text: "Hello"})
			if err == nil || strings.Contains(err.Error(), "SMTP_ADDRESS") {
				t.Fatalf("header validation did not reject before connecting: %v", err)
			}
		})
	}
}

func TestSMTPRejectsInvalidSenderAndEndpoint(t *testing.T) {
	message := model.MailMessage{To: "recipient@example.com", Subject: "Welcome"}
	cfg := config.Default().SMTP
	cfg.Address, cfg.From = "127.0.0.1:25", "invalid-address"
	if err := New(cfg).Send(context.Background(), message); err == nil {
		t.Fatal("invalid SMTP sender accepted")
	}
	cfg.From, cfg.Address = "sender@example.com", "missing-port"
	if err := New(cfg).Send(context.Background(), message); err == nil {
		t.Fatal("invalid SMTP endpoint accepted")
	}
}

func TestSMTPRequiresExplicitInsecureConfiguration(t *testing.T) {
	// Окружение не должно менять уже созданный клиент.
	t.Setenv("SMTP_ALLOW_INSECURE", "true")
	for _, allow := range []bool{false, true} {
		connection, peer := net.Pipe()
		_ = connection.SetDeadline(time.Now().Add(2 * time.Second))
		_ = peer.SetDeadline(time.Now().Add(2 * time.Second))
		done := make(chan error, 1)
		go func() {
			defer peer.Close()
			if _, err := fmt.Fprint(peer, "220 relay.local ESMTP ready\r\n"); err != nil {
				done <- err
				return
			}
			command, err := bufio.NewReader(peer).ReadString('\n')
			if err != nil {
				done <- err
				return
			}
			if !strings.HasPrefix(command, "EHLO ") {
				done <- fmt.Errorf("unexpected command %q", command)
				return
			}
			_, err = fmt.Fprint(peer, "250 relay.local\r\n")
			done <- err
		}()
		client, err := smtp.NewClient(connection, "relay.local")
		if err != nil {
			t.Fatal(err)
		}
		cfg := config.Default().SMTP
		cfg.AllowInsecure = allow
		err = New(cfg).configure(client, "relay.local")
		_ = client.Close()
		if serverErr := <-done; serverErr != nil {
			t.Fatal(serverErr)
		}
		if allow && err != nil {
			t.Fatalf("explicit local insecure mode: %v", err)
		}
		if !allow && (err == nil || !strings.Contains(err.Error(), "STARTTLS")) {
			t.Fatalf("missing TLS was accepted: %v", err)
		}
	}
}

func TestMailEncodingAndLineEndings(t *testing.T) {
	var buffer bytes.Buffer
	message := model.MailMessage{To: "player@example.com", Subject: "Подтверждение почты", Text: "Первая\nВторая\r\nТретья"}
	if err := writeMessage(&buffer, "sender@example.com", message); err != nil {
		t.Fatal(err)
	}
	raw := buffer.String()
	if strings.Contains(strings.ReplaceAll(raw, "\r\n", ""), "\n") {
		t.Fatal("message contains bare LF")
	}
	for _, line := range strings.Split(raw, "\r\n") {
		if strings.HasPrefix(line, "Subject: ") {
			decoded, err := new(mime.WordDecoder).DecodeHeader(strings.TrimPrefix(line, "Subject: "))
			if err != nil || decoded != message.Subject {
				t.Fatalf("invalid RFC2047 subject: %q (%v)", decoded, err)
			}
			return
		}
	}
	t.Fatal("missing subject")
}

func TestCancellationInterruptsSMTPDialogue(t *testing.T) {
	connection, peer := net.Pipe()
	_ = connection.SetDeadline(time.Now().Add(3 * time.Second))
	_ = peer.SetDeadline(time.Now().Add(3 * time.Second))
	defer peer.Close()
	waitingForReply := make(chan struct{})
	serverDone := make(chan error, 1)
	go func() {
		if _, err := fmt.Fprint(peer, "220 relay.local ESMTP ready\r\n"); err != nil {
			serverDone <- err
			return
		}
		reader := bufio.NewReader(peer)
		if _, err := reader.ReadString('\n'); err != nil {
			serverDone <- err
			return
		}
		if _, err := fmt.Fprint(peer, "250 relay.local\r\n"); err != nil {
			serverDone <- err
			return
		}
		command, err := reader.ReadString('\n')
		if err != nil || !strings.HasPrefix(command, "MAIL FROM:") {
			serverDone <- fmt.Errorf("unexpected MAIL command %q: %v", command, err)
			return
		}
		close(waitingForReply)
		// Не отвечаем на MAIL: клиент должен закрыть соединение по отмене контекста.
		_, err = reader.ReadString('\n')
		if err == nil {
			err = fmt.Errorf("client continued SMTP after cancellation")
		} else {
			err = nil
		}
		serverDone <- err
	}()
	client, err := smtp.NewClient(connection, "relay.local")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sent := make(chan error, 1)
	go func() {
		sent <- New(config.Default().SMTP).send(ctx, client, "sender@example.com", model.MailMessage{To: "player@example.com", Subject: "Welcome"})
	}()
	select {
	case <-waitingForReply:
		cancel()
	case <-time.After(time.Second):
		t.Fatal("SMTP test did not reach MAIL dialogue")
	}
	select {
	case err := <-sent:
		if err == nil {
			t.Fatal("cancelled SMTP send unexpectedly succeeded")
		}
	case <-time.After(time.Second):
		t.Fatal("SMTP send ignored cancellation and waited for socket timeout")
	}
	if err := <-serverDone; err != nil {
		t.Fatal(err)
	}
}
