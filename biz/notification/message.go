package notification

import "gotenancy/core/types"

type Message struct {
	TenantID types.TenantID
	Channel  string
	To       string
	Subject  string
	Body     string
	Metadata map[string]string
}
