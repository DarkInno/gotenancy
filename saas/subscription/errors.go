package subscription

import "errors"

var (
	// ErrSubscriptionNotFound reports that a subscription does not exist.
	ErrSubscriptionNotFound = errors.New("gotenancy/subscription: subscription not found")

	// ErrSubscriptionAlreadyExists reports that a tenant already has a subscription.
	ErrSubscriptionAlreadyExists = errors.New("gotenancy/subscription: subscription already exists")

	// ErrInvalidSubscription reports invalid subscription metadata.
	ErrInvalidSubscription = errors.New("gotenancy/subscription: invalid subscription")

	// ErrInvalidTransition reports an invalid subscription lifecycle transition.
	ErrInvalidTransition = errors.New("gotenancy/subscription: invalid transition")
)
