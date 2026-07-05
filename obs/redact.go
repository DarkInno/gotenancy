package obs

import "strings"

var sensitiveKeys = map[string]struct{}{
	"password":     {},
	"secret":       {},
	"token":        {},
	"api_key":      {},
	"apikey":       {},
	"access_token": {},
}

// Redact removes sensitive values from a string map.
func Redact(fields map[string]string) map[string]string {
	if fields == nil {
		return nil
	}

	redacted := make(map[string]string, len(fields))
	for key, value := range fields {
		if _, ok := sensitiveKeys[strings.ToLower(key)]; ok {
			redacted[key] = "[REDACTED]"
			continue
		}
		redacted[key] = value
	}
	return redacted
}
