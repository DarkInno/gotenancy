package subscription

import (
	"context"

	"gotenancy/core/types"
)

// BillingEvent describes a subscription change for external billing systems.
type BillingEvent struct {
	TenantID types.TenantID
	Action   string
	FromPlan string
	ToPlan   string
	Status   Status
}

// BillingHook receives subscription lifecycle changes.
type BillingHook func(context.Context, BillingEvent) error
