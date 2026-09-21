package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"github.com/semenov-timur/weatherbot/internal/config"
	"github.com/semenov-timur/weatherbot/internal/telegram"
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

	bot, err := telegram.New(cfg.TelegramToken, log)
	if err != nil {
		log.Error("create bot", slog.Any("error", err))
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := bot.Run(ctx); err != nil {
		log.Error("bot failed", slog.Any("error", err))
	}
}
