package entity

import "time"

type Club struct {
	ID          string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   time.Time
	Name        string `gorm:"not null"`
	Description string
}
