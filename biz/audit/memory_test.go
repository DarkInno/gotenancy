package audit

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestMemoryStoreRecordAndList(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 7, 5, 12, 0, 0, 0, time.UTC)
	store := NewMemoryStore(WithClock(func() time.Time { return now }))
	event := Event{TenantID: "tenant-a", ActorID: "u1", Action: "orders.create", Resource: "order:1", Metadata: map[string]string{"ip": "127.0.0.1"}}

	if err := store.Record(ctx, event); err != nil {
		t.Fatalf("Record() error = %v", err)
	}
	event.Metadata["ip"] = "changed"

	events, err := store.List(ctx, "tenant-a")
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(events) != 1 || events[0].CreatedAt != now || events[0].Metadata["ip"] != "127.0.0.1" {
		t.Fatalf("List() = %+v, want copied event", events)
	}
	if other, err := store.List(ctx, "tenant-b"); err != nil || len(other) != 0 {
		t.Fatalf("List(other) = %+v, %v; want empty nil", other, err)
	}
}

func TestMemoryStoreValidation(t *testing.T) {
	if err := NewMemoryStore().Record(context.Background(), Event{}); !errors.Is(err, ErrInvalidEvent) {
		t.Fatalf("Record(invalid) error = %v, want ErrInvalidEvent", err)
	}
}
