package tool

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
)

// Middleware wraps a tool [Handler] to add cross-cutting behavior (logging, redaction, metrics).
// The first middleware in [Registry.Use] is the outermost layer (runs first on the way in).
type Middleware func(next Handler) Handler

// Policy runs after the tool name is resolved and arguments are valid JSON, before the handler.
// Return a non-nil error to block the invocation (e.g. blocklisted tool or argument pattern).
type Policy func(ctx context.Context, name string, args json.RawMessage) error

// Use appends global middleware applied to every tool execution (after validation).
func (r *Registry) Use(mw ...Middleware) {
	if r == nil || len(mw) == 0 {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.middleware = append(r.middleware, mw...)
}

// UsePolicy appends policies evaluated in order before middleware and the handler.
func (r *Registry) UsePolicy(p ...Policy) {
	if r == nil || len(p) == 0 {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.policy = append(r.policy, p...)
}

// BlockToolNames returns a [Policy] that rejects invocations whose tool name is in names.
func BlockToolNames(names ...string) Policy {
	set := make(map[string]struct{}, len(names))
	for _, n := range names {
		n = strings.TrimSpace(n)
		if n != "" {
			set[n] = struct{}{}
		}
	}
	return func(ctx context.Context, name string, args json.RawMessage) error {
		_ = ctx
		if _, bad := set[name]; bad {
			return ErrPolicyBlocked
		}
		return nil
	}
}

// ErrPolicyBlocked is returned when a [Policy] rejects a tool call.
var ErrPolicyBlocked = errors.New("tool: blocked by policy")
