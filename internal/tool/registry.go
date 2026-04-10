// Package tool implements native tool registration, JSON Schema validation, and execution.
package tool

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/lethuan127/centrai-agent/internal/session"
)

// Handler runs a validated tool invocation and returns a string result for the model.
type Handler func(ctx context.Context, args json.RawMessage) (string, error)

// Definition describes a tool for the model provider.
type Definition struct {
	Name        string
	Description string
	// Parameters is a JSON Schema object for the arguments object (e.g. {"type":"object",...}).
	Parameters json.RawMessage
}

// RegisteredTool pairs schema metadata with a handler.
type RegisteredTool struct {
	Def     Definition
	Schema  *jsonschema.Schema
	Handler Handler
}

var (
	// ErrUnknownTool is returned when the model requests a name that was not registered.
	ErrUnknownTool = errors.New("tool: unknown tool name")
	// ErrValidation is returned when arguments do not match the tool schema.
	ErrValidation = errors.New("tool: argument validation failed")
)

// Registry holds named tools and dispatches execution.
type Registry struct {
	mu    sync.RWMutex
	tools map[string]*RegisteredTool
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry {
	return &Registry{tools: make(map[string]*RegisteredTool)}
}

// Register adds a tool. The name must be unique; Parameters must be a valid JSON Schema document.
func (r *Registry) Register(def Definition, h Handler) error {
	if def.Name == "" {
		return errors.New("tool: empty name")
	}
	var doc any
	if err := json.Unmarshal(def.Parameters, &doc); err != nil {
		return fmt.Errorf("tool: parse schema JSON for %q: %w", def.Name, err)
	}
	c := jsonschema.NewCompiler()
	loc := "https://centrai.invalid/schema/" + def.Name + ".json"
	if err := c.AddResource(loc, doc); err != nil {
		return fmt.Errorf("tool: add schema resource for %q: %w", def.Name, err)
	}
	sch, err := c.Compile(loc)
	if err != nil {
		return fmt.Errorf("tool: compile schema for %q: %w", def.Name, err)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tools[def.Name] = &RegisteredTool{
		Def:     def,
		Schema:  sch,
		Handler: h,
	}
	return nil
}

// Definitions returns provider-ready tool metadata (OpenAI-style function tools).
// Tool order is sorted by name for deterministic prompts.
func (r *Registry) Definitions() []Definition {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.tools))
	for n := range r.tools {
		names = append(names, n)
	}
	sort.Strings(names)
	out := make([]Definition, 0, len(names))
	for _, n := range names {
		out = append(out, r.tools[n].Def)
	}
	return out
}

// Execute validates arguments, applies timeout, and runs the handler for name.
func (r *Registry) Execute(ctx context.Context, name string, argsJSON string, timeout time.Duration) (string, error) {
	r.mu.RLock()
	t, ok := r.tools[name]
	r.mu.RUnlock()
	if !ok {
		return "", fmt.Errorf("%w: %q", ErrUnknownTool, name)
	}

	var instance any
	if err := json.Unmarshal([]byte(argsJSON), &instance); err != nil {
		return "", fmt.Errorf("%w: invalid JSON arguments: %w", ErrValidation, err)
	}
	if err := t.Schema.Validate(instance); err != nil {
		return "", fmt.Errorf("%w: %w", ErrValidation, err)
	}
	raw := json.RawMessage(argsJSON)

	runCtx := ctx
	var cancel context.CancelFunc
	if timeout > 0 {
		runCtx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	return t.Handler(runCtx, raw)
}

// OpenAIFunctionTools builds the wire payload fragment used by OpenAI-compatible chat APIs.
func OpenAIFunctionTools(defs []Definition) []map[string]any {
	out := make([]map[string]any, 0, len(defs))
	for _, d := range defs {
		var params any
		if err := json.Unmarshal(d.Parameters, &params); err != nil {
			params = map[string]any{"type": "object"}
		}
		out = append(out, map[string]any{
			"type": "function",
			"function": map[string]any{
				"name":        d.Name,
				"description": d.Description,
				"parameters":  params,
			},
		})
	}
	return out
}

// ToolMessagesFromResults builds session tool messages in call order.
func ToolMessagesFromResults(toolCallIDs []string, results []string) []session.Message {
	msgs := make([]session.Message, len(toolCallIDs))
	for i := range toolCallIDs {
		msgs[i] = session.Message{
			Role:       session.RoleTool,
			Content:    results[i],
			ToolCallID: toolCallIDs[i],
		}
	}
	return msgs
}
