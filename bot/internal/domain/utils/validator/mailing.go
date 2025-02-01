package validator

import "unicode/utf8"

func MailingText(text string, _ map[string]interface{}) bool {
	return utf8.RuneCountInString(text) <= 500
}
