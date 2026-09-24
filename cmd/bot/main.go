package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"github.com/semenov-timur/weatherbot/internal/config"
	"github.com/semenov-timur/weatherbot/internal/domain"
	"github.com/semenov-timur/weatherbot/internal/telegram"
	"github.com/semenov-timur/weatherbot/internal/weather/openweather"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	cfg, err := config.Load()
	if err != nil {
		log.Error("load config", slog.Any("error", err))
		os.Exit(1)
	}

	weatherClient, err := openweather.New(cfg.OpenWeatherAPIKey, log)
	if err != nil {
		log.Error("connect to weather api", slog.Any("error", err))
		os.Exit(1)
	}

	loc := domain.Location{Lat: 55.7558, Lon: 37.6173, Name: "Москва"}

	bot, err := telegram.New(cfg.TelegramToken, weatherClient, loc, log)
	if err != nil {
		log.Error("create bot", slog.Any("error", err))
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := bot.Run(ctx); err != nil {
		log.Error("bot failed", slog.Any("error", err))
		os.Exit(1)
	}
}
