package service

import (
	"bytes"
	"context"
	"fmt"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"
	qr "github.com/Badsnus/cu-clubs-bot/bot/pkg/qrcode"
	"github.com/google/uuid"
	tele "gopkg.in/telebot.v3"
)

type qrUserService interface {
	Get(ctx context.Context, userID int64) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) (*entity.User, error)
}

type QrService struct {
	userService qrUserService
	bot         *tele.Bot
	qrChat      *tele.Chat
	qrCFG       qr.Config
	logoPath    string
	botName     string
}

func NewQrService(bot *tele.Bot, qrCFG qr.Config, userService qrUserService, qrChatID int64, logoPath string) (*QrService, error) {
	chat, err := bot.ChatByID(qrChatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get qr chat: %v", err)
	}
	qrCFG.LogoPath = logoPath
	return &QrService{
		userService: userService,
		bot:         bot,
		qrChat:      chat,
		qrCFG:       qrCFG,
		botName:     bot.Me.Username,
	}, nil
}

func (s *QrService) GetUserQR(ctx context.Context, userID int64) (qr tele.File, err error) {
	user, err := s.userService.Get(ctx, userID)
	if err != nil {
		return qr, err
	}
	if user.QRFileID != "" {
		qr, err = s.bot.FileByID(user.QRFileID)
		if err != nil {
			return qr, err
		}
		return qr, nil
	}

	qrCodeID := uuid.New().String()
	link := fmt.Sprintf("https://t.me/%s?start=userQR_%s", s.botName, qrCodeID)
	cfg := s.qrCFG
	cfg.Content = link
	qrData, err := cfg.Generate()
	if err != nil {
		return qr, err
	}

	tempQR := tele.FromReader(bytes.NewReader(qrData))
	qrMsg, err := s.bot.Send(s.qrChat, &tele.Photo{
		File: tempQR,
	})
	if err != nil {
		return qr, err
	}
	user.QRFileID = qrMsg.Photo.FileID
	user.QRCodeID = qrCodeID
	_, err = s.userService.Update(ctx, user)
	if err != nil {
		return qr, err
	}

	qr, err = s.bot.FileByID(qrMsg.Photo.FileID)
	if err != nil {
		return qr, err
	}

	return qr, err
}

func (s *QrService) RevokeUserQR(ctx context.Context, userID int64) error {
	user, err := s.userService.Get(ctx, userID)
	if err != nil {
		return err
	}
	if user.QRFileID != "" {
		user.QRFileID = ""
		user.QRCodeID = ""
		_, err = s.userService.Update(ctx, user)
		if err != nil {
			return err
		}
	}
	return nil
}
