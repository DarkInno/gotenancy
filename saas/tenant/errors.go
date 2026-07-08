package tenant

import "github.com/DarkInno/gotenancy"

var (
	// ErrInvalidState reports an invalid tenant lifecycle transition.
	ErrInvalidState = gotenancy.ErrInvalidState

	// ErrHostRequired reports that an operation requires host-side context.
	ErrHostRequired = gotenancy.ErrHostRequired
)
