package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLokiIgnore_Matching(t *testing.T) {
	tempDir := t.TempDir()
	ignoreContent := `
# Comment line
*.log
*.tmp
bin/
node_modules/
secret.txt
docs/*.md
build/out/
`
	ignorePath := filepath.Join(tempDir, ".lokiignore")
	if err := os.WriteFile(ignorePath, []byte(ignoreContent), 0644); err != nil {
		t.Fatalf("Failed to write test .lokiignore: %v", err)
	}

	ig := LoadLokiIgnore(tempDir)

	tests := []struct {
		relPath  string
		expected bool
	}{
		{"app.log", true},
		{"logs/app.log", true},
		{"test.tmp", true},
		{"bin/loki.exe", true},
		{"bin/sub/tool", true},
		{"node_modules/express/index.js", true},
		{"secret.txt", true},
		{"sub/secret.txt", true},
		{"docs/readme.md", true},
		{"build/out/app.exe", true},
		{"main.go", false},
		{"src/app.js", false},
		{"docs/readme.txt", false},
		{".loki/config", true},
		{".git/HEAD", true},
	}

	for _, tt := range tests {
		got := ig.IsIgnored(tt.relPath)
		if got != tt.expected {
			t.Errorf("IsIgnored(%q) = %v; want %v", tt.relPath, got, tt.expected)
		}
	}
}
