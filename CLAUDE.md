# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

AI Signal Intelligence system that detects important technological signals from multiple information sources and delivers them as actionable insights.

## Value Proposition

Tech professionals face an impossible amount of information — dozens of sources, hundreds of daily posts, constant noise. The real problem is not access to information, it's **knowing what actually matters**.

This system acts as an intelligent filter: it monitors the sources where relevant tech signals appear first (research papers, trending repos, community discussions, new products), scores them by objective engagement metrics, and delivers only what's worth reading — directly to Telegram, on demand.

The result: stay informed like a CTO without spending hours reading feeds.

## Product Definition (v1)

**Goal:** Stay connected with the latest tech trends and AI best practices, as a CTO would.

**Sources (implemented):**
- HackerNews — filtered by score threshold (upvotes ≥ 100)
- arXiv (cs.AI + cs.LG) — filtered by Claude API relevance evaluation
- TechCrunch — editorial, always passes
- Product Hunt — RSS, always passes

**Sources (roadmap):**
- GitHub Trending — filtered by daily stars

**Output:** Telegram bot — title + published date + link per signal

**Trigger:** `/ultimas_novedades` Telegram command (on-demand, not scheduled)

**Filtering:** Each source has its own `ScoreThreshold`. Collectors compute the score. arXiv uses Claude API (Haiku) to evaluate relevance since it has no engagement metrics.

**Deduplication:** RawFeeds are deduplicated by `(source_id, external_id)` — each item is fetched and evaluated only once ever.

## Tech Stack

- **Language:** Go 1.22+
- **Database:** SQLite via `modernc.org/sqlite` (no CGO). Swappable to PostgreSQL via repository interfaces.
- **Telegram:** `go-telegram-bot-api/v5`
- **LLM:** Claude Haiku (`claude-haiku-4-5-20251001`) via `anthropic-sdk-go` — used only for arXiv relevance
- **RSS parsing:** `gofeed`

## Commands

```bash
# Run
source .env && go run ./cmd/bot

# Build
go build -o bin/bot ./cmd/bot

# Verify compilation
go build ./...
```

## Environment Variables

| Variable | Required | Description |
|---|---|---|
| `TELEGRAM_BOT_TOKEN` | Yes | From @BotFather |
| `TELEGRAM_CHAT_ID` | Yes | Your Telegram user ID |
| `ANTHROPIC_API_KEY` | Yes | From console.anthropic.com |
| `DB_PATH` | No | SQLite file path (default: `signals.db`) |

## Architecture Principles

Clean Architecture — domain has zero knowledge of infrastructure details (SQLite, Telegram, HTTP). All external dependencies are behind interfaces defined by the domain.

- **Domain layer:** entities (Source, RawFeed, Signal) and repository interfaces. No imports from infrastructure.
- **Use case layer:** application logic (fetch, filter, deliver). Depends only on domain interfaces.
- **Infrastructure layer:** concrete implementations — SQLite repositories, HTTP collectors per source, Telegram sender, Claude client. Swapping SQLite for PostgreSQL only touches this layer.

SOLID and Clean Code apply throughout. Prefer small, focused functions and explicit interfaces over concrete dependencies.

## Project Structure

```
tech-signal-detectors/
├── cmd/bot/main.go              # entrypoint — wires all components, seeds sources
├── internal/
│   ├── domain/
│   │   ├── source.go            # Source entity (includes ScoreThreshold)
│   │   ├── rawfeed.go           # RawFeed entity
│   │   ├── signal.go            # Signal entity
│   │   └── repository.go       # persistence interfaces
│   ├── usecase/
│   │   ├── fetch.go             # Collector interface + FetchUseCase
│   │   ├── filter.go            # RelevanceEvaluator interface + FilterUseCase
│   │   └── deliver.go           # Notifier interface + DeliverUseCase
│   └── infrastructure/
│       ├── persistence/sqlite/  # db.go (shared conn + schema) + 3 repo files
│       ├── collector/           # hackernews.go, arxiv.go, techcrunch.go, producthunt.go
│       ├── llm/claude.go        # RelevanceEvaluator via Claude Haiku
│       └── telegram/            # bot.go (command handler), notifier.go
├── go.mod
├── .env.example
└── CLAUDE.md
```

