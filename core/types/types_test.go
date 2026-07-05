package types

import (
	"errors"
	"testing"
)

func TestTenantIDString(t *testing.T) {
	id := TenantID("tenant-a")
	if got := id.String(); got != "tenant-a" {
		t.Fatalf("TenantID.String() = %q, want %q", got, "tenant-a")
	}
}

func TestParseTenantIDStringStrategy(t *testing.T) {
	id, err := ParseTenantID(" tenant-a ", TenantIDStrategyString)
	if err != nil {
		t.Fatalf("ParseTenantID() error = %v", err)
	}
	if id != "tenant-a" {
		t.Fatalf("ParseTenantID() = %q, want %q", id, "tenant-a")
	}
}

func TestParseTenantIDIntStrategy(t *testing.T) {
	id, err := ParseTenantID("42", TenantIDStrategyInt)
	if err != nil {
		t.Fatalf("ParseTenantID() error = %v", err)
	}
	if id != "42" {
		t.Fatalf("ParseTenantID() = %q, want %q", id, "42")
	}

	value, err := id.Int64()
	if err != nil {
		t.Fatalf("TenantID.Int64() error = %v", err)
	}
	if value != 42 {
		t.Fatalf("TenantID.Int64() = %d, want 42", value)
	}

	if got := NewTenantIDFromInt(7); got != "7" {
		t.Fatalf("NewTenantIDFromInt() = %q, want %q", got, "7")
	}
}

func TestParseTenantIDUUIDStrategy(t *testing.T) {
	id, err := ParseTenantID("A0EebC99-9C0B-4EF8-BB6D-6BB9BD380A11", TenantIDStrategyUUID)
	if err != nil {
		t.Fatalf("ParseTenantID() error = %v", err)
	}
	if id != "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11" {
		t.Fatalf("ParseTenantID() = %q, want canonical lowercase UUID", id)
	}
}

func TestParseTenantIDRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		strategy TenantIDStrategy
		wantErr  error
	}{
		{name: "empty", raw: " ", strategy: TenantIDStrategyString, wantErr: ErrEmptyTenantID},
		{name: "bad int", raw: "abc", strategy: TenantIDStrategyInt, wantErr: ErrInvalidTenantID},
		{name: "bad uuid", raw: "not-a-uuid", strategy: TenantIDStrategyUUID, wantErr: ErrInvalidTenantID},
		{name: "unknown strategy", raw: "tenant-a", strategy: "snowflake", wantErr: ErrInvalidTenantID},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseTenantID(tt.raw, tt.strategy)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ParseTenantID() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestReservedEnumValues(t *testing.T) {
	statuses := []TenantStatus{
		TenantStatusPending,
		TenantStatusActive,
		TenantStatusSuspended,
		TenantStatusSoftDeleted,
		TenantStatusHardDeleted,
	}
	for _, status := range statuses {
		if status == "" {
			t.Fatal("tenant status must not be empty")
		}
	}

	sides := []MultiTenancySide{MultiTenancySideHost, MultiTenancySideTenant}
	for _, side := range sides {
		if side == "" {
			t.Fatal("multi-tenancy side must not be empty")
		}
	}
}
