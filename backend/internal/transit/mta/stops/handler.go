package stops

import (
	"net/http"

	"github.com/go-chi/chi"
)

type StopHandler struct {
	stopService *StopService
}

func NewStopHandler(ss *StopService) *StopHandler {
	return &StopHandler{stopService: ss}
}

func StopRoutes(h *StopHandler) http.Handler {
	r := chi.NewRouter()
	r.Get("/", h.GetAll)
	r.Get("/{id}", h.Get)
	return r
}

func (h *StopHandler) Get(w http.ResponseWriter, r *http.Request) {
}

func (h *StopHandler) GetAll(w http.ResponseWriter, r *http.Request) {
}
