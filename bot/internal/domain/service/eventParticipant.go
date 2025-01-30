package service

import (
	"bytes"
	"context"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/dto"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/utils/location"
	"github.com/Badsnus/cu-clubs-bot/bot/pkg/logger/types"
	"github.com/xuri/excelize/v2"
	"strconv"
	"strings"
	"time"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"
)

type EventParticipantStorage interface {
	Create(ctx context.Context, eventParticipant *entity.EventParticipant) (*entity.EventParticipant, error)
	Get(ctx context.Context, eventID string, userID int64) (*entity.EventParticipant, error)
	Update(ctx context.Context, eventParticipant *entity.EventParticipant) (*entity.EventParticipant, error)
	Delete(ctx context.Context, eventID string, userID int64) error
	GetByEventID(ctx context.Context, eventID string) ([]entity.EventParticipant, error)
	CountByEventID(ctx context.Context, eventID string) (int64, error)
	CountVisitedByEventID(ctx context.Context, eventID string) (int64, error)
	GetUserEvents(ctx context.Context, userID int64, limit, offset int) ([]dto.UserEvent, error)
	CountUserEvents(ctx context.Context, userID int64) (int64, error)
}

type eventParticipantEventStorage interface {
	GetUpcomingEvents(ctx context.Context, before time.Time) ([]entity.Event, error)
}

type userStorage interface {
	GetUsersByEventID(ctx context.Context, eventID string) ([]entity.User, error)
}

type eventParticipantSMTPClient interface {
	Send(to string, body, message string, subject string, file *bytes.Buffer)
}

type EventParticipantService struct {
	logger *types.Logger

	storage                    EventParticipantStorage
	eventStorage               eventParticipantEventStorage
	userStorage                userStorage
	eventParticipantSMTPClient eventParticipantSMTPClient

	passEmail string
}

func NewEventParticipantService(
	logger *types.Logger,
	storage EventParticipantStorage,
	eventStorage eventParticipantEventStorage,
	userStorage userStorage,
	eventParticipantSMTPClient eventParticipantSMTPClient,
	passEmail string,
) *EventParticipantService {
	return &EventParticipantService{
		logger: logger,

		storage:                    storage,
		eventStorage:               eventStorage,
		userStorage:                userStorage,
		eventParticipantSMTPClient: eventParticipantSMTPClient,

		passEmail: passEmail,
	}
}

func (s *EventParticipantService) Register(ctx context.Context, eventID string, userID int64) (*entity.EventParticipant, error) {
	return s.storage.Create(ctx, &entity.EventParticipant{
		UserID:  userID,
		EventID: eventID,
	})
}

func (s *EventParticipantService) Get(ctx context.Context, eventID string, userID int64) (*entity.EventParticipant, error) {
	return s.storage.Get(ctx, eventID, userID)
}

func (s *EventParticipantService) Update(ctx context.Context, eventParticipant *entity.EventParticipant) (*entity.EventParticipant, error) {
	return s.storage.Update(ctx, eventParticipant)
}

func (s *EventParticipantService) Delete(ctx context.Context, eventID string, userID int64) error {
	return s.storage.Delete(ctx, eventID, userID)
}

func (s *EventParticipantService) GetByEventID(ctx context.Context, eventID string) ([]entity.EventParticipant, error) {
	return s.storage.GetByEventID(ctx, eventID)
}

func (s *EventParticipantService) CountByEventID(ctx context.Context, eventID string) (int, error) {
	count, err := s.storage.CountByEventID(ctx, eventID)
	return int(count), err
}

func (s *EventParticipantService) CountVisitedByEventID(ctx context.Context, eventID string) (int, error) {
	count, err := s.storage.CountVisitedByEventID(ctx, eventID)
	return int(count), err
}

func (s *EventParticipantService) GetUserEvents(ctx context.Context, userID int64, limit, offset int) ([]dto.UserEvent, error) {
	return s.storage.GetUserEvents(ctx, userID, limit, offset)
}

func (s *EventParticipantService) CountUserEvents(ctx context.Context, userID int64) (int64, error) {
	return s.storage.CountUserEvents(ctx, userID)
}

func (s *EventParticipantService) StartPassScheduler() {
	s.logger.Info("Starting pass scheduler")
	go func() {
		ticker := time.NewTicker(45 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			ctx := context.Background()
			s.checkAndSend(ctx)
		}
	}()
}

func (s *EventParticipantService) checkAndSend(ctx context.Context) {
	s.logger.Debugf("Checking for events starting in the next 25 hours")
	now := time.Now().In(location.Location())

	events, err := s.eventStorage.GetUpcomingEvents(ctx, now.Add(72*time.Hour))
	if err != nil {
		s.logger.Errorf("failed to get upcoming events: %v", err)
		return
	}

	for _, event := range events {
		eventStartTime := event.StartTime.In(location.Location())
		weekday := eventStartTime.Weekday()

		// Determine notification time
		var notificationTime time.Time
		if weekday == time.Sunday || weekday == time.Monday {
			notificationTime = time.Date(eventStartTime.Year(), eventStartTime.Month(), eventStartTime.Day(), 12, 0, 0, 0, location.Location())
		} else {
			notificationTime = time.Date(eventStartTime.Year(), eventStartTime.Month(), eventStartTime.Day(), 0, 0, 0, 0, eventStartTime.Location()).Add(-24 * time.Hour).Add(16 * time.Hour)
		}

		// Check if it's time to send the notification
		if now.Before(notificationTime) || now.After(notificationTime.Add(1*time.Hour)) {
			continue
		}

		var participants []entity.User
		participants, err = s.userStorage.GetUsersByEventID(ctx, event.ID)
		if err != nil {
			s.logger.Errorf("failed to get participants for event %s: %v", event.ID, err)
			continue
		}

		var buf *bytes.Buffer
		buf, err = participantsToXLSX(participants)

		s.eventParticipantSMTPClient.Send(s.passEmail, "Event passes", "Event passes", "Event passes", buf)
	}
}

func participantsToXLSX(users []entity.User) (*bytes.Buffer, error) {
	f := excelize.NewFile()

	sheet := "Sheet1"
	_ = f.SetCellValue(sheet, "A1", "Фамилия")
	_ = f.SetCellValue(sheet, "B1", "Имя")
	_ = f.SetCellValue(sheet, "C1", "Отчество")
	_ = f.SetCellValue(sheet, "D1", "Username")
	for i, user := range users {
		if user.Role == entity.Student {
			continue
		}
		fio := strings.Split(user.FIO, " ")

		row := i + 2
		_ = f.SetCellValue(sheet, "A"+strconv.Itoa(row), fio[0])
		_ = f.SetCellValue(sheet, "B"+strconv.Itoa(row), fio[1])
		_ = f.SetCellValue(sheet, "C"+strconv.Itoa(row), fio[2])
		_ = f.SetCellValue(sheet, "D"+strconv.Itoa(row), user.Username)
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}

	return &buf, nil
}
