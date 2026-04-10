package httpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lethuan127/centrai-agent/internal/agent"
	"github.com/lethuan127/centrai-agent/internal/model"
	"github.com/lethuan127/centrai-agent/internal/session"
	"github.com/lethuan127/centrai-agent/internal/store/memory"
	"github.com/lethuan127/centrai-agent/internal/tool"
)

type fakeModel struct{}

func (f *fakeModel) StreamChat(ctx context.Context, req *model.ChatRequest, onChunk func(model.StreamChunk) error) (*model.ChatResult, error) {
	if onChunk != nil {
		_ = onChunk(model.StreamChunk{DeltaText: "x"})
	}
	return &model.ChatResult{Message: session.Message{Role: session.RoleAssistant, Content: "hi"}}, nil
}

func TestHandler_runStreamsChunks(t *testing.T) {
	reg := tool.NewRegistry()
	r := agent.NewRunner(memory.New(), &fakeModel{}, reg, agent.Options{MaxSteps: 4})
	h := Handler(Options{Runner: r})
	body := bytes.NewBufferString(`{"session_id":"s","message":"m"}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/run", body)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	s := rec.Body.String()
	if !strings.Contains(s, "event: chunk") || !strings.Contains(s, `"delta":"x"`) {
		t.Fatalf("body: %s", s)
	}
	if !strings.Contains(s, "event: done") {
		t.Fatalf("missing done: %s", s)
	}
}

func TestHealthz(t *testing.T) {
	h := Handler(Options{Runner: agent.NewRunner(memory.New(), &fakeModel{}, tool.NewRegistry(), agent.Options{})})
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatal(rec.Code)
	}
	req = httptest.NewRequest(http.MethodHead, "/healthz", nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 200 || rec.Body.Len() != 0 {
		t.Fatalf("head: code=%d body=%d", rec.Code, rec.Body.Len())
	}
	req = httptest.NewRequest(http.MethodPost, "/healthz", nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("post healthz: %d", rec.Code)
	}
}

func TestRunRequestJSON(t *testing.T) {
	var r RunRequest
	if err := json.Unmarshal([]byte(`{"session_id":"a","message":"b"}`), &r); err != nil {
		t.Fatal(err)
	}
}
