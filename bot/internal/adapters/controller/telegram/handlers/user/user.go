package user

import (
	"context"
	"errors"
	"github.com/Badsnus/cu-clubs-bot/bot/cmd/bot"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/controller/telegram/handlers/menu"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/database/postgres"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/database/redis/codes"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/database/redis/emails"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/common/errorz"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/service"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/utils/banner"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/utils/location"
	"github.com/Badsnus/cu-clubs-bot/bot/pkg/generator"
	"github.com/Badsnus/cu-clubs-bot/bot/pkg/logger/types"
	"github.com/Badsnus/cu-clubs-bot/bot/pkg/smtp"
	"github.com/nlypage/intele"
	"github.com/spf13/viper"
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/layout"
	"gorm.io/gorm"
	"strconv"
	"strings"
	"time"
)

type userService interface {
	Create(ctx context.Context, user entity.User) (*entity.User, error)
	Get(ctx context.Context, userID int64) (*entity.User, error)
	GetByQRCodeID(ctx context.Context, qrCodeID string) (*entity.User, error)
	SendAuthCode(ctx context.Context, email string) (string, string, error)
	Update(ctx context.Context, user *entity.User) (*entity.User, error)
	GetUserEvents(ctx context.Context, userID int64, limit, offset int) ([]entity.Event, error)
}

type qrCodesGenerator interface {
	Generate() (string, string, error)
	Delete(filePath string) error
}

type eventService interface {
	Get(ctx context.Context, id string) (*entity.Event, error)
	GetWithPagination(ctx context.Context, limit, offset int, order string, role entity.Role) ([]entity.Event, error)
	Count(ctx context.Context) (int64, error)
}

type eventParticipantService interface {
	Register(ctx context.Context, userID int64, eventID string) (*entity.EventParticipant, error)
	Get(ctx context.Context, eventID string, userID int64) (*entity.EventParticipant, error)
	CountByEventID(ctx context.Context, eventID string) (int, error)
}

type Handler struct {
	userService             userService
	eventService            eventService
	eventParticipantService eventParticipantService

	menuHandler *menu.Handler

	qrCodesGenerator qrCodesGenerator
	codesStorage     *codes.Storage
	emailsStorage    *emails.Storage
	input            *intele.InputManager
	layout           *layout.Layout
	logger           *types.Logger
}

func New(b *bot.Bot) *Handler {
	userStorage := postgres.NewUserStorage(b.DB)
	studentDataStorage := postgres.NewStudentDataStorage(b.DB)
	eventStorage := postgres.NewEventStorage(b.DB)
	eventParticipantStorage := postgres.NewEventParticipantStorage(b.DB)

	eventPartService := service.NewEventParticipantService(eventParticipantStorage)

	smtpClient := smtp.NewClient(b.SMTPDialer)
	botName := viper.GetString("bot.username")
	qrCodeLogo := viper.GetString("settings.qr.logo-path")
	qrCodeOutputDir := viper.GetString("settings.qr.output-dir")

	return &Handler{
		userService:             service.NewUserService(userStorage, studentDataStorage, eventPartService, smtpClient),
		eventService:            service.NewEventService(eventStorage),
		eventParticipantService: eventPartService,
		qrCodesGenerator:        generator.NewQrCode(generator.CU, qrCodeOutputDir, qrCodeLogo, botName),
		menuHandler:             menu.New(b),
		codesStorage:            b.Redis.Codes,
		emailsStorage:           b.Redis.Emails,
		layout:                  b.Layout,
		input:                   b.Input,
		logger:                  b.Logger,
	}
}

func (h Handler) Hide(c tele.Context) error {
	return c.Delete()
}

