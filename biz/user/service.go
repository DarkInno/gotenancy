package user

import (
	"context"

	"gotenancy/core/types"
)

type Service interface {
	CreateUser(ctx context.Context, user User) error
	GetUser(ctx context.Context, id string) (User, error)
	AddMember(ctx context.Context, member Member) error
	GetMember(ctx context.Context, tenantID types.TenantID, userID string) (Member, error)
	ListMembers(ctx context.Context, tenantID types.TenantID) ([]Member, error)
	RemoveMember(ctx context.Context, tenantID types.TenantID, userID string) error
}
