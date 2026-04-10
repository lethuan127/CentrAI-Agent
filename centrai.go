// Package centrai is the public facade for the CentrAI Agent Go runtime.
//
// Module path: github.com/lethuan127/centrai-agent
package centrai

import (
	"context"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/lethuan127/centrai-agent/internal/agent"
	"github.com/lethuan127/centrai-agent/internal/agentdef"
	"github.com/lethuan127/centrai-agent/internal/mcp"
	"github.com/lethuan127/centrai-agent/internal/model"
	"github.com/lethuan127/centrai-agent/internal/model/openai"
	"github.com/lethuan127/centrai-agent/internal/model/responsesapi"
	"github.com/lethuan127/centrai-agent/internal/prompt"
	"github.com/lethuan127/centrai-agent/internal/session"
	"github.com/lethuan127/centrai-agent/internal/skill"
	"github.com/lethuan127/centrai-agent/internal/store/memory"
	"github.com/lethuan127/centrai-agent/internal/store/sqlite"
	"github.com/lethuan127/centrai-agent/internal/tool"
)

// Session and message types
type (
	Session  = session.Session
	Message  = session.Message
	Role     = session.Role
	ToolCall = session.ToolCall
	Store    = session.Store
)

// Model streaming
type (
	Client      = model.Client
	ChatRequest = model.ChatRequest
	ChatResult  = model.ChatResult
	StreamChunk = model.StreamChunk
)

// Tools
type (
	Registry   = tool.Registry
	Definition = tool.Definition
	Handler    = tool.Handler
	Middleware = tool.Middleware
	Policy     = tool.Policy
)

// OpenAI-compatible HTTP client (chat completions streaming).
type OpenAIConfig = openai.Config

// ResponsesAPIConfig configures the OpenAI Responses API streaming client.
type ResponsesAPIConfig = responsesapi.Config

// SkillLoader resolves skill files for prompt text (see [skill.Loader]).
type SkillLoader = skill.Loader

// SQLiteStore is a file-backed [Store]; call [SQLiteStore.Close] when shutting down.
type SQLiteStore = sqlite.Store

// AgentDefinition is loaded from YAML or Markdown with YAML front matter (see .centrai/agents/ and agentdef).
type AgentDefinition = agentdef.Definition

// McpServerInstruction configures whether MCP server ids are included in the system prompt (see agentdef).
type McpServerInstruction = agentdef.McpServerInstruction

// Orchestrator
type (
	Runner    = agent.Runner
	RunInput  = agent.RunInput
	RunOutput = agent.RunOutput
	Options   = agent.Options
	Event     = agent.Event
	EventKind = agent.EventKind
	Logger    = agent.Logger
)

var (
	LoadAgentDefinition    = agentdef.LoadFile
	NewRunner              = agent.NewRunner
	NewSlogLogger          = agent.NewSlogLogger
	ErrMaxSteps            = agent.ErrMaxSteps
	NewRegistry            = tool.NewRegistry
	MergeSystemWithSession = prompt.MergeSystemWithSession
	FormatSessionForSystem = prompt.FormatSessionForSystem
	BlockToolNames         = tool.BlockToolNames
)

// NewMemoryStore returns an in-memory session.Store for development and tests.
func NewMemoryStore() Store {
	return memory.New()
}

// NewOpenAIClient returns a streaming model.Client for an OpenAI-compatible API.
func NewOpenAIClient(cfg OpenAIConfig) Client {
	return openai.New(cfg)
}

// NewResponsesAPIClient returns a streaming [Client] using OpenAI's Responses API.
func NewResponsesAPIClient(cfg ResponsesAPIConfig) Client {
	return responsesapi.New(cfg)
}

// NewSQLiteStore opens a SQLite-backed session store at path (creates DB if needed).
func NewSQLiteStore(path string) (*SQLiteStore, error) {
	return sqlite.Open(path)
}

// RegisterMCPTools lists tools from an MCP session and registers them on the registry (with optional name prefix).
func RegisterMCPTools(ctx context.Context, session *mcpsdk.ClientSession, reg *Registry, namePrefix string) error {
	return mcp.RegisterRemoteTools(ctx, session, reg, namePrefix)
}
