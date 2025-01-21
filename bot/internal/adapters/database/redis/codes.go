package redis

import (
	"context"
	"errors"
	"fmt"
	"github.com/Badsnus/cu-clubs-bot/bot/cmd/bot"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/common/errorz"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
)

type CodesStorage struct {
	redis *redis.Client
}

func NewCodesStorage(b *bot.Bot) *CodesStorage {
	return &CodesStorage{
		redis: b.CodeRedis,
	}
}

type Code struct {
	Code        string
	CodeContext string
}

func (s *CodesStorage) Get(userID int64) (Code, error) {
	codeData, err := s.redis.Get(context.Background(), fmt.Sprintf("%d", userID)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return Code{}, nil
		}
		return Code{}, err
	}
	codeSlice := strings.Split(codeData, ":")
	if len(codeSlice) == 1 {
		return Code{
			Code:        codeSlice[0],
			CodeContext: "",
		}, nil
	}

	if len(codeSlice) == 2 {
		return Code{
			Code:        codeSlice[0],
			CodeContext: codeSlice[1],
		}, nil
	}

	return Code{}, errorz.ErrInvalidCode
}

func (s *CodesStorage) Set(userID int64, code string, codeContext string, expiration time.Duration) {
	s.redis.Set(context.Background(), fmt.Sprintf("%d", userID), fmt.Sprintf("%s:%s", code, codeContext), expiration)
}

func (s *CodesStorage) Clear(userID int64) {
	s.redis.Del(context.Background(), fmt.Sprintf("%d", userID))
}
