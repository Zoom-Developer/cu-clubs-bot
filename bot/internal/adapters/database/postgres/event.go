package postgres

import (
	"context"
	"time"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/dto"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"
	"gorm.io/gorm"
)

type EventStorage struct {
	db *gorm.DB
}

func NewEventStorage(db *gorm.DB) *EventStorage {
	return &EventStorage{
		db: db,
	}
}

// Create is a function that creates a new event in the database.
func (s *EventStorage) Create(ctx context.Context, event *entity.Event) (*entity.Event, error) {
	err := s.db.WithContext(ctx).Create(&event).Error
	return event, err
}

// Get is a function that gets an event from the database by id.
func (s *EventStorage) Get(ctx context.Context, id string) (*entity.Event, error) {
	var event entity.Event
	err := s.db.WithContext(ctx).Where("id = ?", id).First(&event).Error
	return &event, err
}

func (s *EventStorage) GetMany(ctx context.Context, ids []string) ([]entity.Event, error) {
	var events []entity.Event
	err := s.db.WithContext(ctx).Where("id IN ?", ids).Find(&events).Error
	return events, err
}

// GetAll is a function that gets all events from the database.
func (s *EventStorage) GetAll(ctx context.Context) ([]entity.Event, error) {
	var events []entity.Event
	err := s.db.WithContext(ctx).Find(&events).Error
	return events, err
}

// Update is a function that updates an event in the database.
func (s *EventStorage) Update(ctx context.Context, event *entity.Event) (*entity.Event, error) {
	err := s.db.WithContext(ctx).Save(&event).Error
	return event, err
}

// Count is a function that gets the count of events from the database.
func (s *EventStorage) Count(ctx context.Context, role string) (int64, error) {
	var count int64
	query := s.db.WithContext(ctx).Model(&entity.Event{}).
		Where("registration_end > ?", time.Now()).
		Where("? = ANY(allowed_roles)", role)

	err := query.Count(&count).Error
	return count, err
}

// GetWithPagination is a function that gets a list of events from the database with pagination. (if role is empty, it will return all events)
func (s *EventStorage) GetWithPagination(ctx context.Context, limit, offset int, order string, role string, userID int64) ([]dto.Event, error) {
	var events []struct {
		entity.Event
		IsRegistered bool
	}

	query := s.db.WithContext(ctx).
		Table("events").
		Select("events.*, CASE WHEN ep.user_id IS NOT NULL THEN true ELSE false END as is_registered").
		Joins("LEFT JOIN event_participants ep ON events.id = ep.event_id AND ep.user_id = ?", userID).
		Where("registration_end > ?", time.Now())

	if role != "" {
		query = query.Where("? = ANY(allowed_roles)", role)
	}

	err := query.Order(order).
		Limit(limit).
		Offset(offset).
		Find(&events).Error
	if err != nil {
		return nil, err
	}

	result := make([]dto.Event, len(events))
	for i, event := range events {
		result[i] = dto.NewEventFromEntity(event.Event, event.IsRegistered)
	}

	return result, nil
}
