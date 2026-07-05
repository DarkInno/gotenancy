package quota

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
)

func TestServiceConsumeCheckAndReset(t *testing.T) {
	ctx := context.Background()
	service := NewService(NewMemoryStore())
	limit := Limit{TenantID: "tenant-a", Resource: "api_calls", Limit: 10, Period: PeriodDay}

	usage, err := service.Consume(ctx, limit, 3)
	if err != nil {
		t.Fatalf("Consume(3) error = %v", err)
	}
	if usage.Used != 3 || usage.Limit != 10 {
		t.Fatalf("Consume(3) usage = %+v, want used 3 limit 10", usage)
	}

	if _, err := service.Check(ctx, limit, 8); !errors.Is(err, ErrQuotaExceeded) {
		t.Fatalf("Check(over) error = %v, want ErrQuotaExceeded", err)
	}
	if _, err := service.Consume(ctx, limit, 8); !errors.Is(err, ErrQuotaExceeded) {
		t.Fatalf("Consume(over) error = %v, want ErrQuotaExceeded", err)
	}

	if err := service.Reset(ctx, "tenant-a", "api_calls", PeriodDay); err != nil {
		t.Fatalf("Reset() error = %v", err)
	}
	usage, err = service.Check(ctx, limit, 10)
	if err != nil {
		t.Fatalf("Check(after reset) error = %v", err)
	}
	if usage.Used != 0 {
		t.Fatalf("usage after reset = %d, want 0", usage.Used)
	}
}

func TestServiceValidation(t *testing.T) {
	ctx := context.Background()
	service := NewService(NewMemoryStore())

	if _, err := service.Consume(ctx, Limit{}, 1); !errors.Is(err, ErrInvalidQuota) {
		t.Fatalf("Consume(invalid limit) error = %v, want ErrInvalidQuota", err)
	}
	if _, err := service.Consume(ctx, Limit{TenantID: "tenant-a", Resource: "api", Limit: 1, Period: PeriodDay}, -1); !errors.Is(err, ErrInvalidQuota) {
		t.Fatalf("Consume(negative amount) error = %v, want ErrInvalidQuota", err)
	}
}

func TestServiceConsumeIsAtomicUnderConcurrency(t *testing.T) {
	ctx := context.Background()
	service := NewService(NewMemoryStore())
	limit := Limit{TenantID: "tenant-a", Resource: "api_calls", Limit: 100, Period: PeriodDay}

	const workers = 200
	var wg sync.WaitGroup
	var successes int64
	unexpected := make(chan error, workers)

	wg.Add(workers)
	for range workers {
		go func() {
			defer wg.Done()

			_, err := service.Consume(ctx, limit, 1)
			if err == nil {
				atomic.AddInt64(&successes, 1)
				return
			}
			if !errors.Is(err, ErrQuotaExceeded) {
				unexpected <- err
			}
		}()
	}
	wg.Wait()
	close(unexpected)

	for err := range unexpected {
		t.Errorf("Consume() unexpected error = %v", err)
	}
	if got := atomic.LoadInt64(&successes); got != limit.Limit {
		t.Fatalf("successful consumes = %d, want %d", got, limit.Limit)
	}

	usage, err := service.Check(ctx, limit, 0)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if usage.Used != limit.Limit {
		t.Fatalf("usage.Used = %d, want %d", usage.Used, limit.Limit)
	}
}

func TestMemoryStoreScopesByTenantResourceAndPeriod(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	if _, err := store.Add(ctx, "tenant-a", "api", PeriodDay, 1); err != nil {
		t.Fatalf("Add tenant-a error = %v", err)
	}
	if _, err := store.Add(ctx, "tenant-b", "api", PeriodDay, 2); err != nil {
		t.Fatalf("Add tenant-b error = %v", err)
	}
	if _, err := store.Add(ctx, "tenant-a", "api", PeriodMonth, 3); err != nil {
		t.Fatalf("Add tenant-a month error = %v", err)
	}

	got, err := store.Get(ctx, "tenant-a", "api", PeriodDay)
	if err != nil {
		t.Fatalf("Get tenant-a day error = %v", err)
	}
	if got != 1 {
		t.Fatalf("tenant-a day usage = %d, want 1", got)
	}
}
