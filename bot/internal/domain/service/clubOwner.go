package service

import (
	"context"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/dto"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"
)

type ClubOwnerStorage interface {
	Create(ctx context.Context, clubOwner *entity.ClubOwner) (*entity.ClubOwner, error)
	Delete(ctx context.Context, userID int64, clubID string) error
	Get(ctx context.Context, clubID string, userID int64) (*entity.ClubOwner, error)
	GetByClubID(ctx context.Context, clubID string) ([]entity.ClubOwner, error)
	GetByUserID(ctx context.Context, userID int64) ([]entity.ClubOwner, error)
}

type ClubOwnerService struct {
	storage     ClubOwnerStorage
	userStorage UserStorage
}

func NewClubOwnerService(storage ClubOwnerStorage, userStorage UserStorage) *ClubOwnerService {
	return &ClubOwnerService{
		storage:     storage,
		userStorage: userStorage,
	}
}

func (s *ClubOwnerService) Add(ctx context.Context, userID int64, clubID string) (*entity.ClubOwner, error) {
	return s.storage.Create(ctx, &entity.ClubOwner{UserID: userID, ClubID: clubID})
}

func (s *ClubOwnerService) Remove(ctx context.Context, userID int64, clubID string) error {
	return s.storage.Delete(ctx, userID, clubID)
}

func (s *ClubOwnerService) Get(ctx context.Context, clubID string, userID int64) (*entity.ClubOwner, error) {
	return s.storage.Get(ctx, clubID, userID)
}

func (s *ClubOwnerService) GetByClubID(ctx context.Context, clubID string) ([]dto.ClubOwner, error) {
	clubOwners, err := s.storage.GetByClubID(ctx, clubID)
	if err != nil {
		return nil, err
	}
	var userIDs []int64
	for _, clubOwner := range clubOwners {
		userIDs = append(userIDs, clubOwner.UserID)
	}

	users, err := s.userStorage.GetMany(ctx, userIDs)
	if err != nil {
		return nil, err
	}
	var result []dto.ClubOwner
	for _, user := range users {
		for _, clubOwner := range clubOwners {
			if user.ID == clubOwner.UserID {
				result = append(result, dto.ClubOwner{
					ClubID:   clubOwner.ClubID,
					UserID:   user.ID,
					FIO:      user.FIO,
					Email:    user.Email,
					Role:     user.Role,
					IsBanned: user.IsBanned,
				})
			}
		}
	}
	return result, nil
}

func (s *ClubOwnerService) GetByUserID(ctx context.Context, userID int64) ([]dto.ClubOwner, error) {
	clubOwners, err := s.storage.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var userIDs []int64
	for _, clubOwner := range clubOwners {
		userIDs = append(userIDs, clubOwner.UserID)
	}

	users, err := s.userStorage.GetMany(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	var result []dto.ClubOwner
	for _, user := range users {
		for _, clubOwner := range clubOwners {
			if user.ID == clubOwner.UserID {
				result = append(result, dto.ClubOwner{
					ClubID:   clubOwner.ClubID,
					UserID:   user.ID,
					FIO:      user.FIO,
					Email:    user.Email,
					Role:     user.Role,
					IsBanned: user.IsBanned,
				})
			}
		}
	}
	return result, nil
}
