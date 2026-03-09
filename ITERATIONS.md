# Iterations

Documents the evolution of the project — decisions made, bugs found, and improvements applied.

---

## v0.1 — First working pipeline (2026-03-08)

### What was built
- Full pipeline: HackerNews + arXiv → filter → Telegram
- SQLite persistence with deduplication
- Claude Haiku for arXiv relevance evaluation
- `/ultimas_novedades` command triggers on-demand fetch

### Bugs found and fixed
- **Telegram commands with hyphens not recognized** — Telegram only supports letters, numbers and `_` in commands. Changed `/ultimas-novedades` → `/ultimas_novedades`
- **HackerNews API returning 400** — `numericFilters=points>0` had unencoded `>`. Fixed to `%3E`
- **arXiv RSS empty on weekends** — RSS feed explicitly skips Saturday and Sunday. Switched to the Atom API (`/api/query`) which works 7 days a week
- **`ParseURLWithContext` wrong argument order** — `gofeed` signature is `(url, ctx)` not `(ctx, url)`
- **MarkdownV2 parse errors in Telegram** — special characters (`.`, `-`, `'`) in titles broke the message. Switched to plain text
- **HN fetching all-time top stories** — missing date filter brought posts from 2013–2024. Added `created_at_i > {48h ago}` filter to Algolia query
- **No user feedback when 0 signals** — handler returned `error` only, no way to distinguish 0 results. Changed to `(int, error)` and added "No hay novedades nuevas por ahora." message
- **Two separate bot API instances** — `Bot` and `Notifier` each called `tgbotapi.NewBotAPI` separately (wasteful, double authentication)
- **Claude too permissive** — prompt approved medicine, biology, social science papers. Tightened to CTO-relevant topics: LLMs, ML infra, software dev, cloud, security, physics

### Fixed after first run
- **HN fetching all-time top stories** — added `created_at_i > {48h ago}` filter
- **Claude too permissive** — tightened prompt to CTO-relevant topics (LLMs, ML infra, software dev, cloud, security, physics)

### Known issues pending
- 54 messages sent at once — need max signals limit per execution
- No user feedback while Claude is evaluating arXiv (can take 20+ seconds of silence)
- HN score threshold (100) not yet validated — may need tuning
