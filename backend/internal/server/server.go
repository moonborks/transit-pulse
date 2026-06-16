package server

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"github.com/moonborks/transit-pulse/internal/transit/mta/routes"
	"github.com/moonborks/transit-pulse/internal/transit/mta/shapes"
	"github.com/moonborks/transit-pulse/internal/transit/mta/stops"
	"github.com/moonborks/transit-pulse/internal/transit/mta/trips"
)

type Handlers struct {
	Route *routes.RouteHandler
	Shape *shapes.ShapeHandler
	Stop  *stops.StopHandler
	Trip  *trips.TripHandler
}

func MainRouter(h *Handlers) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Route("/api", func(api chi.Router) {
		api.Mount("/mta", MTARouter(h))
	})
	return r
}

func MTARouter(h *Handlers) http.Handler {
	r := chi.NewRouter()
	r.Mount("/routes", routes.RouteRoutes(h.Route))
	r.Mount("/shapes", shapes.ShapeRoutes(h.Shape))
	r.Mount("/stops", stops.StopRoutes(h.Stop))
	r.Mount("/trips", trips.TripRoutes(h.Trip))
	return r
}
