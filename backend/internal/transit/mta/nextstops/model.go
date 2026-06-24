package nextstops

type NextStop struct {
	StopID      string `json:"stop_id"`
	ShortTripID string `json:"short_trip_id"`
	RouteID     string `json:"route_id"`
	ArrivalTime string `json:"arrival_time"`
}
