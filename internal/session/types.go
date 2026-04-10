// Package session defines chat domain types and storage interfaces for the agent runtime.
package session

// Role is a message role in the conversation.
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// ToolCall is a single tool invocation requested by the assistant.
type ToolCall struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"` // JSON object as string
}

// Message is one chat message in session history.
type Message struct {
	Role Role `json:"role"`

	// Content is plain text for system, user, assistant (when no tools), and tool roles.
	Content string `json:"content,omitempty"`

	// ToolCalls is set when Role == RoleAssistant and the model requested tools.
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`

	// ToolCallID links a RoleTool message to the assistant's ToolCall.ID.
	ToolCallID string `json:"tool_call_id,omitempty"`
}

// Session is durable conversation state for one session ID.
type Session struct {
	ID       string    `json:"id"`
	Messages []Message `json:"messages"`
	// State is optional key/value data surfaced to the model via the system message (see prompt.MergeSystemWithSession).
	State map[string]string `json:"state,omitempty"`
}
