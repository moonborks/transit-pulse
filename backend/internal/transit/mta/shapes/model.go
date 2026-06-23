package shapes

type Shape struct {
	ID       string  `json:"id"`
	Sequence int64   `json:"sequence"`
	Lat      float64 `json:"lat"`
	Lon      float64 `json:"lon"`
}

type TargetShapeKey struct {
	ID       string
	Sequence int64
}
