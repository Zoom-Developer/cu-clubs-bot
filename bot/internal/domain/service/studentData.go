package service

import (
	"context"

	"github.com/Badsnus/cu-clubs-bot/internal/domain/entity"
)

type StudentDataStorage interface {
	GetByLogin(ctx context.Context, login string) (*entity.StudentData, error)
}

type StudentDataService struct {
	studentDataStorage StudentDataStorage
}

func NewStudentDataService(studentDataStorage StudentDataStorage) *StudentDataService {
	return &StudentDataService{
		studentDataStorage: studentDataStorage,
	}
}

func (s *StudentDataService) GetByLogin(ctx context.Context, login string) (*entity.StudentData, error) {
	return s.studentDataStorage.GetByLogin(ctx, login)
}
