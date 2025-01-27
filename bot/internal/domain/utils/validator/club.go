package validator

import (
	"unicode/utf8"
)

func ClubName(name string, _ map[string]interface{}) bool {
	return utf8.RuneCountInString(name) >= 3 && utf8.RuneCountInString(name) <= 30
}

func ClubDescription(description string, _ map[string]interface{}) bool {
	return utf8.RuneCountInString(description) <= 400
}
