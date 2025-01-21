package middlewares

import (
	"context"
	"github.com/Badsnus/cu-clubs-bot/internal/adapters/database/redis/states"
	"strings"

	"github.com/Badsnus/cu-clubs-bot/bot/cmd/bot"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/database/postgres"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/database/redis"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/service"
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/layout"
)

type userService interface {
	Get(ctx context.Context, userID int64) (*entity.User, error)
	UpdateData(ctx context.Context, c tele.Context) (*entity.User, error)
}

type MiddlewareHandler struct {
	bot           *tele.Bot
	layout        *layout.Layout
	userService   userService
	statesStorage *states.Storage
}

func New(b *bot.Bot) *MiddlewareHandler {
	userStorageLocal := postgres.NewUserStorage(b.DB)
	userServiceLocal := service.NewUserService(userStorageLocal, nil, nil)

	return &MiddlewareHandler{
		bot:           b.Bot,
		layout:        b.Layout,
		userService:   userServiceLocal,
		statesStorage: states.NewStorage(b),
	}
}

func (h MiddlewareHandler) Authorized(next tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) error {
		user, err := h.userService.Get(context.Background(), c.Sender().ID)
		if err != nil {
			return c.Send(h.layout.Text(c, "auth_required"))
		}

		if user.IsBanned {
			return c.Send(h.layout.TextLocale(user.Localization, "banned"))
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
