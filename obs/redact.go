package obs

import "strings"

const redactedValue = "[REDACTED]"

var sensitiveKeys = map[string]struct{}{
	"access_token":  {},
	"api_key":       {},
	"apikey":        {},
	"authorization": {},
	"client_secret": {},
	"cookie":        {},
	"id_token":      {},
	"password":      {},
	"private_key":   {},
	"refresh_token": {},
	"secret":        {},
	"set_cookie":    {},
	"token":         {},
}

// Redact removes sensitive values from a string map.
func Redact(fields map[string]string) map[string]string {
	if fields == nil {
		return nil
	}

	redacted := make(map[string]string, len(fields))
	for key, value := range fields {
		if isSensitiveKey(key) {
			redacted[key] = redactedValue
			continue
		}
		redacted[key] = value
	}
	return redacted
}

func isSensitiveKey(key string) bool {
	normalized := strings.ToLower(strings.TrimSpace(key))
	normalized = strings.ReplaceAll(normalized, "-", "_")
	_, ok := sensitiveKeys[normalized]
	return ok
}
