package ses

import "errors"

var (
	// ErrInvalidSESConfig reports an invalid Amazon SES notifier configuration.
	ErrInvalidSESConfig = errors.New("saas/notification: invalid ses config")

	// ErrSESDelivery reports a failed Amazon SES delivery.
	ErrSESDelivery = errors.New("saas/notification: ses delivery failed")
)
