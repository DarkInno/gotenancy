package resolver

import (
	"errors"

	"gotenancy"
)

var (
	// ErrNoTenant reports that no resolver contributor found a tenant identifier.
	ErrNoTenant = gotenancy.ErrNoTenant

	// ErrNilRequest reports a nil HTTP request.
	ErrNilRequest = errors.New("gotenancy/resolver: nil request")

	// ErrInvalidHost reports a host value that cannot be used for domain resolution.
	ErrInvalidHost = errors.New("gotenancy/resolver: invalid host")

	// ErrNilClaimExtractor reports a nil token claim extractor.
	ErrNilClaimExtractor = errors.New("gotenancy/resolver: nil claim extractor")
)
