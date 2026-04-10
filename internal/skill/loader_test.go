package skill

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadPaths_order(t *testing.T) {
	dir := t.TempDir()
	p1 := filepath.Join(dir, "a.md")
	p2 := filepath.Join(dir, "b.md")
	if err := os.WriteFile(p1, []byte("alpha"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p2, []byte("beta"), 0o644); err != nil {
		t.Fatal(err)
	}
	s, err := LoadPaths([]string{p1, p2})
	if err != nil {
		t.Fatal(err)
	}
	if s != "alpha\n\nbeta" {
		t.Fatalf("got %q", s)
	}
}
