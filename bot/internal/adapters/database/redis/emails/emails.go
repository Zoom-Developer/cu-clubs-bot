package emails

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/common/errorz"
	"github.com/redis/go-redis/v9"
)

type Storage struct {
	redis *redis.Client
}

func NewStorage(client *redis.Client) *Storage {
	return &Storage{
		redis: client,
	}
}

type Email struct {
	Email        string
	EmailContext string
}

func (s *Storage) Get(userID int64) (Email, error) {
	emailData, err := s.redis.Get(context.Background(), fmt.Sprintf("%d", userID)).Result()
	if err != nil {
		return Email{}, err
	}
	emailSlice := strings.Split(emailData, ":")
	if len(emailSlice) == 1 {
		return Email{
			Email:        emailSlice[0],
			EmailContext: "",
		}, nil
	}

	if len(emailSlice) == 2 {
		return Email{
			Email:        emailSlice[0],
			EmailContext: emailSlice[1],
		}, nil
	}

	return Email{}, errorz.ErrInvalidCode
}

func (s *Storage) Set(userID int64, code string, codeContext string, expiration time.Duration) {
	s.redis.Set(context.Background(), fmt.Sprintf("%d", userID), fmt.Sprintf("%s:%s", code, codeContext), expiration)
}

func (s *Storage) Clear(userID int64) {
	s.redis.Del(context.Background(), fmt.Sprintf("%d", userID))
}
