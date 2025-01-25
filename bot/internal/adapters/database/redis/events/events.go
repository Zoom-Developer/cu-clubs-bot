package events

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"
	"time"

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

func (s *Storage) Get(userID int64) (entity.Event, error) {
	eventBytes, err := s.redis.Get(context.Background(), fmt.Sprintf("%d", userID)).Result()
	if err != nil {
		return entity.Event{}, nil
	}

	var event entity.Event
	if err = json.Unmarshal([]byte(eventBytes), &event); err != nil {
		return entity.Event{}, err
	}

	return event, nil
}

func (s *Storage) Set(userID int64, event entity.Event, expiration time.Duration) {
	eventBytes, _ := json.Marshal(event)
	s.redis.Set(context.Background(), fmt.Sprintf("%d", userID), eventBytes, expiration)
}

func (s *Storage) Clear(userID int64) {
	s.redis.Del(context.Background(), fmt.Sprintf("%d", userID))
}
