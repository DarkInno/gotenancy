package notification

import (
	"context"
	"strings"
)

// ChannelRouter dispatches messages to a notifier registered for message.Channel.
type ChannelRouter struct {
	routes map[string]Notifier
}

var _ Notifier = (*ChannelRouter)(nil)

// NewChannelRouter creates a channel-based router.
func NewChannelRouter(routes map[string]Notifier) (*ChannelRouter, error) {
	if len(routes) == 0 {
		return nil, ErrInvalidNotifierConfig
	}

	copied := make(map[string]Notifier, len(routes))
	for channel, notifier := range routes {
		channel = normalizeChannel(channel)
		if channel == "" || notifier == nil {
			return nil, ErrInvalidNotifierConfig
		}
		if _, exists := copied[channel]; exists {
			return nil, ErrInvalidNotifierConfig
		}
		copied[channel] = notifier
	}
	return &ChannelRouter{routes: copied}, nil
}

// Send dispatches message to its channel notifier.
func (router *ChannelRouter) Send(ctx context.Context, message Message) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if router == nil {
		return ErrNilNotifier
	}
	if err := message.Validate(); err != nil {
		return err
	}

	notifier, ok := router.routes[normalizeChannel(message.Channel)]
	if !ok {
		return ErrUnsupportedChannel
	}
	return notifier.Send(ctx, message)
}

func normalizeChannel(channel string) string {
	return strings.TrimSpace(channel)
}
