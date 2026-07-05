package cache

import (
	"context"
	"time"
)

// TenantCache scopes all cache keys by tenant or explicitly allowed host global keys.
type TenantCache struct {
	next    Cache
	builder KeyBuilder
}

// Option configures TenantCache.
type Option func(*TenantCache)

// WithHostGlobalKeys allows host context to access global cache keys.
func WithHostGlobalKeys(allow bool) Option {
	return func(cache *TenantCache) {
		cache.builder.AllowHostGlobal = allow
	}
}

// NewTenantCache creates a scoped cache wrapper.
func NewTenantCache(next Cache, opts ...Option) *TenantCache {
	cache := &TenantCache{next: next}
	for _, opt := range opts {
		if opt != nil {
			opt(cache)
		}
	}
	return cache
}

// Get reads a scoped cache key.
func (cache *TenantCache) Get(ctx context.Context, key string) ([]byte, bool, error) {
	scoped, err := cache.builder.Build(ctx, key)
	if err != nil {
		return nil, false, err
	}
	return cache.next.Get(ctx, scoped)
}

// Set writes a scoped cache key.
func (cache *TenantCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	scoped, err := cache.builder.Build(ctx, key)
	if err != nil {
		return err
	}
	return cache.next.Set(ctx, scoped, value, ttl)
}

// Delete removes a scoped cache key.
func (cache *TenantCache) Delete(ctx context.Context, key string) error {
	scoped, err := cache.builder.Build(ctx, key)
	if err != nil {
		return err
	}
	return cache.next.Delete(ctx, scoped)
}
