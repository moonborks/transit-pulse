package trips

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi"

	"github.com/moonborks/transit-pulse/internal/web"
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
	r.Get("/{id}", h.GetTrip)
	return r
}

func (h *TripHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	trips, err := h.tripService.GetAll(r.Context())
	if err != nil {
		web.WriteError(
			w,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"unable to retrieve trips from table",
		)
		return
	}

	if err := web.WriteJson(w, http.StatusOK, trips); err != nil {
		slog.Error("writing response json")
	}
}

func (h *TripHandler) GetTrip(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	trip, err := h.tripService.GetTrip(r.Context(), id)
	if err != nil {
		web.WriteError(
			w,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"unable to retrieve the specified trip",
		)
		return
	}

	if err := web.WriteJson(w, http.StatusOK, trip); err != nil {
		slog.Error("writing response json")
	}
}
