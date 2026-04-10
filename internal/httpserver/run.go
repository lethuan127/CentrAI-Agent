// Package httpserver exposes an optional HTTP ingress that maps JSON requests to [agent.Runner] with streaming responses.
package httpserver

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/lethuan127/centrai-agent/internal/agent"
	"github.com/lethuan127/centrai-agent/internal/model"
)

// RunRequest is the JSON body for POST /v1/run.
type RunRequest struct {
	SessionID string `json:"session_id"`
	Message   string `json:"message"`
	System    string `json:"system,omitempty"`
	TraceID   string `json:"trace_id,omitempty"`
	MaxSteps  int    `json:"max_steps,omitempty"`
	Model     string `json:"model,omitempty"`
}

// Options configures the HTTP handler.
type Options struct {
	Runner *agent.Runner
	// DefaultSystem is used when the request body omits system.
	DefaultSystem string
}

// Handler returns an http.Handler that serves POST /v1/run (SSE streaming of assistant text).
func Handler(opt Options) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/run", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if opt.Runner == nil {
			http.Error(w, "runner not configured", http.StatusInternalServerError)
			return
		}
		var body RunRequest
		if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&body); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(body.SessionID) == "" || strings.TrimSpace(body.Message) == "" {
			http.Error(w, "session_id and message required", http.StatusBadRequest)
			return
		}

		trace := strings.TrimSpace(body.TraceID)
		if trace == "" {
			trace = strings.TrimSpace(r.Header.Get("traceparent"))
		}

		fl, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no")

		ctx := r.Context()
		out, err := opt.Runner.Run(ctx, &agent.RunInput{
			SessionID:   body.SessionID,
			UserMessage: body.Message,
			System:      firstNonEmpty(body.System, opt.DefaultSystem),
			MaxSteps:    body.MaxSteps,
			Model:       body.Model,
			TraceID:     trace,
			OnModelChunk: func(c model.StreamChunk) error {
				writeSSE(w, fl, "chunk", map[string]any{"delta": c.DeltaText})
				return nil
			},
		})
		if err != nil {
			writeSSE(w, fl, "error", map[string]any{"message": err.Error()})
			return
		}
		writeSSE(w, fl, "done", map[string]any{
			"assistant":  out.Assistant,
			"session_id": out.SessionID,
			"run_id":     out.RunID,
			"steps_used": out.StepsUsed,
			"truncated":  out.Truncated,
		})
	})
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
		if r.Method != http.MethodHead {
			_, _ = w.Write([]byte("ok\n"))
		}
	})
	return mux
}

func firstNonEmpty(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return b
}

func writeSSE(w http.ResponseWriter, fl http.Flusher, event string, payload any) {
	b, err := json.Marshal(payload)
	if err != nil {
		b = []byte(`{"error":"encode_failed"}`)
	}
	_, _ = w.Write([]byte("event: " + event + "\n"))
	_, _ = w.Write([]byte("data: "))
	_, _ = w.Write(b)
	_, _ = w.Write([]byte("\n\n"))
	fl.Flush()
}
