package handlers

import (
	"context"
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
	"net/mail"
	"regexp"
	"slices"
	"strings"
)

type onEventUserService interface {
	Get(ctx context.Context, userID int64) (*entity.User, error)
	Create(ctx context.Context, user entity.User) (*entity.User, error)
}

type OnEventHandler struct {
	layout        *layout.Layout
	bot           *tele.Bot
	userService   onEventUserService
	statesStorage *redis.StatesStorage
	codesStorage  *redis.CodesStorage
}

func NewOnEventHandler(b *bot.Bot) *OnEventHandler {
	userStorage := postgres.NewUserStorage(b.DB)

	return &OnEventHandler{
		layout:        b.Layout,
		bot:           b.Bot,
		userService:   service.NewUserService(userStorage),
		statesStorage: redis.NewStatesStorage(b),
		codesStorage:  redis.NewCodesStorage(b),
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
	case states.WaitingExternalUserFio:
		return h.onExternalUserFio(c)
	case states.WaitingGrantUserFio:
		return h.onGrantUserFio(c)
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
		h.layout.Markup(c, "replyOpenMenu"),
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
		h.layout.Markup(c, "replyOpenMenu"),
	)
}

func validateFio(fio string) bool {
	if splitFio := strings.Split(fio, " "); len(splitFio) != 3 {
		return false
	}
	re := regexp.MustCompile(`^[А-ЯЁ][а-яё]+(?:-[А-ЯЁ][а-яё]+)? [А-ЯЁ][а-яё]+ [А-ЯЁ][а-яё]+$`)
	if !re.MatchString(strings.TrimSpace(fio)) {
		return false
	}

	return true
}

func (h *OnEventHandler) onStudentEmail(c tele.Context) error {
	email := c.Message().Text
	if !validateEmail(email) {
		return c.Send(
			h.layout.Text(c, "invalid_email"),
		)
	}

	return nil
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
	validDomains := viper.GetStringSlice("bot.valid-email-domains")

	for _, domain := range validDomains {
		if strings.HasSuffix(email, domain) {
			return true
		}
	}
	return false
}
