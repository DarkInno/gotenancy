package notification

import (
	"context"
	"errors"
	"testing"
)

func TestMemoryNotifier(t *testing.T) {
	notifier := NewMemoryNotifier()
	message := Message{TenantID: "tenant-a", Channel: "email", To: "user@example.com", Subject: "Hi", Metadata: map[string]string{"kind": "welcome"}}
	if err := notifier.Send(context.Background(), message); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	message.Metadata["kind"] = "changed"

	messages := notifier.Messages()
	if len(messages) != 1 || messages[0].Metadata["kind"] != "welcome" {
		t.Fatalf("Messages() = %+v, want copied welcome message", messages)
	}
	messages[0].Metadata["kind"] = "mutated"
	if again := notifier.Messages(); again[0].Metadata["kind"] != "welcome" {
		t.Fatal("Messages() returned mutable internal state")
	}
}

func TestMemoryNotifierValidation(t *testing.T) {
	if err := NewMemoryNotifier().Send(context.Background(), Message{}); !errors.Is(err, ErrInvalidMessage) {
		t.Fatalf("Send(invalid) error = %v, want ErrInvalidMessage", err)
	}
	if err := (*MemoryNotifier)(nil).Send(context.Background(), Message{}); !errors.Is(err, ErrNilNotifier) {
		t.Fatalf("nil Send() error = %v, want ErrNilNotifier", err)
	}
	if messages := (*MemoryNotifier)(nil).Messages(); messages != nil {
		t.Fatalf("nil Messages() = %+v, want nil", messages)
	}
}
