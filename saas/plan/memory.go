package plan

import (
	"context"
	"sync"
)

var _ Service = (*MemoryService)(nil)

// MemoryService is a thread-safe in-memory plan service.
type MemoryService struct {
	mu    sync.RWMutex
	plans map[string]Plan
}

// NewMemoryService creates an empty plan service.
func NewMemoryService() *MemoryService {
	return &MemoryService{plans: map[string]Plan{}}
}

// Create inserts a plan.
func (service *MemoryService) Create(ctx context.Context, plan Plan) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := validatePlan(plan); err != nil {
		return err
	}

	service.mu.Lock()
	defer service.mu.Unlock()

	if _, ok := service.plans[plan.ID]; ok {
		return ErrPlanAlreadyExists
	}
	service.plans[plan.ID] = clonePlan(plan)
	return nil
}

// Get returns a plan by ID.
func (service *MemoryService) Get(ctx context.Context, id string) (Plan, error) {
	if err := ctx.Err(); err != nil {
		return Plan{}, err
	}
	if id == "" {
		return Plan{}, ErrInvalidPlan
	}

	service.mu.RLock()
	defer service.mu.RUnlock()

	plan, ok := service.plans[id]
	if !ok {
		return Plan{}, ErrPlanNotFound
	}
	return clonePlan(plan), nil
}

// Update replaces a plan.
func (service *MemoryService) Update(ctx context.Context, plan Plan) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := validatePlan(plan); err != nil {
		return err
	}

	service.mu.Lock()
	defer service.mu.Unlock()

	if _, ok := service.plans[plan.ID]; !ok {
		return ErrPlanNotFound
	}
	service.plans[plan.ID] = clonePlan(plan)
	return nil
}

// Delete removes a plan.
func (service *MemoryService) Delete(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if id == "" {
		return ErrInvalidPlan
	}

	service.mu.Lock()
	defer service.mu.Unlock()

	if _, ok := service.plans[id]; !ok {
		return ErrPlanNotFound
	}
	delete(service.plans, id)
	return nil
}

func validatePlan(plan Plan) error {
	if plan.ID == "" || plan.Name == "" {
		return ErrInvalidPlan
	}
	for _, feature := range plan.Features {
		if feature.Key == "" {
			return ErrInvalidPlan
		}
	}
	for _, quota := range plan.Quotas {
		if quota.Resource == "" || quota.Limit < 0 {
			return ErrInvalidPlan
		}
		if quota.Period == "" {
			return ErrInvalidPlan
		}
	}
	return nil
}

func clonePlan(plan Plan) Plan {
	features := make([]Feature, len(plan.Features))
	for i, feature := range plan.Features {
		features[i] = Feature{
			Key:     feature.Key,
			Enabled: feature.Enabled,
			Config:  cloneStringMap(feature.Config),
		}
	}

	quotas := make([]Quota, len(plan.Quotas))
	copy(quotas, plan.Quotas)

	return Plan{
		ID:       plan.ID,
		Name:     plan.Name,
		Features: features,
		Quotas:   quotas,
	}
}

func cloneStringMap(values map[string]string) map[string]string {
	if values == nil {
		return nil
	}
	cloned := make(map[string]string, len(values))
	for key, value := range values {
		cloned[key] = value
	}
	return cloned
}
