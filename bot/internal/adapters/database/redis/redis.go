package redis

import (
	"context"
	"fmt"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/database/redis/callbacks"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/database/redis/codes"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/database/redis/emails"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/database/redis/events"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/database/redis/states"
	"github.com/redis/go-redis/v9"
)

type Client struct {
	States    *states.Storage
	Codes     *codes.Storage
	Emails    *emails.Storage
	Events    *events.Storage
	Callbacks *callbacks.Storage
}

type Options struct {
	Host     string
	Port     string
	Password string
}

func New(opts Options) (*Client, error) {
	stateRedis := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", opts.Host, opts.Port),
		Password: opts.Password,
		DB:       0,
	})
	if err := stateRedis.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping state storage: %w", err)
	}

	codeRedis := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", opts.Host, opts.Port),
		Password: opts.Password,
		DB:       1,
	})
	if err := codeRedis.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping codes storage: %w", err)
	}

	emailRedis := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", opts.Host, opts.Port),
		Password: opts.Password,
		DB:       2,
	})
	if err := emailRedis.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping email storage: %w", err)
	}
	eventsRedis := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", opts.Host, opts.Port),
		Password: opts.Password,
		DB:       3,
	})
	if err := eventsRedis.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping events storage: %w", err)
	}
	callbacksRedis := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", opts.Host, opts.Port),
		Password: opts.Password,
		DB:       4,
	})
	if err := callbacksRedis.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping callbacks storage: %w", err)
	}

	return &Client{
		States:    states.NewStorage(stateRedis),
		Codes:     codes.NewStorage(codeRedis),
		Emails:    emails.NewStorage(emailRedis),
		Events:    events.NewStorage(eventsRedis),
		Callbacks: callbacks.NewStorage(callbacksRedis),
	}, nil
}
