package memory

import (
	"context"
	"sync"

	"github.com/lethuan127/centrai-agent/internal/session"
)

// Store is an in-memory session.Store for development and tests.
type Store struct {
	mu   sync.RWMutex
	data map[string]*session.Session
}

// New returns an empty memory store.
func New() *Store {
	return &Store{data: make(map[string]*session.Session)}
}

// Load implements session.Store.
func (s *Store) Load(ctx context.Context, id string) (*session.Session, error) {
	_ = ctx
	s.mu.RLock()
	defer s.mu.RUnlock()
	if existing, ok := s.data[id]; ok {
		return cloneSession(existing), nil
	}
	return &session.Session{ID: id, Messages: nil}, nil
}

// Save implements session.Store.
func (s *Store) Save(ctx context.Context, sess *session.Session) error {
	_ = ctx
	if sess == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[sess.ID] = cloneSession(sess)
	return nil
}

func cloneSession(src *session.Session) *session.Session {
	if src == nil {
		return nil
	}
	out := &session.Session{
		ID:       src.ID,
		Messages: make([]session.Message, len(src.Messages)),
		State:    nil,
	}
	if len(src.State) > 0 {
		out.State = make(map[string]string, len(src.State))
		for k, v := range src.State {
			out.State[k] = v
		}
	}
	copy(out.Messages, src.Messages)
	for i := range out.Messages {
		if len(out.Messages[i].ToolCalls) > 0 {
			out.Messages[i].ToolCalls = append([]session.ToolCall(nil), out.Messages[i].ToolCalls...)
		}
	}
	return out
}
