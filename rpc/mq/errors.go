package mq

import "errors"

var (
	// ErrInvalidHeaders reports a nil message-header wrapper.
	ErrInvalidHeaders = errors.New("saas/rpc/mq: invalid headers")
)
