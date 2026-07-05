package feature

import "errors"

var (
	// ErrInvalidFeature reports invalid feature input.
	ErrInvalidFeature = errors.New("gotenancy/feature: invalid feature")

	// ErrFeatureNotFound reports a missing feature flag.
	ErrFeatureNotFound = errors.New("gotenancy/feature: feature not found")
)
