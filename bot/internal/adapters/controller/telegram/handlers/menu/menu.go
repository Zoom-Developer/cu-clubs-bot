package menu

import (
	"github.com/Badsnus/cu-clubs-bot/bot/cmd/bot"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/utils"
	"github.com/Badsnus/cu-clubs-bot/bot/pkg/logger/types"
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/layout"
)

type Handler struct {
	layout *layout.Layout
	logger *types.Logger
}

func New(b *bot.Bot) *Handler {
	return &Handler{
		logger: b.Logger,
		layout: b.Layout,
	}
}

func (h Handler) SendMenu(c tele.Context) error {
	isAdmin := utils.IsAdmin(c.Sender().ID)

	menuMarkup := h.layout.Markup(c, "mainMenu:menu")
	if isAdmin {
		menuMarkup.InlineKeyboard = append(menuMarkup.InlineKeyboard, []tele.InlineButton{*h.layout.Button(c, "mainMenu:admin_menu").Inline()})
	}

	h.logger.Infof("(user: %d) send main menu (isAdmin=%t)", c.Sender().ID, isAdmin)
	return c.Send(
		h.layout.Text(c, "main_menu_text", c.Sender().Username),
		menuMarkup,
	)
}

func (h Handler) EditMenu(c tele.Context) error {
	isAdmin := utils.IsAdmin(c.Sender().ID)

	menuMarkup := h.layout.Markup(c, "mainMenu:menu")
	if isAdmin {
		menuMarkup.InlineKeyboard = append(menuMarkup.InlineKeyboard, []tele.InlineButton{*h.layout.Button(c, "mainMenu:admin_menu").Inline()})
	}

	h.logger.Infof("(user: %d) edit main menu (isAdmin=%t)", c.Sender().ID, isAdmin)
	return c.Edit(
		h.layout.Text(c, "main_menu_text", c.Sender().Username),
		menuMarkup,
	)
}
