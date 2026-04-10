package openai

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lethuan127/centrai-agent/internal/model"
	"github.com/lethuan127/centrai-agent/internal/session"
)

func TestStreamChatPlainText(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("path %s", r.URL.Path)
		}
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("no flusher")
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		lines := []string{
			`data: {"choices":[{"delta":{"content":"Hel"}}]}`,
			`data: {"choices":[{"delta":{"content":"lo"}}]}`,
			`data: [DONE]`,
		}
		for _, line := range lines {
			fmt.Fprintf(w, "%s\n\n", line)
			flusher.Flush()
		}
	}))
	defer srv.Close()

	c := New(Config{
		BaseURL: srv.URL,
		Model:   "test-model",
	})

	var got []string
	res, err := c.StreamChat(context.Background(), &model.ChatRequest{
		Model: "test-model",
		Messages: []session.Message{
			{Role: session.RoleUser, Content: "hi"},
		},
	}, func(sc model.StreamChunk) error {
		if sc.DeltaText != "" {
			got = append(got, sc.DeltaText)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Message.Role != session.RoleAssistant || res.Message.Content != "Hello" {
		t.Fatalf("message %+v", res.Message)
	}
	if strings.Join(got, "") != "Hello" {
		t.Fatalf("chunks %v", got)
	}
}

func TestStreamChatToolCall(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("no flusher")
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		lines := []string{
			`data: {"choices":[{"delta":{"tool_calls":[{"index":0,"id":"call_1","type":"function","function":{"name":"echo","arguments":""}}]}}]}`,
			`data: {"choices":[{"delta":{"tool_calls":[{"index":0,"function":{"arguments":"{\"x\":1}"}}]}}]}`,
			`data: [DONE]`,
		}
		for _, line := range lines {
			fmt.Fprintf(w, "%s\n\n", line)
			flusher.Flush()
		}
	}))
	defer srv.Close()

	c := New(Config{BaseURL: srv.URL, Model: "m"})
	res, err := c.StreamChat(context.Background(), &model.ChatRequest{
		Messages: []session.Message{{Role: session.RoleUser, Content: "go"}},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Message.ToolCalls) != 1 {
		t.Fatalf("tool calls %+v", res.Message.ToolCalls)
	}
	tc := res.Message.ToolCalls[0]
	if tc.ID != "call_1" || tc.Name != "echo" || tc.Arguments != `{"x":1}` {
		t.Fatalf("bad tool call %+v", tc)
	}
}
