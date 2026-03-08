package telegram

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	cmdUltimasNovedades = "/ultimas-novedades"
	msgBuscando         = "Buscando novedades..."
	msgSinNovedades     = "No hay novedades por ahora."
)

// Bot listens for Telegram commands and delegates to a handler function.
type Bot struct {
	api     *tgbotapi.BotAPI
	chatID  int64
	handler func(ctx context.Context) error
}

// NewBot creates a new Bot using the provided token, chat ID, and handler.
func NewBot(token string, chatID int64, handler func(ctx context.Context) error) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("telegram bot: create bot api: %w", err)
	}

	return &Bot{
		api:     api,
		chatID:  chatID,
		handler: handler,
	}, nil
}

// Start begins the polling loop. It blocks until ctx is cancelled.
func (b *Bot) Start(ctx context.Context) error {
	updateCfg := tgbotapi.NewUpdate(0)
	updateCfg.Timeout = 60

	updates := b.api.GetUpdatesChan(updateCfg)

	for {
		select {
		case <-ctx.Done():
			b.api.StopReceivingUpdates()
			return ctx.Err()

		case update, ok := <-updates:
			if !ok {
				return nil
			}

			if update.Message == nil {
				continue
			}

			if update.Message.Command() != "ultimas-novedades" {
				continue
			}

			b.sendText(msgBuscando)

			if err := b.handler(ctx); err != nil {
				b.sendText(fmt.Sprintf("Error: %v", err))
				continue
			}

			// The handler is responsible for sending individual notifications via the
			// Notifier. If it returns nil without sending anything the user is informed.
			b.sendText(msgSinNovedades)
		}
	}
}

// sendText sends a plain-text message to the configured chat, logging any error.
func (b *Bot) sendText(text string) {
	msg := tgbotapi.NewMessage(b.chatID, text)
	if _, err := b.api.Send(msg); err != nil {
		fmt.Printf("telegram bot: send message: %v\n", err)
	}
}
