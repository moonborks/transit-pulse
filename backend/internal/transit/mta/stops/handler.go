package stops

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi"

	"github.com/moonborks/transit-pulse/internal/web"
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
	r.Get("/{id}", h.GetStop)
	return r
}

func (h *StopHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	stops, err := h.stopService.GetAll(r.Context())
	if err != nil {
		web.WriteError(
			w,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"unable to retrieve stops from table",
		)
		return
	}

	if err := web.WriteJson(w, http.StatusOK, stops); err != nil {
		slog.Error("writing response json")
	}
}

func (h *StopHandler) GetStop(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	stop, err := h.stopService.GetStop(r.Context(), id)
	if err != nil {
		web.WriteError(
			w,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"unable to retrieve the specified stop",
		)
		return
	}

	if err := web.WriteJson(w, http.StatusOK, stop); err != nil {
		slog.Error("writing response json")
	}
}
