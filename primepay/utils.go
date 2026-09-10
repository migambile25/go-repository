package primepay

import (
	"strings"
)

func normalizePhoneNumber(phoneNumber string) string {

	if strings.HasPrefix(phoneNumber, "0") {
		return "255" + phoneNumber[1:]
	}

	if strings.HasPrefix(phoneNumber, "255") {
		return phoneNumber
	}

	if strings.HasPrefix(phoneNumber, "+") {
		return phoneNumber[1:]
	}

	return phoneNumber
}
