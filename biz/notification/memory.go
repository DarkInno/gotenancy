package notification

import (
	"context"
	"sync"
)

var _ Notifier = (*MemoryNotifier)(nil)

type MemoryNotifier struct {
	mu       sync.Mutex
	messages []Message
}

func NewMemoryNotifier() *MemoryNotifier {
	return &MemoryNotifier{}
}

func (notifier *MemoryNotifier) Send(ctx context.Context, message Message) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if notifier == nil {
		return ErrNilNotifier
	}
	if err := message.Validate(); err != nil {
		return err
	}

	notifier.mu.Lock()
	defer notifier.mu.Unlock()

	notifier.messages = append(notifier.messages, message.Clone())
	return nil
}

func (notifier *MemoryNotifier) Messages() []Message {
	if notifier == nil {
		return nil
	}

	notifier.mu.Lock()
	defer notifier.mu.Unlock()

	messages := make([]Message, len(notifier.messages))
	for i, message := range notifier.messages {
		messages[i] = message.Clone()
	}
	return messages
}
