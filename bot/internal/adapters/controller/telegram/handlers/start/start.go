package start

import (
	"context"
	"errors"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/utils/banner"
	"strings"

	"github.com/Badsnus/cu-clubs-bot/bot/cmd/bot"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/controller/telegram/handlers/menu"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/database/postgres"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/database/redis/codes"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/database/redis/emails"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/service"
	"github.com/Badsnus/cu-clubs-bot/bot/pkg/logger/types"
	"github.com/Badsnus/cu-clubs-bot/bot/pkg/smtp"
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/layout"
	"gorm.io/gorm"
)

type userService interface {
	Create(ctx context.Context, user entity.User) (*entity.User, error)
	Get(ctx context.Context, userID int64) (*entity.User, error)
	GetByQRCodeID(ctx context.Context, qrCodeID string) (*entity.User, error)
	SendAuthCode(ctx context.Context, email string) (string, string, error)
	Update(ctx context.Context, user *entity.User) (*entity.User, error)
}

type Handler struct {
	userService userService

	menuHandler *menu.Handler

	codesStorage  *codes.Storage
	emailsStorage *emails.Storage
	layout        *layout.Layout
	logger        *types.Logger
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
		logger:        b.Logger,
	}
}

func (h *Handler) Start(c tele.Context) error {
	h.logger.Infof("(user: %d) press start button", c.Sender().ID)

	_ = c.Delete()

	user, err := h.userService.Get(context.Background(), c.Sender().ID)

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		h.logger.Errorf("(user: %d) error while getting user from db: %v", c.Sender().ID, err)
		return c.Send(
			h.layout.Text(c, "technical_issues", err.Error()),
			h.layout.Markup(c, "core:hide"),
		)
	}

	payload := strings.Split(c.Message().Payload, "_")

	if len(payload) < 2 {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Send(
				banner.Auth.Caption(h.layout.Text(c, "personal_data_agreement_text")),
				h.layout.Markup(c, "auth:personalData:agreementMenu"),
			)
		}
		if user.IsBanned {
			return c.Send(
				h.layout.Text(c, "banned"),
				h.layout.Markup(c, "core:hide"),
			)
		}
		return h.menuHandler.SendMenu(c)
	}

	var payloadType, data string
	if len(payload) == 2 {
		payloadType, data = payload[0], payload[1]
	}

	switch payloadType {
	case "auth":
		return h.auth(c, data)
	case "qr":
		return h.qr(c, data)
	default:
		return c.Send(
			h.layout.Text(c, "something_went_wrong"),
			h.layout.Markup(c, "core:hide"),
		)
	}
}
