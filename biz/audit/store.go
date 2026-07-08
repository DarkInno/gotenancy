package audit

import (
	"context"

	"github.com/DarkInno/gotenancy/core/types"
)

type Store interface {
	Record(ctx context.Context, event Event) error
	List(ctx context.Context, tenantID types.TenantID) ([]Event, error)
}
