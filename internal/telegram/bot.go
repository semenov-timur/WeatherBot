// Package telegram является адаптером к API Telegram.
package telegram

import (
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// Bot представляет собой обертку над Telegram-ботом,
// инкапсулирующую логику работы с API и логирование.
type Bot struct {
	api *bot.Bot
	log *slog.Logger
}

// New создает и инициализирует новый экземпляр [Bot] с указанным токеном.
//
// В случае ошибки создания бота, функция маскирует токен в сообщении об ошибке.
func New(token string, log *slog.Logger) (*Bot, error) {
	if log == nil {
		return nil, fmt.Errorf("no logger provided")
	}

	b := &Bot{log: log}
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
// На данный момент метод реализует простейшее эхо: принимает текстовые сообщения
// и отправляет их обратно пользователю в тот же чат.
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

	b.log.DebugContext(
		ctx, "incoming message",
		slog.Int64("chat_id", update.Message.Chat.ID),
		slog.Int("text_len", len(update.Message.Text)),
	)

	_, err := api.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   update.Message.Text,
	})
	if err != nil {
		b.log.WarnContext(
			ctx, "send message",
			slog.Int64("chat_id", update.Message.Chat.ID),
			slog.Int("text_len", len(update.Message.Text)),
			slog.Any("error", err),
		)
	}
}
