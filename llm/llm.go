package llm

import "context"

// LLM abstracts language model operations for fact extraction and response generation.
type LLM interface {
	ExtractFacts(ctx context.Context, username string, messages []string) ([]string, error)
	CompactFacts(ctx context.Context, username string, facts []string) ([]string, error)
	GenerateResponse(ctx context.Context, username string, userMessage string, userFacts []string, exampleQuotes []string) (string, error)
}
