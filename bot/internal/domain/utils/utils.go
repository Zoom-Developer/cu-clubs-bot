package utils

import (
	"github.com/spf13/viper"
	"slices"
)

func IsAdmin(userID int64) bool {
	return slices.Contains(viper.GetIntSlice("bot.admin-ids"), int(userID))
}
