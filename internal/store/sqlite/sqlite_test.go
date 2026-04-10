package sqlite

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/lethuan127/centrai-agent/internal/session"
)

func TestStore_roundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "t.db")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	ctx := context.Background()
	id := "sess-1"
	sess := &session.Session{
		ID: id,
		Messages: []session.Message{
			{Role: session.RoleUser, Content: "hi"},
		},
		State: map[string]string{"k": "v"},
	}
	if err := s.Save(ctx, sess); err != nil {
		t.Fatal(err)
	}
	got, err := s.Load(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != id || len(got.Messages) != 1 || got.Messages[0].Content != "hi" {
		t.Fatalf("load mismatch: %+v", got)
	}
	if got.State["k"] != "v" {
		t.Fatalf("state: %+v", got.State)
	}
}

func TestSave_rejectsEmptyID(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(filepath.Join(dir, "x.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	err = s.Save(context.Background(), &session.Session{ID: "", Messages: nil})
	if err == nil {
		t.Fatal("expected error for empty session id")
	}
}
