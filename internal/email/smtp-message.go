package email

import (
	"fmt"
	"strings"
)

func BuildMessage(p emailPayload) []byte {
	var headers []string

	headers = append(headers,
		fmt.Sprintf("From: %s", FROM_EMAIL),
		fmt.Sprintf("To: %s", TO_EMAIL),
		fmt.Sprintf("Subject: %s", p.Subject),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
	)

	email := strings.Join(headers, "\r\n") +
		"\r\n\r\n" +
		"From: " + p.ContactAddress +
		"\n\n" +
		"Message:\n" + p.Msg +
		"\r\n"

	return []byte(email)
}
