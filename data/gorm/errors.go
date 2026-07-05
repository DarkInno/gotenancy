package gormtenant

import (
	"errors"

	"gotenancy"
)

var (
	// ErrTenantFieldNotFound reports that a model does not expose the configured tenant field.
	ErrTenantFieldNotFound = errors.New("gotenancy/gorm: tenant field not found")

	// ErrTenantMismatch reports that a model already contains a different tenant ID.
	ErrTenantMismatch = gotenancy.ErrTenantMismatch

	// ErrUnscopedRequiresHost reports that Unscoped is forbidden in tenant context.
	ErrUnscopedRequiresHost = errors.New("gotenancy/gorm: unscoped requires host context")

	// ErrRawRequiresHost reports that raw SQL requires explicit host context.
	ErrRawRequiresHost = errors.New("gotenancy/gorm: raw SQL requires host context")
)
