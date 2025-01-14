package postgres

import (
	"context"
	"github.com/Badsnus/cu-clubs-bot/internal/domain/entity"
	"gorm.io/gorm"
)

type userStorage struct {
	db *gorm.DB
}

func NewUserStorage(db *gorm.DB) *userStorage {
	return &userStorage{
		db: db,
	}
}

// Create is a function that creates a new user in the database.
func (s *userStorage) Create(ctx context.Context, user *entity.User) (*entity.User, error) {
	err := s.db.WithContext(ctx).Create(&user).Error
	return user, err
}

// Get is a function that gets a user from the database by id.
func (s *userStorage) Get(ctx context.Context, id uint) (*entity.User, error) {
	var user entity.User
	err := s.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	return &user, err
}

// GetAll is a function that gets all users from the database.
func (s *userStorage) GetAll(ctx context.Context) ([]entity.User, error) {
	var users []entity.User
	err := s.db.WithContext(ctx).Find(&users).Error
	return users, err
}

// Update is a function that updates a user in the database.
func (s *userStorage) Update(ctx context.Context, user *entity.User) (*entity.User, error) {
	err := s.db.WithContext(ctx).Save(&user).Error
	return user, err
}

// Count is a function that gets the count of users from the database.
func (s *userStorage) Count(ctx context.Context) (int64, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&entity.User{}).Count(&count).Error
	return count, err
}

// GetWithPagination is a function that gets a list of users from the database with pagination.
func (s *userStorage) GetWithPagination(ctx context.Context, offset, limit int, order string) ([]entity.User, error) {
	var users []entity.User
	err := s.db.WithContext(ctx).Order(order).Offset(offset).Limit(limit).Find(&users).Error
	return users, err
}
