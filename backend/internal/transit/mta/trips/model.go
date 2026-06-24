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

type TripShapeBounds struct {
	ShapeID       string
	StartShapeSeq int
	EndShapeSeq   int
}

type TripStopKey struct {
	ShortTripID string
	StopID      string
}

type TripSequenceKey struct {
	ShortTripID string
	Sequence    int64
}

type TripTrainLocationAPI struct {
	TripID     string  `json:"trip_id"`
	RouteID    string  `json:"route_id"`
	Lat        float64 `json:"lat"`
	Lon        float64 `json:"lon"`
	Bearing    float64 `json:"bearing"`
	NextStopID string  `json:"next_stop_id"`
}

type PrevStopInfo struct {
	PrevStopID              string
	PrevStationStopSequence int64
	PrevDepartureTime       string
}

type TrainContext struct {
	ShortTripID             string
	RouteID                 string
	NextStopID              string
	PrevStopID              string
	NextStationStopSequence int64
	PrevStationStopSequence int64
	NextArrivalTime         string
	PrevDepartureTime       string
	ProgressPercentage      float64
	CurrentShapeSequence    int64
}

type ShapeRange struct {
	PrevShapeSequence int64
	NextShapeSequence int64
}

type TrainCoordinates struct {
	Lat float64
	Lon float64
}
