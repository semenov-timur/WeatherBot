package telegram

import (
	"fmt"

	"github.com/semenov-timur/weatherbot/internal/domain"
)

func formatSnapshot(s domain.Snapshot) string {
	return fmt.Sprintf("%s, сейчас %s\n%+.0f°, ощущается %+.0f°\n%s", s.Location.Name, conditionEmoji(s.Condition), s.Temp, s.FeelsLike, s.Description)
}

func conditionEmoji(condition string) string {
	switch condition {
	case "Clear":
		return "☀️"
	case "Clouds":
		return "☁️"
	case "Rain":
		return "🌧️"
	case "Snow":
		return "❄️"
	case "Thunderstorm":
		return "🌩️"
	default:
		return ""
	}
}