func (h Handler) qrCode(c tele.Context) error {
	h.logger.Infof("(user: %d) requested QR code", c.Sender().ID)

	user, err := h.userService.Get(context.Background(), c.Sender().ID)
	if err != nil {
		h.logger.Errorf("(user: %d) error while getting user from db: %v", c.Sender().ID, err)
		return c.Send(
			banner.Menu.Caption(h.layout.Text(c, "technical_issues", err.Error())),
			h.layout.Markup(c, "mainMenu:back"),
		)
	}

	var file tele.File
	if user.QRCodeID != "" {
		h.logger.Infof("(user: %d) existing QR code found, sending...", c.Sender().ID)
		file, err = c.Bot().FileByID(user.QRFileID)
		if err != nil {
			h.logger.Errorf("(user: %d) failed to retrieve existing QR file: %v", c.Sender().ID, err)
			return c.Edit(
				banner.Menu.Caption(h.layout.Text(c, "technical_issues", err.Error())),
				h.layout.Markup(c, "mainMenu:back"),
			)
		}

		return c.Edit(
			&tele.Photo{
				File:    file,
				Caption: h.layout.Text(c, "qr_text"),
			},
			h.layout.Markup(c, "mainMenu:back"),
		)
	}
	loading, _ := c.Bot().Send(c.Chat(), h.layout.Text(c, "loading"))
	h.logger.Infof("(user: %d) generating new QR code...", c.Sender().ID)
	qrID, qrFilePath, err := h.qrCodesGenerator.Generate()
	if err != nil {
		h.logger.Errorf("(user: %d) failed to generate QR code: %v", c.Sender().ID, err)
		return c.Edit(
			h.layout.Text(c, "technical_issues", err.Error()),
		)
	}

	qrChatID := viper.GetInt64("bot.qr.chat-id")
	qrImg := &tele.Photo{File: tele.FromDisk(qrFilePath)}
	_, err = c.Bot().Send(&tele.Chat{ID: qrChatID}, qrImg)
	if err != nil {
		h.logger.Errorf("(user: %d) failed to send QR code to admin chat: %v", c.Sender().ID, err)
		return c.Edit(
			banner.Menu.Caption(h.layout.Text(c, "technical_issues", err.Error())),
			h.layout.Markup(c, "mainMenu:back"),
		)
	}

	user.QRCodeID = qrID
	user.QRFileID = qrImg.FileID
	user, err = h.userService.Update(context.Background(), user)
	if err != nil {
		h.logger.Errorf("(user: %d) failed to update user data: %v", c.Sender().ID, err)
		return c.Edit(
			banner.Menu.Caption(h.layout.Text(c, "technical_issues", err.Error())),
			h.layout.Markup(c, "mainMenu:back"),
		)
	}

	if err = h.qrCodesGenerator.Delete(qrFilePath); err != nil {
		h.logger.Errorf("(user: %d) failed to delete temporary QR code file: %v", c.Sender().ID, err)
		return c.Edit(
			banner.Menu.Caption(h.layout.Text(c, "technical_issues", err.Error())),
			h.layout.Markup(c, "mainMenu:back"),
		)
	}

	h.logger.Infof("(user: %d) sending QR code to user...", c.Sender().ID)

	file, err = c.Bot().FileByID(user.QRFileID)
	if err != nil {
		h.logger.Errorf("(user: %d) failed to retrieve final QR file: %v", c.Sender().ID, err)
		return c.Edit(
			banner.Menu.Caption(h.layout.Text(c, "technical_issues", err.Error())),
			h.layout.Markup(c, "mainMenu:back"),
		)
	}

	_ = c.Bot().Delete(loading)
	return c.Edit(
		&tele.Photo{
			File:    file,
			Caption: h.layout.Text(c, "qr_text"),
		},
		h.layout.Markup(c, "mainMenu:back"),
	)
}

