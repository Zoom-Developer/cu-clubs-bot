package user

import (
	"context"
	"errors"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/database/redis/codes"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/database/redis/emails"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/utils/validator"
	"github.com/nlypage/intele/collector"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	tele "gopkg.in/telebot.v3"
	"strings"
)

func (h Handler) onStart(c tele.Context) error {
	_, err := h.userService.Get(context.Background(), c.Sender().ID)
	if err != nil {
		authCode := c.Message().Payload
		if authCode == "" {
			return c.Send(
				h.layout.Text(c, "personal_data_agreement_text"),
				h.layout.Markup(c, "auth:personalData:agreementMenu"),
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
				h.layout.Text(c, "technical_issues", err.Error()),
			)
		}

		h.codesStorage.Clear(c.Sender().ID)
		h.emailsStorage.Clear(c.Sender().ID)

		return h.menuHandler.SendMenu(c)
	}

	return h.menuHandler.SendMenu(c)
}

func (h Handler) onDeclinePersonalDataAgreement(c tele.Context) error {
	return c.Edit(
		h.layout.Text(c, "decline_personal_data_agreement_text"),
	)
}

func (h Handler) onAcceptPersonalDataAgreement(c tele.Context) error {
	return c.Edit(
		h.layout.Text(c, "auth_menu_text"),
		h.layout.Markup(c, "auth:menu"),
	)
}

func (h Handler) onExternalUserAuth(c tele.Context) error {
	inputCollector := collector.New()
	_ = c.Edit(
		h.layout.Text(c, "fio_request"),
		h.layout.Markup(c, "auth:backToMenu"),
	)
	inputCollector.Collect(c.Message())

	var (
		fio  string
		done bool
	)
	for {
		message, canceled, err := h.input.Get(context.Background(), c.Sender().ID, 0)
		if message != nil {
			inputCollector.Collect(message)
		}
		switch {
		case canceled:
			_ = inputCollector.Clear(c, collector.ClearOptions{IgnoreErrors: true, ExcludeLast: true})
			return nil
		case err != nil:
			_ = inputCollector.Send(c,
				h.layout.Text(c, "input_error", h.layout.Text(c, "fio_request")),
				h.layout.Markup(c, "auth:backToMenu"),
			)
		case !validator.Fio(message.Text):
			_ = inputCollector.Send(c,
				h.layout.Text(c, "invalid_user_fio"),
				h.layout.Markup(c, "auth:backToMenu"),
			)
		case validator.Fio(message.Text):
			fio = message.Text
			_ = inputCollector.Clear(c, collector.ClearOptions{IgnoreErrors: true})
			done = true
		}
		if done {
			break
		}
	}

	user := entity.User{
		ID:   c.Sender().ID,
		Role: entity.ExternalUser,
		FIO:  fio,
	}
	_, err := h.userService.Create(context.Background(), user)
	if err != nil {
		return c.Send(
			h.layout.Text(c, "technical_issues", err.Error()),
		)
	}

	return h.menuHandler.SendMenu(c)
}

func (h Handler) onGrantUserAuth(c tele.Context) error {
	grantChatID := int64(viper.GetInt("bot.grant-chat-id"))
	member, err := c.Bot().ChatMemberOf(&tele.Chat{ID: grantChatID}, &tele.User{ID: c.Sender().ID})
	if err != nil {
		return c.Send(
			h.layout.Text(c, "technical_issues", err.Error()),
		)
	}

	if member.Role != tele.Creator && member.Role != tele.Administrator && member.Role != tele.Member {
		return c.Edit(
			h.layout.Text(c, "grant_user_required"),
			h.layout.Markup(c, "auth:backToMenu"),
		)
	}

	inputCollector := collector.New()
	_ = c.Edit(
		h.layout.Text(c, "fio_request"),
		h.layout.Markup(c, "auth:backToMenu"),
	)
	inputCollector.Collect(c.Message())

	var (
		fio  string
		done bool
	)
	for {
		message, canceled, errGet := h.input.Get(context.Background(), c.Sender().ID, 0)
		if message != nil {
			inputCollector.Collect(message)
		}
		switch {
		case canceled:
			_ = inputCollector.Clear(c, collector.ClearOptions{IgnoreErrors: true, ExcludeLast: true})
			return nil
		case errGet != nil:
			_ = inputCollector.Send(c,
				h.layout.Text(c, "input_error", h.layout.Text(c, "fio_request")),
				h.layout.Markup(c, "auth:backToMenu"),
			)
		case !validator.Fio(message.Text):
			_ = inputCollector.Send(c,
				h.layout.Text(c, "invalid_user_fio"),
				h.layout.Markup(c, "auth:backToMenu"),
			)
		case validator.Fio(message.Text):
			fio = message.Text
			_ = inputCollector.Clear(c, collector.ClearOptions{IgnoreErrors: true})
			done = true
		}
		if done {
			break
		}
	}

	user := entity.User{
		ID:   c.Sender().ID,
		Role: entity.GrantUser,
		FIO:  fio,
	}
	_, err = h.userService.Create(context.Background(), user)
	if err != nil {
		return c.Send(
			h.layout.Text(c, "technical_issues", err.Error()),
		)
	}

	return h.menuHandler.SendMenu(c)
}

