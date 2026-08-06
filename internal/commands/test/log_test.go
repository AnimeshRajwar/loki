package test

import (
	"loki/internal/commands"
	"os"
	"testing"
)

func TestLog_PrintsLog(t *testing.T) {
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)

	tempDir := t.TempDir()
	os.Chdir(tempDir)
	commands.Init()
	commands.Commit([]string{"-m", "first commit"})

	output := CaptureOutput(func() {
		commands.Log()
	})

	if output == "" {
		t.Errorf("Expected log output, got empty string")
	}
}
