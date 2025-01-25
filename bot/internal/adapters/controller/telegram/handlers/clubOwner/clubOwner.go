package clubowner

import (
	"context"
	"errors"
	"github.com/Badsnus/cu-clubs-bot/bot/cmd/bot"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/database/postgres"
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
	"gorm.io/gorm"
	"strconv"
	"strings"
)

type clubService interface {
	Get(ctx context.Context, id string) (*entity.Club, error)
	GetByOwnerID(ctx context.Context, id int64) ([]entity.Club, error)
	Update(ctx context.Context, club *entity.Club) (*entity.Club, error)
}

type clubOwnerService interface {
	GetByClubID(ctx context.Context, clubID string) ([]dto.ClubOwner, error)
	Get(ctx context.Context, clubID string, userID int64) (*entity.ClubOwner, error)
	Add(ctx context.Context, userID int64, clubID string) (*entity.ClubOwner, error)
	Update(ctx context.Context, clubOwner *entity.ClubOwner) (*entity.ClubOwner, error)
}

type userService interface {
	Get(ctx context.Context, userID int64) (*entity.User, error)
}

type Handler struct {
	layout *layout.Layout
	logger *types.Logger
	input  *intele.InputManager

	clubService      clubService
	clubOwnerService clubOwnerService
	userService      userService
}

func NewHandler(b *bot.Bot) *Handler {
	clubStorage := postgres.NewClubStorage(b.DB)
	clubOwnerStorage := postgres.NewClubOwnerStorage(b.DB)
	userStorage := postgres.NewUserStorage(b.DB)

	return &Handler{
		layout: b.Layout,
		logger: b.Logger,
		input:  b.Input,

		clubService:      service.NewClubService(clubStorage),
		clubOwnerService: service.NewClubOwnerService(clubOwnerStorage, userStorage),
		userService:      service.NewUserService(userStorage, nil, nil),
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
		banner.Menu.Caption(h.layout.Text(c, "clubs_list", clubsCount)),
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
			banner.Menu.Caption(h.layout.Text(c, "technical_issues", err.Error())),
			h.layout.Markup(c, "mainMenu:back"),
		)
	}

	club, err := h.clubService.Get(context.Background(), clubID)
	if err != nil {
		h.logger.Errorf("(user: %d) error while get club: %v", c.Sender().ID, err)
		return c.Send(
			banner.Menu.Caption(h.layout.Text(c, "technical_issues", err.Error())),
			h.layout.Markup(c, "mainMenu:back"),
		)
	}

	clubs, err := h.clubService.GetByOwnerID(context.Background(), c.Sender().ID)
	if err != nil {
		h.logger.Errorf("(user: %d) error while get clubs: %v", c.Sender().ID, err)
		return c.Send(
			banner.Menu.Caption(h.layout.Text(c, "clubs", err.Error())),
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
		banner.Menu.Caption(h.layout.Text(c, "club_owner_club_menu_text", struct {
			Club   entity.Club
			Owners []dto.ClubOwner
		}{
			Club:   *club,
			Owners: clubOwners,
		})),
		menuMarkup,
	)
}

func (h Handler) clubSettings(c tele.Context) error {
	if c.Callback().Data == "" {
		return errorz.ErrInvalidCallbackData
	}
	h.logger.Infof("(user: %d) edit club settings (club_id=%s)", c.Sender().ID, c.Callback().Data)

	club, err := h.clubService.Get(context.Background(), c.Callback().Data)
	if err != nil {
		h.logger.Errorf("(user: %d) error while get club: %v", c.Sender().ID, err)
		return c.Send(
			banner.Menu.Caption(h.layout.Text(c, "technical_issues", err.Error())),
			h.layout.Markup(c, "mainMenu:back"),
		)
	}

	clubOwners, err := h.clubOwnerService.GetByClubID(context.Background(), c.Callback().Data)
	if err != nil {
		h.logger.Errorf("(user: %d) error while get club owners: %v", c.Sender().ID, err)
		return c.Send(
			banner.Menu.Caption(h.layout.Text(c, "technical_issues", err.Error())),
			h.layout.Markup(c, "mainMenu:back"),
		)
	}

	return c.Edit(
		banner.ClubOwner.Caption(h.layout.Text(c, "club_settings_text", struct {
			Club   entity.Club
			Owners []dto.ClubOwner
		}{
			Club:   *club,
			Owners: clubOwners,
		})),
		h.layout.Markup(c, "clubOwner:club:settings", struct {
			ID string
		}{
			ID: club.ID,
		}),
	)
}

