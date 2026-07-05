package cache

import (
	"context"
	"sync"
	"time"
)

var _ Cache = (*Memory)(nil)

// Memory is a thread-safe in-memory cache.
type Memory struct {
	mu         sync.Mutex
	now        func() time.Time
	maxEntries int
	entries    map[string]entry
}

type entry struct {
	value     []byte
	createdAt time.Time
	expiresAt time.Time
}

// NewMemory creates an empty memory cache.
func NewMemory() *Memory {
	return newMemoryWithClock(time.Now)
}

// NewBoundedMemory creates a memory cache with a maximum number of entries.
func NewBoundedMemory(maxEntries int) (*Memory, error) {
	return newBoundedMemoryWithClock(time.Now, maxEntries)
}

func newMemoryWithClock(now func() time.Time) *Memory {
	cache, _ := newBoundedMemoryWithClock(now, 0)
	return cache
}

func newBoundedMemoryWithClock(now func() time.Time, maxEntries int) (*Memory, error) {
	if maxEntries < 0 {
		return nil, ErrInvalidCacheSize
	}
	return &Memory{
		now:        now,
		maxEntries: maxEntries,
		entries:    map[string]entry{},
	}, nil
}

// Get returns a cache value.
func (cache *Memory) Get(ctx context.Context, key string) ([]byte, bool, error) {
	if err := ctx.Err(); err != nil {
		return nil, false, err
	}

	cache.mu.Lock()
	defer cache.mu.Unlock()

	entry, ok := cache.entries[key]
	if !ok {
		return nil, false, nil
	}
	if !entry.expiresAt.IsZero() && !entry.expiresAt.After(cache.now()) {
		delete(cache.entries, key)
		return nil, false, nil
	}
	return cloneBytes(entry.value), true, nil
}

// Set stores a cache value.
func (cache *Memory) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	now := cache.now()
	entry := entry{value: cloneBytes(value), createdAt: now}
	if ttl > 0 {
		entry.expiresAt = now.Add(ttl)
	}

	cache.mu.Lock()
	defer cache.mu.Unlock()

	if _, ok := cache.entries[key]; !ok && cache.maxEntries > 0 && len(cache.entries) >= cache.maxEntries {
		cache.evictExpiredLocked(now)
	}
	if _, ok := cache.entries[key]; !ok && cache.maxEntries > 0 && len(cache.entries) >= cache.maxEntries {
		cache.evictOldestLocked()
	}
	cache.entries[key] = entry
	return nil
}

// Delete removes a cache value.
func (cache *Memory) Delete(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	cache.mu.Lock()
	defer cache.mu.Unlock()

	delete(cache.entries, key)
	return nil
}

func (cache *Memory) evictExpiredLocked(now time.Time) {
	for key, entry := range cache.entries {
		if !entry.expiresAt.IsZero() && !entry.expiresAt.After(now) {
			delete(cache.entries, key)
		}
	}
}

func (cache *Memory) evictOldestLocked() {
	var (
		oldestKey string
		oldest    time.Time
	)
	for key, entry := range cache.entries {
		if oldestKey == "" || entry.createdAt.Before(oldest) {
			oldestKey = key
			oldest = entry.createdAt
		}
	}
	if oldestKey != "" {
		delete(cache.entries, oldestKey)
	}
}

func cloneBytes(value []byte) []byte {
	if value == nil {
		return nil
	}
	cloned := make([]byte, len(value))
	copy(cloned, value)
	return cloned
}
