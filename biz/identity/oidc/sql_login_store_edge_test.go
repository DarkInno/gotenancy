package oidc

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"testing"
	"time"
)

func TestSQLLoginStoreHelpersHandleZeroValuesAndMalformedRows(t *testing.T) {
	var store SQLLoginStore
	if got := store.effectiveTTL(); got != DefaultLoginTTL {
		t.Fatalf("effectiveTTL() = %s, want %s", got, DefaultLoginTTL)
	}
	if store.currentTime().IsZero() {
		t.Fatal("currentTime() returned zero time")
	}
	if got := nullableLoginString(""); got != nil {
		t.Fatalf("nullableLoginString(empty) = %#v, want nil", got)
	}
	if got := nullableLoginString("user-a"); got != "user-a" {
		t.Fatalf("nullableLoginString(value) = %#v, want user-a", got)
	}

	raw, err := marshalLoginRoles(nil)
	if err != nil {
		t.Fatalf("marshalLoginRoles(nil) error = %v", err)
	}
	if raw != "[]" {
		t.Fatalf("marshalLoginRoles(nil) = %q, want []", raw)
	}
	roles, err := unmarshalLoginRoles("")
	if err != nil {
		t.Fatalf("unmarshalLoginRoles(empty) error = %v", err)
	}
	if !reflect.DeepEqual(roles, []string{}) {
		t.Fatalf("unmarshalLoginRoles(empty) = %#v, want empty slice", roles)
	}
	if _, err := unmarshalLoginRoles("not-json"); err == nil {
		t.Fatal("unmarshalLoginRoles(invalid) error = nil, want error")
	}

	scanErr := errors.New("scan failed")
	if _, err := scanLogin(loginScannerFunc(func(...any) error { return scanErr })); !errors.Is(err, scanErr) {
		t.Fatalf("scanLogin(scan failure) error = %v, want scan failure", err)
	}
	if _, err := scanLogin(loginScannerFunc(func(dest ...any) error {
		*(dest[0].(*string)) = "state"
		*(dest[1].(*string)) = "https://issuer.example.com/authorize"
		*(dest[2].(*string)) = "nonce"
		*(dest[3].(*string)) = "verifier"
		*(dest[4].(*string)) = ""
		*(dest[6].(*string)) = "[]"
		*(dest[7].(*time.Time)) = time.Now()
		return nil
	})); !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("scanLogin(invalid row) error = %v, want ErrInvalidConfig", err)
	}
}

func TestOIDCHelpersHandleInvalidInputsAndNilClient(t *testing.T) {
	if _, err := randomURLValue(0); !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("randomURLValue(0) error = %v, want ErrInvalidConfig", err)
	}
	if isLoopbackHost("example.com") {
		t.Fatal("isLoopbackHost(example.com) = true, want false")
	}
	if _, err := callbackValuesFromRequest(nil); !errors.Is(err, ErrInvalidCallback) {
		t.Fatalf("callbackValuesFromRequest(nil) error = %v, want ErrInvalidCallback", err)
	}

	ctx := context.Background()
	var nilClient *Client
	if got := nilClient.context(ctx); got != ctx {
		t.Fatal("nil client context() did not preserve context")
	}
	client := &Client{httpClient: &http.Client{}}
	if got := client.context(ctx); got == ctx {
		t.Fatal("configured client context() did not attach the HTTP client")
	}
}
