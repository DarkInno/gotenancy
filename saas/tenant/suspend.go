package tenant

import (
	"context"

	"gotenancy/core/types"
)

// Suspend transitions an Active tenant to Suspended.
func (manager *Manager) Suspend(ctx context.Context, id types.TenantID) (types.Tenant, error) {
	return manager.transition(ctx, id, "suspend", suspendTransitions)
}
