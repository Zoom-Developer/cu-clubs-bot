package redis

import (
	"context"
	"errors"
	"fmt"
	"github.com/Badsnus/cu-clubs-bot/cmd/bot"
	"github.com/Badsnus/cu-clubs-bot/internal/domain/common/errorz"
	"time"
)

type StatesStorage struct {
	redis *redis.Client
}

func NewStatesStorage(b *bot.Bot) *StatesStorage {
	return &StatesStorage{
		redis: b.Redis,
	}
}

type State struct {
	State        string
	StateContext string
}

func (s *StatesStorage) Get(userID int64) (State, error) {
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

	return State{}, errorz.InvalidState
}

func (s *StatesStorage) Set(userID int64, state string, stateContext string) {
	s.redis.Set(context.Background(), fmt.Sprintf("%d", userID), fmt.Sprintf("%s:%s", state, stateContext), time.Minute*45)
}

func (s *StatesStorage) Clear(userID int64) {
	s.redis.Del(context.Background(), fmt.Sprintf("%d", userID))
}
