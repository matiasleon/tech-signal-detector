package telegram

import (
	"context"
	"fmt"
	"strings"

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

// Send formats the title and url as a MarkdownV2 message and sends it to the configured chat.
func (n *Notifier) Send(_ context.Context, title, url string) error {
	text := fmt.Sprintf("*%s*\n%s", escapeMarkdownV2(title), url)

	msg := tgbotapi.NewMessage(n.chatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdownV2

	if _, err := n.bot.Send(msg); err != nil {
		return fmt.Errorf("telegram notifier: send message: %w", err)
	}

	return nil
}

// escapeMarkdownV2 escapes special characters required by Telegram's MarkdownV2 format.
// See: https://core.telegram.org/bots/api#markdownv2-style
func escapeMarkdownV2(s string) string {
	specialChars := `\_*[]()~` + "`" + `>#+-=|{}.!`
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if strings.ContainsRune(specialChars, r) {
			b.WriteRune('\\')
		}
		b.WriteRune(r)
	}
	return b.String()
}
