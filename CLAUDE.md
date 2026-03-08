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

**Sources:**
- HackerNews — filtered by upvotes + comments
- arXiv — filtered by LLM relevance on title/abstract (cs.AI, cs.LG focus)
- GitHub Trending — filtered by daily stars
- Product Hunt — filtered by upvotes and daily ranking
- TechCrunch — filtered by category/tags (no engagement metrics)

**Output:** Telegram bot — title + link per signal

**Trigger:** `/ultimas-novedades` Telegram command (on-demand, not scheduled)

**Filtering:** Simple threshold-based per source metrics for v1. arXiv uses LLM to evaluate relevance.

## Tech Stack

- **Language:** Go
- **Database:** SQLite (local/dev) → PostgreSQL (cloud). Swappable via repository interfaces.
- **Telegram:** `go-telegram-bot-api`
- **LLM:** Claude API via `anthropic-sdk-go` (used to evaluate arXiv relevance)
- **RSS parsing:** `gofeed`

## Architecture Principles

Clean Architecture — domain has zero knowledge of infrastructure details (SQLite, Telegram, HTTP). All external dependencies are behind interfaces defined by the domain.

- **Domain layer:** entities (Source, RawFeed, Signal) and repository interfaces. No imports from infrastructure.
- **Use case layer:** application logic (fetch, filter, send). Depends only on domain interfaces.
- **Infrastructure layer:** concrete implementations — SQLite repositories, HTTP collectors per source, Telegram sender, Claude client. Swapping SQLite for PostgreSQL only touches this layer.

SOLID and Clean Code apply throughout. Prefer small, focused functions and explicit interfaces over concrete dependencies.

## Project Structure

```
tech-signal-detectors/
├── cmd/
│   └── bot/
│       └── main.go          # entrypoint
├── internal/
│   ├── domain/
│   │   ├── source.go        # Source entity
│   │   ├── rawfeed.go       # RawFeed entity
│   │   ├── signal.go        # Signal entity
│   │   └── repository.go    # persistence interfaces
│   ├── usecase/
│   │   ├── fetch.go         # fetch from sources
│   │   ├── filter.go        # scoring and filtering
│   │   └── deliver.go       # send signals via Telegram
│   └── infrastructure/
│       ├── persistence/
│       │   └── sqlite/      # SQLite implementation of repositories
│       ├── collector/
│       │   ├── hackernews.go
│       │   ├── arxiv.go
│       │   ├── github.go
│       │   ├── producthunt.go
│       │   └── techcrunch.go
│       ├── llm/
│       │   └── claude.go    # Claude API client
│       └── telegram/
│           └── bot.go       # bot and command handlers
├── go.mod
└── CLAUDE.md
```

**Roadmap:**
- v2: Like/dislike feedback loop to personalize signal ranking over time
- v3: Web UI

## Domain Entities

**Source** — configured information sources
- id, name, type (hn/arxiv/github/producthunt/techcrunch), url, enabled

**RawFeed** — everything fetched before filtering
- id, source_id, external_id, title, url, published_at, score, fetched_at

**Signal** — items that passed the filter and were sent to the user
- id, raw_feed_id, relevance_score, sent_at, created_at
