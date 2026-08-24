package core

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

type LokiIgnore struct {
	patterns []string
}

// LoadLokiIgnore reads .lokiignore from the given repo root directory.
func LoadLokiIgnore(repoRoot string) *LokiIgnore {
	ignorePath := filepath.Join(repoRoot, ".lokiignore")
	return LoadLokiIgnoreFile(ignorePath)
}

// LoadLokiIgnoreFile parses a .lokiignore file.
func LoadLokiIgnoreFile(filePath string) *LokiIgnore {
	ig := &LokiIgnore{patterns: []string{}}
	file, err := os.Open(filePath)
	if err != nil {
		return ig
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Normalize slashes in patterns
		line = filepath.ToSlash(line)
		ig.patterns = append(ig.patterns, line)
	}
	return ig
}

// IsIgnored checks if a given relative path (relative to repo root) matches any ignore pattern.
func (ig *LokiIgnore) IsIgnored(relPath string) bool {
	if relPath == "" || ig == nil {
		return false
	}

	cleanPath := filepath.ToSlash(filepath.Clean(relPath))
	if cleanPath == "." || cleanPath == "" {
		return false
	}

	// Always ignore .loki and .git internal directories
	if cleanPath == ".loki" || strings.HasPrefix(cleanPath, ".loki/") ||
		cleanPath == ".git" || strings.HasPrefix(cleanPath, ".git/") {
		return true
	}

	parts := strings.Split(cleanPath, "/")

	for _, pattern := range ig.patterns {
		if matchPattern(pattern, cleanPath, parts) {
			return true
		}
	}

	return false
}

func matchPattern(pattern, cleanPath string, parts []string) bool {
	// Handle trailing slash for directory patterns e.g. "bin/" or "node_modules/"
	isDirPattern := strings.HasSuffix(pattern, "/")
	patternBase := strings.TrimSuffix(pattern, "/")
	patternBase = strings.TrimPrefix(patternBase, "/")

	filename := parts[len(parts)-1]

	// 1. If pattern has no slashes except optional trailing slash (e.g., "*.log", "node_modules", "bin/")
	if !strings.Contains(patternBase, "/") {
		// Wildcard/exact match against filename (e.g. *.log matches app.log)
		if match, _ := filepath.Match(patternBase, filename); match {
			return true
		}

		// Exact or wildcard match against any directory component in path (e.g. node_modules in node_modules/express/index.js)
		for i, part := range parts {
			// If it's a directory pattern or matching a non-final part (a directory component)
			if match, _ := filepath.Match(patternBase, part); match {
				if isDirPattern || i < len(parts)-1 || len(parts) == 1 {
					return true
				}
			}
		}
		return false
	}

	// 2. Pattern contains path separators (e.g. "docs/*.md", "build/output.bin", "a/b/")
	// Match against full relative path
	if match, _ := filepath.Match(patternBase, cleanPath); match {
		return true
	}

	// Check prefix match for directory patterns (e.g., "build/" matching "build/app.exe" or "build/sub/app.exe")
	if isDirPattern || strings.HasSuffix(pattern, "/*") {
		if strings.HasPrefix(cleanPath, patternBase+"/") || cleanPath == patternBase {
			return true
		}
	}

	// Match path components against patternBase
	if strings.HasPrefix(cleanPath, patternBase+"/") {
		return true
	}

	return false
}
