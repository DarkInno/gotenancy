package quota

import "errors"

var (
	// ErrInvalidQuota reports invalid quota input.
	ErrInvalidQuota = errors.New("gotenancy/quota: invalid quota")

	// ErrQuotaExceeded reports a quota limit violation.
	ErrQuotaExceeded = errors.New("gotenancy/quota: quota exceeded")
)
