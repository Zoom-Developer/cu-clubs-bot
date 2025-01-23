package user

import (
	"context"
	"github.com/Badsnus/cu-clubs-bot/bot/cmd/bot"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/controller/telegram/handlers/menu"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/database/postgres"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/database/redis/codes"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/database/redis/emails"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/service"
	"github.com/Badsnus/cu-clubs-bot/bot/pkg/smtp"
	"github.com/nlypage/intele"

	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/layout"
)

type userService interface {
	Create(ctx context.Context, user entity.User) (*entity.User, error)
	Get(ctx context.Context, userID int64) (*entity.User, error)
	SendAuthCode(ctx context.Context, email string) (string, string, error)
}

type Handler struct {
	userService userService

	menuHandler *menu.Handler

	codesStorage  *codes.Storage
	emailsStorage *emails.Storage
	input         *intele.InputManager
	layout        *layout.Layout
}

func New(b *bot.Bot) *Handler {
	userStorage := postgres.NewUserStorage(b.DB)
	studentDataStorage := postgres.NewStudentDataStorage(b.DB)
	smtpClient := smtp.NewClient(b.SMTPDialer)

	return &Handler{
		userService:   service.NewUserService(userStorage, studentDataStorage, smtpClient),
		menuHandler:   menu.New(b),
		codesStorage:  b.Redis.Codes,
		emailsStorage: b.Redis.Emails,
		layout:        b.Layout,
		input:         b.Input,
	}
}

func (h Handler) Hide(c tele.Context) error {
	return c.Delete()
}
