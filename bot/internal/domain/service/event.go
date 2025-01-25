package service

import (
	"context"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"
)

type EventStorage interface {
	Create(ctx context.Context, event *entity.Event) (*entity.Event, error)
	Get(ctx context.Context, id uint) (*entity.Event, error)
	GetMany(ctx context.Context, ids []int64) ([]entity.Event, error)
	GetAll(ctx context.Context) ([]entity.Event, error)
	Update(ctx context.Context, event *entity.Event) (*entity.Event, error)
	Count(ctx context.Context) (int64, error)
	GetWithPagination(ctx context.Context, offset, limit int, order string) ([]entity.Event, error)
}

type EventService struct {
	eventStorage EventStorage
}

func NewEventService(storage EventStorage) *EventService {
	return &EventService{
		eventStorage: storage,
	}
}

func (s *EventService) Create(ctx context.Context, event *entity.Event) (*entity.Event, error) {
	return s.eventStorage.Create(ctx, event)
}
func (s *EventService) Get(ctx context.Context, id uint) (*entity.Event, error) {
	return s.eventStorage.Get(ctx, id)
}
func (s *EventService) GetMany(ctx context.Context, ids []int64) ([]entity.Event, error) {
	return s.eventStorage.GetMany(ctx, ids)
}
func (s *EventService) GetAll(ctx context.Context) ([]entity.Event, error) {
	return s.eventStorage.GetAll(ctx)
}
func (s *EventService) Update(ctx context.Context, event *entity.Event) (*entity.Event, error) {
	return s.eventStorage.Update(ctx, event)
}
func (s *EventService) Count(ctx context.Context) (int64, error) {
	return s.eventStorage.Count(ctx)
}
func (s *EventService) GetWithPagination(ctx context.Context, offset, limit int, order string) ([]entity.Event, error) {
	return s.eventStorage.GetWithPagination(ctx, offset, limit, order)
}