func (h Handler) editName(c tele.Context) error {
	h.logger.Infof("(user: %d) edit club name", c.Sender().ID)

	club, err := h.clubService.Get(context.Background(), c.Callback().Data)
	if err != nil {
		h.logger.Errorf("(user: %d) error while get club: %v", c.Sender().ID, err)
		return c.Send(
			banner.ClubOwner.Caption(h.layout.Text(c, "technical_issues", err.Error())),
			h.layout.Markup(c, "mainMenu:back"),
		)
	}

	inputCollector := collector.New()
	_ = c.Edit(
		banner.ClubOwner.Caption(h.layout.Text(c, "input_club_name")),
		h.layout.Markup(c, "clubOwner:club:settings:back", struct {
			ID string
		}{
			ID: club.ID,
		}),
	)
	inputCollector.Collect(c.Message())

	var (
		clubName string
		done     bool
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
			h.logger.Errorf("(user: %d) error while input club name: %v", c.Sender().ID, errGet)
			_ = inputCollector.Send(c,
				banner.ClubOwner.Caption(h.layout.Text(c, "input_error", h.layout.Text(c, "input_club_name"))),
				h.layout.Markup(c, "clubOwner:club:settings:back", struct {
					ID string
				}{
					ID: club.ID,
				}),
			)
		case message == nil:
			_ = inputCollector.Send(c,
				banner.ClubOwner.Caption(h.layout.Text(c, "input_error", h.layout.Text(c, "input_club_description"))),
				h.layout.Markup(c, "clubOwner:club:settings:back", struct {
					ID string
				}{
					ID: club.ID,
				}),
			)
		case !validator.ClubName(message.Text):
			_ = inputCollector.Send(c,
				banner.ClubOwner.Caption(h.layout.Text(c, "invalid_club_name")),
				h.layout.Markup(c, "clubOwner:club:settings:back", struct {
					ID string
				}{
					ID: club.ID,
				}),
			)
		case validator.ClubName(message.Text):
			clubName = message.Text
			_ = inputCollector.Clear(c, collector.ClearOptions{IgnoreErrors: true})
			done = true
		}
		if done {
			break
		}
	}

	club.Name = clubName
	_, err = h.clubService.Update(context.Background(), club)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return c.Send(
				banner.Menu.Caption(h.layout.Text(c, "club_already_exists")),
				h.layout.Markup(c, "clubOwner:club:settings:back", struct {
					ID string
				}{
					ID: club.ID,
				}),
			)
		}

		h.logger.Errorf("(user: %d) error while update club name: %v", c.Sender().ID, err)
		return c.Send(
			banner.ClubOwner.Caption(h.layout.Text(c, "technical_issues", err.Error())),
			h.layout.Markup(c, "clubOwner:club:settings:back", struct {
				ID string
			}{
				ID: club.ID,
			}),
		)
	}

	return c.Send(
		banner.ClubOwner.Caption(h.layout.Text(c, "name_changed")),
		h.layout.Markup(c, "clubOwner:club:settings:back", struct {
			ID string
		}{
			ID: club.ID,
		}),
	)
}

func (h Handler) editDescription(c tele.Context) error {
	h.logger.Infof("(user: %d) edit club description", c.Sender().ID)

	club, err := h.clubService.Get(context.Background(), c.Callback().Data)
	if err != nil {
		h.logger.Errorf("(user: %d) error while get club: %v", c.Sender().ID, err)
		return c.Send(
			banner.ClubOwner.Caption(h.layout.Text(c, "technical_issues", err.Error())),
			h.layout.Markup(c, "mainMenu:back"),
		)
	}

	inputCollector := collector.New()
	_ = c.Edit(
		banner.ClubOwner.Caption(h.layout.Text(c, "input_club_description")),
		h.layout.Markup(c, "clubOwner:club:settings:back", struct {
			ID string
		}{
			ID: club.ID,
		}),
	)
	inputCollector.Collect(c.Message())

	var (
		clubDescription string
		done            bool
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
			h.logger.Errorf("(user: %d) error while input club name: %v", c.Sender().ID, errGet)
			_ = inputCollector.Send(c,
				banner.ClubOwner.Caption(h.layout.Text(c, "input_error", h.layout.Text(c, "input_club_description"))),
				h.layout.Markup(c, "clubOwner:club:settings:back", struct {
					ID string
				}{
					ID: club.ID,
				}),
			)
		case message == nil:
			_ = inputCollector.Send(c,
				banner.ClubOwner.Caption(h.layout.Text(c, "input_error", h.layout.Text(c, "input_club_description"))),
				h.layout.Markup(c, "clubOwner:club:settings:back", struct {
					ID string
				}{
					ID: club.ID,
				}),
			)
		case !validator.ClubDescription(message.Text, nil):
			_ = inputCollector.Send(c,
				banner.ClubOwner.Caption(h.layout.Text(c, "invalid_club_description")),
				h.layout.Markup(c, "clubOwner:club:settings:back", struct {
					ID string
				}{
					ID: club.ID,
				}),
			)
		case validator.ClubDescription(message.Text, nil):
			clubDescription = message.Text
			_ = inputCollector.Clear(c, collector.ClearOptions{IgnoreErrors: true})
			done = true
		}
		if done {
			break
		}
	}

	club.Description = clubDescription
	_, err = h.clubService.Update(context.Background(), club)
	if err != nil {
		h.logger.Errorf("(user: %d) error while update club description: %v", c.Sender().ID, err)
		return c.Send(
			banner.ClubOwner.Caption(h.layout.Text(c, "technical_issues", err.Error())),
			h.layout.Markup(c, "clubOwner:club:settings:back", struct {
				ID string
			}{
				ID: club.ID,
			}),
		)
	}

	return c.Send(
		banner.ClubOwner.Caption(h.layout.Text(c, "description_changed")),
		h.layout.Markup(c, "clubOwner:club:settings:back", struct {
			ID string
		}{
			ID: club.ID,
		}),
	)
}

