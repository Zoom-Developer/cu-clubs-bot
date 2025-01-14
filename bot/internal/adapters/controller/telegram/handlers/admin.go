package handlers

import (
	"context"
	"github.com/Badsnus/cu-clubs-bot/cmd/bot"
	"github.com/Badsnus/cu-clubs-bot/internal/adapters/database/postgres"
	"github.com/Badsnus/cu-clubs-bot/internal/adapters/database/redis"
	"github.com/Badsnus/cu-clubs-bot/internal/adapters/database/redis/state"
	"github.com/Badsnus/cu-clubs-bot/internal/domain/common/errorz"
	"github.com/Badsnus/cu-clubs-bot/internal/domain/entity"
	"github.com/Badsnus/cu-clubs-bot/internal/domain/service"
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/layout"
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

type AdminHandler struct {
	layout           *layout.Layout
	adminUserService adminUserService
	statesStorage    *redis.StatesStorage
	bot              *bot.Bot
}

func NewAdminHandler(b *bot.Bot) *AdminHandler {
	userStorage := postgres.NewUserStorage(b.DB)

	return &AdminHandler{
		layout:           b.Layout,
		adminUserService: service.NewUserService(userStorage),
		statesStorage:    redis.NewStatesStorage(b),
		bot:              b,
	}
}

func (h AdminHandler) AdminMenu(c tele.Context) error {
	return c.Send(
		h.layout.Text(c, "admin_text", c.Sender().Username),
		h.layout.Markup(c, "adminMenu"),
	)
}

func (h AdminHandler) BackToAdminMenu(c tele.Context) error {
	return c.Edit(
		h.layout.Text(c, "admin_text", c.Sender().Username),
		h.layout.Markup(c, "adminMenu"),
	)
}

func (h AdminHandler) UsersList(c tele.Context) error {
	var (
		p        int
		prevPage int
		nextPage int
		err      error
		rows     []tele.Row
		menuRow  tele.Row
	)
	p = 0
	if c.Callback().Data != "manage_users" {
		p, err = strconv.Atoi(c.Callback().Data)
		if err != nil {
			return err
		}
	}

	usersCount, err := h.adminUserService.Count(context.Background())
	if err != nil {
		return err
	}

	users, err := h.adminUserService.GetWithPagination(context.Background(), p*5, 5, "created_at DESC")
	if err != nil {
		return err
	}

	markup := h.bot.NewMarkup()
	for _, user := range users {
		rows = append(rows, markup.Row(*h.bot.Layout.Button(c, "user", struct {
			ID       uint
			Banned   bool
			Username string
			Page     int
		}{
			ID:       user.ID,
			Banned:   user.Banned,
			Username: user.Username,
			Page:     p,
		})))
	}
	pagesCount := int(usersCount) / 6
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
		*h.bot.Layout.Button(c, "usersPrev", struct {
			Page int
		}{
			Page: prevPage,
		}),
		*h.bot.Layout.Button(c, "pageCounter", struct {
			Page       int
			PagesCount int
		}{
			Page:       p + 1,
			PagesCount: pagesCount + 1,
		}),
		*h.bot.Layout.Button(c, "usersNext", struct {
			Page int
		}{
			Page: nextPage,
		}),
	)

	rows = append(
		rows,
		menuRow,
		markup.Row(*h.bot.Layout.Button(c, "backToAdminMenu", h.layout.Text(c, "back"))))

	markup.Inline(rows...)

	return c.Edit(
		h.bot.Layout.Text(c, "manage_users_text"),
		markup,
	)
}

func (h AdminHandler) ManageUser(c tele.Context) error {
	callbackData := strings.Split(c.Callback().Data, " ")
	if len(callbackData) != 2 {
		return errorz.InvalidCallbackData
	}
	intUserId, err := strconv.Atoi(callbackData[0])
	if err != nil {
		return err
	}
	intPage, err := strconv.Atoi(callbackData[1])
	if err != nil {
		return err
	}

	user, err := h.adminUserService.Get(context.Background(), int64(intUserId))
	if err != nil {
		return err
	}

	return c.Edit(
		h.layout.Text(c, "user_information", struct {
			User entity.User
		}{
			User: *user,
		}),
		h.layout.Markup(c, "manageUser", struct {
			User entity.User
			Page int
		}{
			User: *user,
			Page: intPage,
		}),
	)
}

func (h AdminHandler) BanUser(c tele.Context) error {
	callbackData := strings.Split(c.Callback().Data, " ")
	if len(callbackData) != 2 {
		return errorz.InvalidCallbackData
	}
	intUserId, err := strconv.Atoi(callbackData[0])
	if err != nil {
		return err
	}

	_, err = h.adminUserService.Ban(context.Background(), int64(intUserId))
	if err != nil {
		return err
	}
	return h.ManageUser(c)
}

func (h AdminHandler) Mailing(c tele.Context) error {
	h.statesStorage.Set(c.Sender().ID, state.WaitingForMailingContent, "")
	return c.Edit(
		h.layout.Text(c, "mailing_text"),
		h.layout.Markup(c, "backToAdminMenu"),
	)
}

func (h AdminHandler) SendMailing(c tele.Context) error {
	users, err := h.adminUserService.GetAll(context.Background())
	if err != nil {
		return err
	}

	loading, err := h.bot.Send(c.Chat(), h.layout.Text(c, "loading"))
	if err != nil {
		return err
	}

	for _, user := range users {
		chat, errGetChat := h.bot.ChatByID(int64(user.ID))
		if errGetChat == nil {
			_, _ = h.bot.Copy(
				chat,
				c.Message(),
			)
		}
	}

	_ = c.Delete()
	_ = h.bot.Delete(loading)

	err = c.Send(h.layout.Text(c, "mailing_successful"))
	if err != nil {
		return err
	}
	return h.AdminMenu(c)
}
