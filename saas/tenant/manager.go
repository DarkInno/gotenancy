package tenant

import (
	"context"

	tenantctx "gotenancy/core/context"
	"gotenancy/core/store"
	"gotenancy/core/types"
)

var _ Service = (*Manager)(nil)

// Manager implements tenant lifecycle operations.
type Manager struct {
	store      store.Store
	generateID IDGenerator
	seed       Seeder
	audit      Auditor
}

// New creates a tenant manager.
func New(store store.Store, opts ...Option) *Manager {
	manager := &Manager{
		store:      store,
		generateID: defaultIDGenerator,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(manager)
		}
	}
	return manager
}

// Create creates a Pending tenant and runs the optional seeder.
func (manager *Manager) Create(ctx context.Context, input CreateInput) (types.Tenant, error) {
	id := input.ID
	if id == "" {
		generated, err := manager.generateID(ctx)
		if err != nil {
			return types.Tenant{}, err
		}
		id = generated
	}

	tenant := types.Tenant{
		ID:     id,
		Name:   input.Name,
		Status: types.TenantStatusPending,
		PlanID: input.PlanID,
		Config: cloneConfig(input.Config),
	}
	if err := manager.store.Create(ctx, tenant); err != nil {
		return types.Tenant{}, err
	}
	if manager.seed != nil {
		if err := manager.seed(tenantctx.WithTenant(ctx, tenant), tenant); err != nil {
			_ = manager.store.Delete(ctx, tenant.ID)
			return types.Tenant{}, err
		}
	}
	if err := manager.emit(ctx, Event{TenantID: tenant.ID, Action: "create", To: tenant.Status}); err != nil {
		return types.Tenant{}, err
	}
	return tenant, nil
}

// Get returns tenant metadata.
func (manager *Manager) Get(ctx context.Context, id types.TenantID) (types.Tenant, error) {
	return manager.store.Get(ctx, id)
}

// Update updates tenant metadata without changing lifecycle status.
func (manager *Manager) Update(ctx context.Context, input UpdateInput) (types.Tenant, error) {
	current, err := manager.store.Get(ctx, input.ID)
	if err != nil {
		return types.Tenant{}, err
	}

	current.Name = input.Name
	current.PlanID = input.PlanID
	current.Config = cloneConfig(input.Config)
	if err := manager.store.Update(ctx, current); err != nil {
		return types.Tenant{}, err
	}
	if err := manager.emit(ctx, Event{TenantID: current.ID, Action: "update", From: current.Status, To: current.Status}); err != nil {
		return types.Tenant{}, err
	}
	return current, nil
}

// Delete soft-deletes tenant metadata.
func (manager *Manager) Delete(ctx context.Context, id types.TenantID) error {
	_, err := manager.SoftDelete(ctx, id)
	return err
}

func (manager *Manager) transition(ctx context.Context, id types.TenantID, action string, allowed map[types.TenantStatus]types.TenantStatus) (types.Tenant, error) {
	current, err := manager.store.Get(ctx, id)
	if err != nil {
		return types.Tenant{}, err
	}

	next, ok := allowed[current.Status]
	if !ok {
		return types.Tenant{}, ErrInvalidState
	}

	updated := current
	updated.Status = next
	if err := manager.store.Update(ctx, updated); err != nil {
		return types.Tenant{}, err
	}
	if err := manager.emit(ctx, Event{TenantID: id, Action: action, From: current.Status, To: next}); err != nil {
		return types.Tenant{}, err
	}
	return updated, nil
}

func (manager *Manager) emit(ctx context.Context, event Event) error {
	if manager.audit == nil {
		return nil
	}
	return manager.audit(ctx, event)
}

func cloneConfig(config map[string]string) map[string]string {
	if config == nil {
		return nil
	}
	cloned := make(map[string]string, len(config))
	for key, value := range config {
		cloned[key] = value
	}
	return cloned
}
