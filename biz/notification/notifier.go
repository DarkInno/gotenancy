package notification

import "context"

type Notifier interface {
	Send(ctx context.Context, message Message) error
}

// NotifierFunc adapts a function to Notifier.
type NotifierFunc func(ctx context.Context, message Message) error

// Send calls fn.
func (fn NotifierFunc) Send(ctx context.Context, message Message) error {
	if fn == nil {
		return ErrNilNotifier
	}
	return fn(ctx, message)
}
