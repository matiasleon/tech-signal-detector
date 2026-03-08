package llm

import (
	"context"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/matiasleonperalta/tech-signal-detectors/internal/domain"
)

// RelevanceEvaluator uses the Claude API to determine whether a raw feed
// is relevant to a tech professional interested in software development,
// AI, and emerging technology trends.
type RelevanceEvaluator struct {
	client *anthropic.Client
	model  string
}

// NewRelevanceEvaluator returns a RelevanceEvaluator backed by Claude.
// It uses "claude-haiku-4-5-20251001" by default — fast and cheap for classification.
func NewRelevanceEvaluator(apiKey string) *RelevanceEvaluator {
	client := anthropic.NewClient(option.WithAPIKey(apiKey))
	return &RelevanceEvaluator{
		client: &client,
		model:  "claude-haiku-4-5-20251001",
	}
}

// Evaluate asks Claude whether the feed title is relevant for a tech
// professional and returns true if the answer contains "yes".
func (e *RelevanceEvaluator) Evaluate(ctx context.Context, feed domain.RawFeed) (bool, error) {
	prompt := fmt.Sprintf(
		"Is the following paper or article relevant for a tech professional interested in software development, AI, and emerging technology trends?\n\nTitle: %s\n\nRespond with ONLY \"yes\" or \"no\".",
		feed.Title,
	)

	message, err := e.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(e.model),
		MaxTokens: 10,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	})
	if err != nil {
		return false, fmt.Errorf("claude evaluate: %w", err)
	}

	if len(message.Content) == 0 {
		return false, fmt.Errorf("claude evaluate: empty response content")
	}

	text := message.Content[0].Text
	return strings.Contains(strings.ToLower(text), "yes"), nil
}
