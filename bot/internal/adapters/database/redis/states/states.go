package states

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Badsnus/cu-clubs-bot/bot/cmd/bot"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/common/errorz"
	"github.com/redis/go-redis/v9"
)

type Storage struct {
	redis *redis.Client
}

func NewStorage(b *bot.Bot) *Storage {
	return &Storage{
		redis: b.StateRedis,
	}
}

type State struct {
	State        string
	StateContext string
}

func (s *Storage) Get(userID int64) (State, error) {
	stateData, err := s.redis.Get(context.Background(), fmt.Sprintf("%d", userID)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return State{}, nil
		}
		return State{}, err
	}
	stateSlice := strings.Split(stateData, ":")
	if len(stateSlice) == 1 {
		return State{
			State:        stateSlice[0],
			StateContext: "",
		}, nil
	}

	if len(stateSlice) == 2 {
		return State{
			State:        stateSlice[0],
			StateContext: stateSlice[1],
		}, nil
	}

	return State{}, errorz.ErrInvalidState
}

func (s *Storage) Set(userID int64, state string, stateContext string, expiration time.Duration) {
	s.redis.Set(context.Background(), fmt.Sprintf("%d", userID), fmt.Sprintf("%s:%s", state, stateContext), expiration)
}

func (s *Storage) Clear(userID int64) {
	s.redis.Del(context.Background(), fmt.Sprintf("%d", userID))
}
