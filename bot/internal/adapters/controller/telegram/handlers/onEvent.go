package handlers

//import (
//	"context"
//	"slices"
//
//	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/database/redis/codes"
//	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/database/redis/emails"
//	"github.com/Badsnus/cu-clubs-bot/bot/pkg/smtp"
//
//	"github.com/Badsnus/cu-clubs-bot/bot/cmd/bot"
//	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/database/postgres"
//	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/database/redis/states"
//	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"
//	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/service"
//	"github.com/spf13/viper"
//	tele "gopkg.in/telebot.v3"
//	"gopkg.in/telebot.v3/layout"
//)
//
//type onEventUserService interface {
//	Get(ctx context.Context, userID int64) (*entity.User, error)
//	Create(ctx context.Context, user entity.User) (*entity.User, error)
//	SendAuthCode(ctx context.Context, email string) (string, string, error)
//}
//
//type OnEventHandler struct {
//	layout        *layout.Layout
//	bot           *tele.Bot
//	userService   onEventUserService
//	statesStorage *states.Storage
//	codesStorage  *codes.Storage
//	emailsStorage *emails.Storage
//}
//
//func NewOnEventHandler(b *bot.Bot) *OnEventHandler {
//	userStorage := postgres.NewUserStorage(b.DB)
//	studentDataStorage := postgres.NewStudentDataStorage(b.DB)
//	smtpClient := smtp.NewClient(b.SMTPDialer)
//
//	return &OnEventHandler{
//		layout:      b.Layout,
//		bot:         b.Bot,
//		userService: service.NewUserService(userStorage, studentDataStorage, smtpClient),
//		//statesStorage: states.NewStorage(b),
//		//codesStorage:  codes.NewStorage(b),
//		//emailsStorage: emails.NewStorage(b),
//	}
//}
//
////func (h *OnEventHandler) OnMedia(c tele.Context) error {
////	stateData, errGetState := h.statesStorage.Get(c.Sender().ID)
////	if errGetState != nil {
////		return errGetState
////	}
////
////	switch stateData.State {
////	case state.WaitingForMailingContent:
////		return h.onMailing(c)
////	default:
////		return c.Send(h.layout.Text(c, "unknown_command"))
////	}
////}
//
//func (h *OnEventHandler) checkForAdmin(userID int64) bool {
//	return slices.Contains(viper.GetIntSlice("bot.admin-ids"), int(userID))
//}
