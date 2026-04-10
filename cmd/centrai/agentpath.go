package main

import (
	"os"
	"path/filepath"
	"strings"
)

// resolveAgentFile maps a CLI -agent argument to a file path.
// If arg is an existing file, it is returned as-is.
// If arg is a bare id (no slashes, not absolute)—e.g. "example" or "example.md"—it resolves to
// .centrai/agents/<name>.md, .yaml, or .yml when present; otherwise .centrai/agents/<name>.md for a clear error.
func resolveAgentFile(arg string) string {
	arg = strings.TrimSpace(arg)
	if arg == "" {
		return ""
	}
	if fi, err := os.Stat(arg); err == nil && !fi.IsDir() {
		return arg
	}
	if !isBareAgentRef(arg) {
		return arg
	}
	base := filepath.Base(arg)
	ext := filepath.Ext(base)
	stem := strings.TrimSuffix(base, ext)
	if ext == "" {
		stem = base
	}
	var candidates []string
	switch {
	case ext == "":
		for _, e := range []string{".md", ".yaml", ".yml"} {
			candidates = append(candidates, filepath.Join(".centrai", "agents", stem+e))
		}
	case ext == ".md" || ext == ".yaml" || ext == ".yml":
		candidates = append(candidates, filepath.Join(".centrai", "agents", base))
	default:
		candidates = append(candidates, filepath.Join(".centrai", "agents", base))
	}
	for _, p := range candidates {
		if fi, err := os.Stat(p); err == nil && !fi.IsDir() {
			return p
		}
	}
	if ext == "" {
		return filepath.Join(".centrai", "agents", stem+".md")
	}
	return filepath.Join(".centrai", "agents", base)
}

func isBareAgentRef(arg string) bool {
	if filepath.IsAbs(arg) {
		return false
	}
	return !strings.Contains(arg, "/") && !strings.Contains(arg, `\`)
}
