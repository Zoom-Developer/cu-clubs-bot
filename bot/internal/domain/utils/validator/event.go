package validator

import (
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/utils/location"
	"strconv"
	"time"
	"unicode/utf8"
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

	moscowTime := currentTime.In(location.Location())

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

	if !endTime.In(location.Location()).After(startTime) {
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

	return registeredEndTime.In(location.Location()).Add(22 * time.Hour).Before(startTime)
}

func EventAfterRegistrationText(afterRegistrationText string, _ map[string]interface{}) bool {
	if afterRegistrationText == "skip" {
		return true
	}
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
