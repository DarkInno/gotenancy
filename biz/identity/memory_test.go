package identity

import (
	"context"
	"errors"
	"testing"
)

func TestMemoryStoreTenantIsolationAndConflict(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()

	link := Link{
		TenantID:      "tenant-a",
		UserID:        "u1",
		Provider:      ProviderGoogle,
		Subject:       "sub-1",
		Email:         "u1@example.com",
		EmailVerified: true,
		Metadata:      map[string]string{"org": "a"},
	}
	if err := store.Link(ctx, link); err != nil {
		t.Fatalf("Link() error = %v", err)
	}
	link.Metadata["org"] = "mutated"

	got, err := store.GetByExternal(ctx, "tenant-a", ProviderGoogle, "sub-1")
	if err != nil {
		t.Fatalf("GetByExternal() error = %v", err)
	}
	if got.UserID != "u1" || got.Metadata["org"] != "a" {
		t.Fatalf("GetByExternal() = %+v, want cloned tenant-a link", got)
	}

	if _, err := store.GetByExternal(ctx, "tenant-b", ProviderGoogle, "sub-1"); !errors.Is(err, ErrIdentityNotFound) {
		t.Fatalf("GetByExternal(other tenant) error = %v, want ErrIdentityNotFound", err)
	}

	conflict := got
	conflict.UserID = "u2"
	if err := store.Link(ctx, conflict); !errors.Is(err, ErrIdentityConflict) {
		t.Fatalf("Link(conflict) error = %v, want ErrIdentityConflict", err)
	}
}

func TestMemoryStoreGetByUserSortsLinks(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()

	links := []Link{
		{TenantID: "tenant-a", UserID: "u1", Provider: ProviderMicrosoft, Subject: "sub-2", Email: "u1@example.com"},
		{TenantID: "tenant-a", UserID: "u1", Provider: ProviderGoogle, Subject: "sub-1", Email: "u1@example.com"},
		{TenantID: "tenant-b", UserID: "u1", Provider: ProviderGitHub, Subject: "sub-3", Email: "u1@example.com"},
	}
	for _, link := range links {
		if err := store.Link(ctx, link); err != nil {
			t.Fatalf("Link() error = %v", err)
		}
	}

	got, err := store.GetByUser(ctx, "tenant-a", "u1")
	if err != nil {
		t.Fatalf("GetByUser() error = %v", err)
	}
	if len(got) != 2 || got[0].Provider != ProviderGoogle || got[1].Provider != ProviderMicrosoft {
		t.Fatalf("GetByUser() = %+v, want tenant-a links sorted by provider", got)
	}
}
