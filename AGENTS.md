# Fogelbot

Fogelbot is a Discord bot implemented in Go. Legacy Python artifacts may still exist in the repository, but the active implementation is the Go code under the root packages in this directory.

## What The Project Does

The bot watches Discord messages and replies with a Fogelbot-style response when a trigger matches. Triggers can depend on message content, user identity, sentiment, language detection, cooldown state, and random chance.

LLM features are optional. When enabled, the bot can learn lightweight facts about users from their messages and use those facts to generate more personalized responses. Core trigger behavior should still work when no LLM is configured.

## How To Run And Configure It

Configuration is environment-driven. The current Go config lives in `config/config.go`.

- Required: `DISCORD_TOKEN`
- Optional: `ELECTROSHK_ID`, `MODRIVER_ID`
- Optional LLM: `GEMINI_API_KEY`
- Optional fact storage path: `FACT_DB_PATH` (defaults to a local bbolt database file)

Common commands:

- `make build`
- `./fogelbot`
- `make test`
- `make test-race`
- `docker compose up --build -d`

Tests do not require Discord credentials or a `.env` file.

## Codebase Map

- `main.go`: process entrypoint
- `bot/`: Discord session setup, event handlers, dependency wiring
- `config/`: environment loading and shared timing/sentiment constants
- `triggers/`: trigger interfaces, registry, DSL, and trigger implementations
- `state/`: cooldown tracking
- `sentiment/`: sentiment analysis wrapper
- `language/`: language detection wrapper
- `facts/`: optional fact buffering and persistence
- `llm/`: optional LLM integration

## How The Bot Works

Start with this path when orienting yourself:

`main.go` -> `config.Load()` -> `bot.New()` -> `bot.OnMessage()` -> `triggers.RegisterAll()` / `Registry.Process()`

Important behavioral rules:

- Trigger execution is first-match-wins.
- Trigger registration order lives in `triggers/definitions.go`.
- Some triggers are code-driven; simpler ones use the DSL builder in `triggers/builder.go`.
- Cooldowns live in `state/CooldownManager`.
- The bot ignores its own messages in `bot/handlers.go`.

## How To Work On It Effectively

- If you need to change when the bot replies, inspect `triggers/definitions.go` first.
- If you add or adjust a DSL trigger, remember the evaluation order is match -> user -> sentiment -> probability -> cooldown -> response.
- If you add a new trigger type, prefer keeping it isolated in `triggers/` and registering it in one place.
- If you change trigger order, treat that as a behavior change because earlier triggers can suppress later ones.
- If you touch optional LLM behavior, keep failure modes non-fatal; the bot is designed to keep running when LLM calls fail.

## Testing Guidance

Most behavior can be tested without Discord or external services.

- Trigger tests use hand-written fakes in `triggers/testhelpers_test.go`.
- The test context builder is the quickest way to exercise trigger behavior.
- When testing registry behavior, control random rolls explicitly so probability-gated triggers do not fire accidentally.
- Prefer focused unit tests around trigger precedence, cooldowns, and probability thresholds when changing reply behavior.

## Practical Notes For Future Agents

- Do not spend time on leftover Python files unless the user explicitly asks about the legacy implementation.
- Prefer reading Go files over legacy docs if they disagree.
- For behavior changes, verify both the trigger definition and its tests; several tests depend on deterministic random values and registration order.
