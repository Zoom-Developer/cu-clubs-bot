package handlers

import (
	"context"

	"github.com/Badsnus/cu-clubs-bot/bot/cmd/bot"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/database/postgres"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/database/redis/states"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/service"
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/layout"
)

type adminUserService interface {
	Count(ctx context.Context) (int64, error)
	GetWithPagination(ctx context.Context, offset, limit int, order string) ([]entity.User, error)
	Get(ctx context.Context, userID int64) (*entity.User, error)
	Ban(ctx context.Context, userID int64) (*entity.User, error)
	GetAll(ctx context.Context) ([]entity.User, error)
}

type AdminHandler struct {
	layout           *layout.Layout
	adminUserService adminUserService
	statesStorage    *states.Storage
	bot              *bot.Bot
}

func NewAdminHandler(b *bot.Bot) *AdminHandler {
	userStorage := postgres.NewUserStorage(b.DB)

	return &AdminHandler{
		layout:           b.Layout,
		adminUserService: service.NewUserService(userStorage, nil, nil),
		statesStorage:    states.NewStorage(b),
		bot:              b,
	}
}

func (h AdminHandler) AdminMenu(c tele.Context) error {
	return c.Send(
		h.layout.Text(c, "admin_text", c.Sender().Username),
		h.layout.Markup(c, "adminMenu"),
	)
}

func (h AdminHandler) BackToAdminMenu(c tele.Context) error {
	return c.Edit(
		h.layout.Text(c, "admin_text", c.Sender().Username),
		h.layout.Markup(c, "adminMenu"),
	)
}
