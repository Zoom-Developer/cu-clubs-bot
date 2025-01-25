package clubowner

import (
	"context"
	"github.com/Badsnus/cu-clubs-bot/bot/cmd/bot"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/database/postgres"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/database/redis/events"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/common/errorz"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/dto"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/service"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/utils/banner"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/utils/validator"
	"github.com/Badsnus/cu-clubs-bot/bot/pkg/logger/types"
	"github.com/nlypage/intele"
	"github.com/nlypage/intele/collector"
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/layout"
	"slices"
	"strconv"
	"strings"
	"time"
)

type clubService interface {
	Get(ctx context.Context, id string) (*entity.Club, error)
	GetByOwnerID(ctx context.Context, id int64) ([]entity.Club, error)
}

type clubOwnerService interface {
	GetByClubID(ctx context.Context, clubID string) ([]dto.ClubOwner, error)
}

type eventService interface {
	Create(ctx context.Context, event *entity.Event) (*entity.Event, error)
	Update(ctx context.Context, event *entity.Event) (*entity.Event, error)
}

type Handler struct {
	layout *layout.Layout
	logger *types.Logger
	input  *intele.InputManager

	eventsStorage *events.Storage

	clubService      clubService
	clubOwnerService clubOwnerService
	eventService     eventService
}

func NewHandler(b *bot.Bot) *Handler {
	clubStorage := postgres.NewClubStorage(b.DB)
	clubOwnerStorage := postgres.NewClubOwnerStorage(b.DB)
	userStorage := postgres.NewUserStorage(b.DB)
	eventStorage := postgres.NewEventStorage(b.DB)

	return &Handler{
		layout: b.Layout,
		logger: b.Logger,
		input:  b.Input,

		eventsStorage: b.Redis.Events,

		clubService:      service.NewClubService(clubStorage),
		clubOwnerService: service.NewClubOwnerService(clubOwnerStorage, userStorage),
		eventService:     service.NewEventService(eventStorage),
	}
}

func (h Handler) clubsList(c tele.Context) error {
	h.logger.Infof("(user: %d) edit clubs list", c.Sender().ID)

	var (
		err        error
		clubs      []entity.Club
		rows       []tele.Row
		clubsCount int
	)

	clubs, err = h.clubService.GetByOwnerID(context.Background(), c.Sender().ID)
	if err != nil {
		return err
	}
	clubsCount = len(clubs)

	markup := c.Bot().NewMarkup()
	for _, club := range clubs {
		rows = append(rows, markup.Row(*h.layout.Button(c, "clubOwner:myClubs:club", struct {
			ID   string
			Name string
		}{
			ID:   club.ID,
			Name: club.Name,
		})))
	}

	rows = append(
		rows,
		markup.Row(*h.layout.Button(c, "mainMenu:back")),
	)

	markup.Inline(rows...)

	h.logger.Debugf("(user: %d) club owner clubs list", c.Sender().ID)

	return c.Edit(
		banner.ClubOwner.Caption(h.layout.Text(c, "clubs_list", clubsCount)),
		markup,
	)
}

