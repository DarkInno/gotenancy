package notification

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestWebhookNotifierSend(t *testing.T) {
	now := time.Unix(1700000000, 0).UTC()
	secret := []byte("secret")
	var got WebhookPayload
	var gotSignature string
	var gotTimestamp string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" || r.Header.Get("Accept") != "application/json" {
			t.Fatalf("headers content-type=%q accept=%q, want application/json", r.Header.Get("Content-Type"), r.Header.Get("Accept"))
		}
		if r.Header.Get(webhookTenantHeader) != "tenant-a" || r.Header.Get(webhookChannelHeader) != ChannelWebhook {
			t.Fatalf("tenant/channel headers = %q/%q", r.Header.Get(webhookTenantHeader), r.Header.Get(webhookChannelHeader))
		}
		if r.Header.Get(idempotencyKeyHeader) != "msg-1" {
			t.Fatalf("idempotency header = %q, want msg-1", r.Header.Get(idempotencyKeyHeader))
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("authorization header = %q, want bearer token", r.Header.Get("Authorization"))
		}
		gotSignature = r.Header.Get(WebhookSignatureHeader)
		gotTimestamp = r.Header.Get(WebhookTimestampHeader)
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	notifier, err := NewWebhookNotifier(WebhookConfig{
		Endpoint: server.URL,
		Headers:  map[string]string{"Authorization": "Bearer token"},
		Secret:   secret,
		Now:      func() time.Time { return now },
	})
	if err != nil {
		t.Fatalf("NewWebhookNotifier() error = %v", err)
	}
	message := testMessage(ChannelWebhook)
	message.ID = "msg-1"
	message.Body = "hello"
	message.Metadata = map[string]string{"kind": "welcome"}

	if err := notifier.Send(context.Background(), message); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if got.ID != "msg-1" || got.TenantID != "tenant-a" || got.Channel != ChannelWebhook || got.Metadata["kind"] != "welcome" {
		t.Fatalf("payload = %+v, want tenant webhook payload", got)
	}
	raw, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if gotTimestamp != "1700000000" || gotSignature != signWebhook(gotTimestamp, raw, secret) {
		t.Fatalf("signature timestamp=%q signature=%q, want fixed HMAC", gotTimestamp, gotSignature)
	}
}

func TestWebhookNotifierValidation(t *testing.T) {
	if _, err := NewWebhookNotifier(WebhookConfig{}); !errors.Is(err, ErrInvalidWebhookConfig) {
		t.Fatalf("NewWebhookNotifier(empty) error = %v, want ErrInvalidWebhookConfig", err)
	}
	if _, err := NewWebhookNotifier(WebhookConfig{Endpoint: "http://example.com/hook"}); !errors.Is(err, ErrInvalidWebhookConfig) {
		t.Fatalf("NewWebhookNotifier(insecure) error = %v, want ErrInvalidWebhookConfig", err)
	}
	if _, err := NewWebhookNotifier(WebhookConfig{Endpoint: "http://example.com/hook", AllowInsecureHTTP: true}); err != nil {
		t.Fatalf("NewWebhookNotifier(allow insecure) error = %v", err)
	}
	if _, err := NewWebhookNotifier(WebhookConfig{Endpoint: "https://user:pass@example.com/hook"}); !errors.Is(err, ErrInvalidWebhookConfig) {
		t.Fatalf("NewWebhookNotifier(userinfo) error = %v, want ErrInvalidWebhookConfig", err)
	}
	if _, err := NewWebhookNotifier(WebhookConfig{Endpoint: "https://example.com/hook", Headers: map[string]string{"Bad\r\nHeader": "value"}}); !errors.Is(err, ErrInvalidWebhookConfig) {
		t.Fatalf("NewWebhookNotifier(bad header) error = %v, want ErrInvalidWebhookConfig", err)
	}
	if err := (*WebhookNotifier)(nil).Send(context.Background(), testMessage(ChannelWebhook)); !errors.Is(err, ErrNilNotifier) {
		t.Fatalf("nil Send() error = %v, want ErrNilNotifier", err)
	}
}

func TestWebhookNotifierStatusError(t *testing.T) {
	for _, tt := range []struct {
		status    int
		retryable bool
	}{
		{status: http.StatusBadRequest, retryable: false},
		{status: http.StatusTooManyRequests, retryable: true},
		{status: http.StatusInternalServerError, retryable: true},
	} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "provider says no", tt.status)
		}))
		notifier, err := NewWebhookNotifier(WebhookConfig{Endpoint: server.URL})
		if err != nil {
			server.Close()
			t.Fatalf("NewWebhookNotifier() error = %v", err)
		}
		err = notifier.Send(context.Background(), testMessage(ChannelWebhook))
		server.Close()
		if !errors.Is(err, ErrWebhookDelivery) {
			t.Fatalf("Send(status %d) error = %v, want ErrWebhookDelivery", tt.status, err)
		}
		var statusErr *WebhookStatusError
		if !errors.As(err, &statusErr) {
			t.Fatalf("Send(status %d) error = %v, want WebhookStatusError", tt.status, err)
		}
		if statusErr.StatusCode != tt.status || statusErr.Retryable() != tt.retryable || DefaultRetryIf(err) != tt.retryable {
			t.Fatalf("status error = %+v retryable=%v default=%v, want status %d retryable %v", statusErr, statusErr.Retryable(), DefaultRetryIf(err), tt.status, tt.retryable)
		}
		if statusErr.Body == "" {
			t.Fatalf("status error body empty, want provider body captured")
		}
	}
}

func TestWebhookNotifierRejectsWrongChannel(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	defer server.Close()

	notifier, err := NewWebhookNotifier(WebhookConfig{Endpoint: server.URL, Channel: "sms"})
	if err != nil {
		t.Fatalf("NewWebhookNotifier() error = %v", err)
	}
	if err := notifier.Send(context.Background(), testMessage(ChannelWebhook)); !errors.Is(err, ErrUnsupportedChannel) {
		t.Fatalf("Send(wrong channel) error = %v, want ErrUnsupportedChannel", err)
	}
}
