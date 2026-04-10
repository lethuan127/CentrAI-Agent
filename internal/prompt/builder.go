// Package prompt assembles provider-ready payloads from instructions and session history.
package prompt

import (
	"github.com/lethuan127/centrai-agent/internal/model"
	"github.com/lethuan127/centrai-agent/internal/session"
	"github.com/lethuan127/centrai-agent/internal/tool"
)

// Builder merges system instructions, history, and tool definitions.
type Builder struct{}

// NewBuilder returns a Builder.
func NewBuilder() *Builder {
	return &Builder{}
}

// Build constructs a model.ChatRequest with system message first, then session messages.
func (b *Builder) Build(system string, history []session.Message, defs []tool.Definition, modelName string) *model.ChatRequest {
	msgs := make([]session.Message, 0, len(history)+1)
	if system != "" {
		msgs = append(msgs, session.Message{Role: session.RoleSystem, Content: system})
	}
	msgs = append(msgs, history...)

	return &model.ChatRequest{
		Model:    modelName,
		Messages: msgs,
		Tools:    tool.OpenAIFunctionTools(defs),
	}
}