func (h Handler) eventsList(c tele.Context) error {
	const eventsOnPage = 10
	h.logger.Infof("(user: %d) edit events list", c.Sender().ID)

	var (
		p           int
		prevPage    int
		nextPage    int
		err         error
		eventsCount int64
		events      []entity.Event
		rows        []tele.Row
		menuRow     tele.Row
	)
	if c.Callback().Unique != "mainMenu_events" {
		p, err = strconv.Atoi(c.Callback().Data)
		if err != nil {
			return errorz.ErrInvalidCallbackData
		}
	}

	user, err := h.userService.Get(context.Background(), c.Sender().ID)
	if err != nil {
		h.logger.Errorf("(user: %d) error while getting user from db: %v", c.Sender().ID, err)
		return c.Edit(
			banner.Events.Caption(h.layout.Text(c, "technical_issues", err.Error())),
			h.layout.Markup(c, "mainMenu:back"),
		)
	}

	eventsCount, err = h.eventService.Count(context.Background())
	if err != nil {
		h.logger.Errorf("(user: %d) error while get events count: %v", c.Sender().ID, err)
		return c.Edit(
			banner.Events.Caption(h.layout.Text(c, "technical_issues", err.Error())),
			h.layout.Markup(c, "mainMenu:back"),
		)
	}

	events, err = h.eventService.GetWithPagination(context.Background(), eventsOnPage, p*eventsOnPage, "created_at DESC", user.Role)
	if err != nil {
		h.logger.Errorf(
			"(user: %d) error while get events (offset=%d, limit=%d, order=%s, role=%s): %v",
			c.Sender().ID,
			p*eventsOnPage,
			eventsOnPage,
			user.Role.String(),
			"created_at DESC",
			err,
		)
		return c.Edit(
			banner.Events.Caption(h.layout.Text(c, "technical_issues", err.Error())),
			h.layout.Markup(c, "mainMenu:back"),
		)
	}

	markup := c.Bot().NewMarkup()
	for _, event := range events {
		rows = append(rows, markup.Row(*h.layout.Button(c, "user:events:event", struct {
			ID   string
			Name string
			Page int
		}{
			ID:   event.ID,
			Name: event.Name,
			Page: p,
		})))
	}
	pagesCount := int(eventsCount) / (eventsOnPage + 1)
	if p == 0 {
		prevPage = pagesCount
	} else {
		prevPage = p - 1
	}

	if p >= pagesCount {
		nextPage = 0
	} else {
		nextPage = p + 1
	}

	menuRow = append(menuRow,
		*h.layout.Button(c, "user:events:prev_page", struct {
			Page int
		}{
			Page: prevPage,
		}),
		*h.layout.Button(c, "core:page_counter", struct {
			Page       int
			PagesCount int
		}{
			Page:       p + 1,
			PagesCount: pagesCount + 1,
		}),
		*h.layout.Button(c, "user:events:next_page", struct {
			Page int
		}{
			Page: nextPage,
		}),
	)

	rows = append(
		rows,
		menuRow,
		markup.Row(*h.layout.Button(c, "mainMenu:back")),
	)

	markup.Inline(rows...)

	h.logger.Debugf(
		"(user: %d) user events list (pages_count=%d, page=%d, events_count=%d, next_page=%d, prev_page=%d)",
		c.Sender().ID,
		pagesCount,
		p,
		eventsCount,
		nextPage,
		prevPage,
	)

	_ = c.Edit(
		banner.Events.Caption(h.layout.Text(c, "events_list")),
		markup,
	)
	return nil
}

