package cache

import (
	"context"
	"strings"

	tenantctx "gotenancy/core/context"
)

const (
	tenantPrefix = "t:"
	globalPrefix = "g:"
)

// KeyBuilder creates scoped cache keys.
type KeyBuilder struct {
	AllowHostGlobal bool
}

// Build returns a scoped key for ctx.
func (builder KeyBuilder) Build(ctx context.Context, key string) (string, error) {
	if key == "" || strings.HasPrefix(key, tenantPrefix) || strings.HasPrefix(key, globalPrefix) {
		return "", ErrUnsafeKey
	}

	if tenant, ok := tenantctx.FromContext(ctx); ok {
		return tenantPrefix + tenant.ID.String() + ":" + key, nil
	}
	if tenantctx.IsHost(ctx) {
		if !builder.AllowHostGlobal {
			return "", ErrHostGlobalKeyNotAllowed
		}
		return globalPrefix + key, nil
	}
	return "", ErrNoTenant
}
