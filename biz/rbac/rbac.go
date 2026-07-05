package rbac

import "gotenancy/core/types"

type Permission string

type Role struct {
	TenantID    types.TenantID
	Key         string
	Permissions []Permission
}
