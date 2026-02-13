package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// BaseModel contains common fields for all models
type BaseModel struct {
	ID        uuid.UUID    `json:"id"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
	DeletedAt sql.NullTime `json:"deleted_at,omitempty"`
}

// IsDeleted returns true if the model is soft-deleted
func (b *BaseModel) IsDeleted() bool {
	return b.DeletedAt.Valid
}