func (h Handler) event(c tele.Context) error {
	callbackData := strings.Split(c.Callback().Data, " ")
	if len(callbackData) != 2 {
		return errorz.ErrInvalidCallbackData
	}
	eventID := callbackData[0]
	page := callbackData[1]
	h.logger.Infof("(user: %d) edit event (event_id=%s)", c.Sender().ID, eventID)

	event, err := h.eventService.Get(context.Background(), eventID)
	if err != nil {
		h.logger.Errorf("(user: %d) error while get event: %v", c.Sender().ID, err)
		return c.Edit(
			banner.Events.Caption(h.layout.Text(c, "technical_issues", err.Error())),
			h.layout.Markup(c, "user:events:back", struct {
				Page string
			}{
				Page: page,
			}),
		)
	}
	var registered bool
	_, errGetParticipant := h.eventParticipantService.Get(context.Background(), eventID, c.Sender().ID)
	if errGetParticipant != nil {
		if !errors.Is(errGetParticipant, gorm.ErrRecordNotFound) {
			h.logger.Errorf("(user: %d) error while get participant: %v", c.Sender().ID, errGetParticipant)
			return c.Edit(
				banner.Events.Caption(h.layout.Text(c, "technical_issues", errGetParticipant.Error())),
				h.layout.Markup(c, "user:events:back", struct {
					Page string
				}{
					Page: page,
				}),
			)
		}
	} else {
		registered = true
	}

	participantsCount, err := h.eventParticipantService.CountByEventID(context.Background(), eventID)
	if err != nil {
		h.logger.Errorf("(user: %d) error while get participants count: %v", c.Sender().ID, err)
		return c.Edit(
			banner.Events.Caption(h.layout.Text(c, "technical_issues", err.Error())),
			h.layout.Markup(c, "user:events:back", struct {
				Page string
			}{
				Page: page,
			}),
		)
	}

	if c.Callback().Unique == "event_register" {
		if !registered {
			if (event.MaxParticipants == 0 || participantsCount < event.MaxParticipants) && event.RegistrationEnd.After(time.Now().In(location.Location)) {
				_, err = h.eventParticipantService.Register(context.Background(), c.Sender().ID, eventID)
				if err != nil {
					h.logger.Errorf("(user: %d) error while register to event: %v", c.Sender().ID, err)
					return c.Edit(
						banner.Events.Caption(h.layout.Text(c, "technical_issues", err.Error())),
						h.layout.Markup(c, "user:events:back", struct {
							Page string
						}{
							Page: page,
						}),
					)
				}
				registered = true

			} else {
				switch {
				case event.RegistrationEnd.Before(time.Now().In(location.Location)):
					return c.Respond(&tele.CallbackResponse{
						Text:      h.layout.Text(c, "registration_ended"),
						ShowAlert: true,
					})
				case event.MaxParticipants > 0 && participantsCount >= event.MaxParticipants:
					return c.Respond(&tele.CallbackResponse{
						Text:      h.layout.Text(c, "max_participants_reached"),
						ShowAlert: true,
					})
				}
			}
		}
	}

	_ = c.Edit(
		banner.Events.Caption(h.layout.Text(c, "event_text", struct {
			Name                  string
			Description           string
			Location              string
			StartTime             string
			EndTime               string
			RegistrationEnd       string
			MaxParticipants       int
			AfterRegistrationText string
			IsRegistered          bool
		}{
			Name:                  event.Name,
			Description:           event.Description,
			Location:              event.Location,
			StartTime:             event.StartTime.Format("02.01.2006 15:04"),
			EndTime:               event.EndTime.Format("02.01.2006 15:04"),
			RegistrationEnd:       event.RegistrationEnd.Format("02.01.2006 15:04"),
			MaxParticipants:       event.MaxParticipants,
			AfterRegistrationText: event.AfterRegistrationText,
			IsRegistered:          registered,
		})),
		h.layout.Markup(c, "user:events:event", struct {
			ID           string
			Page         string
			IsRegistered bool
		}{
			ID:           eventID,
			Page:         page,
			IsRegistered: registered,
		}))
	return nil
}