func (h Handler) clubMenu(c tele.Context) error {
	clubID := c.Callback().Data
	if clubID == "" {
		return errorz.ErrInvalidCallbackData
	}

	h.logger.Infof("(user: %d) edit club menu (club_id=%s)", c.Sender().ID, clubID)

	clubOwners, err := h.clubOwnerService.GetByClubID(context.Background(), clubID)
	if err != nil {
		h.logger.Errorf("(user: %d) error while get club owners: %v", c.Sender().ID, err)
		return c.Send(
			banner.ClubOwner.Caption(h.layout.Text(c, "technical_issues", err.Error())),
			h.layout.Markup(c, "mainMenu:back"),
		)
	}

	club, err := h.clubService.Get(context.Background(), clubID)
	if err != nil {
		h.logger.Errorf("(user: %d) error while get club: %v", c.Sender().ID, err)
		return c.Send(
			banner.ClubOwner.Caption(h.layout.Text(c, "technical_issues", err.Error())),
			h.layout.Markup(c, "mainMenu:back"),
		)
	}

	clubs, err := h.clubService.GetByOwnerID(context.Background(), c.Sender().ID)
	if err != nil {
		h.logger.Errorf("(user: %d) error while get clubs: %v", c.Sender().ID, err)
		return c.Send(
			banner.ClubOwner.Caption(h.layout.Text(c, "clubs", err.Error())),
			h.layout.Markup(c, "mainMenu:back"),
		)
	}

	menuMarkup := h.layout.Markup(c, "clubOwner:club:menu", struct {
		ID string
	}{
		ID: clubID,
	})

	if len(clubs) == 1 {
		menuMarkup.InlineKeyboard = append(menuMarkup.InlineKeyboard, []tele.InlineButton{*h.layout.Button(c, "mainMenu:back").Inline()})
	}
	if len(clubs) > 1 {
		menuMarkup.InlineKeyboard = append(menuMarkup.InlineKeyboard, []tele.InlineButton{*h.layout.Button(c, "clubOwner:myClubs:back").Inline()})
	}

	return c.Edit(
		banner.ClubOwner.Caption(h.layout.Text(c, "club_owner_club_menu_text", struct {
			Club   entity.Club
			Owners []dto.ClubOwner
		}{
			Club:   *club,
			Owners: clubOwners,
		})),
		menuMarkup,
	)
}

