package session

import "context"

// Store persists session state. Implementations live under internal/store/.
type Store interface {
	// Load returns the session for id, or an empty session (same id, no messages) if none exists yet.
	Load(ctx context.Context, id string) (*Session, error)
	// Save persists the full session snapshot (including appended messages).
	Save(ctx context.Context, s *Session) error
}
