package notification

import (
	"context"
	"time"
)

// TimeoutNotifier applies a per-message timeout to a notifier.
type TimeoutNotifier struct {
	notifier Notifier
	timeout  time.Duration
}

var _ Notifier = (*TimeoutNotifier)(nil)

// NewTimeoutNotifier wraps notifier with a send timeout.
func NewTimeoutNotifier(notifier Notifier, timeout time.Duration) (*TimeoutNotifier, error) {
	if notifier == nil {
		return nil, ErrNilNotifier
	}
	if timeout <= 0 {
		return nil, ErrInvalidNotifierConfig
	}
	return &TimeoutNotifier{notifier: notifier, timeout: timeout}, nil
}

// Send applies the configured timeout around the wrapped notifier.
func (notifier *TimeoutNotifier) Send(ctx context.Context, message Message) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if notifier == nil {
		return ErrNilNotifier
	}
	if err := message.Validate(); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, notifier.timeout)
	defer cancel()
	return notifier.notifier.Send(ctx, message)
}
