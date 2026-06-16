package routes

import (
	"net/http"

	"github.com/go-chi/chi"
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
}

func (h *RouteHandler) GetRoute(w http.ResponseWriter, r *http.Request) {
}
