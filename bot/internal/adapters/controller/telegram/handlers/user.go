package handlers

import (
	"context"

	"github.com/Badsnus/cu-clubs-bot/cmd/bot"
	"github.com/Badsnus/cu-clubs-bot/internal/adapters/database/postgres"
	"github.com/Badsnus/cu-clubs-bot/internal/domain/entity"
	"github.com/Badsnus/cu-clubs-bot/internal/domain/service"
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/layout"
)

type userService interface {
	Get(ctx context.Context, userID int64) (*entity.User, error)
	UpdateLocalisation(ctx context.Context, userID int64, localisation string) (*entity.User, error)
}

type UserHandler struct {
	userService userService

	layout            *layout.Layout
	subscriptionPrice float64
}

func NewUserHandler(b *bot.Bot) *UserHandler {
	userStorage := postgres.NewUserStorage(b.DB)

	return &UserHandler{
		userService: service.NewUserService(userStorage),
		layout:      b.Layout,
	}
}

func (h UserHandler) OnLocalisation(c tele.Context) error {
	user, err := h.userService.UpdateLocalisation(context.Background(), c.Sender().ID, c.Callback().Data)
	if err != nil {
		return err
	}

	_ = c.Send(h.layout.MarkupLocale(user.Localisation, "replyOpenMenu"))

	return c.Edit(
		h.layout.TextLocale(user.Localisation, "main_menu_text", c.Sender().Username),
		h.layout.MarkupLocale(user.Localisation, "mainMenu"),
	)
}

func (h UserHandler) OnStart(c tele.Context) error {
	return c.Send(
		h.layout.Text(c, "start"),
		h.layout.Markup(c, "replyOpenMenu"),
	)
}

func (h UserHandler) SendMainMenu(c tele.Context) error {
	return c.Send(
		h.layout.Text(c, "main_menu_text", c.Sender().Username),
		h.layout.Markup(c, "mainMenu"),
	)
}

func (h UserHandler) EditMainMenu(c tele.Context) error {
	return c.Edit(
		h.layout.Text(c, "main_menu_text", c.Sender().Username),
		h.layout.Markup(c, "mainMenu"),
	)
}

func (h UserHandler) Hide(c tele.Context) error {
	return c.Delete()
}

func (h UserHandler) Information(c tele.Context) error {
	return c.Edit(
		h.layout.Text(c, "info_text"),
		h.layout.Markup(c, "information"),
	)
}

func (h UserHandler) ChangeLocalisation(c tele.Context) error {
	return c.Edit(
		"<b>🌍 Выберите предпочтительный язык</b>\n"+
			"<b>🌍 Choose your preferred language</b>\n",
		h.layout.MarkupLocale("en", "pickLanguage"),
	)
}

func (h UserHandler) Profile(c tele.Context) error {
	user, err := h.userService.Get(context.Background(), c.Sender().ID)
	if err != nil {
		return err
	}

	return c.Edit(
		h.layout.Text(c, "profile_text", struct {
			User    entity.User
			Expires int
		}{
			User: *user,
		}),
		h.layout.Markup(c, "profile"),
	)
}

func (h UserHandler) SubscriptionInfo(c tele.Context) error {
	return c.Edit(
		h.layout.Text(c, "subscription_text", struct {
			Price float64
		}{
			Price: h.subscriptionPrice,
		}),
		h.layout.Markup(c, "subscriptionInfo"),
	)
}
