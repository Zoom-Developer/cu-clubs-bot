package entity

import "time"

type NotificationType string

const (
	NotificationTypeDay  NotificationType = "day"
	NotificationTypeHour NotificationType = "hour"
)

// EventNotification represents a notification that has been sent to a user
type EventNotification struct {
	ID        string           `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	EventID   string           `gorm:"not null"`
	UserID    int64            `gorm:"not null"`
	Type      NotificationType `gorm:"not null"`
	CreatedAt time.Time        `gorm:"not null"`

	Event Event `gorm:"foreignKey:EventID"`
	User  User  `gorm:"foreignKey:UserID"`
}