func (h Handler) createEvent(c tele.Context) error {
	clubID := c.Callback().Data
	if clubID == "" {
		return errorz.ErrInvalidCallbackData
	}

	h.logger.Infof("(user: %d) create new event request(club=%s)", c.Sender().ID, clubID)

	club, err := h.clubService.Get(context.Background(), clubID)
	if err != nil {
		h.logger.Errorf("(user: %d) error while get clubs: %v", c.Sender().ID, err)
		return c.Send(
			banner.ClubOwner.Caption(h.layout.Text(c, "technical_issues", err.Error())),
			h.layout.Markup(c, "clubOwner:club:back", struct {
				ID string
			}{
				ID: club.ID,
			}),
		)
	}

	inputCollector := collector.New()
	inputCollector.Collect(c.Message())

	isFirst := true

	var steps []struct {
		promptKey  string
		errorKey   string
		result     *string
		validator  func(string, map[string]interface{}) bool
		paramsFunc func(map[string]interface{}) map[string]interface{}
	}

	steps = []struct {
		promptKey  string
		errorKey   string
		result     *string
		validator  func(string, map[string]interface{}) bool
		paramsFunc func(map[string]interface{}) map[string]interface{}
	}{
		{
			promptKey:  "input_event_name",
			errorKey:   "invalid_event_name",
			result:     new(string),
			validator:  validator.EventName,
			paramsFunc: nil,
		},

		{
			promptKey:  "input_event_description",
			errorKey:   "invalid_event_description",
			result:     new(string),
			validator:  validator.EventDescription,
			paramsFunc: nil,
		},

		{
			promptKey:  "input_event_location",
			errorKey:   "invalid_event_location",
			result:     new(string),
			validator:  validator.EventLocation,
			paramsFunc: nil,
		},

		{
			promptKey:  "input_event_start_time",
			errorKey:   "invalid_event_start_time",
			result:     new(string),
			validator:  validator.EventStartTime,
			paramsFunc: nil,
		},

		{
			promptKey: "input_event_end_time",
			errorKey:  "invalid_event_end_time",
			result:    new(string),
			validator: validator.EventEndTime,
			paramsFunc: func(params map[string]interface{}) map[string]interface{} {
				if params == nil {
					params = make(map[string]interface{})
				}
				params["startTime"] = *steps[3].result
				return params
			},
		},
		{
			promptKey: "input_event_registered_end_time",
			errorKey:  "invalid_event_registered_end_time",
			result:    new(string),
			validator: validator.EventRegisteredEndTime,
			paramsFunc: func(params map[string]interface{}) map[string]interface{} {
				if params == nil {
					params = make(map[string]interface{})
				}
				params["startTime"] = *steps[3].result
				return params
			},
		},
		{
			promptKey:  "input_after_registration_text",
			errorKey:   "invalid_after_registration_text",
			result:     new(string),
			validator:  validator.EventAfterRegistrationText,
			paramsFunc: nil,
		},
		{
			promptKey:  "input_max_participants",
			errorKey:   "invalid_max_participants",
			result:     new(string),
			validator:  validator.MaxParticipants,
			paramsFunc: nil,
		},
		{
			promptKey:  "input_expected_participants",
			errorKey:   "invalid_expected_participants",
			result:     new(string),
			validator:  validator.ExpectedParticipants,
			paramsFunc: nil,
		},
	}

	for _, step := range steps {
		done := false

		var params map[string]interface{}
		if step.paramsFunc != nil {
			params = step.paramsFunc(params)
		}

		if isFirst {
			_ = c.Edit(
				banner.ClubOwner.Caption(h.layout.Text(c, step.promptKey)),
				h.layout.Markup(c, "clubOwner:club:back", struct {
					ID string
				}{
					ID: club.ID,
				}),
			)
		} else {
			_ = inputCollector.Send(c,
				banner.ClubOwner.Caption(h.layout.Text(c, step.promptKey)),
				h.layout.Markup(c, "clubOwner:club:back", struct {
					ID string
				}{
					ID: club.ID,
				}),
			)
		}
		isFirst = false

		for !done {
			message, canceled, errGet := h.input.Get(context.Background(), c.Sender().ID, 0)
			if message != nil {
				inputCollector.Collect(message)
			}
			switch {
			case canceled:
				_ = inputCollector.Clear(c, collector.ClearOptions{IgnoreErrors: true, ExcludeLast: true})
				return nil
			case errGet != nil:
				h.logger.Errorf("(user: %d) error while input step (%s): %v", c.Sender().ID, step.promptKey, errGet)
				_ = inputCollector.Send(c,
					banner.ClubOwner.Caption(h.layout.Text(c, "input_error", h.layout.Text(c, step.promptKey))),
					h.layout.Markup(c, "clubOwner:club:back", struct {
						ID string
					}{
						ID: club.ID,
					}),
				)
			case !step.validator(message.Text, params):
				_ = inputCollector.Send(c,
					banner.ClubOwner.Caption(h.layout.Text(c, step.errorKey)),
					h.layout.Markup(c, "clubOwner:club:back", struct {
						ID string
					}{
						ID: club.ID,
					}),
				)
			case step.validator(message.Text, params):
				*step.result = message.Text
				_ = inputCollector.Clear(c, collector.ClearOptions{IgnoreErrors: true})
				done = true
			}
		}
	}

	// Результаты ввода
	const timeLayout = "02.01.2006 15:04"

	var (
		eventName                    string
		eventDescription             string
		eventLocation                string
		eventStartTime               time.Time
		eventStartTimeStr            string
		eventEndTime                 time.Time
		eventEndTimeStr              string
		eventRegistrationEndTime     time.Time
		eventRegistrationEndTimeStr  string
		eventAfterRegistrationText   string
		eventMaxParticipants         int
		eventMaxExpectedParticipants int
	)

	eventName = *steps[0].result
	eventDescription = *steps[1].result
	if *steps[1].result == "skip" {
		eventDescription = ""
	}
	eventLocation = *steps[2].result
	eventStartTime, _ = time.Parse(timeLayout, *steps[3].result)
	eventStartTimeStr = *steps[3].result
	eventEndTime, _ = time.Parse(timeLayout, *steps[4].result)
	eventEndTimeStr = *steps[4].result
	if *steps[4].result == "skip" {
		eventEndTime = time.Time{}
		eventEndTimeStr = ""
	}
	eventRegistrationEndTime, err = time.Parse(timeLayout, *steps[5].result)
	eventRegistrationEndTimeStr = *steps[5].result
	eventAfterRegistrationText = *steps[6].result
	eventMaxParticipants, _ = strconv.Atoi(*steps[7].result)
	eventMaxExpectedParticipants, _ = strconv.Atoi(*steps[8].result)

	event := entity.Event{
		ClubID:                club.ID,
		Name:                  eventName,
		Description:           eventDescription,
		AfterRegistrationText: eventAfterRegistrationText,
		Location:              eventLocation,
		StartTime:             eventStartTime,
		EndTime:               eventEndTime,
		RegistrationEnd:       eventRegistrationEndTime,
		MaxParticipants:       eventMaxParticipants,
		ExpectedParticipants:  eventMaxExpectedParticipants,
	}

	h.eventsStorage.Set(c.Sender().ID, event, 0)

	markup := h.layout.Markup(c, "clubOwner:createClub:confirm", struct {
		ID string
	}{
		ID: clubID,
	})
	for _, role := range club.AllowedRoles {
		markup.InlineKeyboard = append(
			[][]tele.InlineButton{{*h.layout.Button(c, "clubOwner:create_event:role", struct {
				Role     entity.Role
				ID       string
				RoleName string
				Allowed  bool
			}{
				Role:     entity.Role(role),
				ID:       club.ID,
				RoleName: h.layout.Text(c, role),
				Allowed:  slices.Contains(event.AllowedRoles, role),
			}).Inline()}},
			markup.InlineKeyboard...,
		)
	}

	return c.Send(
		banner.ClubOwner.Caption(h.layout.Text(c, "event_confirmation", struct {
			Name                  string
			Description           string
			Location              string
			StartTime             string
			EndTime               string
			RegistrationEnd       string
			AfterRegistrationText string
			MaxParticipants       int
			ExpectedParticipants  int
		}{
			Name:                  event.Name,
			Description:           event.Description,
			Location:              event.Location,
			StartTime:             eventStartTimeStr,
			EndTime:               eventEndTimeStr,
			RegistrationEnd:       eventRegistrationEndTimeStr,
			AfterRegistrationText: event.AfterRegistrationText,
			MaxParticipants:       event.MaxParticipants,
			ExpectedParticipants:  event.ExpectedParticipants,
		})),
		markup,
	)
}

