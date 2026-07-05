package types

// TenantStatus describes the lifecycle state of a tenant.
type TenantStatus string

const (
	TenantStatusPending     TenantStatus = "pending"
	TenantStatusActive      TenantStatus = "active"
	TenantStatusSuspended   TenantStatus = "suspended"
	TenantStatusSoftDeleted TenantStatus = "soft_deleted"
	TenantStatusHardDeleted TenantStatus = "hard_deleted"
)
