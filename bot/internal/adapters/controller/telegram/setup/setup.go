package setup

import (
	"github.com/Badsnus/cu-clubs-bot/bot/cmd/bot"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/controller/telegram/handlers/admin"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/controller/telegram/handlers/menu"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/controller/telegram/handlers/middlewares"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/controller/telegram/handlers/user"
	"github.com/spf13/viper"
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/middleware"
)

func Setup(b *bot.Bot) {
	// Pre-setup and global middlewares
	middle := middlewares.New(b)
	userHandler := user.New(b)
	menuHandler := menu.New(b)
	adminHandler := admin.New(b)

	//onEventHandler := handlers.NewOnEventHandler(b)
	if viper.GetBool("settings.debug") {
		b.Use(middleware.Logger())
	}
	b.Use(b.Layout.Middleware("ru"))
	b.Use(middleware.AutoRespond())
	b.Handle(tele.OnText, b.Input.Handler())
	b.Handle(tele.OnMedia, b.Input.Handler())
	b.Use(middle.ResetInputOnBack)

	// Setup handlers
	//User:
	userHandler.AuthSetup(b.Group())
	b.Use(middle.Authorized)
	b.Handle(b.Layout.TextLocale("ru", "open_main_menu"), menuHandler.SendMenu)
	b.Handle(b.Layout.Callback("mainMenu:back"), menuHandler.EditMenu)

	//Admin:
	admins := viper.GetIntSlice("bot.admin-ids")
	adminsInt64 := make([]int64, len(admins))
	for i, v := range admins {
		adminsInt64[i] = int64(v)
	}
	b.Use(middleware.Whitelist(adminsInt64...))
	adminHandler.AdminSetup(b.Group())

	//adminHandler := handlers.NewAdminHandler(b)
	//b.Handle("/admin", adminHandler.AdminMenu)
	//b.Handle(b.Layout.Callback("backToAdminMenu"), adminHandler.BackToAdminMenu)
	//b.Handle(b.Layout.Callback("manageUsers"), adminHandler.UsersList)
	//b.Handle(b.Layout.Callback("backToUsersList"), adminHandler.UsersList)
	//b.Handle(b.Layout.Callback("user"), adminHandler.ManageUser)
	//b.Handle(b.Layout.Callback("banUser"), adminHandler.BanUser)
	//b.Handle(b.Layout.Callback("mailing"), adminHandler.Mailing)
	//b.Handle(b.Layout.Callback("confirmMailing"), adminHandler.SendMailing)
	//b.Handle(b.Layout.Callback("cancelMailing"), adminHandler.BackToAdminMenu)
}
