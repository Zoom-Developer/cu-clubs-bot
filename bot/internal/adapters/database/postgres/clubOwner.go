package postgres

import (
	"context"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"
	"gorm.io/gorm"
)

type ClubOwnerStorage struct {
	db *gorm.DB
}

func NewClubOwnerStorage(db *gorm.DB) *ClubOwnerStorage {
	return &ClubOwnerStorage{
		db: db,
	}
}

func (s *ClubOwnerStorage) Create(ctx context.Context, clubOwner *entity.ClubOwner) (*entity.ClubOwner, error) {
	err := s.db.WithContext(ctx).Create(&clubOwner).Error
	return clubOwner, err
}

func (s *ClubOwnerStorage) Delete(ctx context.Context, userID int64, clubID string) error {
	err := s.db.WithContext(ctx).Where("club_id = ? AND user_id = ?", clubID, userID).Delete(&entity.ClubOwner{}).Error
	return err
}

func (s *ClubOwnerStorage) Get(ctx context.Context, clubID string, userID int64) (*entity.ClubOwner, error) {
	var clubOwner entity.ClubOwner
	err := s.db.WithContext(ctx).Where("club_id = ? AND user_id = ?", clubID, userID).First(&clubOwner).Error
	return &clubOwner, err
}

func (s *ClubOwnerStorage) Update(ctx context.Context, clubOwner *entity.ClubOwner) (*entity.ClubOwner, error) {
	err := s.db.WithContext(ctx).Save(&clubOwner).Error
	return clubOwner, err
}

func (s *ClubOwnerStorage) GetByClubID(ctx context.Context, clubID string) ([]entity.ClubOwner, error) {
	var clubOwners []entity.ClubOwner
	err := s.db.WithContext(ctx).Where("club_id = ?", clubID).Find(&clubOwners).Error
	return clubOwners, err
}

func (s *ClubOwnerStorage) GetByUserID(ctx context.Context, userID int64) ([]entity.ClubOwner, error) {
	var clubOwners []entity.ClubOwner
	err := s.db.WithContext(ctx).Where("user_id = ?", userID).Find(&clubOwners).Error
	return clubOwners, err
}
