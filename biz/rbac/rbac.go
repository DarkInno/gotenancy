package rbac

import "github.com/DarkInno/gotenancy/core/types"

type Permission string

type Role struct {
	TenantID    types.TenantID
	Key         string
	Permissions []Permission
}
