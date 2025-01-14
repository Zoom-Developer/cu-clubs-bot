package scheduler

//type subscriptionService interface {
//	GetExpiring(ctx context.Context) ([]entity.Subscription, error)
//}
//
//type userService interface {
//	Get(ctx context.Context, id int64) (*entity.User, error)
//}
//
//type SubscriptionScheduler struct {
//	subscriptionService subscriptionService
//	userService         userService
//
//	layout *layout.Layout
//	bot    *tele.Bot
//}
//
//func NewSubscriptionScheduler(bot *bot.Bot) *SubscriptionScheduler {
//	subscriptionStorage := postgres.NewSubscriptionStorage(bot.DB)
//	userStorage := postgres.NewUserStorage(bot.DB)
//
//	return &SubscriptionScheduler{
//		subscriptionService: service.NewSubscriptionService(subscriptionStorage),
//		userService:         service.NewUserService(userStorage),
//		layout:              bot.Layout,
//		bot:                 bot.Bot,
//	}
//}
//
//func (s *SubscriptionScheduler) periodicallySendExpiringNotifications(ctx context.Context) {
//	ticker := time.NewTicker(github.com/Badsnus/cu-clubs-bot0 * time.Minute)
//	for {
//		select {
//		case <-ticker.C:
//			log.Println("Sending expiring subscription notifications...")
//			s.sendExpiringNotifications(ctx)
//		case <-ctx.Done():
//			return
//		}
//	}
//}
//
//func (s *SubscriptionScheduler) sendExpiringNotifications(ctx context.Context) {
//	subs, err := s.subscriptionService.GetExpiring(ctx)
//	if err != nil {
//		log.Printf("Error getting expiring subscriptions: %v", err)
//		return
//	}
//
//	for _, sub := range subs {
//		user, errGetUser := s.userService.Get(ctx, int64(sub.UserID))
//		if errGetUser != nil {
//			log.Printf("Error getting expiring user (id: %d): %v", sub.UserID, err)
//			continue
//		}
//
//		chat, errGetChat := s.bot.ChatByID(int64(user.ID))
//		if errGetChat != nil {
//			continue
//		}
//		_, err = s.bot.Send(chat,
//			s.layout.TextLocale(user.Localisation, "subscription_expiring", sub),
//			s.layout.MarkupLocale(user.Localisation, "hide"),
//		)
//
//		if err != nil {
//			log.Printf("Error sending expiring subscription notify: %v", err)
//			continue
//		}
//	}
//}
//
//func (s *SubscriptionScheduler) Start() {
//	log.Println("Starting subscription notify scheduler...")
//	go s.periodicallySendExpiringNotifications(context.Background())
//}
