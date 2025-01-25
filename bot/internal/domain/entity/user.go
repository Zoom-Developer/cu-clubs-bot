package entity

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type Role string

type Roles []Role

func (r *Roles) Value() (driver.Value, error) {
	return json.Marshal(r)
}

func (r *Roles) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal roles: %v", value)
	}
	return json.Unmarshal(bytes, r)
}

const (
	ExternalUser Role = "external_user"
	GrantUser    Role = "grant_user"
	Student      Role = "student"
)

type User struct {
	ID           int64 `gorm:"primaryKey"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Localization string `gorm:"default:ru"`
	Role         Role   `gorm:"not null"`
	Email        string `gorm:"uniqueIndex:idx_users_email,where:email <> ''"`
	FIO          string `gorm:"not null"`
	QRCodeID     string
	QRFileID     string
	IsBanned     bool   `gorm:"default:false"`
	Clubs        []Club `gorm:"many2many:club_owners;foreignKey:ID;joinForeignKey:UserID;References:ID;JoinReferences:ClubID"`
}

type ClubOwner struct {
	UserID    int64  `gorm:"primaryKey"`
	ClubID    string `gorm:"primaryKey;type:uuid"`
	CreatedAt time.Time
}

type EventParticipant struct {
	EventID   string `gorm:"primaryKey;type:uuid"`
	UserID    int64  `gorm:"primaryKey"`
	CreatedAt time.Time
	IsUserQr  bool
	IsEventQr bool
}
