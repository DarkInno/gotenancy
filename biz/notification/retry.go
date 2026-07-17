package notification

import (
	"context"
	"errors"
	"time"
)

const (
	defaultRetryAttempts = 3
	defaultRetryBackoff  = 100 * time.Millisecond
	defaultMaxBackoff    = 2 * time.Second
)

// RetryConfig configures RetryNotifier.
type RetryConfig struct {
	Attempts   int
	Backoff    time.Duration
	MaxBackoff time.Duration
	RetryIf    func(error) bool
	Sleep      func(context.Context, time.Duration) error
}

// RetryNotifier retries transient notification delivery failures.
type RetryNotifier struct {
	notifier Notifier
	config   RetryConfig
}

var _ Notifier = (*RetryNotifier)(nil)

// NewRetryNotifier wraps notifier with bounded retry behavior.
func NewRetryNotifier(notifier Notifier, config RetryConfig) (*RetryNotifier, error) {
	if notifier == nil {
		return nil, ErrNilNotifier
	}
	if config.Attempts < 0 || config.Backoff < 0 || config.MaxBackoff < 0 {
		return nil, ErrInvalidNotifierConfig
	}
	if config.Backoff > 0 && config.MaxBackoff > 0 && config.MaxBackoff < config.Backoff {
		return nil, ErrInvalidNotifierConfig
	}

	config = normalizeRetryConfig(config)
	if config.Attempts <= 0 || config.Backoff < 0 || config.MaxBackoff < config.Backoff {
		return nil, ErrInvalidNotifierConfig
	}
	return &RetryNotifier{notifier: notifier, config: config}, nil
}

// Send retries message delivery according to config.
func (notifier *RetryNotifier) Send(ctx context.Context, message Message) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if notifier == nil {
		return ErrNilNotifier
	}
	if err := message.Validate(); err != nil {
		return err
	}

	var lastErr error
	for attempt := 1; attempt <= notifier.config.Attempts; attempt++ {
		lastErr = notifier.notifier.Send(ctx, message)
		if lastErr == nil {
			return nil
		}
		if ctxErr := ctx.Err(); ctxErr != nil {
			return errors.Join(lastErr, ctxErr)
		}
		if attempt == notifier.config.Attempts || !notifier.config.RetryIf(lastErr) {
			return lastErr
		}
		if err := notifier.config.Sleep(ctx, retryDelay(notifier.config, attempt)); err != nil {
			return errors.Join(lastErr, err)
		}
	}
	return lastErr
}

// DefaultRetryIf reports whether err should be retried by default.
func DefaultRetryIf(err error) bool {
	if err == nil {
		return false
	}
	var retryable retryableError
	if errors.As(err, &retryable) {
		return retryable.Retryable()
	}
	return !errors.Is(err, ErrInvalidMessage) &&
		!errors.Is(err, ErrInvalidNotifierConfig) &&
		!errors.Is(err, ErrInvalidResendConfig) &&
		!errors.Is(err, ErrInvalidSMTPConfig) &&
		!errors.Is(err, ErrInvalidWebhookConfig) &&
		!errors.Is(err, ErrNilNotifier) &&
		!errors.Is(err, ErrTLSRequired) &&
		!errors.Is(err, ErrUnsupportedChannel)
}

type retryableError interface {
	Retryable() bool
}

func normalizeRetryConfig(config RetryConfig) RetryConfig {
	if config.Attempts == 0 {
		config.Attempts = defaultRetryAttempts
	}
	if config.Backoff == 0 {
		config.Backoff = defaultRetryBackoff
	}
	if config.MaxBackoff == 0 {
		config.MaxBackoff = defaultMaxBackoff
	}
	if config.MaxBackoff < config.Backoff {
		config.MaxBackoff = config.Backoff
	}
	if config.RetryIf == nil {
		config.RetryIf = DefaultRetryIf
	}
	if config.Sleep == nil {
		config.Sleep = sleep
	}
	return config
}

func retryDelay(config RetryConfig, attempt int) time.Duration {
	delay := config.Backoff
	for i := 1; i < attempt; i++ {
		if delay >= config.MaxBackoff/2 {
			return config.MaxBackoff
		}
		delay *= 2
	}
	if delay > config.MaxBackoff {
		return config.MaxBackoff
	}
	return delay
}

func sleep(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		return ctx.Err()
	}

	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
