package rpc

import "errors"

var (
	ErrNoTenantMetadata = errors.New("gotenancy/rpc: no tenant metadata")
	ErrInvalidCarrier   = errors.New("gotenancy/rpc: invalid carrier")
)
