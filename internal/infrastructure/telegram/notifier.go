package telegram

import (
	"context"
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Notifier sends Telegram messages, implementing usecase.Notifier.
type Notifier struct {
	bot    *tgbotapi.BotAPI
	chatID int64
}

// NewNotifier creates a new Notifier using the provided bot token and chat ID.
func NewNotifier(token string, chatID int64) (*Notifier, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("telegram notifier: create bot: %w", err)
	}

	return &Notifier{
		bot:    bot,
		chatID: chatID,
	}, nil
}

// Send formats the title, date and url as a message and sends it to the configured chat.
func (n *Notifier) Send(_ context.Context, title, url string, publishedAt time.Time) error {
	text := fmt.Sprintf("%s\n📅 %s | 🔗 %s", title, publishedAt.Format("02 Jan 2006"), url)

	msg := tgbotapi.NewMessage(n.chatID, text)

	if _, err := n.bot.Send(msg); err != nil {
		return fmt.Errorf("telegram notifier: send message: %w", err)
	}

	return nil
}