func (h Handler) onStudentAuth(c tele.Context) error {
	inputCollector := collector.New()
	_ = c.Edit(
		h.layout.Text(c, "email_request"),
		h.layout.Markup(c, "auth:backToMenu"),
	)
	inputCollector.Collect(c.Message())

	var (
		email string
		done  bool
	)
	for {
		message, canceled, errGet := h.input.Get(context.Background(), c.Sender().ID, 0)
		if message != nil {
			inputCollector.Collect(message)
		}
		switch {
		case canceled:
			return nil
		case errGet != nil:
			_ = inputCollector.Send(c,
				h.layout.Text(c, "input_error", h.layout.Text(c, "email_request")),
				h.layout.Markup(c, "auth:backToMenu"),
			)
		case !validator.Email(message.Text):
			_ = inputCollector.Send(c,
				h.layout.Text(c, "invalid_email"),
				h.layout.Markup(c, "auth:backToMenu"),
			)
		case validator.Email(message.Text):
			email = message.Text
			done = true
		}
		if done {
			break
		}
	}

	_, err := h.codesStorage.Get(c.Sender().ID)
	if err != nil && !errors.Is(err, redis.Nil) {
		return c.Send(
			h.layout.Text(c, "technical_issues", err.Error()),
		)
	}
	var data, code string
	if errors.Is(err, redis.Nil) {
		data, code, err = h.userService.SendAuthCode(context.Background(), email)
		if err != nil {
			return c.Send(
				h.layout.Text(c, "technical_issues", err.Error()),
			)
		}

		h.emailsStorage.Set(c.Sender().ID, email, "", viper.GetDuration("bot.session.email-ttl"))
		h.codesStorage.Set(c.Sender().ID, code, data, viper.GetDuration("bot.session.auth-ttl"))

		return c.Send(
			h.layout.Text(c, "email_auth_link_sent"),
			h.layout.Markup(c, "auth:resendMenu"),
		)
	}

	return c.Send(
		h.layout.Text(c, "resend_timeout"),
		h.layout.Markup(c, "auth:resendMenu"),
	)
}

func (h Handler) onBackToAuthMenu(c tele.Context) error {
	return c.Edit(
		h.layout.Text(c, "auth_menu_text"),
		h.layout.Markup(c, "auth:menu"),
	)
}

func (h Handler) onResendEmailConfirmationCode(c tele.Context) error {
	_, err := h.codesStorage.Get(c.Sender().ID)
	if err != nil && !errors.Is(err, redis.Nil) {
		return c.Send(
			h.layout.Text(c, "technical_issues", err.Error()),
		)
	}

	var data, code string
	var email emails.Email
	if errors.Is(err, redis.Nil) {
		email, err = h.emailsStorage.Get(c.Sender().ID)
		if err != nil && !errors.Is(err, redis.Nil) {
			return c.Send(
				h.layout.Text(c, "technical_issues", err.Error()),
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
				h.layout.Text(c, "technical_issues", err.Error()),
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

func (h Handler) AuthSetup(group *tele.Group) {
	group.Handle("/start", h.onStart)
	group.Handle(h.layout.Callback("auth:personalData:accept"), h.onAcceptPersonalDataAgreement)
	group.Handle(h.layout.Callback("auth:personalData:decline"), h.onDeclinePersonalDataAgreement)
	group.Handle(h.layout.Callback("auth:external_user"), h.onExternalUserAuth)
	group.Handle(h.layout.Callback("auth:grant_user"), h.onGrantUserAuth)
	group.Handle(h.layout.Callback("auth:student"), h.onStudentAuth)
	group.Handle(h.layout.Callback("auth:resend_email"), h.onResendEmailConfirmationCode)
	group.Handle(h.layout.Callback("auth:back_to_menu"), h.onBackToAuthMenu)
}
