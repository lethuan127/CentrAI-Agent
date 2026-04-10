package responsesapi

import (
	"fmt"
	"strings"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
	"github.com/openai/openai-go/v3/shared"

	"github.com/lethuan127/centrai-agent/internal/model"
	"github.com/lethuan127/centrai-agent/internal/session"
)

func chatRequestToParams(req *model.ChatRequest, defaultModel string) (responses.ResponseNewParams, error) {
	msgs := req.Messages
	var p responses.ResponseNewParams
	p.Model = shared.ResponsesModel(defaultModel)

	start := 0
	if len(msgs) > 0 && msgs[0].Role == session.RoleSystem {
		p.Instructions = openai.String(msgs[0].Content)
		start = 1
	}

	items, err := sessionMessagesToInput(msgs[start:])
	if err != nil {
		return p, err
	}
	p.Input = responses.ResponseNewParamsInputUnion{OfInputItemList: items}

	if len(req.Tools) > 0 {
		tools, err := toolsFromChatMaps(req.Tools)
		if err != nil {
			return p, err
		}
		p.Tools = tools
	}

	p.Store = openai.Bool(false)
	return p, nil
}

func sessionMessagesToInput(msgs []session.Message) (responses.ResponseInputParam, error) {
	var out responses.ResponseInputParam
	for _, m := range msgs {
		switch m.Role {
		case session.RoleUser:
			out = append(out, responses.ResponseInputItemParamOfMessage(m.Content, responses.EasyInputMessageRoleUser))
		case session.RoleAssistant:
			if strings.TrimSpace(m.Content) != "" {
				out = append(out, responses.ResponseInputItemParamOfMessage(m.Content, responses.EasyInputMessageRoleAssistant))
			}
			for _, tc := range m.ToolCalls {
				args := tc.Arguments
				if args == "" {
					args = "{}"
				}
				out = append(out, responses.ResponseInputItemParamOfFunctionCall(args, tc.ID, tc.Name))
			}
		case session.RoleTool:
			out = append(out, responses.ResponseInputItemParamOfFunctionCallOutput(m.ToolCallID, m.Content))
		default:
			return nil, fmt.Errorf("responsesapi: unsupported role %q", m.Role)
		}
	}
	return out, nil
}

func toolsFromChatMaps(tools []map[string]any) ([]responses.ToolUnionParam, error) {
	out := make([]responses.ToolUnionParam, 0, len(tools))
	for _, t := range tools {
		if t["type"] != "function" {
			continue
		}
		fn, _ := t["function"].(map[string]any)
		if fn == nil {
			continue
		}
		name, _ := fn["name"].(string)
		if name == "" {
			return nil, fmt.Errorf("responsesapi: tool missing name")
		}
		desc, _ := fn["description"].(string)
		params, _ := fn["parameters"].(map[string]any)
		if params == nil {
			params = map[string]any{"type": "object"}
		}
		out = append(out, responses.ToolUnionParam{
			OfFunction: &responses.FunctionToolParam{
				Name:        name,
				Parameters:  params,
				Strict:      openai.Bool(true),
				Description: openai.String(desc),
			},
		})
	}
	return out, nil
}

func responseToAssistant(r responses.Response) (session.Message, error) {
	var b strings.Builder
	var calls []session.ToolCall
	for _, item := range r.Output {
		switch v := item.AsAny().(type) {
		case responses.ResponseOutputMessage:
			for _, c := range v.Content {
				if strings.TrimSpace(c.Type) == "output_text" {
					b.WriteString(c.Text)
				}
			}
		case responses.ResponseFunctionToolCall:
			args := v.Arguments
			if args == "" {
				args = "{}"
			}
			id := v.CallID
			if id == "" {
				id = v.ID
			}
			calls = append(calls, session.ToolCall{
				ID:        id,
				Name:      v.Name,
				Arguments: args,
			})
		default:
			// ignore other output kinds (reasoning, built-in tools, etc.)
		}
	}
	return session.Message{
		Role:      session.RoleAssistant,
		Content:   b.String(),
		ToolCalls: calls,
	}, nil
}
