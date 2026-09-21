package domain

import "time"

type Location struct {
	Lat  float64
	Lon  float64
	Name string
}

type Snapshot struct {
	Location    Location
	ObservedAt  time.Time
	Temp        float64
	FeelsLike   float64
	Humidity    int
	Pressure    int
	WindSpeed   float64
	Condition   string
	Description string
}

func (s Snapshot) IsWet() bool {
	switch s.Condition {
	case "Rain", "Snow", "Drizzle", "Thunderstorm":
		return true
	default:
		return false
	}
}
