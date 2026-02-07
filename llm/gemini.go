package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const defaultBaseURL = "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash-lite:generateContent"

// Gemini implements LLM using the Google AI Studio API.
type Gemini struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewGemini creates a Gemini client with the given API key.
func NewGemini(apiKey string) *Gemini {
	return &Gemini{
		apiKey:  apiKey,
		baseURL: defaultBaseURL,
		client:  http.DefaultClient,
	}
}

// gemini API request/response types

type geminiRequest struct {
	Contents         []geminiContent  `json:"contents"`
	GenerationConfig generationConfig `json:"generationConfig"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type generationConfig struct {
	ResponseMIMEType string          `json:"responseMimeType"`
	ResponseSchema   json.RawMessage `json:"responseSchema"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

// JSON response schemas for structured outputs
type factsResponse struct {
	Facts []string `json:"facts"`
}

type generateResponseBody struct {
	Response string `json:"response"`
}

var factsSchema = json.RawMessage(`{
	"type": "object",
	"properties": {
		"facts": {
			"type": "array",
			"items": {"type": "string"}
		}
	},
	"required": ["facts"]
}`)

var responseSchema = json.RawMessage(`{
	"type": "object",
	"properties": {
		"response": {"type": "string"}
	},
	"required": ["response"]
}`)

func (g *Gemini) call(ctx context.Context, prompt string, schema json.RawMessage) (string, error) {
	reqBody := geminiRequest{
		Contents: []geminiContent{
			{Parts: []geminiPart{{Text: prompt}}},
		},
		GenerationConfig: generationConfig{
			ResponseMIMEType: "application/json",
			ResponseSchema:   schema,
		},
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.baseURL+"?key="+g.apiKey, bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var gemResp geminiResponse
	if err := json.Unmarshal(body, &gemResp); err != nil {
		return "", fmt.Errorf("unmarshal response: %w", err)
	}

	if len(gemResp.Candidates) == 0 || len(gemResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("empty response from API")
	}

	return gemResp.Candidates[0].Content.Parts[0].Text, nil
}

func (g *Gemini) ExtractFacts(ctx context.Context, username string, messages []string) ([]string, error) {
	prompt := fmt.Sprintf(
		"Given the following messages sent by user %q, infer some facts about them. "+
			"Return only new, non-obvious facts. Messages:\n%s",
		username, strings.Join(messages, "\n"),
	)

	text, err := g.call(ctx, prompt, factsSchema)
	if err != nil {
		return nil, err
	}

	var result factsResponse
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		return nil, fmt.Errorf("parse facts response: %w", err)
	}
	return result.Facts, nil
}

func (g *Gemini) CompactFacts(ctx context.Context, username string, facts []string) ([]string, error) {
	prompt := fmt.Sprintf(
		"Here are facts about user %q:\n%s\n\n"+
			"Compact these down to the 7 most salient facts, merging related ones. "+
			"Return the result as a JSON object with a \"facts\" array.",
		username, strings.Join(facts, "\n"),
	)

	text, err := g.call(ctx, prompt, factsSchema)
	if err != nil {
		return nil, err
	}

	var result factsResponse
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		return nil, fmt.Errorf("parse compact response: %w", err)
	}
	return result.Facts, nil
}

func (g *Gemini) GenerateResponse(ctx context.Context, username string, userMessage string, userFacts []string, exampleQuotes []string) (string, error) {
	factsSection := "No known facts."
	if len(userFacts) > 0 {
		factsSection = strings.Join(userFacts, "\n")
	}

	quotesSection := strings.Join(exampleQuotes, "\n")

	prompt := fmt.Sprintf(
		"You are Fogelbot, an annoying, dismissive, and obnoxious Discord bot that is humorous. "+
			"Some examples of messages you have sent in the past:\n%s\n\n"+
			"Here are some facts you know about the user %q:\n%s\n\n"+
			"The user just sent this message: %q\n\n"+
			"Respond to it as Fogelbot. Keep it short (1-2 sentences). Be dismissive and funny. "+
			"Use the facts about the user to make your response more personal when relevant.",
		quotesSection, username, factsSection, userMessage,
	)

	text, err := g.call(ctx, prompt, responseSchema)
	if err != nil {
		return "", err
	}

	var result generateResponseBody
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		return "", fmt.Errorf("parse generate response: %w", err)
	}
	return result.Response, nil
}
