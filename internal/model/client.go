// Package model defines the streaming LLM client boundary and provider adapters.
package model

import (
	"context"

	"github.com/lethuan127/centrai-agent/internal/session"
)

// StreamChunk is an incremental piece of assistant output during streaming.
type StreamChunk struct {
	DeltaText string
}

// ChatRequest is a single model round (chat + optional tools).
type ChatRequest struct {
	Model    string
	Messages []session.Message
	// Tools is the OpenAI-compatible "tools" array (already JSON-ready).
	Tools []map[string]any
}

// ChatResult is the final assistant step after the stream completes.
type ChatResult struct {
	Message session.Message
}

// Client streams chat completions and returns a final assistant message per round.
type Client interface {
	StreamChat(ctx context.Context, req *ChatRequest, onChunk func(StreamChunk) error) (*ChatResult, error)
}
