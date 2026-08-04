package email

import (
	"encoding/json"
	"net/http"
)

type emailPayload struct {
	ContactAddress string
	Cc             string
	Subject        string
	Msg            string
}

func Hanlder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()

	var payload emailPayload
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&payload); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	if payload.Msg == "" || payload.ContactAddress == "" {
		http.Error(w, "message and from are required", http.StatusBadRequest)
		return
	}

	emailMsg := BuildMessage(payload)

	transport := SMTPSender{}
	if err := transport.Send(emailMsg, emailRecipients(payload)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func emailRecipients(payload emailPayload) []string {
	return []string{FROM_EMAIL, TO_EMAIL, payload.Cc}
}
