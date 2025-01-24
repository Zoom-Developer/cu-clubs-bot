package service

import (
	"context"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"
)

type ClubStorage interface {
	Create(ctx context.Context, club *entity.Club) (*entity.Club, error)
	GetWithPagination(ctx context.Context, offset, limit int, order string) ([]entity.Club, error)
	Get(ctx context.Context, id string) (*entity.Club, error)
	Update(ctx context.Context, club *entity.Club) (*entity.Club, error)
	Delete(ctx context.Context, id string) error
	Count(ctx context.Context) (int64, error)
}

type ClubService struct {
	storage ClubStorage
}

func NewClubService(storage ClubStorage) *ClubService {
	return &ClubService{
		storage: storage,
	}
}

func (s *ClubService) Create(ctx context.Context, club *entity.Club) (*entity.Club, error) {
	return s.storage.Create(ctx, club)
}

func (s *ClubService) GetWithPagination(ctx context.Context, offset, limit int, order string) ([]entity.Club, error) {
	return s.storage.GetWithPagination(ctx, offset, limit, order)
}

func (s *ClubService) Get(ctx context.Context, id string) (*entity.Club, error) {
	return s.storage.Get(ctx, id)
}

func (s *ClubService) Update(ctx context.Context, club *entity.Club) (*entity.Club, error) {
	return s.storage.Update(ctx, club)
}

func (s *ClubService) Delete(ctx context.Context, id string) error {
	return s.storage.Delete(ctx, id)
}

func (s *ClubService) Count(ctx context.Context) (int64, error) {
	return s.storage.Count(ctx)
}
