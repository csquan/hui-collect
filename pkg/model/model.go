package model

import (
	"database/sql"
)

// Model
// Deprecated
// time.Time AutoMigrate to int64 failed. plan b
type Model struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt int64
	UpdatedAt int64
	DeletedAt sql.NullInt64 `gorm:"index"`
}
