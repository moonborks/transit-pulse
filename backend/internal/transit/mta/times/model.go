package times

type Time struct {
	TripID        string `json:"trip_id"`
	StopID        string `json:"stop_id"`
	ArrivalTime   string `json:"arrival_time"`
	DepartureTime string `json:"departure_time"`
	StopSequence  int64  `json:"stop_sequence"`
}
