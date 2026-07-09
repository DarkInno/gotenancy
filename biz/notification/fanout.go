package notification

import (
	"context"
	"errors"
)

// FanoutNotifier sends each message to multiple notifiers in order.
type FanoutNotifier struct {
	notifiers []Notifier
}

var _ Notifier = (*FanoutNotifier)(nil)

// NewFanoutNotifier creates a sequential fanout notifier.
func NewFanoutNotifier(notifiers ...Notifier) (*FanoutNotifier, error) {
	if len(notifiers) == 0 {
		return nil, ErrInvalidNotifierConfig
	}

	copied := make([]Notifier, len(notifiers))
	for i, notifier := range notifiers {
		if notifier == nil {
			return nil, ErrInvalidNotifierConfig
		}
		copied[i] = notifier
	}
	return &FanoutNotifier{notifiers: copied}, nil
}

// Send delivers message to all configured notifiers and joins delivery errors.
func (notifier *FanoutNotifier) Send(ctx context.Context, message Message) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if notifier == nil {
		return ErrNilNotifier
	}
	if err := message.Validate(); err != nil {
		return err
	}

	var sendErr error
	for _, target := range notifier.notifiers {
		if err := ctx.Err(); err != nil {
			return errors.Join(sendErr, err)
		}
		sendErr = errors.Join(sendErr, target.Send(ctx, message))
	}
	return sendErr
}
