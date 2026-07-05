package audit

import (
	"context"
	"sort"
	"sync"
	"time"

	"gotenancy/core/types"
)

var _ Store = (*MemoryStore)(nil)

type MemoryStore struct {
	mu     sync.RWMutex
	events []Event
	now    func() time.Time
}

type Option func(*MemoryStore)

func WithClock(clock func() time.Time) Option {
	return func(store *MemoryStore) {
		if clock != nil {
			store.now = clock
		}
	}
}

func NewMemoryStore(opts ...Option) *MemoryStore {
	store := &MemoryStore{now: time.Now}
	for _, opt := range opts {
		if opt != nil {
			opt(store)
		}
	}
	return store
}

func (store *MemoryStore) Record(ctx context.Context, event Event) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if event.TenantID == "" || event.Action == "" || event.Resource == "" {
		return ErrInvalidEvent
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = store.now()
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	store.events = append(store.events, cloneEvent(event))
	return nil
}

func (store *MemoryStore) List(ctx context.Context, tenantID types.TenantID) ([]Event, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if tenantID == "" {
		return nil, ErrInvalidEvent
	}

	store.mu.RLock()
	defer store.mu.RUnlock()

	events := []Event{}
	for _, event := range store.events {
		if event.TenantID == tenantID {
			events = append(events, cloneEvent(event))
		}
	}
	sort.Slice(events, func(i, j int) bool {
		return events[i].CreatedAt.Before(events[j].CreatedAt)
	})
	return events, nil
}

func cloneEvent(event Event) Event {
	if event.Metadata == nil {
		return event
	}
	metadata := make(map[string]string, len(event.Metadata))
	for key, value := range event.Metadata {
		metadata[key] = value
	}
	event.Metadata = metadata
	return event
}
