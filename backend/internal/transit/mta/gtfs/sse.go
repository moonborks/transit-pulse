package gtfs

import (
	"fmt"
	"net/http"
)

type SSE struct {
	TripChannel chan string
}

func NewSSEChannel(tripChannel chan string) *SSE {
	return &SSE{TripChannel: tripChannel}
}

func (c *SSE) TripEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
	}

	for {
		select {
		case msg := <-c.TripChannel:
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}
