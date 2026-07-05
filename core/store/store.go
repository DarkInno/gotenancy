package store

import (
	"context"

	"gotenancy/core/types"
)

// Store persists tenant metadata.
type Store interface {
	Get(ctx context.Context, id types.TenantID) (types.Tenant, error)
	List(ctx context.Context, filter ListFilter) ([]types.Tenant, error)
	Create(ctx context.Context, tenant types.Tenant) error
	Update(ctx context.Context, tenant types.Tenant) error
	Delete(ctx context.Context, id types.TenantID) error
}

// ListFilter restricts tenant list queries.
type ListFilter struct {
	Statuses []types.TenantStatus
	Limit    int
	Offset   int
}

func (filter ListFilter) matches(tenant types.Tenant) bool {
	if len(filter.Statuses) == 0 {
		return true
	}

	for _, status := range filter.Statuses {
		if tenant.Status == status {
			return true
		}
	}
	return false
}

func (filter ListFilter) validate() error {
	if filter.Limit < 0 || filter.Offset < 0 {
		return ErrInvalidListFilter
	}
	if filter.Offset > 0 && filter.Limit == 0 {
		return ErrInvalidListFilter
	}
	return nil
}

func pageTenants(tenants []types.Tenant, filter ListFilter) []types.Tenant {
	if filter.Offset >= len(tenants) {
		return []types.Tenant{}
	}

	start := filter.Offset
	end := len(tenants)
	if filter.Limit > 0 && start+filter.Limit < end {
		end = start + filter.Limit
	}
	return tenants[start:end]
}
