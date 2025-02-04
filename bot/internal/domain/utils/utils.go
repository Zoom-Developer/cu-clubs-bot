package utils

import (
	"github.com/spf13/viper"
	tele "gopkg.in/telebot.v3"
	"slices"
)

func IsAdmin(userID int64) bool {
	return slices.Contains(viper.GetIntSlice("bot.admin-ids"), int(userID))
}

func ChangeMessageText(msg *tele.Message, text string) interface{} {
	if msg.Photo != nil {
		msg.Photo.Caption = text
		return msg.Photo
	}
	if msg.Video != nil {
		msg.Video.Caption = text
		return msg.Video
	}
	if msg.Audio != nil {
		msg.Audio.Caption = text
		return msg.Audio
	}
	if msg.Document != nil {
		msg.Document.Caption = text
		return msg.Document
	}
	return text
}

func GetMessageText(msg *tele.Message) string {
	switch {
	case msg.Text != "":
		return msg.Text
	case msg.Caption != "":
		return msg.Caption
	default:
		return ""
	}
}
