package service

import (
	"context"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"
)

type EventParticipantStorage interface {
	Create(ctx context.Context, eventParticipant *entity.EventParticipant) (*entity.EventParticipant, error)
	Get(ctx context.Context, eventID string, userID int64) (*entity.EventParticipant, error)
	Update(ctx context.Context, eventParticipant *entity.EventParticipant) (*entity.EventParticipant, error)
	Delete(ctx context.Context, eventID string, userID int64) error
	GetByEventID(ctx context.Context, eventID string) ([]entity.EventParticipant, error)
	CountByEventID(ctx context.Context, eventID string) (int64, error)
}

type EventParticipantService struct {
	storage EventParticipantStorage
}

func NewEventParticipantService(storage EventParticipantStorage) *EventParticipantService {
	return &EventParticipantService{storage}
}

func (s *EventParticipantService) Create(ctx context.Context, eventParticipant *entity.EventParticipant) (*entity.EventParticipant, error) {
	return s.storage.Create(ctx, eventParticipant)
}

func (s *EventParticipantService) Get(ctx context.Context, eventID string, userID int64) (*entity.EventParticipant, error) {
	return s.storage.Get(ctx, eventID, userID)
}

func (s *EventParticipantService) Update(ctx context.Context, eventParticipant *entity.EventParticipant) (*entity.EventParticipant, error) {
	return s.storage.Update(ctx, eventParticipant)
}

func (s *EventParticipantService) Delete(ctx context.Context, eventID string, userID int64) error {
	return s.storage.Delete(ctx, eventID, userID)
}

func (s *EventParticipantService) GetByEventID(ctx context.Context, eventID string) ([]entity.EventParticipant, error) {
	return s.storage.GetByEventID(ctx, eventID)
}

func (s *EventParticipantService) CountByEventID(ctx context.Context, eventID string) (int64, error) {
	return s.storage.CountByEventID(ctx, eventID)
}
