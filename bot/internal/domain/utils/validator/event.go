package validator

import (
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/Badsnus/cu-clubs-bot/bot/pkg/logger"
	"github.com/spf13/viper"
)

func EventName(name string, _ map[string]interface{}) bool {
	return utf8.RuneCountInString(name) >= 5 && utf8.RuneCountInString(name) <= 30
}

func EventDescription(description string, _ map[string]interface{}) bool {
	return utf8.RuneCountInString(description) <= 150
}

func EventLocation(location string, _ map[string]interface{}) bool {
	return utf8.RuneCountInString(location) >= 5 && utf8.RuneCountInString(location) <= 150
}

func EventStartTime(start string, _ map[string]interface{}) bool {
	const layout = "02.01.2006 15:04"

	startTime, err := time.Parse(layout, start)
	if err != nil {
		return false
	}

	currentTime := time.Now()

	moscowLocation, err := time.LoadLocation(viper.GetString("settings.timezone"))
	if err != nil {
		logger.Log.Errorf("error while load time location: %v", err)
		return false
	}
	moscowTime := currentTime.In(moscowLocation)

	tomorrow := moscowTime.Add(time.Hour * time.Duration(24))

	return !startTime.Before(tomorrow)
}

func EventEndTime(end string, params map[string]interface{}) bool {
	const layout = "02.01.2006 15:04"

	if end == "skip" {
		return true
	}

	startTimeStr, ok := params["startTime"].(string)
	if !ok {
		return false
	}
	startTime, _ := time.Parse(layout, startTimeStr)

	endTime, err := time.Parse(layout, end)
	if err != nil {
		return false
	}

	if !endTime.After(startTime) {
		return false
	}

	return true
}

func EventRegisteredEndTime(registeredEnd string, params map[string]interface{}) bool {
	const layout = "02.01.2006 15:04"

	startTimeStr, ok := params["startTime"].(string)
	if !ok {
		return false
	}
	startTime, _ := time.Parse(layout, startTimeStr)
	registeredEndTime, err := time.Parse(layout, registeredEnd)
	if err != nil {
		return false
	}

	return registeredEndTime.Add(22 * time.Hour).Before(startTime)
}

func EventAfterRegistrationText(afterRegistrationText string, _ map[string]interface{}) bool {
	return utf8.RuneCountInString(afterRegistrationText) >= 10 && utf8.RuneCountInString(afterRegistrationText) <= 200
}

func MaxParticipants(maxParticipants string, _ map[string]interface{}) bool {
	_, err := strconv.Atoi(maxParticipants)
	return err == nil
}

func ExpectedParticipants(expectedParticipants string, _ map[string]interface{}) bool {
	_, err := strconv.Atoi(expectedParticipants)
	return err == nil
}
