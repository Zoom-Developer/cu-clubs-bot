package dto

import (
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"
)

type EventUser struct {
	ID        int64
	Username  string
	Email     string
	FIO       string
	UserVisit bool
}

func NewEventUserFromEntity(user entity.User, userVisit bool) EventUser {
	return EventUser{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		FIO:       user.FIO,
		UserVisit: userVisit,
	}
}
