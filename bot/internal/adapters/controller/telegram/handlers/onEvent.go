package handlers

import (
	"context"
	"github.com/Badsnus/cu-clubs-bot/bot/cmd/bot"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/database/postgres"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/database/redis"
	states "github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/database/redis/state"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/common/errorz"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/service"
	"github.com/spf13/viper"
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/layout"
	"net/mail"
	"regexp"
	"slices"
	"strings"
)

type onEventUserService interface {
	Get(ctx context.Context, userID int64) (*entity.User, error)
	Create(ctx context.Context, user entity.User) (*entity.User, error)
	SendAuthCode(ctx context.Context, email string) (string, string, error)
}

type OnEventHandler struct {
	layout        *layout.Layout
	bot           *tele.Bot
	userService   onEventUserService
	statesStorage *states.Storage
	codesStorage  *codes.Storage
	emailsStorage *emails.Storage
}

func NewOnEventHandler(b *bot.Bot) *OnEventHandler {
	userStorage := postgres.NewUserStorage(b.DB)
	studentDataStorage := postgres.NewStudentDataStorage(b.DB)
	smtpClient := smtp.NewClient(b.SMTPDialer)

	return &OnEventHandler{
		layout:        b.Layout,
		bot:           b.Bot,
		userService:   service.NewUserService(userStorage, studentDataStorage, smtpClient),
		statesStorage: states.NewStorage(b),
		codesStorage:  codes.NewStorage(b),
		emailsStorage: emails.NewStorage(b),
	}
}

func (h *OnEventHandler) OnText(c tele.Context) error {
	stateData, errGetState := h.statesStorage.Get(c.Sender().ID)
	if errGetState != nil {
		return errGetState
	}

	switch stateData.State {
	case state.WaitingForMailingContent:
		return h.onMailing(c)
	case state.WaitingExternalUserFio:
		return h.onExternalUserFio(c)
	case state.WaitingGrantUserFio:
		return h.onGrantUserFio(c)
	case state.WaitingStudentEmail:
		return h.onStudentEmail(c)
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
	case state.WaitingForMailingContent:
		return h.onMailing(c)
	default:
		return c.Send(h.layout.Text(c, "unknown_command"))
	}
}

func (h *OnEventHandler) onMailing(c tele.Context) error {
	if !h.checkForAdmin(c.Sender().ID) {
		return errorz.ErrForbidden
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

func (h *OnEventHandler) onExternalUserFio(c tele.Context) error {
	fio := c.Message().Text
	if splitFio := strings.Split(fio, " "); len(splitFio) != 3 {
		return c.Send(
			h.layout.Text(c, "invalid_user_fio"),
		)
	}
	re := regexp.MustCompile(`^[А-ЯЁ][а-яё]+(?:-[А-ЯЁ][а-яё]+)? [А-ЯЁ][а-яё]+ [А-ЯЁ][а-яё]+$`)
	if !re.MatchString(strings.TrimSpace(fio)) {
		return c.Send(
			h.layout.Text(c, "invalid_user_fio"),
		)
	}

	user := entity.User{
		ID:   c.Sender().ID,
		Role: entity.ExternalUser,
		FIO:  fio,
	}
	_, err := h.userService.Create(context.Background(), user)
	if err != nil {
		return c.Send(
			h.layout.Text(c, "technical_issues"),
		)
	}

	h.statesStorage.Clear(c.Sender().ID)

	return c.Send(
		h.layout.Text(c, "start"),
		h.layout.Markup(c, "mainMenu:open"),
	)
}

func (h *OnEventHandler) onGrantUserFio(c tele.Context) error {
	fio := c.Message().Text

	if !validateFio(fio) {
		return c.Send(
			h.layout.Text(c, "invalid_user_fio"),
		)
	}

	user := entity.User{
		ID:   c.Sender().ID,
		Role: entity.GrantUser,
		FIO:  fio,
	}
	_, err := h.userService.Create(context.Background(), user)
	if err != nil {
		return c.Send(
			h.layout.Text(c, "technical_issues"),
		)
	}

	h.statesStorage.Clear(c.Sender().ID)

	return c.Send(
		h.layout.Text(c, "start"),
		h.layout.Markup(c, "mainMenu:open"),
	)
}

func validateFio(fio string) bool {
	if splitFio := strings.Split(fio, " "); len(splitFio) != 3 {
		return false
	}
	re := regexp.MustCompile(`^[А-ЯЁ][а-яё]+(?:-[А-ЯЁ][а-яё]+)? [А-ЯЁ][а-яё]+ [А-ЯЁ][а-яё]+$`)
	return re.MatchString(strings.TrimSpace(fio))
}

func (h *OnEventHandler) onStudentEmail(c tele.Context) error {
	email := c.Message().Text
	if !validateEmail(email) {
		return c.Send(
			h.layout.Text(c, "invalid_email"),
		)
	}
	_, err := h.codesStorage.Get(c.Sender().ID)
	if err != nil && !errors.Is(err, redis.Nil) {
		return c.Send(
			h.layout.Text(c, "technical_issues"),
		)
	}
	var data, code string
	if errors.Is(err, redis.Nil) {
		data, code, err = h.userService.SendAuthCode(context.Background(), email)
		if err != nil {
			return c.Send(
				h.layout.Text(c, "technical_issues"),
			)
		}

		h.emailsStorage.Set(c.Sender().ID, email, "", viper.GetDuration("bot.session.email-ttl"))
		h.codesStorage.Set(c.Sender().ID, code, data, viper.GetDuration("bot.session.auth-ttl"))

		return c.Send(
			h.layout.Text(c, "email_confirmation_code_request"),
			h.layout.Markup(c, "auth:resendMenu"),
		)
	}

	return c.Send(
		h.layout.Text(c, "resend_timeout"),
		h.layout.Markup(c, "auth:resendMenu"),
	)
}

func validateEmail(email string) bool {
	if !validateEmailFormat(email) {
		return false
	}

	if !validateEmailDomain(email) {
		return false
	}

	return true
}

func validateEmailFormat(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func validateEmailDomain(email string) bool {
	validDomains := viper.GetStringSlice("bot.auth.valid-email-domains")

	for _, domain := range validDomains {
		if strings.HasSuffix(email, domain) {
			return true
		}
	}
	return false
}
