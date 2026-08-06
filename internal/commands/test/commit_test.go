package test

import (
	"bytes"
	"loki/internal/commands"
	"os"
	"testing"
)

func TestCommit_DefaultMessage(t *testing.T) {
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)

	tempDir := t.TempDir()
	os.Chdir(tempDir)
	commands.Init()

	output := CaptureOutput(func() {
		commands.Commit([]string{})
	})

	if output == "" {
		t.Errorf("Expected commit output, got empty string")
	}
}

func TestCommit_WithMessage(t *testing.T) {
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)

	tempDir := t.TempDir()
	os.Chdir(tempDir)
	commands.Init()

	output := CaptureOutput(func() {
		commands.Commit([]string{"-m", "test commit"})
	})

	if output == "" || !contains(output, "Committed:") {
		t.Errorf("Expected commit output with hash, got: %s", output)
	}
}

func contains(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}
