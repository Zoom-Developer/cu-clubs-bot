package entity

import (
	"gorm.io/gorm"
	"time"
)

type Event struct {
	ID                    string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	CreatedAt             time.Time
	UpdatedAt             time.Time
	DeletedAt             gorm.DeletedAt
	ClubID                string `gorm:"not null;type:uuid"`
	Club                  Club
	Name                  string `gorm:"not null"`
	Description           string `gorm:"not null"`
	AfterRegistrationText string
	Location              string    `gorm:"not null"`
	StartTime             time.Time `gorm:"not null"`
	EndTime               time.Time
	RegistrationEnd       time.Time `gorm:"not null"`
	MaxParticipants       int
	ExpectedParticipants  int
}
