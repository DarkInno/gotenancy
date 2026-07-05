package gormtenant

import (
	"context"

	"gorm.io/gorm"
)

// BulkCreate creates a slice after applying the tenant context through the plugin callbacks.
func BulkCreate(ctx context.Context, db *gorm.DB, values interface{}) *gorm.DB {
	return db.WithContext(ctx).Create(values)
}
