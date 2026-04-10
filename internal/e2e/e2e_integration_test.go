package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/lethuan127/centrai-agent"
	"github.com/lethuan127/centrai-agent/internal/httpserver"
	"github.com/lethuan127/centrai-agent/internal/model"
	"github.com/lethuan127/centrai-agent/internal/session"
	"github.com/lethuan127/centrai-agent/internal/tool"
)

// scriptedModel replays fixed ChatResult rounds and optionally records the last request.
type scriptedModel struct {
	rounds  []model.ChatResult
	i       int
	lastReq *model.ChatRequest
}

func (s *scriptedModel) StreamChat(ctx context.Context, req *model.ChatRequest, onChunk func(model.StreamChunk) error) (*model.ChatResult, error) {
	s.lastReq = req
	if s.i >= len(s.rounds) {
		return nil, errors.New("scriptedModel: no more rounds")
	}
	r := s.rounds[s.i]
	s.i++
	if onChunk != nil {
		_ = onChunk(model.StreamChunk{DeltaText: "δ"})
	}
	return &r, nil
}

func registerDemoTools(t *testing.T, reg *tool.Registry) {
	t.Helper()
	echoSchema := json.RawMessage(`{
		"type": "object",
		"properties": { "message": { "type": "string" } },
		"required": ["message"],
		"additionalProperties": false
	}`)
	if err := reg.Register(tool.Definition{
		Name:        "echo",
		Description: "Echo text back.",
		Parameters:  echoSchema,
	}, func(ctx context.Context, args json.RawMessage) (string, error) {
		var in struct {
			Message string `json:"message"`
		}
		if err := json.Unmarshal(args, &in); err != nil {
			return "", err
		}
		out, _ := json.Marshal(map[string]string{"echoed": in.Message})
		return string(out), nil
	}); err != nil {
		t.Fatal(err)
	}
	addSchema := json.RawMessage(`{
		"type": "object",
		"properties": {
			"a": { "type": "number" },
			"b": { "type": "number" }
		},
		"required": ["a", "b"],
		"additionalProperties": false
	}`)
	if err := reg.Register(tool.Definition{
		Name:        "add",
		Description: "Add two numbers.",
		Parameters:  addSchema,
	}, func(ctx context.Context, args json.RawMessage) (string, error) {
		var in struct {
			A float64 `json:"a"`
			B float64 `json:"b"`
		}
		if err := json.Unmarshal(args, &in); err != nil {
			return "", err
		}
		return fmt.Sprintf(`{"sum":%g}`, in.A+in.B), nil
	}); err != nil {
		t.Fatal(err)
	}
}

func demoToolsAppendix() string {
	return `You have native tools:
- echo: repeat a message.
- add: add two numbers.

Call a tool when appropriate; otherwise reply normally.`
}

func systemFromAgentDef(def *centrai.AgentDefinition, demoTools bool) string {
	if def == nil {
		s := "You are a helpful assistant."
		if demoTools {
			s += "\n\n" + demoToolsAppendix()
		}
		return s
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
		base += "\n\n" + demoToolsAppendix()
	}
	return base
}

func newMCPAddSession(t *testing.T, ctx context.Context) *mcpsdk.ClientSession {
	t.Helper()
	server := mcpsdk.NewServer(&mcpsdk.Implementation{Name: "e2e-mcp", Version: "v1"}, nil)
	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "add",
		Description: "add two ints",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"a": map[string]any{"type": "number"},
				"b": map[string]any{"type": "number"},
			},
			"required": []any{"a", "b"},
		},
	}, func(ctx context.Context, req *mcpsdk.CallToolRequest, args map[string]any) (*mcpsdk.CallToolResult, any, error) {
		ai, _ := args["a"].(float64)
		bi, _ := args["b"].(float64)
		sum := int(ai + bi)
		return &mcpsdk.CallToolResult{
			Content: []mcpsdk.Content{&mcpsdk.TextContent{Text: strconv.Itoa(sum)}},
		}, nil, nil
	})
	ct, st := mcpsdk.NewInMemoryTransports()
	if _, err := server.Connect(ctx, st, nil); err != nil {
		t.Fatal(err)
	}
	client := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "e2e-cli", Version: "v1"}, nil)
	sess, err := client.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sess.Close() })
	return sess
}

