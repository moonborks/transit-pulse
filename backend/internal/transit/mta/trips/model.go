package trips

type Trip struct {
	ID          string `json:"id"`
	RouteID     string `json:"route_id"`
	ServiceID   string `json:"service_id"`
	HeadSign    string `json:"head_sign"`
	DirectionID string `json:"direction_id"`
	ShapeID     string `json:"shape_id"`
}
