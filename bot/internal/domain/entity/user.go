package entity

import "gorm.io/gorm"

type User struct {
	gorm.Model
	FirstName    string
	Username     string
	Localisation string
	Banned       bool
	//Subscriptions []Subscription `gorm:"foreignKey:UserID"`
}
