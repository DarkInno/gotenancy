package types

// MultiTenancySide identifies whether an operation runs as host infrastructure
// or inside a tenant boundary.
type MultiTenancySide string

const (
	MultiTenancySideHost   MultiTenancySide = "host"
	MultiTenancySideTenant MultiTenancySide = "tenant"
)
