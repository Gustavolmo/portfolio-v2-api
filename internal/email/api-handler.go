package email

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"golang.org/x/time/rate"
)

type emailPayload struct {
	ContactAddress string
	Subject        string
	Msg            string
}

const MAX_PAYLOAD_BYTES = 16 << 10

var requestLimiter = NewRateLimiter(rate.Every(30*time.Second), 3)

func Handler(w http.ResponseWriter, r *http.Request) {
	err, status := SendEmail(w, r, SMTPSender{})
	if err != nil {
		http.Error(w, err.Error(), status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func SendEmail(w http.ResponseWriter, r *http.Request, transport Sender) (error, int) {
	r.Body = http.MaxBytesReader(w, r.Body, MAX_PAYLOAD_BYTES)
	defer r.Body.Close()

	if r.Method != http.MethodPost {
		return fmt.Errorf("method not allowed"), http.StatusMethodNotAllowed
	}

	ip, err := ClientIP(r)
	if err != nil {
		return fmt.Errorf("invalid remote address"), http.StatusBadRequest
	}

	if !requestLimiter.Allow(ip) {
		w.Header().Set("Retry-After", "3600")
		return fmt.Errorf("too many requests, server cap reached, please try again soon"), http.StatusTooManyRequests
	}

	payload, err, status := validatePayload(r.Body)
	if err != nil {
		return err, status
	}

	if err := transport.Send(BuildMessage(payload), []string{FROM_EMAIL, TO_EMAIL}); err != nil {
		return fmt.Errorf("Unable to send email right now"), http.StatusInternalServerError
	}

	return nil, 0
}

func validatePayload(body io.ReadCloser) (emailPayload, error, int) {
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()

	var payload emailPayload
	if err := decoder.Decode(&payload); err != nil {
		if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
			return payload, fmt.Errorf("request body too large"), http.StatusRequestEntityTooLarge
		}
		return payload, fmt.Errorf("invalid JSON body"), http.StatusBadRequest
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return payload, fmt.Errorf("invalid JSON body"), http.StatusBadRequest
	}

	payload.ContactAddress = strings.TrimSpace(payload.ContactAddress)
	payload.Subject = strings.TrimSpace(payload.Subject)
	payload.Msg = strings.TrimSpace(payload.Msg)

	if payload.Msg == "" || payload.ContactAddress == "" {
		return payload, fmt.Errorf("message and from are required"), http.StatusBadRequest
	}

	if !validHeaderValue(payload.Subject, 160) {
		return payload, fmt.Errorf("invalid subject"), http.StatusBadRequest
	}

	if strings.ContainsAny(payload.ContactAddress, "\r\n") {
		return payload, fmt.Errorf("invalid from address"), http.StatusBadRequest
	}

	contact, err := mail.ParseAddress(payload.ContactAddress)
	if payload.ContactAddress != "" && (err != nil || contact.Address != payload.ContactAddress) {
		return payload, fmt.Errorf("invalid from address"), http.StatusBadRequest
	}

	return payload, nil, 0
}

func validHeaderValue(value string, maxLength int) bool {
	if value == "" || len(value) > maxLength {
		return false
	}

	for _, char := range value {
		if char < 32 || char == 127 {
			return false
		}
	}

	return true
}
