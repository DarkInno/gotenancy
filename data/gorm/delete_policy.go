package gormtenant

// DeletePolicy describes the physical delete behavior for tenant-aware models.
type DeletePolicy int

const (
	// DeleteSoft lets GORM soft-delete support keep rows and mark them deleted.
	DeleteSoft DeletePolicy = iota

	// DeleteHard physically removes rows.
	DeleteHard
)
