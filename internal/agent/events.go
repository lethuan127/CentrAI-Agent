package agent

import (
	"log/slog"
	"time"
)

// NewSlogLogger adapts slog to the Runner Logger interface.
func NewSlogLogger(l *slog.Logger) Logger {
	if l == nil {
		return nil
	}
	return slogLogger{l: l}
}

// EventKind identifies structured lifecycle events for observability.
type EventKind string

const (
	EventRunStart         EventKind = "run_start"
	EventRunEnd           EventKind = "run_end"
	EventModelStreamStart EventKind = "model_stream_start"
	EventModelStreamEnd   EventKind = "model_stream_end"
	EventModelChunk       EventKind = "model_chunk"
	EventToolStart        EventKind = "tool_start"
	EventToolEnd          EventKind = "tool_end"
)

// Event is a narrow structured log record for the run lifecycle.
type Event struct {
	Kind      EventKind
	SessionID string
	RunID     string
	TraceID   string
	Tool      string
	Err       error
	Latency   time.Duration
}

// Logger receives lifecycle events (optional).
type Logger interface {
	LogEvent(e Event)
}

type slogLogger struct {
	l *slog.Logger
}

func (s slogLogger) LogEvent(e Event) {
	if s.l == nil {
		return
	}
	attrs := []any{"event", string(e.Kind), "session_id", e.SessionID}
	if e.RunID != "" {
		attrs = append(attrs, "run_id", e.RunID)
	}
	if e.TraceID != "" {
		attrs = append(attrs, "trace_id", e.TraceID)
	}
	if e.Tool != "" {
		attrs = append(attrs, "tool", e.Tool)
	}
	if e.Latency != 0 {
		attrs = append(attrs, "latency_ms", e.Latency.Milliseconds())
	}
	if e.Err != nil {
		s.l.Error("agent event", append(attrs, "err", e.Err)...)
		return
	}
	s.l.Info("agent event", attrs...)
}
