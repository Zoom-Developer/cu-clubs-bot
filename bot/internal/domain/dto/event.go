package dto

import (
	"time"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"
)

type Event struct {
	ID              string
	Name            string
	Description     string
	ClubID          string
	AllowedRoles    []string
	MaxParticipants int
	RegistrationEnd time.Time
	StartTime       time.Time
	EndTime         time.Time
	IsRegistered    bool
}

func NewEventFromEntity(event entity.Event, isRegistered bool) Event {
	return Event{
		ID:              event.ID,
		Name:            event.Name,
		Description:     event.Description,
		ClubID:          event.ClubID,
		AllowedRoles:    event.AllowedRoles,
		MaxParticipants: event.MaxParticipants,
		RegistrationEnd: event.RegistrationEnd,
		StartTime:       event.StartTime,
		EndTime:         event.EndTime,
		IsRegistered:    isRegistered,
	}
}
