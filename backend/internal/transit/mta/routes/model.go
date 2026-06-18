package routes

type Route struct {
	ID        string `json:"id"`
	ShortName string `json:"short_name"`
	LongName  string `json:"long_name"`
	Type      int64  `json:"type"`
	Color     string `json:"color"`
}
