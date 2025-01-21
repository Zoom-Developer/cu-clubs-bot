package handlers

import (
	"context"
	"github.com/Badsnus/cu-clubs-bot/bot/pkg/logger"
	"time"

	"github.com/Badsnus/cu-clubs-bot/bot/cmd/bot"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/database/postgres"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/database/redis"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/database/redis/state"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/service"
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
	logger.Log.Errorf("User ID: %d", c.Sender().ID)
	_, err := h.userService.Get(context.Background(), c.Sender().ID)
	if err != nil {
		return c.Send(
			h.layout.Text(c, "personal_data_agreement_text"),
			h.layout.Markup(c, "personalData:agreementMenu"),
		)
	}

	return c.Send(
		h.layout.Text(c, "start"),
		h.layout.Markup(c, "mainMenu:open"),
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
		h.layout.Markup(c, "auth:menu"),
	)
}

func (h UserHandler) OnExternalUserAuth(c tele.Context) error {
	h.statesStorage.Set(c.Sender().ID, state.WaitingExternalUserFio, "", time.Minute*45)

	return c.Edit(
		h.layout.Text(c, "fio_request"),
		h.layout.Markup(c, "auth:backToMenu"),
	)
}

func (h UserHandler) OnGrantUserAuth(c tele.Context) error {
	granChatID := int64(viper.GetInt("bot.grant-chat-id"))
	member, err := c.Bot().ChatMemberOf(&tele.Chat{ID: granChatID}, &tele.User{ID: c.Sender().ID})
	if err != nil {
		return c.Send(
			h.layout.Text(c, "technical_issues"),
		)
	}

	if member.Role != tele.Creator && member.Role != tele.Administrator && member.Role != tele.Member {
		return c.Send(
			h.layout.Text(c, "grant_user_required"),
			h.layout.Markup(c, "auth:backToMenu"),
		)
	}

	h.statesStorage.Set(c.Sender().ID, state.WaitingGrantUserFio, "", time.Minute*45)
	return c.Edit(
		h.layout.Text(c, "fio_request"),
		h.layout.Markup(c, "auth:backToMenu"),
	)
}

func (h UserHandler) OnBackToAuthMenu(c tele.Context) error {
	return c.Edit(
		h.layout.Text(c, "auth_menu_text"),
		h.layout.Markup(c, "auth:menu"),
	)
}

func (h UserHandler) SendMainMenu(c tele.Context) error {
	return c.Send(
		h.layout.Text(c, "main_menu_text", c.Sender().Username),
		h.layout.Markup(c, "mainMenu:open"),
	)
}

func (h UserHandler) EditMainMenu(c tele.Context) error {
	return c.Edit(
		h.layout.Text(c, "main_menu_text", c.Sender().Username),
		h.layout.Markup(c, "mainMenu:open"),
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
