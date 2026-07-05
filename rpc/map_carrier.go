package rpc

// MapCarrier is an in-memory metadata carrier.
type MapCarrier map[string]string

// Get returns metadata by key.
func (carrier MapCarrier) Get(key string) (string, bool) {
	value, ok := carrier[key]
	return value, ok
}

// Set stores metadata by key.
func (carrier MapCarrier) Set(key string, value string) {
	carrier[key] = value
}
