// Package mcp bridges Model Context Protocol tool discovery and execution into [tool.Registry].
package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/lethuan127/centrai-agent/internal/tool"
)

// RegisterRemoteTools lists tools from an MCP session and registers them on reg with an optional name prefix.
// Collisions with existing tool names return an error.
func RegisterRemoteTools(ctx context.Context, session *mcpsdk.ClientSession, reg *tool.Registry, namePrefix string) error {
	if session == nil || reg == nil {
		return errors.New("mcp: nil session or registry")
	}
	res, err := session.ListTools(ctx, &mcpsdk.ListToolsParams{})
	if err != nil {
		return fmt.Errorf("mcp: list tools: %w", err)
	}
	prefix := strings.TrimSpace(namePrefix)
	for _, t := range res.Tools {
		localName := prefix + t.Name
		params, err := inputSchemaJSON(t.InputSchema)
		if err != nil {
			return fmt.Errorf("mcp: tool %q schema: %w", t.Name, err)
		}
		def := tool.Definition{
			Name:        localName,
			Description: t.Description,
			Parameters:  params,
		}
		remote := t.Name
		h := func(callCtx context.Context, args json.RawMessage) (string, error) {
			var argObj any
			if err := json.Unmarshal(args, &argObj); err != nil {
				return "", err
			}
			out, err := session.CallTool(callCtx, &mcpsdk.CallToolParams{
				Name:      remote,
				Arguments: argObj,
			})
			if err != nil {
				return "", err
			}
			return callToolResultString(out), nil
		}
		if err := reg.Register(def, h); err != nil {
			return fmt.Errorf("mcp: register %q: %w", localName, err)
		}
	}
	return nil
}

func inputSchemaJSON(schema any) (json.RawMessage, error) {
	if schema == nil {
		return json.RawMessage(`{"type":"object"}`), nil
	}
	switch s := schema.(type) {
	case json.RawMessage:
		if len(s) == 0 {
			return json.RawMessage(`{"type":"object"}`), nil
		}
		return s, nil
	case map[string]any:
		return json.Marshal(s)
	default:
		b, err := json.Marshal(schema)
		if err != nil {
			return nil, err
		}
		return b, nil
	}
}

func callToolResultString(res *mcpsdk.CallToolResult) string {
	if res == nil {
		return ""
	}
	if res.IsError {
		var parts []string
		for _, c := range res.Content {
			if tc, ok := c.(*mcpsdk.TextContent); ok {
				parts = append(parts, tc.Text)
			}
		}
		if len(parts) > 0 {
			return strings.Join(parts, "\n")
		}
		return "tool error"
	}
	var parts []string
	for _, c := range res.Content {
		switch v := c.(type) {
		case *mcpsdk.TextContent:
			parts = append(parts, v.Text)
		default:
			b, _ := json.Marshal(c)
			parts = append(parts, string(b))
		}
	}
	if res.StructuredContent != nil {
		b, _ := json.Marshal(res.StructuredContent)
		parts = append(parts, string(b))
	}
	return strings.Join(parts, "\n")
}
