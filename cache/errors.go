package cache

import (
	"errors"

	"github.com/DarkInno/gotenancy"
)

var (
	// ErrNoTenant reports cache access without tenant or allowed host context.
	ErrNoTenant = gotenancy.ErrNoTenant

	// ErrUnsafeKey reports a key that attempts to bypass tenant scoping.
	ErrUnsafeKey = errors.New("gotenancy/cache: unsafe key")

	// ErrHostGlobalKeyNotAllowed reports host global key access without explicit opt-in.
	ErrHostGlobalKeyNotAllowed = errors.New("gotenancy/cache: host global key not allowed")

	// ErrInvalidCacheSize reports an invalid bounded memory cache size.
	ErrInvalidCacheSize = errors.New("gotenancy/cache: invalid cache size")

	// ErrInvalidRedisConfig reports an invalid Redis cache adapter configuration.
	ErrInvalidRedisConfig = errors.New("gotenancy/cache: invalid redis config")
)
