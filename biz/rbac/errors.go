package rbac

import "errors"

var (
	ErrInvalidRole    = errors.New("gotenancy/rbac: invalid role")
	ErrRoleNotFound   = errors.New("gotenancy/rbac: role not found")
	ErrRoleExists     = errors.New("gotenancy/rbac: role already exists")
	ErrPermissionDeny = errors.New("gotenancy/rbac: permission denied")
)
