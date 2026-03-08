package main

import (
	"context"
	"log"
	"os"
	"strconv"

	"github.com/matiasleonperalta/tech-signal-detectors/internal/domain"
	"github.com/matiasleonperalta/tech-signal-detectors/internal/infrastructure/collector"
	"github.com/matiasleonperalta/tech-signal-detectors/internal/infrastructure/llm"
	"github.com/matiasleonperalta/tech-signal-detectors/internal/infrastructure/persistence/sqlite"
	"github.com/matiasleonperalta/tech-signal-detectors/internal/infrastructure/telegram"
	"github.com/matiasleonperalta/tech-signal-detectors/internal/usecase"
)

func main() {
	botToken := mustEnv("TELEGRAM_BOT_TOKEN")
	chatID := mustInt64Env("TELEGRAM_CHAT_ID")
	anthropicKey := mustEnv("ANTHROPIC_API_KEY")
	dbPath := getEnv("DB_PATH", "signals.db")

	db, err := sqlite.Open(dbPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	sourceRepo := sqlite.NewSourceRepository(db)
	rawFeedRepo := sqlite.NewRawFeedRepository(db)
	signalRepo := sqlite.NewSignalRepository(db)

	if err := seedSources(context.Background(), sourceRepo); err != nil {
		log.Fatalf("seed sources: %v", err)
	}

	evaluator := llm.NewRelevanceEvaluator(anthropicKey)

	notifier, err := telegram.NewNotifier(botToken, chatID)
	if err != nil {
		log.Fatalf("create notifier: %v", err)
	}

	collectors := map[domain.SourceType]usecase.Collector{
		domain.SourceTypeHackerNews: collector.NewHackerNews(),
		domain.SourceTypeArXiv:      collector.NewArXiv(),
	}

	fetch := usecase.NewFetchUseCase(sourceRepo, rawFeedRepo, collectors)
	filter := usecase.NewFilterUseCase(sourceRepo, signalRepo, evaluator)
	deliver := usecase.NewDeliverUseCase(signalRepo, rawFeedRepo, notifier)

	handler := func(ctx context.Context) error {
		feeds, err := fetch.Execute(ctx)
		if err != nil {
			return err
		}
		signals, err := filter.Execute(ctx, feeds)
		if err != nil {
			return err
		}
		return deliver.Execute(ctx, signals)
	}

	bot, err := telegram.NewBot(botToken, chatID, handler)
	if err != nil {
		log.Fatalf("create bot: %v", err)
	}

	log.Println("Bot iniciado. Esperando comandos...")
	if err := bot.Start(context.Background()); err != nil {
		log.Fatalf("bot: %v", err)
	}
}

func seedSources(ctx context.Context, repo domain.SourceRepository) error {
	defaults := []domain.Source{
		{
			ID:             "hackernews",
			Name:           "Hacker News",
			Type:           domain.SourceTypeHackerNews,
			URL:            "",
			Enabled:        true,
			ScoreThreshold: 100,
		},
		{
			ID:             "arxiv-ai",
			Name:           "arXiv AI/ML",
			Type:           domain.SourceTypeArXiv,
			URL:            "https://export.arxiv.org/rss/cs.AI+cs.LG",
			Enabled:        true,
			ScoreThreshold: 0,
		},
	}

	for _, s := range defaults {
		if err := repo.Save(ctx, s); err != nil {
			return err
		}
	}
	return nil
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("missing required env var: %s", key)
	}
	return v
}

func mustInt64Env(key string) int64 {
	v := mustEnv(key)
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		log.Fatalf("invalid %s: %v", key, err)
	}
	return n
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
