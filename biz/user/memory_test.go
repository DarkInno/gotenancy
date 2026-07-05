package user

import (
	"context"
	"errors"
	"testing"
)

func TestMemoryServiceUserAndMembers(t *testing.T) {
	ctx := context.Background()
	service := NewMemoryService()

	if err := service.CreateUser(ctx, User{ID: "u1", Email: "u1@example.com", Name: "User 1"}); err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
	if err := service.CreateUser(ctx, User{ID: "u1", Email: "u1@example.com"}); !errors.Is(err, ErrUserExists) {
		t.Fatalf("CreateUser(duplicate) error = %v, want ErrUserExists", err)
	}

	member := Member{TenantID: "tenant-a", UserID: "u1", Roles: []string{"admin"}}
	if err := service.AddMember(ctx, member); err != nil {
		t.Fatalf("AddMember() error = %v", err)
	}
	member.Roles[0] = "mutated"

	got, err := service.GetMember(ctx, "tenant-a", "u1")
	if err != nil {
		t.Fatalf("GetMember() error = %v", err)
	}
	if got.Roles[0] != "admin" {
		t.Fatalf("GetMember roles = %#v, want admin copy", got.Roles)
	}

	members, err := service.ListMembers(ctx, "tenant-a")
	if err != nil {
		t.Fatalf("ListMembers() error = %v", err)
	}
	if len(members) != 1 || members[0].UserID != "u1" {
		t.Fatalf("ListMembers() = %+v, want u1", members)
	}

	if _, err := service.GetMember(ctx, "tenant-b", "u1"); !errors.Is(err, ErrMemberNotFound) {
		t.Fatalf("GetMember(other tenant) error = %v, want ErrMemberNotFound", err)
	}
	if err := service.RemoveMember(ctx, "tenant-a", "u1"); err != nil {
		t.Fatalf("RemoveMember() error = %v", err)
	}
}

func TestMemoryServiceValidation(t *testing.T) {
	ctx := context.Background()
	service := NewMemoryService()

	if err := service.CreateUser(ctx, User{}); !errors.Is(err, ErrInvalidUser) {
		t.Fatalf("CreateUser(invalid) error = %v, want ErrInvalidUser", err)
	}
	if err := service.AddMember(ctx, Member{TenantID: "tenant-a", UserID: "missing"}); !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("AddMember(missing user) error = %v, want ErrUserNotFound", err)
	}
}
