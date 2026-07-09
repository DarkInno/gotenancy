package notification

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestMessageValidateAndClone(t *testing.T) {
	message := Message{
		TenantID: "tenant-a",
		Channel:  "email",
		To:       "user@example.com",
		Metadata: map[string]string{"kind": "welcome"},
		Tags:     map[string]string{"tenant": "tenant-a"},
	}
	if err := message.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if err := (Message{}).Validate(); !errors.Is(err, ErrInvalidMessage) {
		t.Fatalf("Validate(empty) error = %v, want ErrInvalidMessage", err)
	}
	message.ID = "msg-1\r\nX-Bad: yes"
	if err := message.Validate(); !errors.Is(err, ErrInvalidMessage) {
		t.Fatalf("Validate(injected id) error = %v, want ErrInvalidMessage", err)
	}

	message.ID = ""
	cloned := message.Clone()
	cloned.Metadata["kind"] = "changed"
	cloned.Tags["tenant"] = "changed"
	if message.Metadata["kind"] != "welcome" {
		t.Fatal("Clone() returned mutable metadata")
	}
	if message.Tags["tenant"] != "tenant-a" {
		t.Fatal("Clone() returned mutable tags")
	}
}

func TestNotifierFunc(t *testing.T) {
	called := false
	fn := NotifierFunc(func(ctx context.Context, message Message) error {
		called = true
		return message.Validate()
	})
	if err := fn.Send(context.Background(), testMessage(ChannelEmail)); err != nil {
		t.Fatalf("NotifierFunc.Send() error = %v", err)
	}
	if !called {
		t.Fatal("NotifierFunc.Send() did not call function")
	}
	if err := (NotifierFunc)(nil).Send(context.Background(), testMessage(ChannelEmail)); !errors.Is(err, ErrNilNotifier) {
		t.Fatalf("nil NotifierFunc.Send() error = %v, want ErrNilNotifier", err)
	}
}

func TestChannelRouter(t *testing.T) {
	email := NewMemoryNotifier()
	sms := NewMemoryNotifier()
	router, err := NewChannelRouter(map[string]Notifier{
		" email ": email,
		"sms":     sms,
	})
	if err != nil {
		t.Fatalf("NewChannelRouter() error = %v", err)
	}

	if err := router.Send(context.Background(), testMessage(ChannelEmail)); err != nil {
		t.Fatalf("Send(email) error = %v", err)
	}
	if len(email.Messages()) != 1 || len(sms.Messages()) != 0 {
		t.Fatalf("router messages email=%d sms=%d, want 1/0", len(email.Messages()), len(sms.Messages()))
	}

	routes := map[string]Notifier{"email": email}
	copiedRouter, err := NewChannelRouter(routes)
	if err != nil {
		t.Fatalf("NewChannelRouter(copy) error = %v", err)
	}
	routes["email"] = sms
	if err := copiedRouter.Send(context.Background(), testMessage(ChannelEmail)); err != nil {
		t.Fatalf("copied router Send() error = %v", err)
	}
	if len(email.Messages()) != 2 || len(sms.Messages()) != 0 {
		t.Fatalf("copied router messages email=%d sms=%d, want 2/0", len(email.Messages()), len(sms.Messages()))
	}
	if err := router.Send(context.Background(), testMessage("push")); !errors.Is(err, ErrUnsupportedChannel) {
		t.Fatalf("Send(push) error = %v, want ErrUnsupportedChannel", err)
	}
}

func TestChannelRouterValidation(t *testing.T) {
	if _, err := NewChannelRouter(nil); !errors.Is(err, ErrInvalidNotifierConfig) {
		t.Fatalf("NewChannelRouter(nil) error = %v, want ErrInvalidNotifierConfig", err)
	}
	if _, err := NewChannelRouter(map[string]Notifier{"": NewMemoryNotifier()}); !errors.Is(err, ErrInvalidNotifierConfig) {
		t.Fatalf("NewChannelRouter(empty channel) error = %v, want ErrInvalidNotifierConfig", err)
	}
	if _, err := NewChannelRouter(map[string]Notifier{"email": nil}); !errors.Is(err, ErrInvalidNotifierConfig) {
		t.Fatalf("NewChannelRouter(nil notifier) error = %v, want ErrInvalidNotifierConfig", err)
	}
	if err := (*ChannelRouter)(nil).Send(context.Background(), testMessage(ChannelEmail)); !errors.Is(err, ErrNilNotifier) {
		t.Fatalf("nil router Send() error = %v, want ErrNilNotifier", err)
	}
}

func TestFanoutNotifier(t *testing.T) {
	first := NewMemoryNotifier()
	second := NewMemoryNotifier()
	fanout, err := NewFanoutNotifier(first, second)
	if err != nil {
		t.Fatalf("NewFanoutNotifier() error = %v", err)
	}

	if err := fanout.Send(context.Background(), testMessage(ChannelEmail)); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if len(first.Messages()) != 1 || len(second.Messages()) != 1 {
		t.Fatalf("fanout messages first=%d second=%d, want 1/1", len(first.Messages()), len(second.Messages()))
	}
}

func TestFanoutNotifierJoinsErrors(t *testing.T) {
	errFirst := errors.New("first failed")
	errSecond := errors.New("second failed")
	fanout, err := NewFanoutNotifier(
		NotifierFunc(func(context.Context, Message) error { return errFirst }),
		NotifierFunc(func(context.Context, Message) error { return errSecond }),
	)
	if err != nil {
		t.Fatalf("NewFanoutNotifier() error = %v", err)
	}

	err = fanout.Send(context.Background(), testMessage(ChannelEmail))
	if !errors.Is(err, errFirst) || !errors.Is(err, errSecond) {
		t.Fatalf("Send() error = %v, want joined first and second errors", err)
	}
}

