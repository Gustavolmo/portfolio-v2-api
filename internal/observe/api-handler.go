package observe

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"
)

type Status struct {
	ConnectedClients int64 `json:"connectedClients"`
	UptimeSeconds    int64 `json:"uptimeSeconds"`
}

var connectedClients atomic.Int64
var startedAt = time.Now()

func Hanlder(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	connectedClients.Add(1)
	defer connectedClients.Add(-1)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	if err := send(w, flusher); err != nil {
		return
	}

	for {
		select {
		case <-ticker.C:
			if err := send(w, flusher); err != nil {
				return
			}

		case <-r.Context().Done():
			return
		}
	}
}

func send(w http.ResponseWriter, flusher http.Flusher) error {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	data, err := json.Marshal(snapshot())
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, "data: %s\n\n", data); err != nil {
		return err
	}

	flusher.Flush()
	return nil
}

func snapshot() Status {
	return Status{
		UptimeSeconds:    int64(time.Since(startedAt).Seconds()),
		ConnectedClients: connectedClients.Load(),
	}
}
