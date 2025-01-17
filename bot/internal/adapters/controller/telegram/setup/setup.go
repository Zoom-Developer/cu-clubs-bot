package setup

import (
	"github.com/Badsnus/cu-clubs-bot/cmd/bot"
	"github.com/Badsnus/cu-clubs-bot/internal/adapters/controller/telegram/handlers"
	"github.com/Badsnus/cu-clubs-bot/internal/adapters/controller/telegram/handlers/middlewares"
	"github.com/spf13/viper"
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/middleware"
)

func Setup(b *bot.Bot) {
	//subscriptionScheduler := scheduler.NewSubscriptionScheduler(b)
	//subscriptionScheduler.Start()

	middle := middlewares.New(b)
	userHandler := handlers.NewUserHandler(b)
	onEventHandler := handlers.NewOnEventHandler(b)

	if viper.GetBool("settings.debug") {
		b.Use(middleware.Logger())
	}

	b.Use(b.Layout.Middleware("ru"))
	b.Use(middleware.AutoRespond())
	b.Use(middle.Authorized)
	//b.Handle(b.Layout.Callback("english"), userHandler.OnLocalisation)
	//b.Handle(b.Layout.Callback("russian"), userHandler.OnLocalisation)
	//b.Use(middle.Localisation)
	//b.Use(b.Layout.Middleware("en", middle.SetupLocalisation))

	b.Handle(tele.OnText, onEventHandler.OnText)
	b.Handle(tele.OnMedia, onEventHandler.OnMedia)
	b.Use(middle.ResetStateOnBack)

	b.Handle("/start", userHandler.OnStart)
	b.Use(middle.Subscribed)
	//b.Handle(b.Layout.Callback("hide"), userHandler.Hide)

	//b.Handle(b.Layout.Callback("backToMainMenu"), userHandler.EditMainMenu)
	//b.Handle(b.Layout.Callback("information"), userHandler.Information)
	//
	//b.Handle(b.Layout.Callback("profile"), userHandler.Profile)

	// Admin:
	//admins := b.Viper.GetIntSlice("bot.admin-ids")
	//adminsInt64 := make([]int64, len(admins))
	//for i, v := range admins {
	//	adminsInt64[i] = int64(v)
	//}

	//b.Use(middleware.Whitelist(adminsInt64...))
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
