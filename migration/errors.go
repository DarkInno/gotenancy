package migration

import "errors"

var (
	// ErrInvalidIdentifier reports an unsafe SQL identifier.
	ErrInvalidIdentifier = errors.New("gotenancy/migration: invalid identifier")

	// ErrInvalidMigration reports incomplete migration input.
	ErrInvalidMigration = errors.New("gotenancy/migration: invalid migration")

	// ErrUnsupportedDialect reports a dialect not supported by the planner.
	ErrUnsupportedDialect = errors.New("gotenancy/migration: unsupported dialect")
)
