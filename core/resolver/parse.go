package resolver

import "github.com/DarkInno/gotenancy/core/types"

func parseTenantID(raw string, strategy types.TenantIDStrategy) (types.TenantID, error) {
	return types.ParseTenantID(raw, strategy)
}
