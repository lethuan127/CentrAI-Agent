// Package responsesapi implements [model.Client] using OpenAI's Responses API with streaming only.
package responsesapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	openaisdk "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"

	"github.com/lethuan127/centrai-agent/internal/model"
)

// Config configures the OpenAI Responses API client (same transport shape as chat completions).
type Config struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
	Model      string
}

// Client implements [model.Client] via Responses API streaming.
type Client struct {
	cfg Config
	sdk openaisdk.Client
}

// New returns a Client.
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

// StreamChat implements [model.Client].
func (c *Client) StreamChat(ctx context.Context, req *model.ChatRequest, onChunk func(model.StreamChunk) error) (*model.ChatResult, error) {
	if req == nil {
		return nil, errors.New("responsesapi: nil request")
	}
	modelName := c.cfg.Model
	if req.Model != "" {
		modelName = req.Model
	}
	if modelName == "" {
		return nil, errors.New("responsesapi: empty model name")
	}

	params, err := chatRequestToParams(req, modelName)
	if err != nil {
		return nil, err
	}

	stream := c.sdk.Responses.NewStreaming(ctx, params)
	defer stream.Close()

	var completed *responses.Response
	for stream.Next() {
		ev := stream.Current()
		switch ev.Type {
		case "response.output_text.delta":
			d := ev.AsResponseOutputTextDelta()
			if onChunk != nil && d.Delta != "" {
				if err := onChunk(model.StreamChunk{DeltaText: d.Delta}); err != nil {
					return nil, err
				}
			}
		case "response.completed":
			completed = &ev.Response
		case "error":
			er := ev.AsError()
			return nil, fmt.Errorf("responsesapi: stream error: %s", er.Message)
		}
	}
	if err := stream.Err(); err != nil {
		return nil, err
	}
	if completed == nil {
		return nil, errors.New("responsesapi: stream ended without completion event")
	}
	msg, err := responseToAssistant(*completed)
	if err != nil {
		return nil, err
	}
	return &model.ChatResult{Message: msg}, nil
}
