package handlers

import (
	"context"
	"github.com/Badsnus/cu-clubs-bot/cmd/bot"
	"github.com/Badsnus/cu-clubs-bot/internal/adapters/database/postgres"
	"github.com/Badsnus/cu-clubs-bot/internal/adapters/database/redis"
	"github.com/Badsnus/cu-clubs-bot/internal/adapters/database/redis/state"
	"github.com/Badsnus/cu-clubs-bot/internal/domain/entity"
	"github.com/Badsnus/cu-clubs-bot/internal/domain/service"
	"github.com/spf13/viper"
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/layout"
)

type userService interface {
	Get(ctx context.Context, userID int64) (*entity.User, error)
}

type UserHandler struct {
	userService userService

	statesStorage *redis.StatesStorage

	layout *layout.Layout
}

func NewUserHandler(b *bot.Bot) *UserHandler {
	userStorage := postgres.NewUserStorage(b.DB)

	return &UserHandler{
		userService:   service.NewUserService(userStorage),
		statesStorage: redis.NewStatesStorage(b),
		layout:        b.Layout,
	}
}

func (h UserHandler) OnStart(c tele.Context) error {
	_, err := h.userService.Get(context.Background(), c.Sender().ID)
	if err != nil {
		return c.Send(
			h.layout.Text(c, "personal_data_agreement_text"),
			h.layout.Markup(c, "personalDataAgreementMenu"),
		)
	}

	return c.Send(
		h.layout.Text(c, "start"),
		h.layout.Markup(c, "replyOpenMenu"),
	)
}

func (h UserHandler) OnDeclinePersonalDataAgreement(c tele.Context) error {
	return c.Edit(
		h.layout.Text(c, "decline_personal_data_agreement_text"),
	)
}

func (h UserHandler) OnAcceptPersonalDataAgreement(c tele.Context) error {
	return c.Edit(
		h.layout.Text(c, "auth_menu_text"),
		h.layout.Markup(c, "authMenu"),
	)
}

func (h UserHandler) OnExternalUserAuth(c tele.Context) error {
	h.statesStorage.Set(c.Sender().ID, state.WaitingExternalUserFio, "")

	return c.Edit(
		h.layout.Text(c, "fio_request"),
		h.layout.Markup(c, "backToAuthMenu"),
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
