// Package agentdef loads declarative agent definitions from YAML (Cursor / Claude Code–style agent files).
package agentdef

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Definition is the supported YAML shape for a single agent (version 1).
// Field names follow common agent-file conventions: identity, instructions, optional model and tool bundles.
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
	// Model overrides the default/chat model when non-empty (e.g. gpt-4o-mini).
	Model string `yaml:"model"`
	// Tools lists built-in tool bundles to enable. Supported: "demo" (echo + add).
	Tools []string `yaml:"tools"`
	// MaxSteps caps model rounds for this agent when set (>0).
	MaxSteps *int `yaml:"max_steps"`
	// Metadata is free-form key/value for hosts (labels, owner, etc.).
	Metadata map[string]string `yaml:"metadata"`
}

// LoadFile reads and validates path (YAML).
func LoadFile(path string) (*Definition, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var d Definition
	if err := yaml.Unmarshal(raw, &d); err != nil {
		return nil, fmt.Errorf("agentdef: parse %s: %w", path, err)
	}
	if err := d.Validate(); err != nil {
		return nil, err
	}
	return &d, nil
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
