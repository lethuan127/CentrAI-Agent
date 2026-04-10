package agent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/lethuan127/centrai-agent/internal/model"
	"github.com/lethuan127/centrai-agent/internal/prompt"
	"github.com/lethuan127/centrai-agent/internal/session"
	"github.com/lethuan127/centrai-agent/internal/tool"
)

// ErrMaxSteps is returned when MaxSteps model rounds are exhausted before completion.
var ErrMaxSteps = errors.New("agent: max steps reached")

// RunInput is one user turn for a session.
type RunInput struct {
	SessionID   string
	UserMessage string
	System      string
	MaxSteps    int
	// Model overrides the client's default model name when non-empty.
	Model string
	// ToolCallTimeout bounds each tool handler (0 = use ctx only).
	ToolCallTimeout time.Duration
	// OmitSessionContext disables appending session id/state into the system message for this run.
	OmitSessionContext bool
}

// RunOutput is the outcome of a completed run (no pending tool calls).
type RunOutput struct {
	SessionID string
	Assistant string
	Messages  []session.Message // full history after run (including new messages)
	StepsUsed int
	Truncated bool // true if stopped due to MaxSteps while tools might still be needed
}

// Runner orchestrates session load, prompt build, streaming model calls, and the tool loop.
type Runner struct {
	store    session.Store
	model    model.Client
	registry *tool.Registry
	prompt   *prompt.Builder
	maxSteps int
	log      Logger
}

// Options configures a Runner.
type Options struct {
	MaxSteps int // default 16 if 0
	Logger   Logger
}

// NewRunner constructs a Runner. maxSteps in Options caps each Run unless RunInput.MaxSteps overrides.
func NewRunner(store session.Store, m model.Client, reg *tool.Registry, opt Options) *Runner {
	ms := opt.MaxSteps
	if ms == 0 {
		ms = 16
	}
	return &Runner{
		store:    store,
		model:    m,
		registry: reg,
		prompt:   prompt.NewBuilder(),
		maxSteps: ms,
		log:      opt.Logger,
	}
}

func (r *Runner) effectiveMax(in *RunInput) int {
	if in.MaxSteps > 0 {
		return in.MaxSteps
	}
	return r.maxSteps
}

// Run executes one user turn: append message, loop model streaming + tools until done or limit.
func (r *Runner) Run(ctx context.Context, in *RunInput) (*RunOutput, error) {
	if in == nil || in.SessionID == "" {
		return nil, errors.New("agent: session id required")
	}
	if in.UserMessage == "" {
		return nil, errors.New("agent: user message required")
	}

	sess, err := r.store.Load(ctx, in.SessionID)
	if err != nil {
		return nil, err
	}
	sess.ID = in.SessionID

	r.emit(Event{Kind: EventRunStart, SessionID: in.SessionID})

	sess.Messages = append(sess.Messages, session.Message{
		Role:    session.RoleUser,
		Content: in.UserMessage,
	})
	if err := r.store.Save(ctx, sess); err != nil {
		return nil, err
	}

	max := r.effectiveMax(in)
	stepsUsed := 0

	for step := 0; step < max; step++ {
		stepsUsed = step + 1
		defs := r.registry.Definitions()
		system := prompt.MergeSystemWithSession(in.System, sess, in.OmitSessionContext)
		req := r.prompt.Build(system, sess.Messages, defs, in.Model)

		streamStart := time.Now()
		r.emit(Event{Kind: EventModelStreamStart, SessionID: in.SessionID})

		res, err := r.model.StreamChat(ctx, req, func(c model.StreamChunk) error {
			r.emit(Event{Kind: EventModelChunk, SessionID: in.SessionID})
			return nil
		})
		r.emit(Event{Kind: EventModelStreamEnd, SessionID: in.SessionID, Latency: time.Since(streamStart)})

		if err != nil {
			r.emit(Event{Kind: EventRunEnd, SessionID: in.SessionID, Err: err})
			return nil, err
		}

		assistant := res.Message
		sess.Messages = append(sess.Messages, assistant)
		if err := r.store.Save(ctx, sess); err != nil {
			return nil, err
		}

		if len(assistant.ToolCalls) == 0 {
			r.emit(Event{Kind: EventRunEnd, SessionID: in.SessionID})
			return &RunOutput{
				SessionID: in.SessionID,
				Assistant: assistant.Content,
				Messages:  append([]session.Message(nil), sess.Messages...),
				StepsUsed: stepsUsed,
				Truncated: false,
			}, nil
		}

		ids := make([]string, len(assistant.ToolCalls))
		results := make([]string, len(assistant.ToolCalls))
		for i, tc := range assistant.ToolCalls {
			t0 := time.Now()
			r.emit(Event{Kind: EventToolStart, SessionID: in.SessionID, Tool: tc.Name})
			out, err := r.registry.Execute(ctx, tc.Name, tc.Arguments, in.ToolCallTimeout)
			r.emit(Event{Kind: EventToolEnd, SessionID: in.SessionID, Tool: tc.Name, Latency: time.Since(t0), Err: err})
			if err != nil {
				r.emit(Event{Kind: EventRunEnd, SessionID: in.SessionID, Err: err})
				return nil, fmt.Errorf("tool %q: %w", tc.Name, err)
			}
			ids[i] = tc.ID
			results[i] = out
		}

		toolMsgs := tool.ToolMessagesFromResults(ids, results)
		sess.Messages = append(sess.Messages, toolMsgs...)
		if err := r.store.Save(ctx, sess); err != nil {
			return nil, err
		}
	}

	r.emit(Event{Kind: EventRunEnd, SessionID: in.SessionID, Err: ErrMaxSteps})
	return &RunOutput{
		SessionID: in.SessionID,
		Messages:  append([]session.Message(nil), sess.Messages...),
		StepsUsed: stepsUsed,
		Truncated: true,
	}, ErrMaxSteps
}

func (r *Runner) emit(e Event) {
	if r.log == nil {
		return
	}
	r.log.LogEvent(e)
}
