package service

import (
	"context"
	"github.com/Badsnus/cu-clubs-bot/internal/domain/entity"
)

type UserStorage interface {
	Create(ctx context.Context, user *entity.User) (*entity.User, error)
	Get(ctx context.Context, id uint) (*entity.User, error)
	GetAll(ctx context.Context) ([]entity.User, error)
	Update(ctx context.Context, user *entity.User) (*entity.User, error)
	Count(ctx context.Context) (int64, error)
	GetWithPagination(ctx context.Context, limit int, offset int, order string) ([]entity.User, error)
}

type UserService struct {
	userStorage UserStorage
}

func NewUserService(userStorage UserStorage) *UserService {
	return &UserService{
		userStorage: userStorage,
	}
}

func (s *UserService) Create(ctx context.Context, c tele.Context) (*entity.User, error) {
	var user entity.User
	user.ID = uint(c.Sender().ID)
	user.FirstName = c.Sender().FirstName
	user.Username = c.Sender().Username
	return s.userStorage.Create(ctx, &user)
}

func (s *UserService) Get(ctx context.Context, userID int64) (*entity.User, error) {
	return s.userStorage.Get(ctx, uint(userID))
}

func (s *UserService) GetAll(ctx context.Context) ([]entity.User, error) {
	return s.userStorage.GetAll(ctx)
}

func (s *UserService) Update(ctx context.Context, user *entity.User) (*entity.User, error) {
	return s.userStorage.Update(ctx, user)
}

func (s *UserService) UpdateData(ctx context.Context, c tele.Context) (*entity.User, error) {
	user, err := s.Get(ctx, c.Sender().ID)
	if err != nil {
		return nil, err
	}
	user.ID = uint(c.Sender().ID)
	user.Username = c.Sender().Username
	user.FirstName = c.Sender().FirstName

	return s.userStorage.Update(ctx, user)
}

func (s *UserService) UpdateLocalisation(ctx context.Context, userID int64, localisation string) (*entity.User, error) {
	user, err := s.Get(ctx, userID)
	if err != nil {
		return nil, err
	}
	user.Localisation = localisation
	return s.userStorage.Update(ctx, user)
}

func (s *UserService) Count(ctx context.Context) (int64, error) {
	return s.userStorage.Count(ctx)
}

func (s *UserService) GetWithPagination(ctx context.Context, limit int, offset int, order string) ([]entity.User, error) {
	return s.userStorage.GetWithPagination(ctx, limit, offset, order)
}

func (s *UserService) Ban(ctx context.Context, userID int64) (*entity.User, error) {
	user, err := s.userStorage.Get(ctx, uint(userID))
	if err != nil {
		return nil, err
	}
	user.Banned = !user.Banned
	return s.userStorage.Update(ctx, user)
}
