package nextstop

type NextStop struct {
	StopID      string `json:"stop_id"`
	TripID      string `json:"trip_id"`
	RouteID     string `json:"route_id"`
	ArrivalTime string `json:"arrival_time"`
}
