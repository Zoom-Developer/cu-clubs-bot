package service

import (
	"context"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"
)

type EventStorage interface {
	Create(ctx context.Context, event *entity.Event) (*entity.Event, error)
	Get(ctx context.Context, id string) (*entity.Event, error)
	GetMany(ctx context.Context, ids []string) ([]entity.Event, error)
	GetAll(ctx context.Context) ([]entity.Event, error)
	Update(ctx context.Context, event *entity.Event) (*entity.Event, error)
	Count(ctx context.Context) (int64, error)
	GetWithPagination(ctx context.Context, limit, offset int, order string, role string) ([]entity.Event, error)
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

func (s *EventService) Get(ctx context.Context, id string) (*entity.Event, error) {
	return s.eventStorage.Get(ctx, id)
}

func (s *EventService) GetMany(ctx context.Context, ids []string) ([]entity.Event, error) {
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

func (s *EventService) GetWithPagination(ctx context.Context, limit, offset int, order string, role entity.Role) ([]entity.Event, error) {
	return s.eventStorage.GetWithPagination(ctx, limit, offset, order, string(role))
}
