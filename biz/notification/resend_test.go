package notification

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestResendNotifierSendEmail(t *testing.T) {
	var payload map[string]any
	var idempotency string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer re_test" {
			t.Fatalf("authorization header = %q, want bearer key", r.Header.Get("Authorization"))
		}
		if r.Header.Get("User-Agent") == "" {
			t.Fatal("missing User-Agent header")
		}
		idempotency = r.Header.Get(idempotencyKeyHeader)
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"email_123"}`))
	}))
	defer server.Close()

	notifier, err := NewResendNotifier(ResendConfig{
		APIKey:   "re_test",
		From:     "Acme <onboarding@example.com>",
		Endpoint: server.URL,
	})
	if err != nil {
		t.Fatalf("NewResendNotifier() error = %v", err)
	}
	message := testMessage(ChannelEmail)
	message.ID = "msg-1"
	message.Body = "welcome"
	message.Metadata = map[string]string{"trace.id": "internal-only"}
	message.Tags = map[string]string{"tenant": "tenant-a"}

	result, err := notifier.SendEmail(context.Background(), message)
	if err != nil {
		t.Fatalf("SendEmail() error = %v", err)
	}
	if result.ID != "email_123" {
		t.Fatalf("SendEmail() ID = %q, want email_123", result.ID)
	}
	if idempotency != "msg-1" {
		t.Fatalf("idempotency = %q, want msg-1", idempotency)
	}
	if payload["from"] != "Acme <onboarding@example.com>" || payload["subject"] != "Hi" || payload["text"] != "welcome" {
		t.Fatalf("payload = %#v, want from subject text", payload)
	}
	if to, ok := payload["to"].([]any); !ok || len(to) != 1 || to[0] != "user@example.com" {
		t.Fatalf("payload to = %#v, want single recipient", payload["to"])
	}
	if tags, ok := payload["tags"].([]any); !ok || len(tags) != 1 {
		t.Fatalf("payload tags = %#v, want one tag", payload["tags"])
	}
}

func TestResendNotifierDoesNotMapMetadataToTags(t *testing.T) {
	var payload map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"email_123"}`))
	}))
	defer server.Close()

	notifier, err := NewResendNotifier(ResendConfig{
		APIKey:   "re_test",
		From:     "noreply@example.com",
		Endpoint: server.URL,
	})
	if err != nil {
		t.Fatalf("NewResendNotifier() error = %v", err)
	}
	message := testMessage(ChannelEmail)
	message.Body = "welcome"
	message.Metadata = map[string]string{"trace.id": "internal-only"}

	if _, err := notifier.SendEmail(context.Background(), message); err != nil {
		t.Fatalf("SendEmail() error = %v", err)
	}
	if _, ok := payload["tags"]; ok {
		t.Fatalf("payload tags = %#v, want absent for metadata-only message", payload["tags"])
	}
}

func TestResendNotifierHTMLBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		if payload["html"] != "<p>welcome</p>" {
			t.Fatalf("payload html = %#v, want html body", payload["html"])
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"email_123"}`))
	}))
	defer server.Close()

	notifier, err := NewResendNotifier(ResendConfig{
		APIKey:     "re_test",
		From:       "noreply@example.com",
		Endpoint:   server.URL,
		BodyFormat: ResendBodyHTML,
	})
	if err != nil {
		t.Fatalf("NewResendNotifier() error = %v", err)
	}
	message := testMessage(ChannelEmail)
	message.Body = "<p>welcome</p>"
	if err := notifier.Send(context.Background(), message); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
}

func TestResendNotifierValidation(t *testing.T) {
	if _, err := NewResendNotifier(ResendConfig{}); !errors.Is(err, ErrInvalidResendConfig) {
		t.Fatalf("NewResendNotifier(empty) error = %v, want ErrInvalidResendConfig", err)
	}
	if _, err := NewResendNotifier(ResendConfig{APIKey: "re_test", From: "noreply@example.com", Endpoint: "http://example.com/emails"}); !errors.Is(err, ErrInvalidResendConfig) {
		t.Fatalf("NewResendNotifier(insecure) error = %v, want ErrInvalidResendConfig", err)
	}
	if _, err := NewResendNotifier(ResendConfig{APIKey: "re_test", From: "noreply@example.com", Endpoint: "http://example.com/emails", AllowInsecureHTTP: true}); err != nil {
		t.Fatalf("NewResendNotifier(allow insecure) error = %v", err)
	}
	if _, err := NewResendNotifier(ResendConfig{APIKey: "re_test", From: "noreply@example.com", Endpoint: "https://user:pass@example.com/emails"}); !errors.Is(err, ErrInvalidResendConfig) {
		t.Fatalf("NewResendNotifier(userinfo) error = %v, want ErrInvalidResendConfig", err)
	}
	if err := (*ResendNotifier)(nil).Send(context.Background(), testMessage(ChannelEmail)); !errors.Is(err, ErrNilNotifier) {
		t.Fatalf("nil Send() error = %v, want ErrNilNotifier", err)
	}
}

func TestResendNotifierMessageValidation(t *testing.T) {
	notifier, err := NewResendNotifier(ResendConfig{
		APIKey:   "re_test",
		From:     "noreply@example.com",
		Endpoint: "http://127.0.0.1/emails",
	})
	if err != nil {
		t.Fatalf("NewResendNotifier() error = %v", err)
	}
	message := testMessage("sms")
	message.Body = "body"
	if _, err := notifier.SendEmail(context.Background(), message); !errors.Is(err, ErrUnsupportedChannel) {
		t.Fatalf("SendEmail(wrong channel) error = %v, want ErrUnsupportedChannel", err)
	}
	message = testMessage(ChannelEmail)
	message.Body = ""
	if _, err := notifier.SendEmail(context.Background(), message); !errors.Is(err, ErrInvalidMessage) {
		t.Fatalf("SendEmail(empty body) error = %v, want ErrInvalidMessage", err)
	}
	message.Body = "body"
	message.Tags = map[string]string{"bad key": "value"}
	if _, err := notifier.SendEmail(context.Background(), message); !errors.Is(err, ErrInvalidMessage) {
		t.Fatalf("SendEmail(bad tags) error = %v, want ErrInvalidMessage", err)
	}
	message.Tags = nil
	message.ID = strings.Repeat("x", defaultResendIDMax+1)
	if _, err := notifier.SendEmail(context.Background(), message); !errors.Is(err, ErrInvalidMessage) {
		t.Fatalf("SendEmail(long idempotency key) error = %v, want ErrInvalidMessage", err)
	}
	message.ID = ""
	message.To = strings.Join(makeRecipients(defaultResendRecipientMax+1), ", ")
	if _, err := notifier.SendEmail(context.Background(), message); !errors.Is(err, ErrInvalidMessage) {
		t.Fatalf("SendEmail(too many recipients) error = %v, want ErrInvalidMessage", err)
	}
}

func TestResendNotifierRejectsEmptyProviderID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":""}`))
	}))
	defer server.Close()

	notifier, err := NewResendNotifier(ResendConfig{APIKey: "re_test", From: "noreply@example.com", Endpoint: server.URL})
	if err != nil {
		t.Fatalf("NewResendNotifier() error = %v", err)
	}
	message := testMessage(ChannelEmail)
	message.Body = "body"

	_, err = notifier.SendEmail(context.Background(), message)
	if !errors.Is(err, ErrResendDelivery) {
		t.Fatalf("SendEmail(empty id) error = %v, want ErrResendDelivery", err)
	}
	if !DefaultRetryIf(err) {
		t.Fatalf("DefaultRetryIf(empty id) = false, want true")
	}
}

func TestResendNotifierStatusError(t *testing.T) {
	for _, tt := range []struct {
		status    int
		retryable bool
	}{
		{status: http.StatusBadRequest, retryable: false},
		{status: http.StatusTooManyRequests, retryable: true},
		{status: http.StatusInternalServerError, retryable: true},
	} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "resend says no", tt.status)
		}))
		notifier, err := NewResendNotifier(ResendConfig{APIKey: "re_test", From: "noreply@example.com", Endpoint: server.URL})
		if err != nil {
			server.Close()
			t.Fatalf("NewResendNotifier() error = %v", err)
		}
		message := testMessage(ChannelEmail)
		message.Body = "body"
		_, err = notifier.SendEmail(context.Background(), message)
		server.Close()
		if !errors.Is(err, ErrResendDelivery) {
			t.Fatalf("SendEmail(status %d) error = %v, want ErrResendDelivery", tt.status, err)
		}
		var statusErr *ResendStatusError
		if !errors.As(err, &statusErr) {
			t.Fatalf("SendEmail(status %d) error = %v, want ResendStatusError", tt.status, err)
		}
		if statusErr.StatusCode != tt.status || statusErr.Retryable() != tt.retryable || DefaultRetryIf(err) != tt.retryable {
			t.Fatalf("status error = %+v retryable=%v default=%v, want status %d retryable %v", statusErr, statusErr.Retryable(), DefaultRetryIf(err), tt.status, tt.retryable)
		}
		if statusErr.Body == "" {
			t.Fatalf("status error body empty, want provider body captured")
		}
	}
}

func makeRecipients(count int) []string {
	recipients := make([]string, count)
	for i := range recipients {
		recipients[i] = "user" + string(rune('a'+i%26)) + "@example.com"
	}
	return recipients
}
