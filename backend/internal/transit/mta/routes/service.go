package routes

import "context"

type RouteService struct {
	routeRepo *RouteRepo
}

func NewRouteService(rr *RouteRepo) *RouteService {
	return &RouteService{routeRepo: rr}
}

func (s *RouteService) GetAll(ctx context.Context) ([]*Route, error) {
	return s.routeRepo.GetAll(ctx)
}

func (s *RouteService) GetRoute(ctx context.Context, id string) (*Route, error) {
	return s.routeRepo.GetRoute(ctx, id)
}