func (h Handler) addOwner(c tele.Context) error {
	if c.Callback().Data == "" {
		return errorz.ErrInvalidCallbackData
	}
	clubID := c.Callback().Data

	inputCollector := collector.New()
	inputCollector.Collect(c.Message())
	club, err := h.clubService.Get(context.Background(), clubID)
	if err != nil {
		_ = inputCollector.Clear(c, collector.ClearOptions{IgnoreErrors: true})
		h.logger.Errorf("(user: %d) error while get club: %v", c.Sender().ID, err)
		return c.Send(
			banner.Menu.Caption(h.layout.Text(c, "technical_issues", err.Error())),
			h.layout.Markup(c, "mainMenu:back"),
		)
	}

	h.logger.Infof("(user: %d) add club owner (club_id=%s)", c.Sender().ID, clubID)
	_ = c.Edit(
		banner.Menu.Caption(h.layout.Text(c, "input_user_id")),
		h.layout.Markup(c, "clubOwner:club:settings:back", struct {
			ID string
		}{
			ID: club.ID,
		}),
	)

	var (
		user *entity.User
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
				banner.Menu.Caption(h.layout.Text(c, "input_error", h.layout.Text(c, "input_user_id"))),
				h.layout.Markup(c, "clubOwner:club:settings:back", struct {
					ID string
				}{
					ID: club.ID,
				}),
			)
		case message == nil:
			_ = inputCollector.Send(c,
				banner.Menu.Caption(h.layout.Text(c, "input_error", h.layout.Text(c, "input_user_id"))),
				h.layout.Markup(c, "clubOwner:club:settings:back", struct {
					ID string
				}{
					ID: club.ID,
				}),
			)
		default:
			userID, err := strconv.ParseInt(message.Text, 10, 64)
			if err != nil {
				_ = inputCollector.Send(c,
					banner.Menu.Caption(h.layout.Text(c, "input_user_id")),
					h.layout.Markup(c, "clubOwner:club:settings:back", struct {
						ID string
					}{
						ID: club.ID,
					}),
				)
				break
			}

			user, err = h.userService.Get(context.Background(), userID)
			if err != nil {
				_ = inputCollector.Send(c,
					banner.Menu.Caption(h.layout.Text(c, "user_not_found", struct {
						ID   int64
						Text string
					}{
						ID:   userID,
						Text: h.layout.Text(c, "input_user_id"),
					})),
					h.layout.Markup(c, "clubOwner:club:settings:back", struct {
						ID string
					}{
						ID: club.ID,
					}),
				)
				break
			}
			done = true
		}
		if done {
			break
		}
	}

	_, err = h.clubOwnerService.Add(context.Background(), user.ID, club.ID)
	if err != nil {
		_ = inputCollector.Clear(c, collector.ClearOptions{IgnoreErrors: true})
		h.logger.Errorf(
			"(user: %d) error while add club owner (club_id=%s, user_id=%d): %v",
			c.Sender().ID,
			clubID,
			user.ID,
			err,
		)
		return c.Send(
			banner.Menu.Caption(h.layout.Text(c, "technical_issues", err.Error())),
			h.layout.Markup(c, "clubOwner:club:settings:back", struct {
				ID string
			}{
				ID: club.ID,
			}),
		)
	}

	h.logger.Infof(
		"(user: %d) club owner added (club_id=%s, user_id=%d)",
		c.Sender().ID,
		clubID,
		user.ID,
	)

	_ = inputCollector.Clear(c, collector.ClearOptions{IgnoreErrors: true})
	return c.Send(
		banner.Menu.Caption(h.layout.Text(c, "club_owner_added", struct {
			Club entity.Club
			User entity.User
		}{
			Club: *club,
			User: *user,
		})),
		h.layout.Markup(c, "clubOwner:club:settings:back", struct {
			ID string
		}{
			ID: club.ID,
		}),
	)
}

