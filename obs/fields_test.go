package obs

import (
	"context"
	"testing"

	tenantctx "github.com/DarkInno/gotenancy/core/context"
	"github.com/DarkInno/gotenancy/core/types"
)

func TestFields(t *testing.T) {
	tenantFields := Fields(tenantctx.WithTenant(context.Background(), types.Tenant{ID: "tenant-a"}))
	if tenantFields[TenantIDField] != "tenant-a" || tenantFields[TenantSideField] != tenantSide {
		t.Fatalf("tenant Fields() = %#v, want tenant-a tenant side", tenantFields)
	}

	hostFields := Fields(tenantctx.WithHost(context.Background()))
	if _, ok := hostFields[TenantIDField]; ok {
		t.Fatalf("host Fields() tenant id = %q, want absent", hostFields[TenantIDField])
	}
	if hostFields[TenantSideField] != hostSide {
		t.Fatalf("host Fields() = %#v, want host side", hostFields)
	}

	empty := Fields(context.Background())
	if len(empty) != 0 {
		t.Fatalf("background Fields() = %#v, want empty", empty)
	}
}

func TestRedact(t *testing.T) {
	input := map[string]string{"tenant_id": "tenant-a", "api_key": "secret", "Password": "pw"}
	got := Redact(input)
	if got["tenant_id"] != "tenant-a" {
		t.Fatalf("Redact() tenant_id = %q, want tenant-a", got["tenant_id"])
	}
	if got["api_key"] != "[REDACTED]" || got["Password"] != "[REDACTED]" {
		t.Fatalf("Redact() = %#v, want sensitive fields redacted", got)
	}

	got["tenant_id"] = "changed"
	if input["tenant_id"] != "tenant-a" {
		t.Fatal("Redact() mutated input")
	}
}
