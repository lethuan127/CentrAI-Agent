// Package agentdef loads declarative agent definitions from YAML or Markdown with YAML front matter.
package agentdef

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Definition is the supported schema for a single agent (version 1), from YAML or Markdown front matter.
// File-based agents use name, description, tools, mcpServers, maxTurns, skills, provider, model (see README in .centrai/agents).
type Definition struct {
	// Version must be 1 for this schema.
	Version int `yaml:"version"`
	// Kind is optional; only "Agent" is recognized when set.
	Kind string `yaml:"kind"`
	// Name is a short id for logging and UX (e.g. code-reviewer).
	Name string `yaml:"name"`
	// Description is optional context, prepended before instructions when present.
	Description string `yaml:"description"`
	// Instructions is the system prompt (role, rules, output format).
	Instructions string `yaml:"instructions"`
	// Provider names the LLM vendor or routing profile (e.g. openai). The CLI currently uses an OpenAI-compatible HTTP client regardless; other values are reserved for future multi-provider wiring.
	Provider string `yaml:"provider"`
	// Model overrides the default/chat model when non-empty (e.g. gpt-4o-mini).
	Model string `yaml:"model"`
	// Tools lists built-in tool bundles to enable. Supported: "demo" (echo + add).
	Tools []string `yaml:"tools"`
	// McpServers lists MCP server ids (e.g. keys from .mcp.json) the host should attach; wire tools with internal/mcp.RegisterRemoteTools.
	McpServers []string `yaml:"mcpServers"`
	// McpServerInstruction optionally controls whether declared MCP server ids are copied into the system message (see [McpServerInstruction]).
	McpServerInstruction *McpServerInstruction `yaml:"mcpServerInstruction,omitempty"`
	// Skills lists skill ids or paths for a future skill loader; optional context may be appended to the system prompt.
	Skills []string `yaml:"skills"`
	// MaxTurns caps model rounds per user turn when set (>0).
	MaxTurns *int `yaml:"maxTurns"`
	// Metadata is free-form key/value for hosts (labels, owner, etc.).
	Metadata map[string]string `yaml:"metadata"`
}

// McpServerInstruction configures MCP-related lines in the LLM system prompt.
type McpServerInstruction struct {
	// Allow when true (or omitted), append declared MCP server ids to the system message when mcpServers is non-empty.
	// When explicitly false, those ids are omitted from the prompt (hosts may still register MCP tools out-of-band).
	Allow *bool `yaml:"allow,omitempty"`
}

// McpServerInstructionAllowed reports whether MCP server declarations should appear in the system message.
// When [Definition.McpServerInstruction] is nil or Allow is nil, this is true (backward compatible).
func (d *Definition) McpServerInstructionAllowed() bool {
	if d == nil || d.McpServerInstruction == nil {
		return true
	}
	if d.McpServerInstruction.Allow == nil {
		return true
	}
	return *d.McpServerInstruction.Allow
}

// TurnLimit returns maxTurns when set and >0, otherwise defaultVal.
func (d *Definition) TurnLimit(defaultVal int) int {
	if d == nil {
		return defaultVal
	}
	if d.MaxTurns != nil && *d.MaxTurns > 0 {
		return *d.MaxTurns
	}
	return defaultVal
}

// LLMMetaAppendix returns optional lines listing provider hint, skills, and MCP server ids for the system prompt (empty if none).
func (d *Definition) LLMMetaAppendix() string {
	if d == nil {
		return ""
	}
	var parts []string
	if p := strings.TrimSpace(d.Provider); p != "" {
		parts = append(parts, "Preferred LLM provider (routing hint): "+p)
	}
	if len(d.Skills) > 0 {
		parts = append(parts, "Declared skill references: "+strings.Join(d.Skills, ", "))
	}
	if len(d.McpServers) > 0 && d.McpServerInstructionAllowed() {
		parts = append(parts, "Declared MCP server ids (host-configured): "+strings.Join(d.McpServers, ", "))
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, "\n")
}

// LoadFile reads and validates path. It accepts:
//   - `.yaml` / `.yml`: entire file is YAML
//   - `.md` / `.markdown`: YAML front matter between --- fences; optional Markdown body is appended to Instructions when front matter leaves it empty
func LoadFile(path string) (*Definition, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	lower := strings.ToLower(path)
	switch {
	case strings.HasSuffix(lower, ".md"), strings.HasSuffix(lower, ".markdown"):
		return loadMarkdown(raw, path)
	default:
		var d Definition
		if err := yaml.Unmarshal(raw, &d); err != nil {
			return nil, fmt.Errorf("agentdef: parse %s: %w", path, err)
		}
		if err := d.Validate(); err != nil {
			return nil, err
		}
		return &d, nil
	}
}

func loadMarkdown(raw []byte, path string) (*Definition, error) {
	front, body, ok := splitYAMLFrontMatter(string(raw))
	if !ok {
		return nil, fmt.Errorf("agentdef: %s: expected YAML front matter starting with ---", path)
	}
	var d Definition
	if err := yaml.Unmarshal([]byte(front), &d); err != nil {
		return nil, fmt.Errorf("agentdef: parse front matter %s: %w", path, err)
	}
	if err := mergeMarkdownBody(&d, body); err != nil {
		return nil, err
	}
	if err := d.Validate(); err != nil {
		return nil, err
	}
	return &d, nil
}

// splitYAMLFrontMatter returns the YAML between the first pair of --- lines and the remainder (Markdown body).
func splitYAMLFrontMatter(s string) (front string, body string, ok bool) {
	s = strings.TrimPrefix(s, "\ufeff")
	lines := strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n")
	if len(lines) < 2 || strings.TrimSpace(lines[0]) != "---" {
		return "", "", false
	}
	var yamlLines []string
	i := 1
	for i < len(lines) {
		if strings.TrimSpace(lines[i]) == "---" {
			break
		}
		yamlLines = append(yamlLines, lines[i])
		i++
	}
	if i >= len(lines) || strings.TrimSpace(lines[i]) != "---" {
		return "", "", false
	}
	rest := lines[i+1:]
	body = strings.TrimSpace(strings.Join(rest, "\n"))
	return strings.Join(yamlLines, "\n"), body, true
}

func mergeMarkdownBody(d *Definition, body string) error {
	body = strings.TrimSpace(body)
	if body == "" {
		return nil
	}
	if strings.TrimSpace(d.Instructions) == "" {
		d.Instructions = body
		return nil
	}
	d.Instructions = strings.TrimRight(d.Instructions, "\n") + "\n\n" + body
	return nil
}

// Validate checks required fields for version 1.
func (d *Definition) Validate() error {
	if d == nil {
		return errors.New("agentdef: nil definition")
	}
	if d.Version != 1 {
		return fmt.Errorf("agentdef: unsupported version %d (need 1)", d.Version)
	}
	if k := strings.TrimSpace(d.Kind); k != "" && !strings.EqualFold(k, "Agent") {
		return fmt.Errorf("agentdef: unsupported kind %q", d.Kind)
	}
	if strings.TrimSpace(d.Instructions) == "" {
		return errors.New("agentdef: instructions is required")
	}
	return nil
}

// WantsDemoTools reports whether the "demo" tool bundle is requested.
func (d *Definition) WantsDemoTools() bool {
	if d == nil {
		return false
	}
	for _, t := range d.Tools {
		if strings.EqualFold(strings.TrimSpace(t), "demo") {
			return true
		}
	}
	return false
}
