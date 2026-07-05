package types

// Tenant is the shared metadata shape used by the core abstractions.
type Tenant struct {
	ID     TenantID
	Name   string
	Status TenantStatus
	PlanID string
	Config map[string]string
}
