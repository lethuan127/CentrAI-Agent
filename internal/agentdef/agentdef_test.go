package agentdef

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func repoCentraiExampleMD(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller")
	}
	root := filepath.Join(filepath.Dir(file), "..", "..")
	return filepath.Join(root, ".centrai", "agents", "example.md")
}

func TestLoadFileExampleMarkdown(t *testing.T) {
	path := repoCentraiExampleMD(t)
	d, err := LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if d.Name != "example-assistant" {
		t.Fatalf("name %q", d.Name)
	}
	if !d.WantsDemoTools() {
		t.Fatal("expected demo tools")
	}
}

func TestLoadMarkdownBodyFillsInstructions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "agent.md")
	content := "---\nversion: 1\n---\n\nHello from body only.\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	d, err := LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(d.Instructions, "Hello from body only") {
		t.Fatalf("instructions %q", d.Instructions)
	}
}

func TestTurnLimit(t *testing.T) {
	a := 3
	d := &Definition{Version: 1, Instructions: "x", MaxTurns: &a}
	if got := d.TurnLimit(100); got != 3 {
		t.Fatalf("TurnLimit = %d want 3", got)
	}
	d2 := &Definition{Version: 1, Instructions: "x"}
	if got := d2.TurnLimit(99); got != 99 {
		t.Fatalf("TurnLimit = %d want 99", got)
	}
}

func TestLoadYAMLFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "agent.yaml")
	yml := "version: 1\ninstructions: \"hi\"\ntools:\n  - demo\n"
	if err := os.WriteFile(path, []byte(yml), 0o600); err != nil {
		t.Fatal(err)
	}
	d, err := LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if d.Instructions != "hi" {
		t.Fatalf("instructions %q", d.Instructions)
	}
	if !d.WantsDemoTools() {
		t.Fatal("demo")
	}
}

func TestValidateMissingInstructions(t *testing.T) {
	err := (&Definition{Version: 1, Instructions: ""}).Validate()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMcpServerInstructionAllow_falseOmitsFromAppendix(t *testing.T) {
	no := false
	d := &Definition{
		Version:              1,
		Instructions:         "do work",
		McpServers:           []string{"fetch"},
		McpServerInstruction: &McpServerInstruction{Allow: &no},
	}
	meta := d.LLMMetaAppendix()
	if strings.Contains(meta, "MCP") || strings.Contains(meta, "fetch") {
		t.Fatalf("expected MCP line omitted, got %q", meta)
	}
}

func TestMcpServerInstructionAllow_trueIncludes(t *testing.T) {
	yes := true
	d := &Definition{
		Version:              1,
		Instructions:         "do work",
		McpServers:           []string{"fetch"},
		McpServerInstruction: &McpServerInstruction{Allow: &yes},
	}
	meta := d.LLMMetaAppendix()
	if !strings.Contains(meta, "fetch") {
		t.Fatalf("expected MCP line, got %q", meta)
	}
}

func TestMcpServerInstruction_nilLegacyIncludesMCP(t *testing.T) {
	d := &Definition{
		Version:      1,
		Instructions: "do work",
		McpServers:   []string{"a"},
	}
	meta := d.LLMMetaAppendix()
	if !strings.Contains(meta, "a") {
		t.Fatalf("expected MCP line, got %q", meta)
	}
}

func TestMcpServerInstruction_yamlUnmarshal(t *testing.T) {
	const y = `
version: 1
instructions: "x"
mcpServers:
  - fetch
mcpServerInstruction:
  allow: false
`
	var d Definition
	if err := yaml.Unmarshal([]byte(y), &d); err != nil {
		t.Fatal(err)
	}
	if d.McpServerInstruction == nil || d.McpServerInstruction.Allow == nil || *d.McpServerInstruction.Allow {
		t.Fatalf("want allow false, got %+v", d.McpServerInstruction)
	}
	if d.McpServerInstructionAllowed() {
		t.Fatal("expected disallowed")
	}
}
