package subscription

import (
	"context"
	"errors"
	"testing"
	"time"

	"gotenancy/core/types"
)

func TestMemoryServiceSubscribe(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 7, 5, 12, 0, 0, 0, time.UTC)
	events := []BillingEvent{}
	service := NewMemoryService(
		WithClock(func() time.Time { return now }),
		WithBillingHook(func(_ context.Context, event BillingEvent) error {
			events = append(events, event)
			return nil
		}),
	)

	got, err := service.Subscribe(ctx, "tenant-a", "starter")
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}
	if got.TenantID != "tenant-a" || got.PlanID != "starter" || got.Status != StatusActive || !got.StartDate.Equal(now) {
		t.Fatalf("Subscribe() = %+v, want active subscription", got)
	}
	if len(events) != 1 || events[0].Action != "subscribe" || events[0].ToPlan != "starter" {
		t.Fatalf("events = %+v, want subscribe event", events)
	}
	if _, err := service.Subscribe(ctx, "tenant-a", "starter"); !errors.Is(err, ErrSubscriptionAlreadyExists) {
		t.Fatalf("Subscribe(duplicate) error = %v, want ErrSubscriptionAlreadyExists", err)
	}
}

func TestMemoryServiceUpgradeDowngradeUnsubscribe(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 7, 5, 12, 0, 0, 0, time.UTC)
	service := NewMemoryService(WithClock(func() time.Time { return now }))
	if _, err := service.Subscribe(ctx, "tenant-a", "starter"); err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}

	upgraded, err := service.Upgrade(ctx, "tenant-a", "pro")
	if err != nil {
		t.Fatalf("Upgrade() error = %v", err)
	}
	if upgraded.PlanID != "pro" || upgraded.Status != StatusActive {
		t.Fatalf("Upgrade() = %+v, want pro active", upgraded)
	}

	downgraded, err := service.Downgrade(ctx, "tenant-a", "starter")
	if err != nil {
		t.Fatalf("Downgrade() error = %v", err)
	}
	if downgraded.PlanID != "starter" || downgraded.Status != StatusActive {
		t.Fatalf("Downgrade() = %+v, want starter active", downgraded)
	}

	now = now.Add(time.Hour)
	cancelled, err := service.Unsubscribe(ctx, "tenant-a")
	if err != nil {
		t.Fatalf("Unsubscribe() error = %v", err)
	}
	if cancelled.Status != StatusCancelled || cancelled.EndDate == nil || !cancelled.EndDate.Equal(now) {
		t.Fatalf("Unsubscribe() = %+v, want cancelled with end date", cancelled)
	}

	if _, err := service.Upgrade(ctx, "tenant-a", "pro"); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("Upgrade(cancelled) error = %v, want ErrInvalidTransition", err)
	}
	if _, err := service.Unsubscribe(ctx, "tenant-a"); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("Unsubscribe(cancelled) error = %v, want ErrInvalidTransition", err)
	}
}

func TestMemoryServiceGetCopiesSubscription(t *testing.T) {
	ctx := context.Background()
	service := NewMemoryService()
	if _, err := service.Subscribe(ctx, "tenant-a", "starter"); err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}
	cancelled, err := service.Unsubscribe(ctx, "tenant-a")
	if err != nil {
		t.Fatalf("Unsubscribe() error = %v", err)
	}

	cancelled.EndDate = nil
	got, err := service.Get(ctx, "tenant-a")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.EndDate == nil {
		t.Fatal("Get() EndDate = nil, want stored end date")
	}

	got.EndDate = nil
	again, err := service.Get(ctx, "tenant-a")
	if err != nil {
		t.Fatalf("Get() again error = %v", err)
	}
	if again.EndDate == nil {
		t.Fatal("mutating returned subscription changed stored end date")
	}
}

func TestMemoryServiceValidationAndMissing(t *testing.T) {
	ctx := context.Background()
	service := NewMemoryService()

	if _, err := service.Subscribe(ctx, "", "starter"); !errors.Is(err, ErrInvalidSubscription) {
		t.Fatalf("Subscribe(empty tenant) error = %v, want ErrInvalidSubscription", err)
	}
	if _, err := service.Subscribe(ctx, "tenant-a", ""); !errors.Is(err, ErrInvalidSubscription) {
		t.Fatalf("Subscribe(empty plan) error = %v, want ErrInvalidSubscription", err)
	}
	if _, err := service.Get(ctx, types.TenantID("missing")); !errors.Is(err, ErrSubscriptionNotFound) {
		t.Fatalf("Get(missing) error = %v, want ErrSubscriptionNotFound", err)
	}
	if _, err := service.Upgrade(ctx, "missing", "pro"); !errors.Is(err, ErrSubscriptionNotFound) {
		t.Fatalf("Upgrade(missing) error = %v, want ErrSubscriptionNotFound", err)
	}
	if _, err := service.Downgrade(ctx, "missing", "starter"); !errors.Is(err, ErrSubscriptionNotFound) {
		t.Fatalf("Downgrade(missing) error = %v, want ErrSubscriptionNotFound", err)
	}
	if _, err := service.Unsubscribe(ctx, "missing"); !errors.Is(err, ErrSubscriptionNotFound) {
		t.Fatalf("Unsubscribe(missing) error = %v, want ErrSubscriptionNotFound", err)
	}
}
