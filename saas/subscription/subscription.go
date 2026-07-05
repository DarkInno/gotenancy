package subscription

import (
	"time"

	"gotenancy/core/types"
)

// Subscription binds a tenant to a plan.
type Subscription struct {
	TenantID  types.TenantID
	PlanID    string
	Status    Status
	StartDate time.Time
	EndDate   *time.Time
}
