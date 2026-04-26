# Fogelbot

A Discord bot that responds to messages with personality-driven quotes and reactions based on keywords, sentiment analysis, language detection, user identity, and random chance. Optionally uses an LLM to learn facts about users and generate personalized responses.

## Setup

### Requirements

- Go 1.21+
- A Discord bot token

### Environment Variables

Create a `.env` file or set these in your environment:

| Variable | Required | Description |
|---|---|---|
| `DISCORD_TOKEN` | Yes | Discord bot token |
| `ELECTROSHK_ID` | No | Discord user ID for user-specific triggers |
| `MODRIVER_ID` | No | Discord user ID for user-specific triggers |
| `GEMINI_API_KEY` | No | Google AI Studio API key for LLM features |
| `FACT_DB_PATH` | No | Path to the bbolt database file (default: `fogelbot.db`) |

### Build and Run

```sh
make build
./fogelbot
```

For a statically linked Linux binary:

```sh
make build-alpine
```

### Testing

```sh
make test          # run all tests
make test-race     # run with race detector
```

No `.env` file or Discord token is needed for tests.

## Docker Compose

The repo includes a `Dockerfile` and `docker-compose.yml` for running the bot as a Compose-managed service.

```sh
docker compose up --build -d
```

This setup:

- loads environment variables from `./.env` through Docker Compose
- stores the optional bbolt fact database in `./fogelbot.db`, mounted at `/data/fogelbot.db`
- does not publish any ports, since the bot only makes outbound connections

## Architecture

### Trigger System

The core abstraction is the `Trigger` interface (`triggers/triggers.go`). Triggers are registered in a `Registry` in priority order, and `Registry.Process()` returns the first match (first-match-wins).

There are two kinds of triggers:

- **Code-driven triggers** implement `Trigger` directly for complex stateful logic (e.g., `FogelTrigger`, `QuickReplyTrigger`, `GeneratedResponseTrigger`).
- **DSL triggers** use a fluent builder for simple data-driven definitions:

```go
r.On("Pizza").Match(`\bpizza\b`).Probability(0.10).Respond("Pizza time!")
```

All trigger registration lives in `triggers/definitions.go` via `RegisterAll()`, which reads top-to-bottom in priority order:

1. Fogel mention (sentiment-aware responses to "fogel" or @mention)
2. Quick reply (10% chance when users reply quickly after the bot)
3. Keyword triggers (5% chance on matching DSL-defined patterns)
4. Language detection (5% chance on non-English messages)
5. User-specific triggers (DSL-defined, targeted at specific users with 0.5%-2.5% gates)
6. Random negative (2.5% chance on negative sentiment messages)
7. Generated response (2.5% chance LLM-generated response, if configured)
8. Random quotes (1% chance of a random quote)

### Message Flow

```
Discord message
  -> bot.OnMessage
  -> Observe message for fact collection (if LLM configured)
  -> Build Context with all dependencies
  -> Registry.Process(ctx)
  -> Send first matching response
```

### LLM Integration (Optional)

When `GEMINI_API_KEY` is set, the bot collects user messages and uses Gemini 2.5 Flash-Lite to:

- **Extract facts** about users from their messages
- **Compact facts** when they exceed 15 per user (down to 7)
- **Generate responses** (2.5% ambient chance, plus a higher-probability branch for direct `@mention`s) using the bot's persona and known user facts

Facts are stored in a local bbolt database. Message buffering flushes on either a size threshold (10 messages) or a time interval (1 minute). All LLM errors are logged and swallowed to avoid affecting core bot functionality.

The LLM is abstracted behind an interface (`llm.LLM`), and all external dependencies are injectable for testing.

### Package Structure

```
main.go              Entry point
bot/                 Discord session, event handlers, wiring
config/              Environment variable loading
triggers/            Trigger interface, registry, DSL builder, all trigger implementations
  triggers.go          Core interfaces (Trigger, Registry, Context, dependency interfaces)
  builder.go           Fluent DSL builder for data-driven triggers
  definitions.go       RegisterAll() — all triggers in priority order
  generated.go         LLM-backed generated response trigger
  responses.go         Response string constants
  fogel.go             Fogel mention trigger
  quickreply.go        Quick reply trigger
  language.go          Language detection trigger
  random.go            Random negative sentiment trigger
llm/                 LLM abstraction
  llm.go               LLM interface
  gemini.go            Google Gemini implementation
facts/               Fact collection and storage
  store.go             bbolt-backed per-user fact storage
  buffer.go            Per-user message buffer with flush logic
  collector.go         Orchestrator: buffer + store + LLM
sentiment/           Sentiment analysis (VADER)
language/            Language detection (Lingua)
state/               Cooldown management
design-docs/         Design documents
```

### Testing

Tests use hand-written fakes (no external test frameworks) defined in `triggers/testhelpers_test.go`. The context builder pattern makes writing tests concise:

```go
ctx := newTestContext("hello").
    WithSentiment(0.5).
    WithRand(newAlwaysRand(0, 0.99)).
    Build()
```

All external dependencies (`SentimentAnalyzer`, `LanguageDetector`, `Rand`, `Clock`, `LLM`, `FactProvider`) are interfaces, so tests run without a Discord connection, API keys, or `.env` file.
