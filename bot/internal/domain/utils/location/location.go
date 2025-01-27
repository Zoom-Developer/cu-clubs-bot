package location

import (
	"github.com/spf13/viper"
	"log"
	"time"
)

var Location *time.Location

func init() {
	var err error
	Location, err = time.LoadLocation(viper.GetString("settings.timezone"))
	if err != nil {
		log.Fatalf("error while load time location: %v", err)
	}
}
