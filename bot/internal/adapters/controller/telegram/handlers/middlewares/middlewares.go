package middlewares

import (
	"context"
	"github.com/nlypage/intele"
	"strings"

	"github.com/Badsnus/cu-clubs-bot/bot/cmd/bot"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/database/postgres"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/service"
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/layout"
)

type userService interface {
	Get(ctx context.Context, userID int64) (*entity.User, error)
	UpdateData(ctx context.Context, c tele.Context) (*entity.User, error)
}

type Handler struct {
	bot         *tele.Bot
	layout      *layout.Layout
	userService userService
	input       *intele.InputManager
}

func New(b *bot.Bot) *Handler {
	userStorageLocal := postgres.NewUserStorage(b.DB)
	userServiceLocal := service.NewUserService(userStorageLocal, nil, nil)

	return &Handler{
		bot:         b.Bot,
		layout:      b.Layout,
		userService: userServiceLocal,
		input:       b.Input,
	}
}

func (h Handler) Authorized(next tele.HandlerFunc) tele.HandlerFunc {
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

//func (h Handler) Localisation(next tele.HandlerFunc) tele.HandlerFunc {
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

//func (h Handler) SetupLocalisation(r tele.Recipient) string {
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

// ResetInputOnBack middleware clears the input state when the back button is pressed.
func (h Handler) ResetInputOnBack(next tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) error {
		if c.Callback() != nil {
			if strings.Contains(c.Callback().Data, "back") {
				h.input.Cancel(c.Sender().ID)
			}
		}
		if c.Message() != nil {
			if strings.HasPrefix(c.Message().Text, "/") {
				h.input.Cancel(c.Sender().ID)
			}
		}

		return next(c)
	}
}
