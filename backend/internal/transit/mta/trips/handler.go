package trips

import (
	"net/http"

	"github.com/go-chi/chi"
)

type TripHandler struct {
	tripService *TripService
}

func NewTripHandler(ts *TripService) *TripHandler {
	return &TripHandler{tripService: ts}
}

func TripRoutes(h *TripHandler) http.Handler {
	r := chi.NewRouter()
	r.Get("/", h.GetAll)
	r.Get("/{id}", h.Get)
	return r
}

func (h *TripHandler) Get(w http.ResponseWriter, r *http.Request) {
}

func (h *TripHandler) GetAll(w http.ResponseWriter, r *http.Request) {
}