func (h Handler) eventAllowedRoles(c tele.Context) error {
	data := strings.Split(c.Callback().Data, " ")
	if len(data) != 2 {
		return errorz.ErrInvalidCallbackData
	}
	clubID, role := data[0], data[1]

	club, err := h.clubService.Get(context.Background(), clubID)
	if err != nil {
		h.logger.Errorf("(user: %d) error while get clubs: %v", c.Sender().ID, err)
		return c.Send(
			banner.ClubOwner.Caption(h.layout.Text(c, "technical_issues", err.Error())),
			h.layout.Markup(c, "clubOwner:club:back", struct {
				ID string
			}{
				ID: club.ID,
			}),
		)
	}

	event, err := h.eventsStorage.Get(c.Sender().ID)
	if err != nil {
		return c.Edit(
			banner.ClubOwner.Caption(h.layout.Text(c, "technical_issues", err.Error())),
			h.layout.Markup(c, "clubOwner:club:back", struct {
				ID string
			}{
				ID: clubID,
			}),
		)
	}

	var (
		contains bool
		roleI    int
	)
	for i, r := range event.AllowedRoles {
		if r == role {
			contains = true
			roleI = i
			break
		}
	}
	if contains {
		event.AllowedRoles = append(event.AllowedRoles[:roleI], event.AllowedRoles[roleI+1:]...)
	} else {
		event.AllowedRoles = append(event.AllowedRoles, role)
	}

	h.eventsStorage.Set(c.Sender().ID, event, 0)

	markup := h.layout.Markup(c, "clubOwner:createClub:confirm", struct {
		ID string
	}{
		ID: club.ID,
	})

	for _, role = range club.AllowedRoles {
		markup.InlineKeyboard = append(
			[][]tele.InlineButton{{*h.layout.Button(c, "clubOwner:create_event:role", struct {
				Role     entity.Role
				ID       string
				RoleName string
				Allowed  bool
			}{
				Role:     entity.Role(role),
				ID:       club.ID,
				RoleName: h.layout.Text(c, role),
				Allowed:  slices.Contains(event.AllowedRoles, role),
			}).Inline()}},
			markup.InlineKeyboard...,
		)
	}
	const timeLayout = "02.01.2006 15:04"

	eventTimeStr := event.EndTime.Format(timeLayout)
	if event.EndTime.IsZero() {
		eventTimeStr = ""
	}

	return c.Edit(
		banner.ClubOwner.Caption(h.layout.Text(c, "event_confirmation", struct {
			Name                  string
			Description           string
			Location              string
			StartTime             string
			EndTime               string
			RegistrationEnd       string
			AfterRegistrationText string
			MaxParticipants       int
			ExpectedParticipants  int
		}{
			Name:                  event.Name,
			Description:           event.Description,
			Location:              event.Location,
			StartTime:             event.StartTime.Format(timeLayout),
			EndTime:               eventTimeStr,
			RegistrationEnd:       event.RegistrationEnd.Format(timeLayout),
			AfterRegistrationText: event.AfterRegistrationText,
			MaxParticipants:       event.MaxParticipants,
			ExpectedParticipants:  event.ExpectedParticipants,
		})),
		markup,
	)
}

