package handlers

import (
	"context"
	"errors"
	"strings"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/database/redis/codes"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/database/redis/emails"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/database/redis/states"
	"github.com/Badsnus/cu-clubs-bot/bot/pkg/logger"
	"github.com/redis/go-redis/v9"

	"github.com/Badsnus/cu-clubs-bot/bot/cmd/bot"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/database/postgres"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/database/redis/state"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/service"
	"github.com/Badsnus/cu-clubs-bot/bot/pkg/smtp"

	"github.com/spf13/viper"
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/layout"
)

type userService interface {
	Create(ctx context.Context, user entity.User) (*entity.User, error)
	Get(ctx context.Context, userID int64) (*entity.User, error)
	SendAuthCode(ctx context.Context, email string) (string, string, error)
}

type UserHandler struct {
	userService userService

	statesStorage *states.Storage
	codesStorage  *codes.Storage
	emailsStorage *emails.Storage

	layout *layout.Layout
}

func NewUserHandler(b *bot.Bot) *UserHandler {
	userStorage := postgres.NewUserStorage(b.DB)
	studentDataStorage := postgres.NewStudentDataStorage(b.DB)
	smtpClient := smtp.NewClient(b.SMTPDialer)

	return &UserHandler{
		userService:   service.NewUserService(userStorage, studentDataStorage, smtpClient),
		statesStorage: states.NewStorage(b),
		codesStorage:  codes.NewStorage(b),
		emailsStorage: emails.NewStorage(b),
		layout:        b.Layout,
	}
}

func (h UserHandler) OnStart(c tele.Context) error {
	logger.Log.Errorf("User ID: %d", c.Sender().ID)
	_, err := h.userService.Get(context.Background(), c.Sender().ID)
	if err != nil {
		authCode := c.Message().Payload
		if authCode == "" {
			return c.Send(
				h.layout.Text(c, "personal_data_agreement_text"),
				h.layout.Markup(c, "personalData:agreementMenu"),
			)
		}

		var code codes.Code
		code, err = h.codesStorage.Get(c.Sender().ID)
		if err != nil {
			return c.Send(
				h.layout.Text(c, "session_expire"),
			)
		}

		if authCode != code.Code {
			return c.Send(
				h.layout.Text(c, "something_went_wrong"),
			)
		}

		data := strings.Split(code.CodeContext, ";")
		email, fio := data[0], data[1]

		user := entity.User{
			ID:    c.Sender().ID,
			Role:  entity.Student,
			Email: email,
			FIO:   fio,
		}

		_, err = h.userService.Create(context.Background(), user)
		if err != nil {
			return c.Send(
				h.layout.Text(c, "technical_issues"),
			)
		}

		h.statesStorage.Clear(c.Sender().ID)
		h.codesStorage.Clear(c.Sender().ID)
		h.emailsStorage.Clear(c.Sender().ID)

		return c.Send(
			h.layout.Text(c, "start"),
			h.layout.Markup(c, "mainMenu:open"),
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
	h.statesStorage.Set(c.Sender().ID, state.WaitingExternalUserFio, "", viper.GetDuration("bot.session.ttl"))

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

	h.statesStorage.Set(c.Sender().ID, state.WaitingGrantUserFio, "", viper.GetDuration("bot.session.ttl"))
	return c.Edit(
		h.layout.Text(c, "fio_request"),
		h.layout.Markup(c, "auth:backToMenu"),
	)
}

func (h UserHandler) OnStudentAuth(c tele.Context) error {
	h.statesStorage.Set(c.Sender().ID, state.WaitingStudentEmail, "", viper.GetDuration("bot.session.ttl"))

	return c.Edit(
		h.layout.Text(c, "email_request"),
		h.layout.Markup(c, "auth:backToMenu"),
	)
}

func (h UserHandler) OnBackToAuthMenu(c tele.Context) error {
	return c.Edit(
		h.layout.Text(c, "auth_menu_text"),
		h.layout.Markup(c, "auth:menu"),
	)
}

func (h UserHandler) OnResendEmailConfirmationCode(c tele.Context) error {
	_, err := h.codesStorage.Get(c.Sender().ID)
	if err != nil && !errors.Is(err, redis.Nil) {
		return c.Send(
			h.layout.Text(c, "technical_issues"),
		)
	}

	var data, code string
	var email emails.Email
	if errors.Is(err, redis.Nil) {
		email, err = h.emailsStorage.Get(c.Sender().ID)
		if err != nil && !errors.Is(err, redis.Nil) {
			return c.Send(
				h.layout.Text(c, "technical_issues"),
			)
		}

		if errors.Is(err, redis.Nil) {
			return c.Send(
				h.layout.Text(c, "session_expire"),
			)
		}

		data, code, err = h.userService.SendAuthCode(context.Background(), email.Email)
		if err != nil {
			return c.Send(
				h.layout.Text(c, "technical_issues"),
			)
		}

		h.emailsStorage.Set(c.Sender().ID, email.Email, "", viper.GetDuration("bot.session.email-ttl"))
		h.codesStorage.Set(c.Sender().ID, code, data, viper.GetDuration("bot.session.auth-ttl"))

		return c.Edit(
			h.layout.Text(c, "email_confirmation_code_request"),
			h.layout.Markup(c, "auth:resendMenu"),
		)
	}

	return c.Edit(
		h.layout.Text(c, "resend_timeout"),
		h.layout.Markup(c, "auth:resendMenu"),
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
