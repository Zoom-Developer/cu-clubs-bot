package admin

import (
	"context"
	"errors"
	"github.com/Badsnus/cu-clubs-bot/bot/cmd/bot"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/database/postgres"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/service"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/utils/validator"
	"github.com/Badsnus/cu-clubs-bot/bot/pkg/intele"
	"github.com/Badsnus/cu-clubs-bot/bot/pkg/intele/collector"
	"github.com/Badsnus/cu-clubs-bot/bot/pkg/logger/types"
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/layout"
	"gorm.io/gorm"
	"strconv"
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

type Handler struct {
	layout *layout.Layout
	logger *types.Logger
	bot    *bot.Bot
	input  *intele.InputManager

	adminUserService adminUserService
	clubService      clubService
}

func New(b *bot.Bot) *Handler {
	userStorage := postgres.NewUserStorage(b.DB)
	clubStorage := postgres.NewClubStorage(b.DB)

	return &Handler{
		layout:           b.Layout,
		logger:           b.Logger,
		bot:              b,
		input:            b.Input,
		adminUserService: service.NewUserService(userStorage, nil, nil),
		clubService:      service.NewClubService(clubStorage),
	}
}

func (h Handler) adminMenu(c tele.Context) error {
	h.logger.Infof("(user: %d) edit admin menu", c.Sender().ID)
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

func (h Handler) AdminSetup(group *tele.Group) {
	group.Handle(h.layout.Callback("mainMenu:admin_menu"), h.adminMenu)
	group.Handle(h.layout.Callback("admin:back_to_menu"), h.adminMenu)
	group.Handle(h.layout.Callback("admin:create_club"), h.createClub)
	group.Handle(h.layout.Callback("admin:clubs"), h.clubsList)
	group.Handle(h.layout.Callback("admin:clubs:prev_page"), h.clubsList)
	group.Handle(h.layout.Callback("admin:clubs:next_page"), h.clubsList)
}
