package postgres

import (
	"context"
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
func (s *EventStorage) Get(ctx context.Context, id uint) (*entity.Event, error) {
	var event entity.Event
	err := s.db.WithContext(ctx).Where("id = ?", id).First(&event).Error
	return &event, err
}

func (s *EventStorage) GetMany(ctx context.Context, ids []int64) ([]entity.Event, error) {
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
func (s *EventStorage) Count(ctx context.Context) (int64, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&entity.Event{}).Count(&count).Error
	return count, err
}

// GetWithPagination is a function that gets a list of events from the database with pagination.
func (s *EventStorage) GetWithPagination(ctx context.Context, offset, limit int, order string) ([]entity.Event, error) {
	var events []entity.Event
	err := s.db.WithContext(ctx).Order(order).Offset(offset).Limit(limit).Find(&events).Error
	return events, err
}
