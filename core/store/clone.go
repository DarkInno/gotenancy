package store

import "gotenancy/core/types"

func cloneTenant(tenant types.Tenant) types.Tenant {
	if tenant.Config == nil {
		return tenant
	}

	cloned := make(map[string]string, len(tenant.Config))
	for key, value := range tenant.Config {
		cloned[key] = value
	}
	tenant.Config = cloned
	return tenant
}
