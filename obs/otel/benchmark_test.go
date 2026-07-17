package otel

import (
	"context"
	"testing"

	tenantctx "github.com/DarkInno/saas/core/context"
	"github.com/DarkInno/saas/core/types"
)

func BenchmarkSpanAttributes(b *testing.B) {
	ctx := tenantctx.WithTenant(context.Background(), types.Tenant{ID: "tenant-a"})
	b.ReportAllocs()
	for range b.N {
		_ = SpanAttributes(ctx)
	}
}
