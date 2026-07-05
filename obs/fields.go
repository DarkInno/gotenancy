package obs

import (
	"context"

	tenantctx "gotenancy/core/context"
)

const (
	// TenantIDField is the standard observability field for tenant ID.
	TenantIDField = "tenant_id"

	// TenantSideField is the standard observability field for tenant side.
	TenantSideField = "tenant_side"

	hostSide   = "host"
	tenantSide = "tenant"
)

// Fields returns tenant observability fields for ctx.
func Fields(ctx context.Context) map[string]string {
	fields := map[string]string{}
	if tenant, ok := tenantctx.FromContext(ctx); ok {
		fields[TenantIDField] = tenant.ID.String()
		fields[TenantSideField] = tenantSide
		return fields
	}
	if tenantctx.IsHost(ctx) {
		fields[TenantSideField] = hostSide
	}
	return fields
}