func (h Handler) confirmEventCreation(c tele.Context) error {
	clubID := c.Callback().Data
	if clubID == "" {
		return errorz.ErrInvalidCallbackData
	}

	event, err := h.eventsStorage.Get(c.Sender().ID)
	if err != nil {
		return c.Edit(
			banner.ClubOwner.Caption(h.layout.Text(c, "technical_issues", err.Error())),
			h.layout.Markup(c, "clubOwner:club:back", struct {
				ID string
			}{
				ID: clubID,
			}),
		)
	}

	_, err = h.eventService.Create(context.Background(), &event)
	if err != nil {
		return c.Edit(
			banner.ClubOwner.Caption(h.layout.Text(c, "technical_issues", err.Error())),
			h.layout.Markup(c, "clubOwner:club:back", struct {
				ID string
			}{
				ID: clubID,
			}),
		)
	}

	return c.Edit(
		banner.ClubOwner.Caption(h.layout.Text(c, "event_created", struct {
			Name string
		}{
			Name: event.Name,
		})),
		h.layout.Markup(c, "clubOwner:club:back", struct {
			ID string
		}{
			ID: clubID,
		}))
}

func (h Handler) ClubOwnerSetup(group *tele.Group) {
	group.Handle(h.layout.Callback("clubOwner:myClubs"), h.clubsList)
	group.Handle(h.layout.Callback("clubOwner:myClubs:back"), h.clubsList)
	group.Handle(h.layout.Callback("clubOwner:myClubs:club"), h.clubMenu)
	group.Handle(h.layout.Callback("clubOwner:club:back"), h.clubMenu)
	group.Handle(h.layout.Callback("clubOwner:club:create_event"), h.createEvent)
	group.Handle(h.layout.Callback("clubOwner:create_event:refill"), h.createEvent)
	group.Handle(h.layout.Callback("clubOwner:create_event:confirm"), h.confirmEventCreation)
	group.Handle(h.layout.Callback("clubOwner:create_event:role"), h.eventAllowedRoles)
}
