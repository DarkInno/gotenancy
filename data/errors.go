package data

import (
	"errors"

	"gotenancy"
)

var (
	// ErrNoTenant reports that a tenant-scoped filter was requested without tenant context.
	ErrNoTenant = gotenancy.ErrNoTenant

	// ErrInvalidFieldName reports an unsafe field name in filter options.
	ErrInvalidFieldName = errors.New("gotenancy/data: invalid field name")
)
