package middlewares

import (
	"context"
	"strings"

	"github.com/Badsnus/cu-clubs-bot/cmd/bot"
	"github.com/Badsnus/cu-clubs-bot/internal/adapters/database/postgres"
	"github.com/Badsnus/cu-clubs-bot/internal/adapters/database/redis"
	"github.com/Badsnus/cu-clubs-bot/internal/domain/entity"
	"github.com/Badsnus/cu-clubs-bot/internal/domain/service"
	"github.com/spf13/viper"
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/layout"
)

type userService interface {
	Get(ctx context.Context, userID int64) (*entity.User, error)
	UpdateData(ctx context.Context, c tele.Context) (*entity.User, error)
	Create(ctx context.Context, c tele.Context) (*entity.User, error)
}

type MiddlewareHandler struct {
	bot           *tele.Bot
	layout        *layout.Layout
	userService   userService
	statesStorage *redis.StatesStorage
}

func New(b *bot.Bot) *MiddlewareHandler {
	userStorageLocal := postgres.NewUserStorage(b.DB)
	userServiceLocal := service.NewUserService(userStorageLocal)

	return &MiddlewareHandler{
		bot:           b.Bot,
		layout:        b.Layout,
		userService:   userServiceLocal,
		statesStorage: redis.NewStatesStorage(b),
	}
}

func (h MiddlewareHandler) Authorized(next tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) error {
		user, err := h.userService.Get(context.Background(), c.Sender().ID)
		if err != nil {
			user, err = h.userService.Create(context.Background(), c)
			if err != nil {
				return err
			}

		} else if user.Username != c.Sender().Username || user.FirstName != c.Sender().FirstName {
			user, err = h.userService.UpdateData(context.Background(), c)
			if err != nil {
				return err
			}
		}

		if user.Banned {
			return c.Send(h.layout.TextLocale(user.Localisation, "banned"))
		}

		return next(c)
	}
}

//func (h MiddlewareHandler) Localisation(next tele.HandlerFunc) tele.HandlerFunc {
//	return func(c tele.Context) error {
//		user, err := h.userService.Get(context.Background(), c.Sender().ID)
//		if err != nil {
//			return err
//		}
//		if user.Localisation == "" {
//			return c.Send(
//				"<b>🌍 Выберите предпочтительный язык</b>\n"+
//					"<b>🌍 Choose your preferred language</b>\n",
//				h.layout.MarkupLocale("en", "pickLanguage"),
//			)
//		}
//
//		return next(c)
//	}
//}

//func (h MiddlewareHandler) SetupLocalisation(r tele.Recipient) string {
//	userID, err := strconv.Atoi(r.Recipient())
//	if err != nil {
//		return ""
//	}
//
//	user, err := h.userService.Get(context.Background(), int64(userID))
//	if err != nil {
//		return ""
//	}
//	return user.Localisation
//}

// ResetStateOnBack middleware clears the input state when the back button is pressed.
func (h MiddlewareHandler) ResetStateOnBack(next tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) error {
		if c.Callback() != nil {
			if strings.Contains(c.Callback().Data, "back") {
				h.statesStorage.Clear(c.Sender().ID)
			}
		}

		return next(c)
	}
}
func (h MiddlewareHandler) Subscribed(next tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) error {
		channel, err := h.bot.ChatByID(viper.GetInt64("bot.channel-id"))
		if err != nil {
			return err
		}

		member, err := c.Bot().ChatMemberOf(channel, c.Sender())
		if err != nil {
			return err
		}

		if member.Role == tele.Creator || member.Role == tele.Administrator || member.Role == tele.Member {
			return next(c)
		}

		return c.Send(
			h.layout.Text(c, "subscribe"),
			h.layout.Markup(c, "joinChannel", struct {
				Url string
			}{
				Url: channel.InviteLink,
			}))
	}
}
