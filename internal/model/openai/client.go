package openai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	openaisdk "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/shared"

	"github.com/lethuan127/centrai-agent/internal/model"
	"github.com/lethuan127/centrai-agent/internal/session"
)

// Config holds settings for the official OpenAI Go SDK (github.com/openai/openai-go/v3).
type Config struct {
	BaseURL    string // e.g. https://api.openai.com/v1
	APIKey     string // optional if OPENAI_API_KEY is set (see openai.NewClient defaults)
	HTTPClient *http.Client
	Model      string
}

// Client implements model.Client using openai-go streaming chat completions.
type Client struct {
	cfg Config
	sdk openaisdk.Client
}

// New returns a Client. BaseURL should be the API root including /v1 for api.openai.com.
func New(cfg Config) *Client {
	opts := make([]option.RequestOption, 0, 4)
	if u := strings.TrimSpace(cfg.BaseURL); u != "" {
		opts = append(opts, option.WithBaseURL(strings.TrimRight(u, "/")))
	}
	if cfg.APIKey != "" {
		opts = append(opts, option.WithAPIKey(cfg.APIKey))
	}
	if cfg.HTTPClient != nil {
		opts = append(opts, option.WithHTTPClient(cfg.HTTPClient))
	}
	sdk := openaisdk.NewClient(opts...)
	return &Client{cfg: cfg, sdk: sdk}
}

// StreamChat implements model.Client via Chat.Completions.NewStreaming.
func (c *Client) StreamChat(ctx context.Context, req *model.ChatRequest, onChunk func(model.StreamChunk) error) (*model.ChatResult, error) {
	if req == nil {
		return nil, errors.New("openai: nil request")
	}
	modelName := c.cfg.Model
	if req.Model != "" {
		modelName = req.Model
	}
	if modelName == "" {
		return nil, errors.New("openai: empty model name")
	}

	rawMsgs, err := json.Marshal(messagesToOpenAI(req.Messages))
	if err != nil {
		return nil, err
	}
	var messages []openaisdk.ChatCompletionMessageParamUnion
	if err := json.Unmarshal(rawMsgs, &messages); err != nil {
		return nil, fmt.Errorf("openai: decode messages: %w", err)
	}

	params := openaisdk.ChatCompletionNewParams{
		Messages: messages,
		Model:    shared.ChatModel(modelName),
	}
	if len(req.Tools) > 0 {
		rawTools, err := json.Marshal(req.Tools)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(rawTools, &params.Tools); err != nil {
			return nil, fmt.Errorf("openai: decode tools: %w", err)
		}
	}

	stream := c.sdk.Chat.Completions.NewStreaming(ctx, params)
	defer stream.Close()

	var acc openaisdk.ChatCompletionAccumulator
	for stream.Next() {
		chunk := stream.Current()
		if onChunk != nil && len(chunk.Choices) > 0 {
			d := chunk.Choices[0].Delta.Content
			if d != "" {
				if err := onChunk(model.StreamChunk{DeltaText: d}); err != nil {
					return nil, err
				}
			}
		}
		if !acc.AddChunk(chunk) {
			return nil, errors.New("openai: stream chunk out of sequence")
		}
	}
	if err := stream.Err(); err != nil {
		return nil, err
	}
	if len(acc.Choices) == 0 {
		return nil, errors.New("openai: empty completion")
	}

	out := completionMessageToSession(acc.Choices[0].Message)
	return &model.ChatResult{Message: out}, nil
}

func completionMessageToSession(m openaisdk.ChatCompletionMessage) session.Message {
	out := session.Message{Role: session.RoleAssistant, Content: m.Content}
	for _, tc := range m.ToolCalls {
		// Prefer flattened fields (filled by ChatCompletionAccumulator); AsFunction()
		// relies on RawJSON and can be empty for accumulated chunks.
		id := tc.ID
		name := tc.Function.Name
		args := tc.Function.Arguments
		if f := tc.AsFunction(); id == "" && f.ID != "" {
			id, name, args = f.ID, f.Function.Name, f.Function.Arguments
		}
		if id == "" || name == "" {
			continue
		}
		if args == "" {
			args = "{}"
		}
		out.ToolCalls = append(out.ToolCalls, session.ToolCall{
			ID:        id,
			Name:      name,
			Arguments: args,
		})
	}
	return out
}
