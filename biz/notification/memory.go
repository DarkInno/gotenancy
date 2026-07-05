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
	if message.TenantID == "" || message.Channel == "" || message.To == "" {
		return ErrInvalidMessage
	}

	notifier.mu.Lock()
	defer notifier.mu.Unlock()

	notifier.messages = append(notifier.messages, cloneMessage(message))
	return nil
}

func (notifier *MemoryNotifier) Messages() []Message {
	notifier.mu.Lock()
	defer notifier.mu.Unlock()

	messages := make([]Message, len(notifier.messages))
	for i, message := range notifier.messages {
		messages[i] = cloneMessage(message)
	}
	return messages
}

func cloneMessage(message Message) Message {
	if message.Metadata == nil {
		return message
	}
	metadata := make(map[string]string, len(message.Metadata))
	for key, value := range message.Metadata {
		metadata[key] = value
	}
	message.Metadata = metadata
	return message
}
