package subscription

import (
	"context"
	"sync"
	"time"

	"github.com/DarkInno/gotenancy/core/types"
)

var _ Service = (*MemoryService)(nil)

// Option configures MemoryService.
type Option func(*MemoryService)

// WithClock sets the clock used for subscription dates.
func WithClock(clock func() time.Time) Option {
	return func(service *MemoryService) {
		if clock != nil {
			service.now = clock
		}
	}
}

// WithBillingHook sets the billing hook.
func WithBillingHook(hook BillingHook) Option {
	return func(service *MemoryService) {
		service.billing = hook
	}
}

// MemoryService is a thread-safe in-memory subscription service.
type MemoryService struct {
	mu            sync.RWMutex
	now           func() time.Time
	billing       BillingHook
	subscriptions map[types.TenantID]Subscription
}

// NewMemoryService creates an empty subscription service.
func NewMemoryService(opts ...Option) *MemoryService {
	service := &MemoryService{
		now:           time.Now,
		subscriptions: map[types.TenantID]Subscription{},
	}
	for _, opt := range opts {
		if opt != nil {
			opt(service)
		}
	}
	return service
}

// Subscribe creates an active subscription for a tenant.
func (service *MemoryService) Subscribe(ctx context.Context, tenantID types.TenantID, planID string) (Subscription, error) {
	if err := ctx.Err(); err != nil {
		return Subscription{}, err
	}
	if tenantID == "" || planID == "" {
		return Subscription{}, ErrInvalidSubscription
	}

	service.mu.Lock()
	defer service.mu.Unlock()

	if _, ok := service.subscriptions[tenantID]; ok {
		return Subscription{}, ErrSubscriptionAlreadyExists
	}

	subscription := Subscription{
		TenantID:  tenantID,
		PlanID:    planID,
		Status:    StatusActive,
		StartDate: service.now(),
	}
	if err := service.emit(ctx, BillingEvent{TenantID: tenantID, Action: "subscribe", ToPlan: planID, Status: subscription.Status}); err != nil {
		return Subscription{}, err
	}
	service.subscriptions[tenantID] = cloneSubscription(subscription)
	return subscription, nil
}

// Unsubscribe cancels an active subscription.
func (service *MemoryService) Unsubscribe(ctx context.Context, tenantID types.TenantID) (Subscription, error) {
	if err := ctx.Err(); err != nil {
		return Subscription{}, err
	}

	service.mu.Lock()
	defer service.mu.Unlock()

	current, ok := service.subscriptions[tenantID]
	if !ok {
		return Subscription{}, ErrSubscriptionNotFound
	}
	if current.Status != StatusActive {
		return Subscription{}, ErrInvalidTransition
	}
	now := service.now()
	current.Status = StatusCancelled
	current.EndDate = &now
	if err := service.emit(ctx, BillingEvent{TenantID: tenantID, Action: "unsubscribe", FromPlan: current.PlanID, Status: current.Status}); err != nil {
		return Subscription{}, err
	}
	service.subscriptions[tenantID] = cloneSubscription(current)
	return cloneSubscription(current), nil
}

// Upgrade changes an active subscription to a higher plan.
func (service *MemoryService) Upgrade(ctx context.Context, tenantID types.TenantID, planID string) (Subscription, error) {
	return service.changePlan(ctx, tenantID, planID, "upgrade")
}

// Downgrade changes an active subscription to a lower plan.
func (service *MemoryService) Downgrade(ctx context.Context, tenantID types.TenantID, planID string) (Subscription, error) {
	return service.changePlan(ctx, tenantID, planID, "downgrade")
}

// Get returns a subscription by tenant ID.
func (service *MemoryService) Get(ctx context.Context, tenantID types.TenantID) (Subscription, error) {
	if err := ctx.Err(); err != nil {
		return Subscription{}, err
	}
	if tenantID == "" {
		return Subscription{}, ErrInvalidSubscription
	}

	service.mu.RLock()
	defer service.mu.RUnlock()

	subscription, ok := service.subscriptions[tenantID]
	if !ok {
		return Subscription{}, ErrSubscriptionNotFound
	}
	return cloneSubscription(subscription), nil
}

func (service *MemoryService) changePlan(ctx context.Context, tenantID types.TenantID, planID string, action string) (Subscription, error) {
	if err := ctx.Err(); err != nil {
		return Subscription{}, err
	}
	if tenantID == "" || planID == "" {
		return Subscription{}, ErrInvalidSubscription
	}

	service.mu.Lock()
	defer service.mu.Unlock()

	current, ok := service.subscriptions[tenantID]
	if !ok {
		return Subscription{}, ErrSubscriptionNotFound
	}
	if current.Status != StatusActive {
		return Subscription{}, ErrInvalidTransition
	}
	fromPlan := current.PlanID
	current.PlanID = planID
	if err := service.emit(ctx, BillingEvent{TenantID: tenantID, Action: action, FromPlan: fromPlan, ToPlan: planID, Status: current.Status}); err != nil {
		return Subscription{}, err
	}
	service.subscriptions[tenantID] = cloneSubscription(current)
	return cloneSubscription(current), nil
}

func (service *MemoryService) emit(ctx context.Context, event BillingEvent) error {
	if service.billing == nil {
		return nil
	}
	return service.billing(ctx, event)
}

func cloneSubscription(subscription Subscription) Subscription {
	if subscription.EndDate == nil {
		return subscription
	}
	endDate := *subscription.EndDate
	subscription.EndDate = &endDate
	return subscription
}
