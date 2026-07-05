package tenant

import (
	"context"

	"gotenancy/core/types"
)

// Restore transitions a Suspended tenant to Active.
func (manager *Manager) Restore(ctx context.Context, id types.TenantID) (types.Tenant, error) {
	return manager.transition(ctx, id, "restore", restoreTransitions)
}
