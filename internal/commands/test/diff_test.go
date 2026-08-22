package test

import (
	"bytes"
	"loki/internal/commands"
	"loki/internal/core"
	"loki/internal/utils"
	"os"
	"testing"
)

func TestDiff_Unstaged(t *testing.T) {
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)

	tempDir := t.TempDir()
	err := os.Chdir(tempDir)
	if err != nil {
		t.Fatal(err)
	}

	commands.Init()

	filename := "test.txt"
	contentA := "line 1\nline 2\nline 3\n"
	err = os.WriteFile(filename, []byte(contentA), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Stage the file
	commands.Add([]string{filename})

	// Modify the file
	contentB := "line 1\nline 2 modified\nline 3\nline 4\n"
	err = os.WriteFile(filename, []byte(contentB), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Capture loki diff (unstaged changes)
	output := CaptureOutput(func() {
		commands.Diff([]string{})
	})

	if !bytes.Contains([]byte(output), []byte("line 2 modified")) {
		t.Errorf("Expected diff to contain added line 'line 2 modified', got:\n%s", output)
	}
	if !bytes.Contains([]byte(output), []byte("line 2")) {
		t.Errorf("Expected diff to contain deleted line 'line 2', got:\n%s", output)
	}
}

func TestDiff_Staged(t *testing.T) {
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)

	tempDir := t.TempDir()
	err := os.Chdir(tempDir)
	if err != nil {
		t.Fatal(err)
	}

	commands.Init()

	filename := "test.txt"
	contentA := "line 1\nline 2\nline 3\n"
	err = os.WriteFile(filename, []byte(contentA), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Stage and commit initial version
	commands.Add([]string{filename})
	commands.Commit([]string{"-m", "initial commit"})

	// Modify the file
	contentB := "line 1\nline 2 modified\nline 3\n"
	err = os.WriteFile(filename, []byte(contentB), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Stage the modification
	commands.Add([]string{filename})

	// Diff --staged should compare Index against HEAD tree
	repo := core.OpenRepository()
	headEntries, _ := repo.GetHeadTreeEntries()
	indexEntries := repo.GetIndexEntries()
	t.Logf("HEAD entries: %+v", headEntries)
	t.Logf("Index entries: %+v", indexEntries)
	for p, h := range headEntries {
		hc, _ := repo.ReadBlob(h)
		t.Logf("HEAD blob for %s (%s): %q", p, h, hc)
	}
	for p, h := range indexEntries {
		ic, _ := repo.ReadBlob(h)
		t.Logf("Index blob for %s (%s): %q", p, h, ic)
	}

	outputStaged := CaptureOutput(func() {
		commands.Diff([]string{"--staged"})
	})
	t.Logf("outputStaged raw: %q", outputStaged)

	if !bytes.Contains([]byte(outputStaged), []byte("line 2 modified")) {
		t.Errorf("Expected staged diff to contain 'line 2 modified', got:\n%s", outputStaged)
	}

	// Unstaged diff should be empty now
	outputUnstaged := CaptureOutput(func() {
		commands.Diff([]string{})
	})

	if outputUnstaged != "" {
		t.Errorf("Expected unstaged diff to be empty, got:\n%s", outputUnstaged)
	}
}

func TestDiff_NoChanges(t *testing.T) {
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)

	tempDir := t.TempDir()
	err := os.Chdir(tempDir)
	if err != nil {
		t.Fatal(err)
	}

	commands.Init()

	filename := "test.txt"
	content := "line 1\nline 2\n"
	err = os.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	commands.Add([]string{filename})

	// Diff (unstaged) should be empty since working dir matches index
	output := CaptureOutput(func() {
		commands.Diff([]string{})
	})
	if output != "" {
		t.Errorf("Expected diff to be empty, got:\n%s", output)
	}

	// Diff --staged should show the added file (empty HEAD vs Index)
	outputStaged := CaptureOutput(func() {
		commands.Diff([]string{"--staged"})
	})
	if !bytes.Contains([]byte(outputStaged), []byte("+line 1")) || !bytes.Contains([]byte(outputStaged), []byte("+line 2")) {
		t.Errorf("Expected staged diff to contain additions, got:\n%s", outputStaged)
	}

	// Commit changes
	commands.Commit([]string{"-m", "initial"})

	// Now Diff --staged should also be empty
	outputStagedAfter := CaptureOutput(func() {
		commands.Diff([]string{"--staged"})
	})
	if outputStagedAfter != "" {
		t.Errorf("Expected staged diff to be empty after commit, got:\n%s", outputStagedAfter)
	}
}

func TestDiff_DeletedFile(t *testing.T) {
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)

	tempDir := t.TempDir()
	err := os.Chdir(tempDir)
	if err != nil {
		t.Fatal(err)
	}

	commands.Init()

	filename := "test.txt"
	content := "line 1\nline 2\n"
	err = os.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	commands.Add([]string{filename})
	commands.Commit([]string{"-m", "initial"})

	// Modify file in index by staging a change
	err = os.WriteFile(filename, []byte("line 1\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	commands.Add([]string{filename})

	// Delete file in working directory
	err = os.Remove(filename)
	if err != nil {
		t.Fatal(err)
	}

	// Diff (unstaged) should show deletion relative to the index (index has "line 1\n", working directory is empty)
	output := CaptureOutput(func() {
		commands.Diff([]string{})
	})
	if !bytes.Contains([]byte(output), []byte("-line 1")) {
		t.Errorf("Expected unstaged diff to show deletion of 'line 1', got:\n%s", output)
	}
}

func TestDiff_UntrackedFile(t *testing.T) {
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)

	tempDir := t.TempDir()
	err := os.Chdir(tempDir)
	if err != nil {
		t.Fatal(err)
	}

	commands.Init()

	// Create an untracked file
	err = os.WriteFile("untracked.txt", []byte("untracked content\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Diff should be empty because the file is not in index
	output := CaptureOutput(func() {
		commands.Diff([]string{})
	})
	if output != "" {
		t.Errorf("Expected diff to be empty for untracked file, got:\n%s", output)
	}
}

func TestDiff_InvalidArgument(t *testing.T) {
	output := CaptureOutput(func() {
		commands.Diff([]string{"--invalid-flag"})
	})
	if !bytes.Contains([]byte(output), []byte("Unknown argument: --invalid-flag")) {
		t.Errorf("Expected unknown argument warning, got:\n%s", output)
	}
}

func TestDebugMyersDiff(t *testing.T) {
	a := []string{"line 1", "line 2", "line 3"}
	b := []string{"line 1", "line 2 modified", "line 3"}
	diffLines := utils.MyersDiff(a, b)
	for i, dl := range diffLines {
		t.Logf("idx=%d Op=%d Text=%q", i, dl.Op, dl.Text)
	}
}
