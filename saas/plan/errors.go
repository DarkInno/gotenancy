package plan

import "errors"

var (
	// ErrPlanNotFound reports that a plan does not exist.
	ErrPlanNotFound = errors.New("gotenancy/plan: plan not found")

	// ErrPlanAlreadyExists reports that a plan already exists.
	ErrPlanAlreadyExists = errors.New("gotenancy/plan: plan already exists")

	// ErrInvalidPlan reports invalid plan metadata.
	ErrInvalidPlan = errors.New("gotenancy/plan: invalid plan")
)
