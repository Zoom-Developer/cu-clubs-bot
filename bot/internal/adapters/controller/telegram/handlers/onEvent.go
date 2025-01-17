package handlers

import (
	"context"
	"slices"

	"github.com/Badsnus/cu-clubs-bot/cmd/bot"
	"github.com/Badsnus/cu-clubs-bot/internal/adapters/database/postgres"
	"github.com/Badsnus/cu-clubs-bot/internal/adapters/database/redis"
	states "github.com/Badsnus/cu-clubs-bot/internal/adapters/database/redis/state"
	"github.com/Badsnus/cu-clubs-bot/internal/domain/common/errorz"
	"github.com/Badsnus/cu-clubs-bot/internal/domain/entity"
	"github.com/Badsnus/cu-clubs-bot/internal/domain/service"
	"github.com/spf13/viper"
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/layout"
)

type onEventUserService interface {
	Get(ctx context.Context, userID int64) (*entity.User, error)
}

type OnEventHandler struct {
	layout        *layout.Layout
	bot           *tele.Bot
	userService   onEventUserService
	statesStorage *redis.StatesStorage
}

func NewOnEventHandler(b *bot.Bot) *OnEventHandler {
	userStorage := postgres.NewUserStorage(b.DB)

	return &OnEventHandler{
		layout:        b.Layout,
		bot:           b.Bot,
		userService:   service.NewUserService(userStorage),
		statesStorage: redis.NewStatesStorage(b),
	}
}

func (h *OnEventHandler) OnText(c tele.Context) error {
	stateData, errGetState := h.statesStorage.Get(c.Sender().ID)
	if errGetState != nil {
		return errGetState
	}

	switch stateData.State {
	case states.WaitingForMailingContent:
		return h.onMailing(c)
	default:
		return c.Send(h.layout.Text(c, "unknown_command"))
	}
}

func (h *OnEventHandler) OnMedia(c tele.Context) error {
	stateData, errGetState := h.statesStorage.Get(c.Sender().ID)
	if errGetState != nil {
		return errGetState
	}

	switch stateData.State {
	case states.WaitingForMailingContent:
		return h.onMailing(c)
	default:
		return c.Send(h.layout.Text(c, "unknown_command"))
	}
}

func (h *OnEventHandler) onMailing(c tele.Context) error {
	if !h.checkForAdmin(c.Sender().ID) {
		return errorz.Forbidden
	}
	h.statesStorage.Clear(c.Sender().ID)
	chat, _ := h.bot.ChatByID(c.Sender().ID)
	_, err := h.bot.Copy(
		chat,
		c.Message(),
		h.layout.Markup(c, "checkMailing"),
	)
	return err
}

func (h *OnEventHandler) checkForAdmin(userID int64) bool {
	return slices.Contains(viper.GetIntSlice("bot.admin-ids"), int(userID))
}
