package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestGemini(url string) *Gemini {
	return &Gemini{
		apiKey:  "test-key",
		baseURL: url,
		client:  http.DefaultClient,
	}
}

func geminiJSONResponse(text string) string {
	resp := geminiResponse{
		Candidates: []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		}{
			{Content: struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			}{
				Parts: []struct {
					Text string `json:"text"`
				}{{Text: text}},
			}},
		},
	}
	data, _ := json.Marshal(resp)
	return string(data)
}

func TestExtractFacts(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("key") != "test-key" {
			t.Errorf("expected API key 'test-key', got %q", r.URL.Query().Get("key"))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(geminiJSONResponse(`{"facts": ["likes pizza", "lives in NYC"]}`)))
	}))
	defer srv.Close()

	g := newTestGemini(srv.URL)
	facts, err := g.ExtractFacts(context.Background(), "alice", []string{"I love pizza", "NYC is great"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(facts) != 2 {
		t.Fatalf("expected 2 facts, got %d", len(facts))
	}
	if facts[0] != "likes pizza" || facts[1] != "lives in NYC" {
		t.Errorf("unexpected facts: %v", facts)
	}
}

func TestCompactFacts(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(geminiJSONResponse(`{"facts": ["pizza lover in NYC"]}`)))
	}))
	defer srv.Close()

	g := newTestGemini(srv.URL)
	facts, err := g.CompactFacts(context.Background(), "alice", []string{"likes pizza", "lives in NYC", "eats pizza often"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(facts) != 1 || facts[0] != "pizza lover in NYC" {
		t.Errorf("unexpected facts: %v", facts)
	}
}

func TestGenerateResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(geminiJSONResponse(`{"response": "Nobody asked, alice."}`)))
	}))
	defer srv.Close()

	g := newTestGemini(srv.URL)
	resp, err := g.GenerateResponse(context.Background(), "alice", "hello", []string{"likes pizza"}, []string{"Rough year for Kanye"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "Nobody asked, alice." {
		t.Errorf("unexpected response: %q", resp)
	}
}

func TestAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal"}`))
	}))
	defer srv.Close()

	g := newTestGemini(srv.URL)
	_, err := g.ExtractFacts(context.Background(), "alice", []string{"hello"})
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestMalformedJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(geminiJSONResponse(`not json`)))
	}))
	defer srv.Close()

	g := newTestGemini(srv.URL)
	_, err := g.ExtractFacts(context.Background(), "alice", []string{"hello"})
	if err == nil {
		t.Fatal("expected error for malformed JSON")
	}
}

func TestEmptyResponse(t *testing.T) {
	resp := geminiResponse{Candidates: nil}
	data, _ := json.Marshal(resp)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}))
	defer srv.Close()

	g := newTestGemini(srv.URL)
	_, err := g.ExtractFacts(context.Background(), "alice", []string{"hello"})
	if err == nil {
		t.Fatal("expected error for empty response")
	}
}

func TestGenerateResponseNoFacts(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(geminiJSONResponse(`{"response": "Whatever."}`)))
	}))
	defer srv.Close()

	g := newTestGemini(srv.URL)
	resp, err := g.GenerateResponse(context.Background(), "bob", "hi", nil, []string{"quote1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "Whatever." {
		t.Errorf("unexpected response: %q", resp)
	}
}
