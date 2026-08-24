package test

import (
	"bytes"
	"loki/internal/commands"
	"os"
	"path/filepath"
	"testing"
)

func TestAdd_IgnoredFile(t *testing.T) {
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)

	tempDir := t.TempDir()
	os.Chdir(tempDir)
	commands.Init()

	// Write .lokiignore
	os.WriteFile(".lokiignore", []byte("*.log\nnode_modules/\n"), 0644)
	os.WriteFile("app.log", []byte("log data"), 0644)
	os.WriteFile("main.go", []byte("package main"), 0644)

	// Attempt adding ignored file
	outputAddLog := CaptureOutput(func() {
		commands.Add([]string{"app.log"})
	})

	if !bytes.Contains([]byte(outputAddLog), []byte("warning: file is ignored by .lokiignore -> app.log")) {
		t.Errorf("Expected ignore warning when adding app.log, got: %s", outputAddLog)
	}

	// Attempt adding valid file
	outputAddGo := CaptureOutput(func() {
		commands.Add([]string{"main.go"})
	})

	if !bytes.Contains([]byte(outputAddGo), []byte("staged: main.go")) {
		t.Errorf("Expected main.go to be staged, got: %s", outputAddGo)
	}
}

func TestStatus_FiltersIgnoredFiles(t *testing.T) {
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)

	tempDir := t.TempDir()
	os.Chdir(tempDir)
	commands.Init()

	os.WriteFile(".lokiignore", []byte("*.log\nbin/\n"), 0644)
	os.WriteFile("app.log", []byte("log data"), 0644)
	os.MkdirAll("bin", 0755)
	os.WriteFile(filepath.Join("bin", "app.exe"), []byte("binary data"), 0644)
	os.WriteFile("readme.md", []byte("# Readme"), 0644)

	outputStatus := CaptureOutput(func() {
		commands.Status()
	})

	// Should show untracked readme.md
	if !bytes.Contains([]byte(outputStatus), []byte("readme.md")) {
		t.Errorf("Expected readme.md in status output, got: %s", outputStatus)
	}

	// Should NOT show app.log or bin/app.exe
	if bytes.Contains([]byte(outputStatus), []byte("app.log")) {
		t.Errorf("Status output should not contain ignored file app.log: %s", outputStatus)
	}
	if bytes.Contains([]byte(outputStatus), []byte("app.exe")) {
		t.Errorf("Status output should not contain ignored file app.exe: %s", outputStatus)
	}
}
