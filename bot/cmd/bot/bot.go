package bot

import (
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/service"
	"github.com/spf13/viper"
	"go.uber.org/zap/zapcore"
	"gopkg.in/gomail.v2"
	"sync"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/config"
	"github.com/Badsnus/cu-clubs-bot/bot/pkg/logger"
	"github.com/Badsnus/cu-clubs-bot/bot/pkg/logger/types"
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
	EmailRedis *redis.Client
	SMTPDialer *gomail.Dialer
	Logger     *types.Logger
}

func New(config *config.Config) (*Bot, error) {
	lt, err := layout.New("telegram.yml")
	if err != nil {
		return nil, err
	}

	settings := lt.Settings()
	botLogger, err := logger.Named("bot")
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
		EmailRedis: config.EmailRedis,
		SMTPDialer: config.SMTPDialer,
		Logger:     botLogger,
	}

	return bot, nil
}

func (b *Bot) Start() {

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		logger.Log.Info("Bot starting")

		if viper.GetBool("settings.logging.log-to-channel") {
			notifyLogger, err := logger.Named("notify")
			if err != nil {
				logger.Log.Errorf("Failed to create notify logger: %v", err)
			} else {
				notifyService := service.NewNotifyService(b.Bot, b.Layout, notifyLogger)
				logHook, err := notifyService.LogHook(
					viper.GetInt64("settings.logging.channel-id"),
					viper.GetString("settings.logging.locale"),
					zapcore.Level(viper.GetInt("settings.logging.channel-log-level")),
				)
				if err != nil {
					logger.Log.Errorf("Failed to create notify log hook: %v", err)
				} else {
					logger.SetLogHook(logHook)
				}
			}
		}
		b.Bot.Start()
	}()

	wg.Wait()
}
