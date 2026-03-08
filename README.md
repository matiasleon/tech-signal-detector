# Tech Signal Detectors

Monitors HackerNews and arXiv to surface the most relevant tech signals and delivers them via Telegram on demand.

## Prerequisites

- [Go 1.22+](https://go.dev/dl/)
- A Telegram bot token (see setup below)
- An Anthropic API key (used to evaluate arXiv paper relevance)

## Setup

### 1. Clone and install dependencies

```bash
git clone https://github.com/matiasleonperalta/tech-signal-detectors
cd tech-signal-detectors
go mod download
```

### 2. Create your Telegram bot

1. Open Telegram and talk to [@BotFather](https://t.me/BotFather)
2. Send `/newbot` and follow the instructions
3. Copy the bot token you receive

### 3. Get your Telegram Chat ID

1. Talk to [@userinfobot](https://t.me/userinfobot) on Telegram
2. It will reply with your user ID — that's your `TELEGRAM_CHAT_ID`

### 4. Get an Anthropic API key

1. Create an account at [console.anthropic.com](https://console.anthropic.com)
2. Go to **API Keys** and create a new key
3. Add credits to your account (usage is minimal — only arXiv papers are evaluated)

### 5. Configure environment variables

```bash
cp .env.example .env
# Edit .env with your actual values
```

### 6. Run

```bash
source .env && go run ./cmd/bot
```

Or build a binary:

```bash
go build -o bin/bot ./cmd/bot
source .env && ./bin/bot
```

## Usage

Once the bot is running, open Telegram and send `/ultimas-novedades` to your bot. It will fetch the latest signals from HackerNews and arXiv, filter them, and reply with the most relevant ones.

## Project structure

```
internal/
├── domain/          # Entities and repository interfaces
├── usecase/         # Application logic (fetch, filter, deliver)
└── infrastructure/  # Collectors, SQLite, Telegram, Claude API
cmd/bot/main.go      # Entrypoint — wires everything together
```
