package subscription

import (
	"context"

	"gotenancy/core/types"
)

// Service manages tenant subscriptions.
type Service interface {
	Subscribe(ctx context.Context, tenantID types.TenantID, planID string) (Subscription, error)
	Unsubscribe(ctx context.Context, tenantID types.TenantID) (Subscription, error)
	Upgrade(ctx context.Context, tenantID types.TenantID, planID string) (Subscription, error)
	Downgrade(ctx context.Context, tenantID types.TenantID, planID string) (Subscription, error)
	Get(ctx context.Context, tenantID types.TenantID) (Subscription, error)
}
