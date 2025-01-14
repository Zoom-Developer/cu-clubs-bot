package postgres

import "github.com/Badsnus/cu-clubs-bot/internal/domain/entity"

// Migrations is a list of all gorm migrations for the database.
var Migrations = []interface{}{
	&entity.User{},
}
