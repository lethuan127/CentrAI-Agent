package openai

import (
	"github.com/lethuan127/centrai-agent/internal/session"
)

// messagesToOpenAI converts session messages to OpenAI chat API message objects.
func messagesToOpenAI(msgs []session.Message) []map[string]any {
	out := make([]map[string]any, 0, len(msgs))
	for _, m := range msgs {
		switch m.Role {
		case session.RoleSystem, session.RoleUser:
			out = append(out, map[string]any{
				"role":    string(m.Role),
				"content": m.Content,
			})
		case session.RoleAssistant:
			om := map[string]any{
				"role":    "assistant",
				"content": nil,
			}
			if m.Content != "" {
				om["content"] = m.Content
			}
			if len(m.ToolCalls) > 0 {
				tcs := make([]map[string]any, len(m.ToolCalls))
				for i, tc := range m.ToolCalls {
					tcs[i] = map[string]any{
						"id":   tc.ID,
						"type": "function",
						"function": map[string]any{
							"name":      tc.Name,
							"arguments": tc.Arguments,
						},
					}
				}
				om["tool_calls"] = tcs
			}
			out = append(out, om)
		case session.RoleTool:
			out = append(out, map[string]any{
				"role":         "tool",
				"tool_call_id": m.ToolCallID,
				"content":      m.Content,
			})
		}
	}
	return out
}
