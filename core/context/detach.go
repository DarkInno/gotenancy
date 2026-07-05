package tenantctx

import "context"

// Detach returns a context without tenant or host state and without parent cancellation.
func Detach(ctx context.Context) context.Context {
	return context.WithValue(context.WithoutCancel(ctx), stateKey{}, state{side: sideNone})
}
