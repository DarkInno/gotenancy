package subscription

// Status describes subscription lifecycle state.
type Status string

const (
	StatusActive    Status = "active"
	StatusExpired   Status = "expired"
	StatusCancelled Status = "cancelled"
)
