package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/lethuan127/centrai-agent/internal/agentdef"
	"github.com/lethuan127/centrai-agent/internal/tool"
)

// registerDemoTools adds small, safe tools for manual testing of the tool loop.
func registerDemoTools(reg *tool.Registry) error {
	echoSchema := json.RawMessage(`{
		"type": "object",
		"properties": {
			"message": {
				"type": "string",
				"description": "Exact text to return to the user"
			}
		},
		"required": ["message"],
		"additionalProperties": false
	}`)
	if err := reg.Register(tool.Definition{
		Name:        "echo",
		Description: "Returns the given message back to the caller. Use when the user wants text repeated or verified.",
		Parameters:  echoSchema,
	}, func(ctx context.Context, args json.RawMessage) (string, error) {
		_ = ctx
		var in struct {
			Message string `json:"message"`
		}
		if err := json.Unmarshal(args, &in); err != nil {
			return "", err
		}
		out, _ := json.Marshal(map[string]string{"echoed": in.Message})
		return string(out), nil
	}); err != nil {
		return err
	}

	addSchema := json.RawMessage(`{
		"type": "object",
		"properties": {
			"a": { "type": "number", "description": "First summand" },
			"b": { "type": "number", "description": "Second summand" }
		},
		"required": ["a", "b"],
		"additionalProperties": false
	}`)
	return reg.Register(tool.Definition{
		Name:        "add",
		Description: "Adds two numbers and returns the sum as JSON.",
		Parameters:  addSchema,
	}, func(ctx context.Context, args json.RawMessage) (string, error) {
		_ = ctx
		var in struct {
			A float64 `json:"a"`
			B float64 `json:"b"`
		}
		if err := json.Unmarshal(args, &in); err != nil {
			return "", err
		}
		sum := in.A + in.B
		return fmt.Sprintf(`{"sum":%g}`, sum), nil
	})
}

const demoToolsInstructionsAppendix = `You have native tools:
- echo: repeat or echo a message the user cares about.
- add: add two numbers when the user asks for arithmetic.

Call a tool when it is the right way to answer; otherwise reply normally.`

func systemPrompt(demoTools bool) string {
	s := "You are a helpful assistant."
	if demoTools {
		s += "\n\n" + demoToolsInstructionsAppendix
	}
	return s
}

// buildSystemPrompt chooses the system message: YAML agent definition wins when -agent is set.
func buildSystemPrompt(def *agentdef.Definition, demoTools bool) string {
	if def == nil {
		return systemPrompt(demoTools)
	}
	var parts []string
	if s := strings.TrimSpace(def.Description); s != "" {
		parts = append(parts, s)
	}
	if s := strings.TrimSpace(def.Instructions); s != "" {
		parts = append(parts, s)
	}
	base := strings.Join(parts, "\n\n")
	if base == "" {
		base = "You are a helpful assistant."
	}
	if meta := def.LLMMetaAppendix(); meta != "" {
		base += "\n\n" + meta
	}
	if demoTools {
		base += "\n\n" + demoToolsInstructionsAppendix
	}
	return base
}
