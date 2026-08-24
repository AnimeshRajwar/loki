package core

import (
	"os"
	"path/filepath"
	"strings"
)

type IgnoreMatcher struct {
	patterns []string
}

func LoadIgnore(repoRoot string) *IgnoreMatcher {
	matcher := &IgnoreMatcher{}
	data, err := os.ReadFile(filepath.Join(repoRoot, ".lokiignore"))
	if err != nil {
		return matcher
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		matcher.patterns = append(matcher.patterns, line)
	}
	return matcher
}

func (m *IgnoreMatcher) Matches(path string) bool {
	if len(m.patterns) == 0 {
		return false
	}

	path = filepath.ToSlash(path)

	for _, pattern := range m.patterns {
		if strings.HasSuffix(pattern, "/") {
			dirPattern := strings.TrimSuffix(pattern, "/")
			if path == dirPattern || strings.HasPrefix(path, dirPattern+"/") {
				return true
			}
		}

		matched, _ := filepath.Match(pattern, path)
		if matched {
			return true
		}

		matchedBase, _ := filepath.Match(pattern, filepath.Base(path))
		if matchedBase {
			return true
		}
	}
	return false
}
