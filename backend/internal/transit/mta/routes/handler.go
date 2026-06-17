package routes

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi"

	"github.com/moonborks/transit-pulse/internal/web"
)

type RouteHandler struct {
	routeService *RouteService
}

func NewRouteHandler(rs *RouteService) *RouteHandler {
	return &RouteHandler{routeService: rs}
}

func RouteRoutes(h *RouteHandler) http.Handler {
	r := chi.NewRouter()
	r.Get("/", h.GetAll)
	r.Get("/{id}", h.GetRoute)
	return r
}

func (h *RouteHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	routes, err := h.routeService.GetAll(r.Context())
	if err != nil {
		web.WriteError(
			w,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"unable to retrieve routes from table",
		)
		return
	}

	if err := web.WriteJson(w, http.StatusOK, routes); err != nil {
		slog.Error("writing response json")
	}
}

func (h *RouteHandler) GetRoute(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	route, err := h.routeService.GetRoute(r.Context(), id)
	if err != nil {
		web.WriteError(
			w,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"unable to retrieve the specified route",
		)
		return
	}

	if err := web.WriteJson(w, http.StatusOK, route); err != nil {
		slog.Error("writing response json")
	}
}
