package notification

import "errors"

var (
	ErrInvalidMessage        = errors.New("gotenancy/notification: invalid message")
	ErrInvalidNotifierConfig = errors.New("gotenancy/notification: invalid notifier config")
	ErrInvalidResendConfig   = errors.New("gotenancy/notification: invalid resend config")
	ErrInvalidSESConfig      = errors.New("gotenancy/notification: invalid ses config")
	ErrInvalidSMTPConfig     = errors.New("gotenancy/notification: invalid smtp config")
	ErrInvalidWebhookConfig  = errors.New("gotenancy/notification: invalid webhook config")
	ErrNilNotifier           = errors.New("gotenancy/notification: nil notifier")
	ErrResendDelivery        = errors.New("gotenancy/notification: resend delivery failed")
	ErrSESDelivery           = errors.New("gotenancy/notification: ses delivery failed")
	ErrUnsupportedChannel    = errors.New("gotenancy/notification: unsupported channel")
	ErrWebhookDelivery       = errors.New("gotenancy/notification: webhook delivery failed")
	ErrTLSRequired           = errors.New("gotenancy/notification: tls required")
)
