# Fogelbot

A Go Discord bot ported from Python. Responds to messages with personality-driven quotes and reactions based on keywords, sentiment analysis, language detection, user identity, and random chance.

## Commands

- `go build ./...` — build
- `go test ./...` — run all tests (no `.env` or Discord token needed)
- `go test -race ./...` — run with race detector

## Architecture

The core abstraction is the `Trigger` interface in `triggers/triggers.go`. Triggers are registered in a `Registry` in priority order, and `Registry.Process()` returns the first match (first-match-wins).

There are two kinds of triggers. **Code-driven triggers** implement `Trigger` directly for complex stateful logic. **DSL triggers** use a fluent builder for simple data-driven definitions:

```go
r.On("Pizza").Match(`\bpizza\b`).Probability(0.10).Respond("Pizza time!")
```

All trigger registration lives in `triggers/definitions.go` via `RegisterAll()`, which reads top-to-bottom in priority order. To add a new trigger, add one line there. DSL conditions evaluate in order: Match -> User -> Sentiment -> Probability -> Cooldown -> Response. `%s` in responses is replaced with the author's @mention.

All external dependencies (`SentimentAnalyzer`, `LanguageDetector`, `Rand`, `Clock`) are interfaces on `Context`, so tests run without a Discord connection or `.env` file. `CooldownManager` also takes a `Clock` for testable time.

Message flow: Discord message -> `bot.OnMessage` -> build `Context` with all dependencies -> `Registry.Process(ctx)` -> send first matching response.

## Testing

Tests use hand-written fakes (no external test frameworks) in `triggers/testhelpers_test.go`. Context builder pattern: `newTestContext("msg").WithSentiment(0.5).WithRand(...).Build()`.

When testing with the full registry, use high float values (0.99) for `newAlwaysRand` to prevent probability-gated triggers from accidentally firing. User-gated DSL triggers don't consume a `Float64` call if the user check fails first.
