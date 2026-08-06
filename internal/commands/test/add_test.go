package test

import (
	"bytes"
	"loki/internal/commands"
	"os"
	"testing"
)

func TestAdd_NoFiles(t *testing.T) {
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)

	tempDir := t.TempDir()
	os.Chdir(tempDir)
	commands.Init()

	output := CaptureOutput(func() {
		commands.Add([]string{})
	})

	if !bytes.Contains([]byte(output), []byte("error: no files specified")) {
		t.Errorf("Expected error message for no files, got: %s", output)
	}
}
