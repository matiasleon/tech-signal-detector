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
- **Claude too permissive** — prompt approved medicine, biology, social science papers. Tightened to CTO-relevant topics: LLMs, ML infra, software dev, cloud, security, physics

### Known issues (carried forward to v0.2)
- 54 messages sent at once — no limit per execution
- HN fills all slots before arXiv gets a chance — early-exit in filter loop biased toward first source processed
- Two separate bot API instances — `Bot` and `Notifier` each called `tgbotapi.NewBotAPI` separately
- No user feedback while Claude evaluates arXiv papers (up to 30s of silence)

---

## v0.2 — Multi-source, date sorting, signal limit (2026-03-08)

### What was added
- **TechCrunch collector** — RSS via `gofeed`, all items pass filter (editorial source, Score=0, Threshold=0)
- **Product Hunt collector** — RSS via `gofeed`, same filter logic as TechCrunch
- **`PublishedAt` field on Signal** — persisted in SQLite, used for sorting
- **Sort by date, then limit** — filter now collects ALL passing signals from all sources, sorts by `PublishedAt` descending, then slices to `maxSignals` (50). Previously had an early-exit that biased results toward whichever source was processed first (HN)
- **Date shown in Telegram message** — format: `{title}\n📅 02 Jan 2006 | 🔗 {url}`

### Design decisions
- **`maxSignals` lives in `FilterUseCase`, not the handler** — it's a domain/business rule ("how many signals to deliver per run"), not an orchestration concern. The use case owns it.
- **TechCrunch always passes filter** — editorial curation is the filter. No score threshold or LLM evaluation needed. Handled in `FilterUseCase.passes()` with a dedicated case.
- **Sort after collecting all sources** — ensures fair representation across sources. The old approach processed sources in map iteration order (non-deterministic in Go) and stopped at the limit, meaning some sources might never contribute.
- **SQLite migration via `ALTER TABLE ... ADD COLUMN`** — used `isDuplicateColumn()` helper to ignore "duplicate column" errors, making the migration idempotent and safe to re-run on existing databases.

### Known issues (still pending)
- Two separate bot API instances — `Bot` and `Notifier` each call `tgbotapi.NewBotAPI`
- HN lets non-tech articles through (score-only filter, no LLM) — e.g. "Yoghurt delivery women"
- GitHub Trending collector not yet implemented
