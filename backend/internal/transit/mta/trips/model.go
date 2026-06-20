package trips

type Trip struct {
	ID          string  `json:"id"`
	DayOfWeek   string  `json:"day_of_week"`
	ShortTripID string  `json:"short_trip_id"`
	RouteID     string  `json:"route_id"`
	ServiceID   string  `json:"service_id"`
	Headsign    string  `json:"headsign"`
	DirectionID string  `json:"direction_id"`
	ShapeID     *string `json:"shape_id"`
}

type FreqDay string

const (
	Everyday FreqDay = "everyday"
	Weekday  FreqDay = "weekday"
	Saturday FreqDay = "saturday"
	Sunday   FreqDay = "sunday"
)

type TripAPI struct {
	RouteID  string  `json:"route_id"`
	Headsign string  `json:"headsign"`
	ShapeID  *string `json:"shape_id"`
}
