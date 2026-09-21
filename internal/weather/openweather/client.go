package openweather

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/semenov-timur/weatherbot/internal/domain"
)

type Client struct {
	apiKey string
	http   *http.Client
	log    *slog.Logger
}

type snapshotDTO struct {
	Weather []struct {
		Main        string `json:"main"`
		Description string `json:"description"`
	} `json:"weather"`

	Main struct {
		Temp      float64 `json:"temp"`
		FeelsLike float64 `json:"feels_like"`
		Pressure  int     `json:"pressure"`
		Humidity  int     `json:"humidity"`
	} `json:"main"`

	Wind struct {
		Speed float64 `json:"speed"`
	} `json:"wind"`

	Name string `json:"name"`
	Dt   int64  `json:"dt"`
}

// New создает и возвращает новый экземпляр [Client] для работы с API OpenWeather.
// Функция настраивает встроенный HTTP-клиент с таймаутом в 10 секунд.
//
// Возвращает ошибку, если apiKey или log равны nil/пустому значению.
func New(apiKey string, log *slog.Logger) (*Client, error) {
	if apiKey == "" {
		return nil, errors.New("no api key provided")
	}
	if log == nil {
		return nil, errors.New("no logger provided")
	}
	return &Client{
		apiKey: apiKey,
		log:    log,
		http:   &http.Client{Timeout: 10 * time.Second},
	}, nil
}

func (c *Client) Current(ctx context.Context, loc domain.Location) (domain.Snapshot, error) {
	params := url.Values{}
	params.Set("lat", strconv.FormatFloat(loc.Lat, 'f', -1, 64))
	params.Set("lon", strconv.FormatFloat(loc.Lon, 'f', -1, 64))
	params.Set("units", "metric")
	params.Set("lang", "ru")
	params.Set("appid", c.apiKey)
	endpoint := "https://api.openweathermap.org/data/2.5/weather?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return domain.Snapshot{}, fmt.Errorf("build request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return domain.Snapshot{}, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return domain.Snapshot{}, fmt.Errorf("unexpected status %s", resp.Status)
	}

	var dto snapshotDTO
	if err := json.NewDecoder(resp.Body).Decode(&dto); err != nil {
		return domain.Snapshot{}, fmt.Errorf("decode response: %w", err)
	}

	snap, err := dto.toDomain(loc)
	if err != nil {
		return domain.Snapshot{}, err
	}
	return snap, nil
}

func (d snapshotDTO) toDomain(loc domain.Location) (domain.Snapshot, error) {
	if len(d.Weather) == 0 {
		return domain.Snapshot{}, errors.New("response contains no weather data")
	}
	return domain.Snapshot{
		Location:    loc,
		ObservedAt:  time.Unix(d.Dt, 0).UTC(),
		Temp:        d.Main.Temp,
		FeelsLike:   d.Main.FeelsLike,
		Humidity:    d.Main.Humidity,
		Pressure:    d.Main.Pressure,
		WindSpeed:   d.Wind.Speed,
		Condition:   d.Weather[0].Main,
		Description: d.Weather[0].Description,
	}, nil
}
