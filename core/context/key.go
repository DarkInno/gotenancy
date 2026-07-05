package tenantctx

type stateKey struct{}

type side uint8

const (
	sideNone side = iota
	sideTenant
	sideHost
)
