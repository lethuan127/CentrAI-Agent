package agent

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/lethuan127/centrai-agent/internal/model"
	"github.com/lethuan127/centrai-agent/internal/session"
	"github.com/lethuan127/centrai-agent/internal/store/memory"
	"github.com/lethuan127/centrai-agent/internal/tool"
)

type fakeModel struct {
	rounds []model.ChatResult
	i      int
}

func (f *fakeModel) StreamChat(ctx context.Context, req *model.ChatRequest, onChunk func(model.StreamChunk) error) (*model.ChatResult, error) {
	if f.i >= len(f.rounds) {
		return nil, errors.New("fakeModel: no more rounds")
	}
	r := f.rounds[f.i]
	f.i++
	if onChunk != nil {
		_ = onChunk(model.StreamChunk{DeltaText: "…"})
	}
	return &r, nil
}

func TestRunnerNoTool(t *testing.T) {
	st := memory.New()
	reg := tool.NewRegistry()
	m := &fakeModel{rounds: []model.ChatResult{{
		Message: session.Message{Role: session.RoleAssistant, Content: "done"},
	}}}
	r := NewRunner(st, m, reg, Options{MaxSteps: 4})

	out, err := r.Run(context.Background(), &RunInput{
		SessionID:   "s1",
		UserMessage: "hello",
		System:      "sys",
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.Assistant != "done" {
		t.Fatalf("assistant %q", out.Assistant)
	}
	if len(out.Messages) != 2 {
		t.Fatalf("messages %d", len(out.Messages))
	}
}

func TestRunnerToolThenText(t *testing.T) {
	st := memory.New()
	reg := tool.NewRegistry()
	schema := json.RawMessage(`{"type":"object","properties":{"n":{"type":"integer"}}}`)
	if err := reg.Register(tool.Definition{
		Name:        "add_one",
		Description: "adds one",
		Parameters:  schema,
	}, func(ctx context.Context, args json.RawMessage) (string, error) {
		var v struct {
			N int `json:"n"`
		}
		if err := json.Unmarshal(args, &v); err != nil {
			return "", err
		}
		b, err := json.Marshal(map[string]int{"result": v.N + 1})
		if err != nil {
			return "", err
		}
		return string(b), nil
	}); err != nil {
		t.Fatal(err)
	}

	m := &fakeModel{rounds: []model.ChatResult{
		{Message: session.Message{
			Role: session.RoleAssistant,
			ToolCalls: []session.ToolCall{{
				ID: "c1", Name: "add_one", Arguments: `{"n":1}`,
			}},
		}},
		{Message: session.Message{Role: session.RoleAssistant, Content: "ok"}},
	}}

	r := NewRunner(st, m, reg, Options{MaxSteps: 4})
	out, err := r.Run(context.Background(), &RunInput{SessionID: "t1", UserMessage: "run"})
	if err != nil {
		t.Fatal(err)
	}
	if out.Assistant != "ok" {
		t.Fatalf("assistant %q", out.Assistant)
	}
	if len(out.Messages) != 4 {
		// user, assistant+tool_calls, tool, assistant
		t.Fatalf("want 4 messages, got %d: %+v", len(out.Messages), out.Messages)
	}
}

func TestRunnerMaxSteps(t *testing.T) {
	st := memory.New()
	reg := tool.NewRegistry()
	schema := json.RawMessage(`{"type":"object","properties":{"n":{"type":"integer"}}}`)
	if err := reg.Register(tool.Definition{Name: "add_one", Parameters: schema}, func(ctx context.Context, args json.RawMessage) (string, error) {
		return `{"result":2}`, nil
	}); err != nil {
		t.Fatal(err)
	}
	m := &fakeModel{rounds: []model.ChatResult{
		{Message: session.Message{
			Role: session.RoleAssistant,
			ToolCalls: []session.ToolCall{{
				ID: "c1", Name: "add_one", Arguments: `{"n":1}`,
			}},
		}},
	}}
	r := NewRunner(st, m, reg, Options{MaxSteps: 1})
	_, err := r.Run(context.Background(), &RunInput{SessionID: "x", UserMessage: "go", MaxSteps: 1})
	if !errors.Is(err, ErrMaxSteps) {
		t.Fatalf("want ErrMaxSteps, got %v", err)
	}
}
