// Package telegram является адаптером к API Telegram.
package telegram

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime/debug"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/semenov-timur/weatherbot/internal/domain"
)

// Bot представляет собой обертку над Telegram-ботом,
// инкапсулирующую логику работы с API и логирование.
type Bot struct {
	api     *bot.Bot
	weather weatherSource
	loc     domain.Location
	log     *slog.Logger
}

type weatherSource interface {
	Current(ctx context.Context, loc domain.Location) (domain.Snapshot, error)
}

// New создает и инициализирует новый экземпляр [Bot] с указанным токеном.
//
// В случае ошибки создания бота, функция маскирует токен в сообщении об ошибке.
func New(token string, weather weatherSource, loc domain.Location, log *slog.Logger) (*Bot, error) {
	if log == nil {
		return nil, errors.New("logger is required")
	}
	if weather == nil {
		return nil, errors.New("weather source is required")
	}

	b := &Bot{weather: weather, loc: loc, log: log}
	handlerFunc := b.handleUpdate

	apiBot, err := bot.New(token, bot.WithDefaultHandler(handlerFunc))
	if err != nil {
		errMsg := strings.ReplaceAll(err.Error(), token, "***")
		return nil, fmt.Errorf("create bot: %s", errMsg)
	}

	b.api = apiBot
	return b, nil
}

// Run запускает бесконечный цикл обработки обновлений (Long Polling) Telegram-бота.
// Метод блокирует выполнение до тех пор, пока переданный контекст ctx не будет отменен.
func (b *Bot) Run(ctx context.Context) error {
	b.log.InfoContext(ctx, "bot started")

	b.api.Start(ctx)

	b.log.InfoContext(ctx, "bot stopped")

	return nil
}

// handleUpdate обрабатывает входящие обновления от Telegram API.
func (b *Bot) handleUpdate(ctx context.Context, api *bot.Bot, update *models.Update) {
	defer func() {
		if r := recover(); r != nil {
			b.log.ErrorContext(
				ctx, "handler panicked",
				slog.Any("panic", r),
				slog.String("stack", string(debug.Stack())),
			)
		}
	}()

	if update == nil {
		return
	}
	if update.Message == nil {
		return
	}
	if update.Message.Text == "" {
		return
	}

	switch update.Message.Text {
	case "/weather", "/now":
		b.handleWeather(ctx, update.Message.Chat.ID)
	default:
		b.log.DebugContext(
			ctx, "incoming message",
			slog.Int64("chat_id", update.Message.Chat.ID),
			slog.Int("text_len", len(update.Message.Text)),
		)
		b.sendMessage(ctx, update.Message.Chat.ID, update.Message.Text)
	}
}

func (b *Bot) handleWeather(ctx context.Context, chatID int64) {
	snapshot, err := b.weather.Current(ctx, b.loc)
	if err != nil {
		b.log.ErrorContext(ctx, "weather service failed", slog.Any("error", err))
		b.sendMessage(ctx, chatID, "Не получилось узнать погоду, попробуй через пару минут")
		return
	}
	b.sendMessage(ctx, chatID, formatSnapshot(snapshot))
}

func (b *Bot) sendMessage(ctx context.Context, chatID int64, text string) {
	_, err := b.api.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   text,
	})
	if err != nil {
		b.log.WarnContext(
			ctx, "send message",
			slog.Int64("chat_id", chatID),
			slog.Int("text_len", len(text)),
			slog.Any("error", err),
		)
	}
}
