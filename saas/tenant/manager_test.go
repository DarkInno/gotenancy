package tenant

import (
	"context"
	"errors"
	"testing"

	tenantctx "gotenancy/core/context"
	"gotenancy/core/store"
	"gotenancy/core/types"
)

func TestCreateTenant(t *testing.T) {
	ctx := context.Background()
	backing := store.NewMemoryStore()
	seeded := false
	events := []Event{}
	manager := New(backing,
		WithIDGenerator(func(context.Context) (types.TenantID, error) {
			return "tenant-a", nil
		}),
		WithSeeder(func(ctx context.Context, tenant types.Tenant) error {
			if got, ok := tenantctx.FromContext(ctx); !ok || got.ID != tenant.ID {
				t.Fatalf("Seeder tenant context = %+v, %v; want tenant", got, ok)
			}
			seeded = true
			return nil
		}),
		WithAuditor(func(_ context.Context, event Event) error {
			events = append(events, event)
			return nil
		}),
	)

	input := CreateInput{Name: "Tenant A", PlanID: "starter", Config: map[string]string{"region": "us"}}
	created, err := manager.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.ID != "tenant-a" || created.Status != types.TenantStatusPending {
		t.Fatalf("Create() = %+v, want generated pending tenant", created)
	}
	if !seeded {
		t.Fatal("Seeder was not called")
	}
	if len(events) != 1 || events[0].Action != "create" || events[0].To != types.TenantStatusPending {
		t.Fatalf("events = %+v, want create event", events)
	}

	input.Config["region"] = "eu"
	stored, err := backing.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("store.Get() error = %v", err)
	}
	if stored.Config["region"] != "us" {
		t.Fatalf("stored config = %q, want us", stored.Config["region"])
	}
}

func TestCreateRollsBackWhenSeederFails(t *testing.T) {
	ctx := context.Background()
	backing := store.NewMemoryStore()
	wantErr := errors.New("seed failed")
	manager := New(backing, WithSeeder(func(context.Context, types.Tenant) error {
		return wantErr
	}))

	_, err := manager.Create(ctx, CreateInput{ID: "tenant-a", Name: "Tenant A"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("Create() error = %v, want seed error", err)
	}
	if _, err := backing.Get(ctx, "tenant-a"); !errors.Is(err, store.ErrTenantNotFound) {
		t.Fatalf("store.Get() after rollback error = %v, want ErrTenantNotFound", err)
	}
}

func TestUpdateTenant(t *testing.T) {
	ctx := context.Background()
	manager := New(store.NewMemoryStore())

	created, err := manager.Create(ctx, CreateInput{ID: "tenant-a", Name: "Tenant A", PlanID: "starter"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	updated, err := manager.Update(ctx, UpdateInput{
		ID:     created.ID,
		Name:   "Tenant A Updated",
		PlanID: "pro",
		Config: map[string]string{"region": "us"},
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.Status != types.TenantStatusPending {
		t.Fatalf("Update() status = %q, want pending unchanged", updated.Status)
	}
	if updated.Name != "Tenant A Updated" || updated.PlanID != "pro" || updated.Config["region"] != "us" {
		t.Fatalf("Update() = %+v, want metadata updated", updated)
	}
}

func TestLifecycleTransitions(t *testing.T) {
	ctx := context.Background()
	manager := New(store.NewMemoryStore())
	created, err := manager.Create(ctx, CreateInput{ID: "tenant-a", Name: "Tenant A"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	active, err := manager.Activate(ctx, created.ID)
	if err != nil {
		t.Fatalf("Activate() error = %v", err)
	}
	if active.Status != types.TenantStatusActive {
		t.Fatalf("Activate() status = %q, want active", active.Status)
	}

	suspended, err := manager.Suspend(ctx, created.ID)
	if err != nil {
		t.Fatalf("Suspend() error = %v", err)
	}
	if suspended.Status != types.TenantStatusSuspended {
		t.Fatalf("Suspend() status = %q, want suspended", suspended.Status)
	}

	restored, err := manager.Restore(ctx, created.ID)
	if err != nil {
		t.Fatalf("Restore() error = %v", err)
	}
	if restored.Status != types.TenantStatusActive {
		t.Fatalf("Restore() status = %q, want active", restored.Status)
	}

	deleted, err := manager.SoftDelete(ctx, created.ID)
	if err != nil {
		t.Fatalf("SoftDelete() error = %v", err)
	}
	if deleted.Status != types.TenantStatusSoftDeleted {
		t.Fatalf("SoftDelete() status = %q, want soft_deleted", deleted.Status)
	}
}

func TestInvalidLifecycleTransitions(t *testing.T) {
	ctx := context.Background()
	manager := New(store.NewMemoryStore())
	created, err := manager.Create(ctx, CreateInput{ID: "tenant-a", Name: "Tenant A"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if _, err := manager.Suspend(ctx, created.ID); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("Suspend(pending) error = %v, want ErrInvalidState", err)
	}
	if _, err := manager.Restore(ctx, created.ID); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("Restore(pending) error = %v, want ErrInvalidState", err)
	}
	if _, err := manager.SoftDelete(ctx, created.ID); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("SoftDelete(pending) error = %v, want ErrInvalidState", err)
	}
}

func TestDeleteAliasesSoftDelete(t *testing.T) {
	ctx := context.Background()
	manager := New(store.NewMemoryStore())
	created, err := manager.Create(ctx, CreateInput{ID: "tenant-a", Name: "Tenant A"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if _, err := manager.Activate(ctx, created.ID); err != nil {
		t.Fatalf("Activate() error = %v", err)
	}

	if err := manager.Delete(ctx, created.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	got, err := manager.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("Get() after delete error = %v", err)
	}
	if got.Status != types.TenantStatusSoftDeleted {
		t.Fatalf("Delete() status = %q, want soft_deleted", got.Status)
	}
}

func TestHardDeleteRequiresHostAndAllowedState(t *testing.T) {
	ctx := context.Background()
	manager := New(store.NewMemoryStore())
	created, err := manager.Create(ctx, CreateInput{ID: "tenant-a", Name: "Tenant A"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if _, err := manager.Activate(ctx, created.ID); err != nil {
		t.Fatalf("Activate() error = %v", err)
	}

	if err := manager.HardDelete(ctx, created.ID); !errors.Is(err, ErrHostRequired) {
		t.Fatalf("HardDelete(tenant ctx) error = %v, want ErrHostRequired", err)
	}

	hostCtx := tenantctx.WithHost(context.Background())
	if err := manager.HardDelete(hostCtx, created.ID); err != nil {
		t.Fatalf("HardDelete(host) error = %v", err)
	}
	if _, err := manager.Get(ctx, created.ID); !errors.Is(err, store.ErrTenantNotFound) {
		t.Fatalf("Get() after hard delete error = %v, want ErrTenantNotFound", err)
	}

	pending, err := manager.Create(ctx, CreateInput{ID: "tenant-b", Name: "Tenant B"})
	if err != nil {
		t.Fatalf("Create(tenant-b) error = %v", err)
	}
	if err := manager.HardDelete(hostCtx, pending.ID); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("HardDelete(pending) error = %v, want ErrInvalidState", err)
	}
}
