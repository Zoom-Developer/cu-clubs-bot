package entity

import (
	"time"
)

type Role string

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
	Warnings  bool   `gorm:"default:false"`
	CreatedAt time.Time
}

type EventParticipant struct {
	EventID   string `gorm:"primaryKey;type:uuid"`
	UserID    int64  `gorm:"primaryKey"`
	CreatedAt time.Time
	IsUserQr  bool
	IsEventQr bool
}
