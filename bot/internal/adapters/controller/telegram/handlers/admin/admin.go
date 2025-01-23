package admin

import (
	"context"
	"errors"
	"github.com/Badsnus/cu-clubs-bot/bot/cmd/bot"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/database/postgres"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/common/errorz"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/dto"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/service"
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

type adminUserService interface {
	Count(ctx context.Context) (int64, error)
	GetWithPagination(ctx context.Context, offset, limit int, order string) ([]entity.User, error)
	Get(ctx context.Context, userID int64) (*entity.User, error)
	Ban(ctx context.Context, userID int64) (*entity.User, error)
	GetAll(ctx context.Context) ([]entity.User, error)
}

type clubService interface {
	Create(ctx context.Context, club *entity.Club) (*entity.Club, error)
	GetWithPagination(ctx context.Context, offset, limit int, order string) ([]entity.Club, error)
	Get(ctx context.Context, id string) (*entity.Club, error)
	Update(ctx context.Context, club *entity.Club) (*entity.Club, error)
	Delete(ctx context.Context, id string) error
	Count(ctx context.Context) (int64, error)
}

type clubOwnerService interface {
	Add(ctx context.Context, userID int64, clubID string) (*entity.ClubOwner, error)
	Remove(ctx context.Context, userID int64, clubID string) error
	Get(ctx context.Context, clubID string, userID int64) (*entity.ClubOwner, error)
	GetByClubID(ctx context.Context, clubID string) ([]dto.ClubOwner, error)
	GetByUserID(ctx context.Context, userID int64) ([]dto.ClubOwner, error)
}

type Handler struct {
	layout *layout.Layout
	logger *types.Logger
	bot    *bot.Bot
	input  *intele.InputManager

	adminUserService adminUserService
	clubService      clubService
	clubOwnerService clubOwnerService
}

func New(b *bot.Bot) *Handler {
	userStorage := postgres.NewUserStorage(b.DB)
	clubStorage := postgres.NewClubStorage(b.DB)
	clubOwnerStorage := postgres.NewClubOwnerStorage(b.DB)

	return &Handler{
		layout:           b.Layout,
		logger:           b.Logger,
		bot:              b,
		input:            b.Input,
		adminUserService: service.NewUserService(userStorage, nil, nil),
		clubService:      service.NewClubService(clubStorage),
		clubOwnerService: service.NewClubOwnerService(clubOwnerStorage, userStorage),
	}
}

func (h Handler) adminMenu(c tele.Context) error {
	h.logger.Infof("(user: %d) edit admin menu", c.Sender().ID)
	commands := h.layout.Commands()
	commands = append(commands, tele.Command{Text: "/ban", Description: h.layout.Text(c, "command_ban")})
	errSetCommands := c.Bot().SetCommands(commands)
	if errSetCommands != nil {
		h.logger.Errorf("(user: %d) error while set admin commands: %v", c.Sender().ID, errSetCommands)
	}

	return c.Edit(
		h.layout.Text(c, "admin_menu_text", c.Sender().Username),
		h.layout.Markup(c, "admin:menu"),
	)
}

func (h Handler) createClub(c tele.Context) error {
	h.logger.Infof("(user: %d) create new club request", c.Sender().ID)
	inputCollector := collector.New()
	_ = c.Edit(
		h.layout.Text(c, "input_club_name"),
		h.layout.Markup(c, "admin:backToMenu"),
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
				h.layout.Text(c, "input_error", h.layout.Text(c, "input_club_name")),
				h.layout.Markup(c, "admin:backToMenu"),
			)
		case !validator.ClubName(message.Text):
			_ = inputCollector.Send(c,
				h.layout.Text(c, "invalid_club_name"),
				h.layout.Markup(c, "admin:backToMenu"),
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

	club, err := h.clubService.Create(context.Background(), &entity.Club{
		Name: clubName,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return c.Send(
				h.layout.Text(c, "club_already_exists"),
				h.layout.Markup(c, "admin:backToMenu"),
			)
		}

		h.logger.Errorf("(user: %d) error while create new club: %v", c.Sender().ID, err)
		return c.Send(
			h.layout.Text(c, "technical_issues", err.Error()),
			h.layout.Markup(c, "admin:backToMenu"),
		)
	}

	h.logger.Infof("(user: %d) new club created: %s", c.Sender().ID, club.Name)
	return c.Send(
		h.layout.Text(c, "club_created", club),
		h.layout.Markup(c, "admin:backToMenu"),
	)
}

func (h Handler) clubsList(c tele.Context) error {
	const clubsOnPage = 10
	h.logger.Infof("(user: %d) edit clubs list", c.Sender().ID)

	var (
		p          int
		prevPage   int
		nextPage   int
		err        error
		clubsCount int64
		clubs      []entity.Club
		rows       []tele.Row
		menuRow    tele.Row
	)
	if c.Callback().Data != "admin_clubs" {
		p, err = strconv.Atoi(c.Callback().Data)
		if err != nil {
			return err
		}
	}

	clubsCount, err = h.clubService.Count(context.Background())
	if err != nil {
		return err
	}

	clubs, err = h.clubService.GetWithPagination(context.Background(), p*clubsOnPage, clubsOnPage, "created_at DESC")
	if err != nil {
		return err
	}

	markup := c.Bot().NewMarkup()
	for _, club := range clubs {
		rows = append(rows, markup.Row(*h.layout.Button(c, "admin:clubs:club", struct {
			ID   string
			Name string
			Page int
		}{
			ID:   club.ID,
			Name: club.Name,
			Page: p,
		})))
	}
	pagesCount := int(clubsCount) / (clubsOnPage + 1)
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
		*h.layout.Button(c, "admin:clubs:prev_page", struct {
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
		*h.layout.Button(c, "admin:clubs:next_page", struct {
			Page int
		}{
			Page: nextPage,
		}),
	)

	rows = append(
		rows,
		menuRow,
		markup.Row(*h.layout.Button(c, "admin:back_to_menu")),
	)

	markup.Inline(rows...)

	h.logger.Debugf("(user: %d) clubs list (pages_count=%d, page=%d, clubs_count=%d, next_page=%d, prev_page=%d)",
		c.Sender().ID,
		pagesCount,
		p,
		clubsCount,
		nextPage,
		prevPage,
	)

	_ = c.Edit(
		h.layout.Text(c, "clubs_list", clubsCount),
		markup,
	)
	return nil
}

func (h Handler) clubMenu(c tele.Context) error {
	callbackData := strings.Split(c.Callback().Data, " ")
	if len(callbackData) != 2 {
		return errorz.ErrInvalidCallbackData
	}
	clubID := callbackData[0]
	page := callbackData[1]

	h.logger.Infof("(user: %d) edit club menu (club_id=%s)", c.Sender().ID, clubID)

	club, err := h.clubService.Get(context.Background(), clubID)
	if err != nil {
		h.logger.Errorf("(user: %d) error while get club: %v", c.Sender().ID, err)
		return c.Send(
			h.layout.Text(c, "technical_issues", err.Error()),
			h.layout.Markup(c, "admin:clubs:back", struct {
				Page string
			}{
				Page: page,
			}),
		)
	}

	clubOwners, err := h.clubOwnerService.GetByClubID(context.Background(), clubID)
	if err != nil {
		h.logger.Errorf("(user: %d) error while get club owners: %v", c.Sender().ID, err)
		return c.Send(
			h.layout.Text(c, "technical_issues", err.Error()),
			h.layout.Markup(c, "admin:clubs:back", struct {
				Page string
			}{
				Page: page,
			}),
		)
	}

	return c.Edit(
		h.layout.Text(c, "admin_club_menu_text", struct {
			Club   entity.Club
			Owners []dto.ClubOwner
		}{
			Club:   *club,
			Owners: clubOwners,
		}),
		h.layout.Markup(c, "admin:club:menu", struct {
			ID   string
			Page string
		}{
			ID:   clubID,
			Page: page,
		}),
	)
}

func (h Handler) addClubOwner(c tele.Context) error {
	callbackData := strings.Split(c.Callback().Data, " ")
	if len(callbackData) != 2 {
		return errorz.ErrInvalidCallbackData
	}
	clubID := callbackData[0]
	page := callbackData[1]

	h.logger.Infof("(user: %d) add club owner (club_id=%s)", c.Sender().ID, clubID)
	inputCollector := collector.New()
	inputCollector.Collect(c.Message())
	_ = c.Edit(
		h.layout.Text(c, "input_user_id"),
		h.layout.Markup(c, "admin:club:back", struct {
			ID   string
			Page string
		}{
			ID:   clubID,
			Page: page,
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
				h.layout.Text(c, "input_error", h.layout.Text(c, "input_user_id")),
				h.layout.Markup(c, "admin:club:back", struct {
					ID   string
					Page string
				}{
					ID:   clubID,
					Page: page,
				}),
			)
		default:
			userID, err := strconv.ParseInt(message.Text, 10, 64)
			if err != nil {
				_ = inputCollector.Send(c,
					h.layout.Text(c, "input_user_id"),
					h.layout.Markup(c, "admin:club:back", struct {
						ID   string
						Page string
					}{
						ID:   clubID,
						Page: page,
					}),
				)
				break
			}

			user, err = h.adminUserService.Get(context.Background(), userID)
			if err != nil {
				_ = inputCollector.Send(c,
					h.layout.Text(c, "user_not_found", struct {
						ID   int64
						Text string
					}{
						ID:   userID,
						Text: h.layout.Text(c, "input_user_id"),
					}),
					h.layout.Markup(c, "admin:club:back", struct {
						ID   string
						Page string
					}{
						ID:   clubID,
						Page: page,
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

	club, err := h.clubService.Get(context.Background(), clubID)
	if err != nil {
		_ = inputCollector.Clear(c, collector.ClearOptions{IgnoreErrors: true})
		h.logger.Errorf("(user: %d) error while get club: %v", c.Sender().ID, err)
		return c.Send(
			h.layout.Text(c, "technical_issues", err.Error()),
			h.layout.Markup(c, "admin:club:back", struct {
				ID   string
				Page string
			}{
				ID:   clubID,
				Page: page,
			}),
		)
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
			h.layout.Text(c, "technical_issues", err.Error()),
			h.layout.Markup(c, "admin:club:back", struct {
				ID   string
				Page string
			}{
				ID:   clubID,
				Page: page,
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
		h.layout.Text(c, "club_owner_added", struct {
			Club entity.Club
			User entity.User
		}{
			Club: *club,
			User: *user,
		}),
		h.layout.Markup(c, "admin:club:back", struct {
			ID   string
			Page string
		}{
			ID:   clubID,
			Page: page,
		}),
	)
}

func (h Handler) removeClubOwner(c tele.Context) error {
	callbackData := strings.Split(c.Callback().Data, " ")
	if len(callbackData) != 2 {
		return errorz.ErrInvalidCallbackData
	}
	clubID := callbackData[0]
	page := callbackData[1]

	h.logger.Infof("(user: %d) remove club owner (club_id=%s)", c.Sender().ID, clubID)
	inputCollector := collector.New()
	inputCollector.Collect(c.Message())
	_ = c.Edit(
		h.layout.Text(c, "input_user_id"),
		h.layout.Markup(c, "admin:club:back", struct {
			ID   string
			Page string
		}{
			ID:   clubID,
			Page: page,
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
				h.layout.Text(c, "input_error", h.layout.Text(c, "input_user_id")),
				h.layout.Markup(c, "admin:club:back", struct {
					ID   string
					Page string
				}{
					ID:   clubID,
					Page: page,
				}),
			)
		default:
			userID, err := strconv.ParseInt(message.Text, 10, 64)
			if err != nil {
				_ = inputCollector.Send(c,
					h.layout.Text(c, "input_user_id"),
					h.layout.Markup(c, "admin:club:back", struct {
						ID   string
						Page string
					}{
						ID:   clubID,
						Page: page,
					}),
				)
				break
			}

			user, err = h.adminUserService.Get(context.Background(), userID)
			if err != nil {
				_ = inputCollector.Send(c,
					h.layout.Text(c, "user_not_found", struct {
						ID   int64
						Text string
					}{
						ID:   userID,
						Text: h.layout.Text(c, "input_user_id"),
					}),
					h.layout.Markup(c, "admin:club:back", struct {
						ID   string
						Page string
					}{
						ID:   clubID,
						Page: page,
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

	club, err := h.clubService.Get(context.Background(), clubID)
	if err != nil {
		_ = inputCollector.Clear(c, collector.ClearOptions{IgnoreErrors: true})
		h.logger.Errorf("(user: %d) error while get club: %v", c.Sender().ID, err)
		return c.Send(
			h.layout.Text(c, "technical_issues", err.Error()),
			h.layout.Markup(c, "admin:club:back", struct {
				ID   string
				Page string
			}{
				ID:   clubID,
				Page: page,
			}),
		)
	}

	err = h.clubOwnerService.Remove(context.Background(), user.ID, club.ID)
	if err != nil {
		_ = inputCollector.Clear(c, collector.ClearOptions{IgnoreErrors: true})
		h.logger.Errorf(
			"(user: %d) error while remove club owner (club_id=%s, user_id=%d): %v",
			c.Sender().ID,
			clubID,
			user.ID,
			err,
		)
		return c.Send(
			h.layout.Text(c, "technical_issues", err.Error()),
			h.layout.Markup(c, "admin:club:back", struct {
				ID   string
				Page string
			}{
				ID:   clubID,
				Page: page,
			}),
		)
	}

	h.logger.Infof(
		"(user: %d) club owner removed (club_id=%s, user_id=%d)",
		c.Sender().ID,
		clubID,
		user.ID,
	)

	_ = inputCollector.Clear(c, collector.ClearOptions{IgnoreErrors: true})
	return c.Send(
		h.layout.Text(c, "club_owner_removed", struct {
			Club entity.Club
			User entity.User
		}{
			Club: *club,
			User: *user,
		}),
		h.layout.Markup(c, "admin:club:back", struct {
			ID   string
			Page string
		}{
			ID:   clubID,
			Page: page,
		}),
	)
}

func (h Handler) deleteClub(c tele.Context) error {
	callbackData := strings.Split(c.Callback().Data, " ")
	if len(callbackData) != 2 {
		return errorz.ErrInvalidCallbackData
	}
	clubID := callbackData[0]
	page := callbackData[1]

	h.logger.Infof("(user: %d) delete club (club_id=%s)", c.Sender().ID, clubID)

	club, err := h.clubService.Get(context.Background(), clubID)
	if err != nil {
		h.logger.Errorf("(user: %d) error while get club: %v", c.Sender().ID, err)
		return c.Edit(
			h.layout.Text(c, "technical_issues", err.Error()),
			h.layout.Markup(c, "admin:clubs:back", struct {
				Page string
			}{
				Page: page,
			}),
		)
	}
	err = h.clubService.Delete(context.Background(), club.ID)
	if err != nil {
		h.logger.Errorf("(user: %d) error while delete club: %v", c.Sender().ID, err)
		return c.Edit(
			h.layout.Text(c, "technical_issues", err.Error()),
			h.layout.Markup(c, "admin:clubs:back", struct {
				Page string
			}{
				Page: page,
			}),
		)
	}

	h.logger.Infof("(user: %d) club deleted (club_id=%s)", c.Sender().ID, clubID)
	return c.Edit(
		h.layout.Text(c, "club_deleted", club),
		h.layout.Markup(c, "admin:clubs:back", struct {
			Page string
		}{
			Page: page,
		}),
	)
}

func (h Handler) banUser(c tele.Context) error {
	_ = c.Delete()
	if c.Message() == nil || c.Message().Payload == "" {
		return c.Send(
			h.layout.Text(c, "invalid_ban_data"),
			h.layout.Markup(c, "core:hide"),
		)
	}
	userID, err := strconv.ParseInt(c.Message().Payload, 10, 64)
	if err != nil {
		return c.Send(
			h.layout.Text(c, "invalid_ban_data"),
			h.layout.Markup(c, "core:hide"),
		)
	}

	h.logger.Infof("(user: %d) attempt ban user: %d", c.Sender().ID, userID)
	if userID == c.Sender().ID {
		return c.Send(
			h.layout.Text(c, "attempt_to_ban_self"),
			h.layout.Markup(c, "core:hide"),
		)
	}
	user, err := h.adminUserService.Ban(context.Background(), userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Send(
				h.layout.Text(c, "user_not_found", struct {
					ID int64
				}{
					ID: userID,
				}),
				h.layout.Markup(c, "core:hide"),
			)
		}
		h.logger.Errorf("(user: %d) error while ban user: %v", c.Sender().ID, err)
		return c.Send(
			h.layout.Text(c, "technical_issues", err.Error()),
			h.layout.Markup(c, "core:hide"),
		)
	}

	if user.IsBanned {
		h.logger.Infof("(user: %d) user banned: %d", c.Sender().ID, userID)
		return c.Send(
			h.layout.Text(c, "user_banned", struct {
				FIO string
				ID  int64
			}{
				FIO: user.FIO,
				ID:  user.ID,
			}),
			h.layout.Markup(c, "core:hide"),
		)
	}

	h.logger.Infof("(user: %d) user unbanned: %d", c.Sender().ID, userID)
	return c.Send(
		h.layout.Text(c, "user_unbanned", struct {
			FIO string
			ID  int64
		}{
			FIO: user.FIO,
			ID:  user.ID,
		}),
		h.layout.Markup(c, "core:hide"),
	)
}

func (h Handler) AdminSetup(group *tele.Group) {
	group.Handle(h.layout.Callback("mainMenu:admin_menu"), h.adminMenu)
	group.Handle(h.layout.Callback("admin:back_to_menu"), h.adminMenu)
	group.Handle(h.layout.Callback("admin:create_club"), h.createClub)
	group.Handle(h.layout.Callback("admin:clubs"), h.clubsList)
	group.Handle(h.layout.Callback("admin:clubs:prev_page"), h.clubsList)
	group.Handle(h.layout.Callback("admin:clubs:next_page"), h.clubsList)
	group.Handle(h.layout.Callback("admin:clubs:back"), h.clubsList)
	group.Handle(h.layout.Callback("admin:clubs:club"), h.clubMenu)
	group.Handle(h.layout.Callback("admin:club:back"), h.clubMenu)
	group.Handle(h.layout.Callback("admin:club:add_owner"), h.addClubOwner)
	group.Handle(h.layout.Callback("admin:club:del_owner"), h.removeClubOwner)
	group.Handle(h.layout.Callback("admin:club:delete"), h.deleteClub)
	group.Handle("/ban", h.banUser)
}
