package postgres

import (
	"context"
	"fmt"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"
	"gorm.io/gorm"
)

type EventParticipantStorage struct {
	db *gorm.DB
}

func NewEventParticipantStorage(db *gorm.DB) *EventParticipantStorage {
	return &EventParticipantStorage{
		db: db,
	}
}

func (s *EventParticipantStorage) Create(ctx context.Context, eventParticipant *entity.EventParticipant) (*entity.EventParticipant, error) {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Check if event exists
		var eventExists int64
		if err := tx.Model(&entity.Event{}).Where("id = ?", eventParticipant.EventID).Count(&eventExists).Error; err != nil {
			return err
		}
		if eventExists == 0 {
			return fmt.Errorf("event with id %s not found", eventParticipant.EventID)
		}

		// Check if user exists
		var userExists int64
		if err := tx.Model(&entity.User{}).Where("id = ?", eventParticipant.UserID).Count(&userExists).Error; err != nil {
			return err
		}
		if userExists == 0 {
			return fmt.Errorf("user with id %d not found", eventParticipant.UserID)
		}

		// Create participant
		return tx.Create(&eventParticipant).Error
	})

	return eventParticipant, err
}

func (s *EventParticipantStorage) Get(ctx context.Context, eventID string, userID int64) (*entity.EventParticipant, error) {
	var eventParticipant entity.EventParticipant
	err := s.db.WithContext(ctx).Where("event_id = ? AND user_id = ?", eventID, userID).First(&eventParticipant).Error
	return &eventParticipant, err
}

func (s *EventParticipantStorage) Update(ctx context.Context, eventParticipant *entity.EventParticipant) (*entity.EventParticipant, error) {
	err := s.db.WithContext(ctx).Save(&eventParticipant).Error
	return eventParticipant, err
}

func (s *EventParticipantStorage) Delete(ctx context.Context, eventID string, userID int64) error {
	err := s.db.WithContext(ctx).Where("event_id = ? AND user_id = ?", eventID, userID).Delete(&entity.EventParticipant{}).Error
	return err
}

func (s *EventParticipantStorage) GetByEventID(ctx context.Context, eventID string) ([]entity.EventParticipant, error) {
	var eventParticipants []entity.EventParticipant
	err := s.db.WithContext(ctx).Where("event_id = ?", eventID).Find(&eventParticipants).Error
	return eventParticipants, err
}

func (s *EventParticipantStorage) CountByEventID(ctx context.Context, eventID string) (int64, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&entity.EventParticipant{}).Where("event_id = ?", eventID).Count(&count).Error
	return count, err
}
