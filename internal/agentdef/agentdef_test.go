package agentdef

import (
	"path/filepath"
	"runtime"
	"testing"
)

func repoAgentsExample(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller")
	}
	// internal/agentdef/*_test.go -> repo root/agents/example.yaml
	root := filepath.Join(filepath.Dir(file), "..", "..")
	return filepath.Join(root, "agents", "example.yaml")
}

func TestLoadFileExample(t *testing.T) {
	path := repoAgentsExample(t)
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

func TestValidateMissingInstructions(t *testing.T) {
	err := (&Definition{Version: 1, Instructions: ""}).Validate()
	if err == nil {
		t.Fatal("expected error")
	}
}
