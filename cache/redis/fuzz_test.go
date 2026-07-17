package redis

import (
	"errors"
	"strings"
	"testing"
)

func FuzzNewFromURL(f *testing.F) {
	f.Add("redis://localhost:6379/0")
	f.Add("redis://cache-user:super-secret@localhost/%zz")

	f.Fuzz(func(t *testing.T, rawURL string) {
		cache, err := NewFromURL(rawURL)
		if err != nil {
			if !errors.Is(err, ErrInvalidRedisConfig) {
				t.Fatalf("NewFromURL(%q) error = %v, want ErrInvalidRedisConfig", rawURL, err)
			}
			if strings.Contains(rawURL, "super-secret") && strings.Contains(err.Error(), "super-secret") {
				t.Fatalf("NewFromURL(%q) error leaked password: %q", rawURL, err)
			}
			return
		}

		if closeErr := cache.Close(); closeErr != nil {
			t.Fatalf("NewFromURL(%q) Close() error = %v", rawURL, closeErr)
		}
	})
}