func (h Handler) myEvents(c tele.Context) error {
	const eventsOnPage = 10
	h.logger.Infof("(user: %d) edit my events list", c.Sender().ID)

	var (
		p           int
		prevPage    int
		nextPage    int
		err         error
		eventsCount int64
		events      []entity.Event
		rows        []tele.Row
		menuRow     tele.Row
	)
	if c.Callback().Unique != "mainMenu_myEvents" {
		p, err = strconv.Atoi(c.Callback().Data)
		if err != nil {
			return errorz.ErrInvalidCallbackData
		}
	}

	user, err := h.userService.Get(context.Background(), c.Sender().ID)
	if err != nil {
		h.logger.Errorf("(user: %d) error while getting user from db: %v", c.Sender().ID, err)
		return c.Edit(
			banner.Events.Caption(h.layout.Text(c, "technical_issues", err.Error())),
			h.layout.Markup(c, "mainMenu:back"),
		)
	}

	eventsCount, err = h.eventService.Count(context.Background())
	if err != nil {
		h.logger.Errorf("(user: %d) error while get events count: %v", c.Sender().ID, err)
		return c.Edit(
			banner.Events.Caption(h.layout.Text(c, "technical_issues", err.Error())),
			h.layout.Markup(c, "mainMenu:back"),
		)
	}

	events, err = h.userService.GetUserEvents(context.Background(), user.ID, eventsOnPage, p*eventsOnPage)
	if err != nil {
		h.logger.Errorf(
			"(user: %d) error while get my events (offset=%d, limit=%d): %v",
			c.Sender().ID,
			p*eventsOnPage,
			eventsOnPage,
			err,
		)
		return c.Edit(
			banner.Events.Caption(h.layout.Text(c, "technical_issues", err.Error())),
			h.layout.Markup(c, "mainMenu:back"),
		)
	}

	markup := c.Bot().NewMarkup()
	for _, event := range events {
		rows = append(rows, markup.Row(*h.layout.Button(c, "user:myEvents:event", struct {
			ID     string
			Name   string
			Page   int
			IsOver bool
		}{
			ID:     event.ID,
			Name:   event.Name,
			Page:   p,
			IsOver: event.IsOver(),
		})))
	}

	pagesCount := int(eventsCount) / (eventsOnPage + 1)
	if p == 0 {
		prevPage = pagesCount
	} else {
		prevPage = p - 1
	}

	if p >= pagesCount {
		nextPage = 0
	} else {
		nextPage = p + 1
	}

	menuRow = append(menuRow,
		*h.layout.Button(c, "user:myEvents:prev_page", struct {
			Page int
		}{
			Page: prevPage,
		}),
		*h.layout.Button(c, "core:page_counter", struct {
			Page       int
			PagesCount int
		}{
			Page:       p + 1,
			PagesCount: pagesCount + 1,
		}),
		*h.layout.Button(c, "user:myEvents:next_page", struct {
			Page int
		}{
			Page: nextPage,
		}),
	)

	rows = append(
		rows,
		menuRow,
		markup.Row(*h.layout.Button(c, "mainMenu:back")),
	)

	markup.Inline(rows...)

	h.logger.Debugf(
		"(user: %d) user my events list (pages_count=%d, page=%d, events_count=%d, next_page=%d, prev_page=%d)",
		c.Sender().ID,
		pagesCount,
		p,
		eventsCount,
		nextPage,
		prevPage,
	)

	_ = c.Edit(
		banner.Events.Caption(h.layout.Text(c, "my_events_list")),
		markup,
	)
	return nil

}

func (h Handler) UserSetup(group *tele.Group) {
	group.Handle(h.layout.Callback("mainMenu:qr"), h.qrCode)

	group.Handle(h.layout.Callback("mainMenu:events"), h.eventsList)
	group.Handle(h.layout.Callback("user:events:prev_page"), h.eventsList)
	group.Handle(h.layout.Callback("user:events:next_page"), h.eventsList)
	group.Handle(h.layout.Callback("user:events:back"), h.eventsList)
	group.Handle(h.layout.Callback("user:events:event"), h.event)
	group.Handle(h.layout.Callback("user:events:event:register"), h.event)

	group.Handle(h.layout.Callback("mainMenu:my_events"), h.myEvents)
	group.Handle(h.layout.Callback("user:myEvents:prev_page"), h.myEvents)
	group.Handle(h.layout.Callback("user:myEvents:next_page"), h.myEvents)
	group.Handle(h.layout.Callback("user:myEvents:back"), h.myEvents)
}
