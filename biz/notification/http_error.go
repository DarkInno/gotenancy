package notification

import (
	"context"
	"errors"
)

type timeoutError interface {
	Timeout() bool
}

// normalizeHTTPError preserves the original HTTP error while making caller
// cancellation and transport timeouts consistently discoverable with errors.Is.
// Older supported Go releases do not always unwrap http.Client timeout errors to
// context.DeadlineExceeded.
func normalizeHTTPError(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}
	if ctxErr := ctx.Err(); ctxErr != nil {
		return errors.Join(err, ctxErr)
	}

	var timeout timeoutError
	if errors.As(err, &timeout) && timeout.Timeout() {
		return errors.Join(err, context.DeadlineExceeded)
	}
	return err
}
