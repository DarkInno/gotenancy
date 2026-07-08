package store

import (
	"context"
	"time"

	"github.com/DarkInno/gotenancy/core/types"
)

var _ Store = (*CachedStore)(nil)

// CachedStore wraps a Store with cache-aside reads and write-through invalidation.
type CachedStore struct {
	next  Store
	cache Cache
	ttl   time.Duration
}

// NewCachedStore creates a cached store decorator.
func NewCachedStore(next Store, cache Cache, ttl time.Duration) (*CachedStore, error) {
	if next == nil {
		return nil, ErrNilStore
	}
	if cache == nil {
		return nil, ErrNilCache
	}

	return &CachedStore{next: next, cache: cache, ttl: ttl}, nil
}

// Get returns tenant metadata from cache when available, otherwise from the wrapped store.
func (store *CachedStore) Get(ctx context.Context, id types.TenantID) (types.Tenant, error) {
	tenant, ok, err := store.cache.Get(ctx, id)
	if err == nil && ok {
		return tenant, nil
	}

	tenant, err = store.next.Get(ctx, id)
	if err != nil {
		return types.Tenant{}, err
	}
	_ = store.cache.Set(ctx, tenant, store.ttl)
	return tenant, nil
}

// List delegates to the wrapped store.
func (store *CachedStore) List(ctx context.Context, filter ListFilter) ([]types.Tenant, error) {
	return store.next.List(ctx, filter)
}

// Create inserts tenant metadata and refreshes the cache.
func (store *CachedStore) Create(ctx context.Context, tenant types.Tenant) error {
	if err := store.next.Create(ctx, tenant); err != nil {
		return err
	}
	return store.cache.Set(ctx, tenant, store.ttl)
}

// Update replaces tenant metadata and refreshes the cache.
func (store *CachedStore) Update(ctx context.Context, tenant types.Tenant) error {
	if err := store.next.Update(ctx, tenant); err != nil {
		return err
	}
	return store.cache.Set(ctx, tenant, store.ttl)
}

// Delete removes tenant metadata and invalidates its cache entry.
func (store *CachedStore) Delete(ctx context.Context, id types.TenantID) error {
	if err := store.next.Delete(ctx, id); err != nil {
		return err
	}
	return store.cache.Delete(ctx, id)
}
