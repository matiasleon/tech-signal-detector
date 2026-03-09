package telegram

import (
	"context"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	msgBuscando     = "Buscando novedades... puede tardar hasta 30 segundos mientras evaluamos papers con IA 🤖"
	msgSinNovedades = "No hay novedades nuevas por ahora."
	msgStart        = "👋 Hola! Soy tu detector de señales tecnológicas.\n\nMonitoreo HackerNews y arXiv para traerte solo lo que vale la pena leer.\n\nComandos disponibles:\n/ultimas_novedades — trae las señales más relevantes"
)

// Bot listens for Telegram commands and delegates to a handler function.
type Bot struct {
	api     *tgbotapi.BotAPI
	chatID  int64
	handler func(ctx context.Context) (int, error)
}

// NewBot creates a new Bot using the provided token, chat ID, and handler.
func NewBot(token string, chatID int64, handler func(ctx context.Context) (int, error)) (*Bot, error) {
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

			if update.Message == nil || !update.Message.IsCommand() {
				continue
			}

			cmd := update.Message.Command()
			log.Printf("[bot] command received: /%s", cmd)

			switch cmd {
			case "start":
				b.sendText(msgStart)

			case "ultimas_novedades":
				b.sendText(msgBuscando)
				count, err := b.handler(ctx)
				if err != nil {
					log.Printf("[bot] handler error: %v", err)
					b.sendText(fmt.Sprintf("Error: %v", err))
				} else if count == 0 {
					b.sendText(msgSinNovedades)
				}
			default:
				log.Printf("[bot] unknown command: /%s", cmd)
			}
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