func TestIntegration_RunnerToolLoopAndSQLitePersistence(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "e2e.db")

	st0, err := centrai.NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	reg := centrai.NewRegistry()
	registerDemoTools(t, reg)

	m := &scriptedModel{rounds: []model.ChatResult{
		{Message: session.Message{
			Role: session.RoleAssistant,
			ToolCalls: []session.ToolCall{{
				ID: "tc1", Name: "add", Arguments: `{"a":2,"b":3}`,
			}},
		}},
		{Message: session.Message{Role: session.RoleAssistant, Content: "done"}},
	}}
	r := centrai.NewRunner(st0, m, reg, centrai.Options{MaxSteps: 8})

	out1, err := r.Run(ctx, &centrai.RunInput{
		SessionID:   "persist-1",
		UserMessage: "compute",
		System:      "You are testing persistence.",
	})
	if err != nil {
		t.Fatal(err)
	}
	if out1.Assistant != "done" {
		t.Fatalf("assistant: %q", out1.Assistant)
	}
	if out1.StepsUsed != 2 {
		t.Fatalf("steps: %d", out1.StepsUsed)
	}
	if err := st0.Close(); err != nil {
		t.Fatal(err)
	}

	st1, err := centrai.NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer st1.Close()

	m2 := &scriptedModel{rounds: []model.ChatResult{
		{Message: session.Message{Role: session.RoleAssistant, Content: "second"}},
	}}
	r2 := centrai.NewRunner(st1, m2, reg, centrai.Options{MaxSteps: 8})
	out2, err := r2.Run(ctx, &centrai.RunInput{
		SessionID:   "persist-1",
		UserMessage: "again",
		System:      "sys",
	})
	if err != nil {
		t.Fatal(err)
	}
	if out2.Assistant != "second" {
		t.Fatalf("assistant: %q", out2.Assistant)
	}
	// Prior user + assistant tool round + tool msg + assistant text + new user + new assistant
	if n := len(out2.Messages); n < 4 {
		t.Fatalf("expected prior history in session, got %d messages", n)
	}
}

func TestIntegration_MCPToolRegisteredAndExecuted(t *testing.T) {
	ctx := context.Background()
	reg := centrai.NewRegistry()
	registerDemoTools(t, reg)

	mcpSess := newMCPAddSession(t, ctx)
	if err := centrai.RegisterMCPTools(ctx, mcpSess, reg, "mcp_"); err != nil {
		t.Fatal(err)
	}

	m := &scriptedModel{rounds: []model.ChatResult{
		{Message: session.Message{
			Role: session.RoleAssistant,
			ToolCalls: []session.ToolCall{{
				ID: "m1", Name: "mcp_add", Arguments: `{"a":10,"b":20}`,
			}},
		}},
		{Message: session.Message{Role: session.RoleAssistant, Content: "sum received"}},
	}}
	r := centrai.NewRunner(centrai.NewMemoryStore(), m, reg, centrai.Options{MaxSteps: 8})
	out, err := r.Run(ctx, &centrai.RunInput{SessionID: "mcp", UserMessage: "use mcp", System: "test"})
	if err != nil {
		t.Fatal(err)
	}
	if out.Assistant != "sum received" {
		t.Fatalf("assistant: %q", out.Assistant)
	}
}

func TestIntegration_AgentDefinitionSkillsAndSystemMerge(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	agentPath := filepath.Join(dir, "agent.md")
	agentBody := `---
version: 1
name: e2e-agent
description: Desc line.
provider: test
tools:
  - demo
mcpServers:
  - notes-server
skills:
  - extra
maxTurns: 8
---
Main instructions from file.
`
	if err := os.WriteFile(agentPath, []byte(agentBody), 0o600); err != nil {
		t.Fatal(err)
	}
	skillFile := filepath.Join(dir, "extra.md")
	if err := os.WriteFile(skillFile, []byte("Skill line from extra."), 0o600); err != nil {
		t.Fatal(err)
	}

	def, err := centrai.LoadAgentDefinition(agentPath)
	if err != nil {
		t.Fatal(err)
	}
	if def.Name != "e2e-agent" {
		t.Fatalf("name %q", def.Name)
	}

	loader := &centrai.SkillLoader{SearchDirs: []string{dir}}
	skillText, err := loader.LoadNamed([]string{"extra"})
	if err != nil {
		t.Fatal(err)
	}
	base := systemFromAgentDef(def, true)
	system := strings.TrimSpace(base + "\n\n" + skillText)

	m := &scriptedModel{rounds: []model.ChatResult{
		{Message: session.Message{Role: session.RoleAssistant, Content: "ok"}},
	}}
	reg := centrai.NewRegistry()
	registerDemoTools(t, reg)

	r := centrai.NewRunner(centrai.NewMemoryStore(), m, reg, centrai.Options{MaxSteps: def.TurnLimit(16)})
	_, err = r.Run(ctx, &centrai.RunInput{
		SessionID:   "skills",
		UserMessage: "hi",
		System:      system,
	})
	if err != nil {
		t.Fatal(err)
	}
	if m.lastReq == nil {
		t.Fatal("expected model to receive request")
	}
	sys := m.lastReq.Messages[0].Content
	if !strings.Contains(sys, "Main instructions from file") {
		t.Fatalf("system missing agent instructions: %q", sys)
	}
	if !strings.Contains(sys, "Skill line from extra.") {
		t.Fatalf("system missing skill text: %q", sys)
	}
	if !strings.Contains(sys, "notes-server") {
		t.Fatalf("system missing MCP meta: %q", sys)
	}
}

