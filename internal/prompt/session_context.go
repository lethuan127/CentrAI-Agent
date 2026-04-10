package prompt

import (
	"sort"
	"strings"

	"github.com/lethuan127/centrai-agent/internal/session"
)

// FormatSessionForSystem returns a markdown fragment listing session id and optional state,
// suitable for appending to the system message (similar in role to Agno’s session state in system context).
func FormatSessionForSystem(s *session.Session) string {
	if s == nil || s.ID == "" {
		return ""
	}
	var b strings.Builder
	b.WriteString("### Session context\n\n")
	b.WriteString("- **session_id**: `")
	b.WriteString(s.ID)
	b.WriteString("`\n")
	if len(s.State) == 0 {
		return b.String()
	}
	b.WriteString("\n#### Session state\n\n")
	keys := make([]string, 0, len(s.State))
	for k := range s.State {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		b.WriteString("- **")
		b.WriteString(k)
		b.WriteString("**: ")
		b.WriteString(s.State[k])
		b.WriteString("\n")
	}
	return b.String()
}

// MergeSystemWithSession appends formatted session context to base system text when s is non-nil and omit is false.
func MergeSystemWithSession(base string, s *session.Session, omit bool) string {
	if omit || s == nil {
		return base
	}
	block := FormatSessionForSystem(s)
	if block == "" {
		return base
	}
	if strings.TrimSpace(base) == "" {
		return block
	}
	return base + "\n\n" + block
}
