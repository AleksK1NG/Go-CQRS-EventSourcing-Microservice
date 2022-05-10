package utils

import "strings"

func CheckErrForMessages(err error, messages ...string) bool {
	for _, message := range messages {
		if strings.Contains(err.Error(), message) {
			return true
		}
	}
	return false
}

func CheckErrForMessagesCaseInSensitive(err error, messages ...string) bool {
	for _, message := range messages {
		if strings.Contains(strings.TrimSpace(strings.ToLower(err.Error())), strings.TrimSpace(strings.ToLower(message))) {
			return true
		}
	}
	return false
}
