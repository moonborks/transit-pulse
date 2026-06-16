package routes

type RouteService struct {
	routeRepo *RouteRepo
}

func NewRouteService(rr *RouteRepo) *RouteService {
	return &RouteService{routeRepo: rr}
}

func (s *RouteService) GetAll() {
}

func (s *RouteService) GetRoute() {
}
