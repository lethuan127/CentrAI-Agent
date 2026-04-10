package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveAgentFileShortID(t *testing.T) {
	dir := t.TempDir()
	agents := filepath.Join(dir, ".centrai", "agents")
	if err := os.MkdirAll(agents, 0o755); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(agents, "example.md")
	if err := os.WriteFile(p, []byte("---\nversion: 1\ninstructions: x\n---\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	old, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(old) }()

	want := filepath.Join(".centrai", "agents", "example.md")
	if got := resolveAgentFile("example"); got != want {
		t.Fatalf("resolveAgentFile(example) = %q want %q", got, want)
	}
	if got := resolveAgentFile("example.md"); got != want {
		t.Fatalf("resolveAgentFile(example.md) = %q want %q", got, want)
	}
}

func TestResolveAgentFileMissingDefaultsToMD(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".centrai", "agents"), 0o755); err != nil {
		t.Fatal(err)
	}
	old, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(old) }()

	want := filepath.Join(".centrai", "agents", "missing.md")
	if got := resolveAgentFile("missing"); got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