func TestFanoutNotifierValidation(t *testing.T) {
	if _, err := NewFanoutNotifier(); !errors.Is(err, ErrInvalidNotifierConfig) {
		t.Fatalf("NewFanoutNotifier(empty) error = %v, want ErrInvalidNotifierConfig", err)
	}
	if _, err := NewFanoutNotifier(nil); !errors.Is(err, ErrInvalidNotifierConfig) {
		t.Fatalf("NewFanoutNotifier(nil) error = %v, want ErrInvalidNotifierConfig", err)
	}
	if err := (*FanoutNotifier)(nil).Send(context.Background(), testMessage(ChannelEmail)); !errors.Is(err, ErrNilNotifier) {
		t.Fatalf("nil fanout Send() error = %v, want ErrNilNotifier", err)
	}
}

func TestRetryNotifierRetriesTransientError(t *testing.T) {
	transient := errors.New("transient")
	attempts := 0
	var delays []time.Duration
	retry, err := NewRetryNotifier(
		NotifierFunc(func(context.Context, Message) error {
			attempts++
			if attempts < 3 {
				return transient
			}
			return nil
		}),
		RetryConfig{
			Attempts:   3,
			Backoff:    10 * time.Millisecond,
			MaxBackoff: 30 * time.Millisecond,
			Sleep: func(ctx context.Context, delay time.Duration) error {
				delays = append(delays, delay)
				return ctx.Err()
			},
		},
	)
	if err != nil {
		t.Fatalf("NewRetryNotifier() error = %v", err)
	}

	if err := retry.Send(context.Background(), testMessage(ChannelEmail)); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if attempts != 3 {
		t.Fatalf("attempts = %d, want 3", attempts)
	}
	if len(delays) != 2 || delays[0] != 10*time.Millisecond || delays[1] != 20*time.Millisecond {
		t.Fatalf("delays = %v, want 10ms and 20ms", delays)
	}
}

func TestRetryNotifierDoesNotRetryPermanentError(t *testing.T) {
	attempts := 0
	retry, err := NewRetryNotifier(
		NotifierFunc(func(context.Context, Message) error {
			attempts++
			return ErrUnsupportedChannel
		}),
		RetryConfig{Attempts: 3, Sleep: func(context.Context, time.Duration) error { return nil }},
	)
	if err != nil {
		t.Fatalf("NewRetryNotifier() error = %v", err)
	}

	err = retry.Send(context.Background(), testMessage(ChannelEmail))
	if !errors.Is(err, ErrUnsupportedChannel) {
		t.Fatalf("Send() error = %v, want ErrUnsupportedChannel", err)
	}
	if attempts != 1 {
		t.Fatalf("attempts = %d, want 1", attempts)
	}
}

func TestRetryNotifierStopsWhenContextCanceled(t *testing.T) {
	transient := errors.New("transient")
	ctx, cancel := context.WithCancel(context.Background())
	retry, err := NewRetryNotifier(
		NotifierFunc(func(context.Context, Message) error { return transient }),
		RetryConfig{
			Attempts: 2,
			Sleep: func(context.Context, time.Duration) error {
				cancel()
				return context.Canceled
			},
		},
	)
	if err != nil {
		t.Fatalf("NewRetryNotifier() error = %v", err)
	}

	err = retry.Send(ctx, testMessage(ChannelEmail))
	if !errors.Is(err, transient) || !errors.Is(err, context.Canceled) {
		t.Fatalf("Send() error = %v, want transient and context canceled", err)
	}
}

func TestRetryNotifierValidation(t *testing.T) {
	if _, err := NewRetryNotifier(nil, RetryConfig{}); !errors.Is(err, ErrNilNotifier) {
		t.Fatalf("NewRetryNotifier(nil) error = %v, want ErrNilNotifier", err)
	}
	if _, err := NewRetryNotifier(NewMemoryNotifier(), RetryConfig{Attempts: -1}); !errors.Is(err, ErrInvalidNotifierConfig) {
		t.Fatalf("NewRetryNotifier(bad attempts) error = %v, want ErrInvalidNotifierConfig", err)
	}
	if _, err := NewRetryNotifier(NewMemoryNotifier(), RetryConfig{Backoff: time.Second, MaxBackoff: time.Millisecond}); !errors.Is(err, ErrInvalidNotifierConfig) {
		t.Fatalf("NewRetryNotifier(bad max backoff) error = %v, want ErrInvalidNotifierConfig", err)
	}
}

func TestTimeoutNotifier(t *testing.T) {
	timeout, err := NewTimeoutNotifier(NotifierFunc(func(ctx context.Context, message Message) error {
		<-ctx.Done()
		return ctx.Err()
	}), time.Nanosecond)
	if err != nil {
		t.Fatalf("NewTimeoutNotifier() error = %v", err)
	}

	if err := timeout.Send(context.Background(), testMessage(ChannelEmail)); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Send() error = %v, want context deadline exceeded", err)
	}
}

func TestTimeoutNotifierValidation(t *testing.T) {
	if _, err := NewTimeoutNotifier(nil, time.Second); !errors.Is(err, ErrNilNotifier) {
		t.Fatalf("NewTimeoutNotifier(nil) error = %v, want ErrNilNotifier", err)
	}
	if _, err := NewTimeoutNotifier(NewMemoryNotifier(), 0); !errors.Is(err, ErrInvalidNotifierConfig) {
		t.Fatalf("NewTimeoutNotifier(0) error = %v, want ErrInvalidNotifierConfig", err)
	}
}

func testMessage(channel string) Message {
	return Message{TenantID: "tenant-a", Channel: channel, To: "user@example.com", Subject: "Hi"}
}