func TestIntegration_HTTPSSEAndTraceHeader(t *testing.T) {
	ctx := context.Background()
	reg := centrai.NewRegistry()
	registerDemoTools(t, reg)
	m := &scriptedModel{rounds: []model.ChatResult{
		{Message: session.Message{Role: session.RoleAssistant, Content: "sse-ok"}},
	}}
	r := centrai.NewRunner(centrai.NewMemoryStore(), m, reg, centrai.Options{MaxSteps: 4})
	h := httpserver.Handler(httpserver.Options{Runner: r, DefaultSystem: "default-sys"})
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)

	body := bytes.NewBufferString(`{"session_id":"http1","message":"hello","max_steps":4}`)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, srv.URL+"/v1/run", body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("traceparent", "00-aaa-bbb-01")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d", resp.StatusCode)
	}
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	s := string(raw)
	if !strings.Contains(s, "event: chunk") || !strings.Contains(s, `"delta":"δ"`) {
		t.Fatalf("missing chunk SSE: %s", s)
	}
	if !strings.Contains(s, "event: done") || !strings.Contains(s, `"assistant":"sse-ok"`) {
		t.Fatalf("missing done SSE: %s", s)
	}

	// Health
	hreq, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+"/healthz", nil)
	hresp, err := http.DefaultClient.Do(hreq)
	if err != nil {
		t.Fatal(err)
	}
	defer hresp.Body.Close()
	if hresp.StatusCode != 200 {
		t.Fatalf("healthz %d", hresp.StatusCode)
	}
}

func TestIntegration_ToolPolicyBlocks(t *testing.T) {
	ctx := context.Background()
	reg := centrai.NewRegistry()
	schema := json.RawMessage(`{"type":"object","properties":{"x":{"type":"string"}}}`)
	if err := reg.Register(tool.Definition{Name: "blocked_tool", Parameters: schema}, func(ctx context.Context, args json.RawMessage) (string, error) {
		return "no", nil
	}); err != nil {
		t.Fatal(err)
	}
	reg.UsePolicy(centrai.BlockToolNames("blocked_tool"))

	m := &scriptedModel{rounds: []model.ChatResult{
		{Message: session.Message{
			Role: session.RoleAssistant,
			ToolCalls: []session.ToolCall{{
				ID: "b1", Name: "blocked_tool", Arguments: `{"x":"y"}`,
			}},
		}},
	}}
	r := centrai.NewRunner(centrai.NewMemoryStore(), m, reg, centrai.Options{MaxSteps: 4})
	_, err := r.Run(ctx, &centrai.RunInput{SessionID: "pol", UserMessage: "go", System: "sys"})
	if err == nil {
		t.Fatal("expected error from blocked tool")
	}
	if !errors.Is(err, tool.ErrPolicyBlocked) && !strings.Contains(err.Error(), "blocked") {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestIntegration_MaxStepsExceeded(t *testing.T) {
	ctx := context.Background()
	reg := centrai.NewRegistry()
	schema := json.RawMessage(`{"type":"object","properties":{"n":{"type":"integer"}}}`)
	if err := reg.Register(tool.Definition{Name: "loop", Parameters: schema}, func(ctx context.Context, args json.RawMessage) (string, error) {
		return `{"ok":true}`, nil
	}); err != nil {
		t.Fatal(err)
	}
	m := &scriptedModel{rounds: []model.ChatResult{
		{Message: session.Message{
			Role:      session.RoleAssistant,
			ToolCalls: []session.ToolCall{{ID: "l1", Name: "loop", Arguments: `{"n":1}`}},
		}},
	}}
	r := centrai.NewRunner(centrai.NewMemoryStore(), m, reg, centrai.Options{MaxSteps: 1})
	_, err := r.Run(ctx, &centrai.RunInput{SessionID: "mx", UserMessage: "x", System: "s", MaxSteps: 1})
	if !errors.Is(err, centrai.ErrMaxSteps) {
		t.Fatalf("want ErrMaxSteps, got %v", err)
	}
}

type captureLogger struct{ kinds []string }

func (c *captureLogger) LogEvent(e centrai.Event) {
	c.kinds = append(c.kinds, string(e.Kind))
}

func TestIntegration_LoggerEvents(t *testing.T) {
	ctx := context.Background()
	log := &captureLogger{}
	reg := centrai.NewRegistry()
	m := &scriptedModel{rounds: []model.ChatResult{
		{Message: session.Message{Role: session.RoleAssistant, Content: "x"}},
	}}
	r := centrai.NewRunner(centrai.NewMemoryStore(), m, reg, centrai.Options{MaxSteps: 4, Logger: log})
	if _, err := r.Run(ctx, &centrai.RunInput{SessionID: "log", UserMessage: "m", System: "s", TraceID: "trace-e2e"}); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, k := range log.kinds {
		if k == "run_start" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected run_start in %v", log.kinds)
	}
}
