package bot

import (
	"sync"

	"github.com/Badsnus/cu-clubs-bot/internal/adapters/config"
	"github.com/Badsnus/cu-clubs-bot/internal/adapters/logger"
	"github.com/redis/go-redis/v9"
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/layout"
	"gorm.io/gorm"
)

type Bot struct {
	*tele.Bot
	Layout     *layout.Layout
	DB         *gorm.DB
	StateRedis *redis.Client
	CodeRedis  *redis.Client
	Logger     *logger.Logger
}

func New(config *config.Config) (*Bot, error) {
	lt, err := layout.New("telegram.yml")
	if err != nil {
		return nil, err
	}

	settings := lt.Settings()
	botLogger, err := logger.GetPrefixed("bot")
	if err != nil {
		return nil, err
	}
	settings.OnError = func(err error, ctx tele.Context) {
		botLogger.Errorf("(user: %d) | unique: %s | Error: %v", ctx.Sender().ID, ctx.Callback().Unique, err)
	}

	b, err := tele.NewBot(settings)
	if err != nil {
		return nil, err
	}

	if cmds := lt.Commands(); cmds != nil {
		if err = b.SetCommands(cmds); err != nil {
			return nil, err
		}
	}

	bot := &Bot{
		Bot:        b,
		Layout:     lt,
		DB:         config.Database,
		StateRedis: config.StateRedis,
		CodeRedis:  config.CodeRedis,
		Logger:     botLogger,
	}

	return bot, nil
}

func (b *Bot) Start() {
	defer logger.Log.Sync()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		logger.Log.Info("Bot starting")
		b.Bot.Start()
	}()

	wg.Wait()
}
