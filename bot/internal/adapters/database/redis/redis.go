package redis

import (
	"context"
	"fmt"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/database/redis/codes"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/database/redis/emails"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/database/redis/events"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/database/redis/states"
	"github.com/redis/go-redis/v9"
)

type Client struct {
	States *states.Storage
	Codes  *codes.Storage
	Emails *emails.Storage
	Events *events.Storage
}

type Options struct {
	Host     string
	Port     string
	Password string
}

func New(opts Options) (*Client, error) {
	stateStorage := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", opts.Host, opts.Port),
		Password: opts.Password,
		DB:       0,
	})
	if err := stateStorage.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping state storage: %w", err)
	}

	codeStorage := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", opts.Host, opts.Port),
		Password: opts.Password,
		DB:       1,
	})
	if err := codeStorage.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping codes storage: %w", err)
	}

	emailStorage := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", opts.Host, opts.Port),
		Password: opts.Password,
		DB:       2,
	})
	if err := emailStorage.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping email storage: %w", err)
	}
	eventsStorage := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", opts.Host, opts.Port),
		Password: opts.Password,
		DB:       3,
	})
	if err := eventsStorage.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping events storage: %w", err)
	}

	return &Client{
		States: states.NewStorage(stateStorage),
		Codes:  codes.NewStorage(codeStorage),
		Emails: emails.NewStorage(emailStorage),
		Events: events.NewStorage(eventsStorage),
	}, nil
}
