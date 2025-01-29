package entity

import (
	"fmt"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/utils/location"
	"github.com/lib/pq"
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
	AllowedRoles          pq.StringArray `gorm:"type:text[]"`
}

func (e *Event) IsOver(additionalTime time.Duration) bool {
	return e.StartTime.Before(time.Now().In(location.Location()).Add(-additionalTime))
}

func (e *Event) Link(botName string) string {
	return fmt.Sprintf("https://t.me/%s?start=event_%s", botName, e.ID)
}
