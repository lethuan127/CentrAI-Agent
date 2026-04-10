package prompt

import (
	"strings"
	"testing"

	"github.com/lethuan127/centrai-agent/internal/session"
)

func TestFormatSessionForSystem(t *testing.T) {
	s := &session.Session{
		ID: "cli",
		State: map[string]string{
			"theme": "dark",
		},
	}
	out := FormatSessionForSystem(s)
	if !strings.Contains(out, "`cli`") {
		t.Fatalf("missing session id: %q", out)
	}
	if !strings.Contains(out, "**theme**") || !strings.Contains(out, "dark") {
		t.Fatalf("missing state: %q", out)
	}
}

func TestMergeSystemWithSession(t *testing.T) {
	s := &session.Session{ID: "s1"}
	out := MergeSystemWithSession("You are helpful.", s, false)
	if !strings.Contains(out, "You are helpful.") || !strings.Contains(out, "`s1`") {
		t.Fatalf("unexpected merge: %q", out)
	}
	if MergeSystemWithSession("only", s, true) != "only" {
		t.Fatal("omit should skip merge")
	}
}
