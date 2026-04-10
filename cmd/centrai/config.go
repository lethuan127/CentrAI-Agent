// Config and flags for cmd/centrai. All process configuration is loaded here or in main;
// internal packages receive explicit structs (see docs/9. code-structure.md §8).
package main

import (
	"flag"
	"os"
)

// Config is populated only from flags and environment (no reads inside internal/).
type Config struct {
	APIKey             string
	BaseURL            string
	Model              string
	Message            string
	Session            string
	Repl               bool
	MaxSteps           int
	OmitSessionContext bool
	DemoTools          bool
	AgentFile          string // path to agent definition: .yaml or .md with YAML front matter (see .centrai/agents/)
	HTTPAddr           string // if non-empty, serve Run API (see api/openapi.yaml) and skip one-shot/REPL unless combined
}

func loadConfig() Config {
	var c Config
	flag.StringVar(&c.BaseURL, "base-url", envOr("OPENAI_BASE_URL", "https://api.openai.com/v1"), "OpenAI-compatible API base URL")
	flag.StringVar(&c.Model, "model", envOr("OPENAI_MODEL", "gpt-4o-mini"), "Model name")
	flag.StringVar(&c.Message, "message", "", "Single user message (non-interactive)")
	flag.StringVar(&c.Session, "session", "cli", "Session id for store keying")
	flag.BoolVar(&c.Repl, "repl", false, "Read lines from stdin until EOF")
	flag.IntVar(&c.MaxSteps, "max-steps", 16, "Max model rounds per run")
	flag.BoolVar(&c.OmitSessionContext, "omit-session-context", false, "Do not inject session id/state into the system message")
	flag.BoolVar(&c.DemoTools, "demo-tools", false, "Register echo/add tools to exercise the tool-calling loop")
	flag.StringVar(&c.AgentFile, "agent", "", "Agent file path, or short id (e.g. example → .centrai/agents/example.md)")
	flag.StringVar(&c.HTTPAddr, "http", "", "Listen address for HTTP ingress (e.g. :8080); empty disables")
	flag.Parse()

	c.APIKey = os.Getenv("OPENAI_API_KEY")
	return c
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
