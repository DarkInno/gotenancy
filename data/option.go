package data

import (
	"fmt"
	"strings"
)

const (
	// DefaultTenantField is the default database field used for tenant isolation.
	DefaultTenantField = "tenant_id"
)

type filterOptions struct {
	tenantField        string
	softDeleteField    string
	includeSoftDeleted bool
}

// FilterOption configures a Filter.
type FilterOption func(*filterOptions) error

// WithTenantField overrides the tenant field name.
func WithTenantField(field string) FilterOption {
	return func(opts *filterOptions) error {
		if !isSafeFieldName(field) {
			return fmt.Errorf("%w: %q", ErrInvalidFieldName, field)
		}
		opts.tenantField = field
		return nil
	}
}

// WithSoftDeleteField adds an IS NULL condition for soft-deleted records.
func WithSoftDeleteField(field string) FilterOption {
	return func(opts *filterOptions) error {
		if !isSafeFieldName(field) {
			return fmt.Errorf("%w: %q", ErrInvalidFieldName, field)
		}
		opts.softDeleteField = field
		return nil
	}
}

// WithIncludeSoftDeleted disables the soft-delete IS NULL condition.
func WithIncludeSoftDeleted(include bool) FilterOption {
	return func(opts *filterOptions) error {
		opts.includeSoftDeleted = include
		return nil
	}
}

func defaultFilterOptions() filterOptions {
	return filterOptions{tenantField: DefaultTenantField}
}

func isSafeFieldName(value string) bool {
	if value == "" {
		return false
	}

	parts := strings.Split(value, ".")
	for _, part := range parts {
		if !isSafeIdentifier(part) {
			return false
		}
	}
	return true
}

func isSafeIdentifier(value string) bool {
	if value == "" {
		return false
	}

	for i, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '_' {
			continue
		}
		if i > 0 && r >= '0' && r <= '9' {
			continue
		}
		return false
	}
	return true
}
