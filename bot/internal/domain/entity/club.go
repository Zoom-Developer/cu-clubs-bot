package entity

import (
	"gorm.io/gorm"
	"time"
)

type Club struct {
	ID          string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt
	Name        string `gorm:"not null;unique"`
	Description string
}