func (h Handler) warnings(c tele.Context) error {
	var (
		clubID string
	)
	if c.Callback().Unique == "clubOwner_club_warnings" {
		h.logger.Infof("(user: %d) club warning settings (club_id=%s)", c.Sender().ID, c.Callback().Data)
		if c.Callback().Data == "" {
			return errorz.ErrInvalidCallbackData
		}
		clubID = c.Callback().Data
	}

	if c.Callback().Unique == "cOwner_warnings" {
		h.logger.Infof("(user: %d) club owner edit warning settings (club_id=%s, user_id=%s)", c.Sender().ID, c.Callback().Data, c.Callback().Data)
		callbackData := strings.Split(c.Callback().Data, " ")
		if len(callbackData) != 2 {
			return errorz.ErrInvalidCallbackData
		}
		clubID = callbackData[0]
		userID, err := strconv.ParseInt(callbackData[1], 10, 64)
		if err != nil {
			return errorz.ErrInvalidCallbackData
		}

		clubOwner, err := h.clubOwnerService.Get(context.Background(), clubID, userID)
		if err != nil {
			h.logger.Errorf("(user: %d) error while get club owner: %v", c.Sender().ID, err)
			return c.Edit(
				banner.Menu.Caption(h.layout.Text(c, "technical_issues", err.Error())),
				h.layout.Markup(c, "clubOwner:club:settings:back", struct {
					ID string
				}{
					ID: clubID,
				}),
			)
		}
		clubOwner.Warnings = !clubOwner.Warnings
		_, err = h.clubOwnerService.Update(context.Background(), clubOwner)
		if err != nil {
			h.logger.Errorf("(user: %d) error while update club owner: %v", c.Sender().ID, err)
			return c.Edit(
				banner.Menu.Caption(h.layout.Text(c, "technical_issues", err.Error())),
				h.layout.Markup(c, "clubOwner:club:settings:back", struct {
					ID string
				}{
					ID: clubID,
				}),
			)
		}
	}

	owners, err := h.clubOwnerService.GetByClubID(context.Background(), clubID)
	if err != nil {
		h.logger.Errorf("(user: %d) error while get club owners: %v", c.Sender().ID, err)
		return c.Edit(
			banner.Menu.Caption(h.layout.Text(c, "technical_issues", err.Error())),
			h.layout.Markup(c, "clubOwner:club:settings:back", struct {
				ID string
			}{
				ID: clubID,
			}),
		)
	}
	warningsMarkup := h.layout.Markup(c, "clubOwner:club:settings:warnings", struct {
		ID string
	}{
		ID: clubID,
	})

	for _, owner := range owners {
		warningsMarkup.InlineKeyboard = append(
			[][]tele.InlineButton{{*h.layout.Button(c, "clubOwner:club:settings:warnings:user", owner).Inline()}},
			warningsMarkup.InlineKeyboard...,
		)
	}

	return c.Edit(
		banner.ClubOwner.Caption(h.layout.Text(c, "warnings_text")),
		warningsMarkup,
	)
}

func (h Handler) ClubOwnerSetup(group *tele.Group) {
	group.Handle(h.layout.Callback("clubOwner:myClubs"), h.clubsList)
	group.Handle(h.layout.Callback("clubOwner:myClubs:back"), h.clubsList)
	group.Handle(h.layout.Callback("clubOwner:myClubs:club"), h.clubMenu)
	group.Handle(h.layout.Callback("clubOwner:club:back"), h.clubMenu)
	group.Handle(h.layout.Callback("clubOwner:club:settings"), h.clubSettings)
	group.Handle(h.layout.Callback("clubOwner:club:settings:back"), h.clubSettings)
	group.Handle(h.layout.Callback("clubOwner:club:settings:edit_name"), h.editName)
	group.Handle(h.layout.Callback("clubOwner:club:settings:edit_description"), h.editDescription)
	group.Handle(h.layout.Callback("clubOwner:club:settings:add_owner"), h.addOwner)
	group.Handle(h.layout.Callback("clubOwner:club:settings:warnings"), h.warnings)
	group.Handle(h.layout.Callback("clubOwner:club:settings:warnings:user"), h.warnings)
}