## Adding a New Source

1. Add `SourceType` constant to `domain/source.go`
2. Implement `usecase.Collector` in `infrastructure/collector/`
3. Register it in `cmd/bot/main.go` collectors map
4. Add seed entry in `seedSources()` with the appropriate `ScoreThreshold`

No changes needed in use cases or domain.

## Domain Entities

**Source** — configured information sources
- id, name, type, url, enabled, score_threshold

**RawFeed** — everything fetched before filtering
- id, source_id, external_id, title, url, score, published_at, fetched_at

**Signal** — items that passed the filter
- id, raw_feed_id, relevance_score, sent_at, created_at, published_at

## Design Philosophy

Always prefer the simplest solution that solves the actual problem. When evaluating options, question assumptions and look for the approach with the least friction for the developer and user.

**Example:** instead of documenting `source .env && go run ./cmd/bot` (two steps, requires shell knowledge), we added `godotenv` so the app loads `.env` automatically — `go run ./cmd/bot` is enough. The library is a small tradeoff for a much simpler experience.

When a workaround feels awkward, stop and ask: is there a cleaner solution?

## Development Methodology

### Parallel development with subagents

This project uses Claude Code subagents to parallelize independent tasks and accelerate development.

**Key facts about subagents:**
- Each subagent gets its own isolated context window — it does NOT inherit the parent conversation
- Subagents automatically load `CLAUDE.md` from the project directory — use this to share context instead of repeating it in every prompt
- Subagents cannot spawn other subagents (no nesting)
- Background subagents (`run_in_background: true`) run concurrently; permissions must be pre-approved upfront

**When to launch subagents in parallel:**
- Multiple files that are independent of each other (e.g. one collector per source, one repo per entity)
- Infrastructure components that don't depend on each other (SQLite + Telegram + Claude client)
- Any set of tasks where the output of one is not the input of another

**When to work sequentially (main conversation):**
- Domain design and architectural decisions — these need back-and-forth with the user
- When one component depends on another being complete (e.g. main.go after all infrastructure exists)
- When reviewing and validating that everything compiles together
- Quick, targeted changes — subagents have startup overhead

**The workflow we follow:**
1. **Define the "what" before the "how"** — agree on entities, interfaces, and flow before touching code
2. **Keep CLAUDE.md updated** — subagents load it automatically, reducing context that needs to be passed in each prompt
3. **Design interfaces in the domain layer first** — use cases and infrastructure implement them
4. **Launch parallel background subagents** for independent implementation tasks
5. **Verify compilation** after all agents complete (`go build ./...`)
6. **Commit working state** before moving to the next layer

**Optimization: use CLAUDE.md as shared context**
Since subagents load CLAUDE.md automatically, keep entities, interfaces, and architectural decisions documented here. This avoids copying domain definitions into every subagent prompt — just reference "see CLAUDE.md for domain entities".

**Example from this project:**
- Domain entities → sequential (foundation for everything, requires design decisions)
- Use cases (fetch, filter, deliver) → parallel (3 background agents)
- Infrastructure (SQLite, Telegram, Claude, HackerNews, arXiv) → parallel (5 background agents)
- main.go → sequential (wires everything together, depends on all infrastructure)

## Iteration History

All bugs found, design decisions, and improvements are documented in [`ITERATIONS.md`](./ITERATIONS.md). Read it to understand *why* things are built the way they are — especially decisions that might look surprising at first (e.g. why TechCrunch always passes, why the filter collects all signals before limiting, why the Atom API instead of arXiv RSS).

## Roadmap

- v2: Like/dislike feedback loop to personalize signal ranking over time
- v3: Web UI
