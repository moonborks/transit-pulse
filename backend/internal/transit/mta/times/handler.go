package times

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi"

	"github.com/moonborks/transit-pulse/internal/web"
)

type TimeHandler struct {
	timeService *TimeService
}

func NewTimeHandler(ts *TimeService) *TimeHandler {
	return &TimeHandler{timeService: ts}
}

func TimeRoutes(h *TimeHandler) http.Handler {
	r := chi.NewRouter()
	r.Get("/", h.GetAll)
	return r
}

func (h *TimeHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	times, err := h.timeService.GetAll(r.Context())
	if err != nil {
		web.WriteError(
			w,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"unable to retrieve times from table",
		)
		return
	}

	if err := web.WriteJson(w, http.StatusOK, times); err != nil {
		slog.Error("writing response json")
	}
}
