// Package centrai is the public facade for the CentrAI Agent Go runtime.
//
// Module path: github.com/lethuan127/centrai-agent
package centrai

import (
	"github.com/lethuan127/centrai-agent/internal/agent"
	"github.com/lethuan127/centrai-agent/internal/agentdef"
	"github.com/lethuan127/centrai-agent/internal/model"
	"github.com/lethuan127/centrai-agent/internal/model/openai"
	"github.com/lethuan127/centrai-agent/internal/prompt"
	"github.com/lethuan127/centrai-agent/internal/session"
	"github.com/lethuan127/centrai-agent/internal/store/memory"
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
)

// OpenAI-compatible HTTP client
type OpenAIConfig = openai.Config

// AgentDefinition is a YAML-loaded agent (see agents/ and agentdef package).
type AgentDefinition = agentdef.Definition

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
)

// NewMemoryStore returns an in-memory session.Store for development and tests.
func NewMemoryStore() Store {
	return memory.New()
}

// NewOpenAIClient returns a streaming model.Client for an OpenAI-compatible API.
func NewOpenAIClient(cfg OpenAIConfig) Client {
	return openai.New(cfg)
}
