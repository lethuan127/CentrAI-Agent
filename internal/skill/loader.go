// Package skill loads instruction fragments from the filesystem for prompt assembly.
package skill

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Loader resolves named skills or explicit paths to instruction text.
type Loader struct {
	// SearchDirs are tried in order for bare names (not containing a path separator).
	SearchDirs []string
}

// LoadPaths reads UTF-8 files in the given order and joins them with a blank line between files.
func LoadPaths(paths []string) (string, error) {
	if len(paths) == 0 {
		return "", nil
	}
	var b strings.Builder
	for i, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			return "", fmt.Errorf("skill: read %q: %w", p, err)
		}
		if i > 0 {
			b.WriteString("\n\n")
		}
		b.Write(data)
	}
	return strings.TrimSpace(b.String()), nil
}

// Resolve returns absolute paths for each name. A name containing a path separator
// is treated as a path (relative to the working directory). Otherwise a match is
// chosen among SearchDirs × candidates (name.md, name.txt, name); if several exist,
// the lexicographically smallest path wins (deterministic).
func (l *Loader) Resolve(names []string) ([]string, error) {
	var out []string
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if strings.Contains(name, string(os.PathSeparator)) || filepath.IsAbs(name) {
			abs, err := filepath.Abs(name)
			if err != nil {
				return nil, err
			}
			if _, err := os.Stat(abs); err != nil {
				return nil, fmt.Errorf("skill: %w", err)
			}
			out = append(out, abs)
			continue
		}
		found, err := l.resolveName(name)
		if err != nil {
			return nil, err
		}
		out = append(out, found)
	}
	return out, nil
}

func (l *Loader) resolveName(name string) (string, error) {
	candidates := []string{name + ".md", name + ".txt", name}
	var dirs []string
	if len(l.SearchDirs) > 0 {
		dirs = append(dirs, l.SearchDirs...)
	} else {
		dirs = []string{"."}
	}
	var matches []string
	for _, dir := range dirs {
		for _, c := range candidates {
			p := filepath.Join(dir, c)
			if st, err := os.Stat(p); err == nil && !st.IsDir() {
				matches = append(matches, p)
			}
		}
	}
	if len(matches) == 0 {
		return "", fmt.Errorf("skill: no file for %q in search dirs", name)
	}
	sort.Strings(matches)
	return matches[0], nil
}

// LoadNamed resolves names then loads file contents (same semantics as [LoadPaths]).
func (l *Loader) LoadNamed(names []string) (string, error) {
	paths, err := l.Resolve(names)
	if err != nil {
		return "", err
	}
	return LoadPaths(paths)
}
