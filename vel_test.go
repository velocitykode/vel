package vel

import (
	"errors"
	"os"
	"testing"
	"time"
)

func TestExecute_Initializes(t *testing.T) {
	rootCmd = nil

	err := Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if rootCmd == nil {
		t.Fatal("Execute() should initialize rootCmd")
	}
}

func TestExecute_Idempotent(t *testing.T) {
	rootCmd = nil

	// First call initializes
	Execute()
	firstCmd := rootCmd

	// Second call should reuse
	Execute()
	if rootCmd != firstCmd {
		t.Error("Execute() should reuse existing rootCmd")
	}
}

func TestRootCmd_RegistersSubcommands(t *testing.T) {
	rootCmd = nil
	Execute()

	cmd := RootCmd()
	commands := make(map[string]bool)
	for _, c := range cmd.Commands() {
		commands[c.Name()] = true
	}

	required := []string{"serve", "build", "migrate", "migrate:fresh", "make:handler", "key:generate"}
	for _, name := range required {
		if !commands[name] {
			t.Errorf("Missing required command: %s", name)
		}
	}
}

func TestVersion_IsSet(t *testing.T) {
	if Version == "" {
		t.Error("Version should not be empty")
	}
}

func TestHasNewerFiles_NoDirectory(t *testing.T) {
	// Non-existent directory should return false
	result := hasNewerFiles("nonexistent_dir_12345", time.Now())
	if result {
		t.Error("hasNewerFiles() should return false for non-existent directory")
	}
}

func TestHasNewerFiles_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Empty directory should return false
	result := hasNewerFiles(tmpDir, time.Now())
	if result {
		t.Error("hasNewerFiles() should return false for empty directory")
	}
}

func TestHasNewerFiles_NoGoFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a non-Go file
	os.WriteFile(tmpDir+"/readme.md", []byte("# Test"), 0644)

	result := hasNewerFiles(tmpDir, time.Now().Add(-time.Hour))
	if result {
		t.Error("hasNewerFiles() should return false when only non-Go files exist")
	}
}

func TestHasNewerFiles_OlderGoFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a Go file
	os.WriteFile(tmpDir+"/main.go", []byte("package main"), 0644)

	// Use a time in the future, so the file is "older"
	result := hasNewerFiles(tmpDir, time.Now().Add(time.Hour))
	if result {
		t.Error("hasNewerFiles() should return false when Go files are older")
	}
}

func TestHasNewerFiles_NewerGoFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a Go file
	os.WriteFile(tmpDir+"/main.go", []byte("package main"), 0644)

	// Use a time in the past, so the file is "newer"
	result := hasNewerFiles(tmpDir, time.Now().Add(-time.Hour))
	if result != true {
		t.Error("hasNewerFiles() should return true when Go files are newer")
	}
}

func TestHasNewerFiles_NestedGoFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create nested directories with Go files
	os.MkdirAll(tmpDir+"/pkg/handlers", 0755)
	os.WriteFile(tmpDir+"/pkg/handlers/user.go", []byte("package handlers"), 0644)

	// Use a time in the past
	result := hasNewerFiles(tmpDir, time.Now().Add(-time.Hour))
	if result != true {
		t.Error("hasNewerFiles() should find newer Go files in subdirectories")
	}
}

func TestNeedsRebuild_NoDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// With no relevant directories, should return false
	result := needsRebuild()
	if result {
		t.Error("needsRebuild() should return false when no relevant directories exist")
	}
}

func TestNeedsRebuild_WithNewerGoMod(t *testing.T) {
	// Save and mock osExecutable to simulate no executable found
	originalOsExecutable := osExecutable
	defer func() { osExecutable = originalOsExecutable }()
	osExecutable = func() (string, error) {
		return "", errors.New("executable not found")
	}

	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create a go.mod file (will be newer than any existing binary)
	os.WriteFile("go.mod", []byte("module test"), 0644)

	// With a newer go.mod and no executable found, should not trigger rebuild
	result := needsRebuild()
	if result {
		t.Error("needsRebuild() should return false when executable cannot be found")
	}
}

func TestRebuildSelf_NoGoMod(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// rebuildSelf should handle failure gracefully (no panic)
	// It will fail because there's no go.mod or cmd/vel
	rebuildSelf()
	// If we get here without panic, the test passes
}

// Additional tests for needsRebuild coverage

func TestNeedsRebuild_WithNewerGoSum(t *testing.T) {
	// Save and mock osExecutable to simulate no executable found
	originalOsExecutable := osExecutable
	defer func() { osExecutable = originalOsExecutable }()
	osExecutable = func() (string, error) {
		return "", errors.New("executable not found")
	}

	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create a go.sum file (will be newer than any existing binary)
	os.WriteFile("go.sum", []byte("github.com/example/pkg v1.0.0 h1:abc\n"), 0644)

	// With a newer go.sum and no executable found, should not trigger rebuild
	result := needsRebuild()
	if result {
		t.Error("needsRebuild() should return false when executable cannot be found")
	}
}

func TestNeedsRebuild_WithNewerFilesInCmdVel(t *testing.T) {
	// Save and mock osExecutable to simulate no executable found
	originalOsExecutable := osExecutable
	defer func() { osExecutable = originalOsExecutable }()
	osExecutable = func() (string, error) {
		return "", errors.New("executable not found")
	}

	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create cmd/vel directory with a Go file
	os.MkdirAll("cmd/vel", 0755)
	os.WriteFile("cmd/vel/main.go", []byte("package main"), 0644)

	// With newer files and no executable found, should not trigger rebuild
	result := needsRebuild()
	if result {
		t.Error("needsRebuild() should return false when executable cannot be found")
	}
}

func TestNeedsRebuild_WithNewerFilesInBootstrap(t *testing.T) {
	// Save and mock osExecutable to simulate no executable found
	originalOsExecutable := osExecutable
	defer func() { osExecutable = originalOsExecutable }()
	osExecutable = func() (string, error) {
		return "", errors.New("executable not found")
	}

	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create bootstrap directory with a Go file
	os.MkdirAll("bootstrap", 0755)
	os.WriteFile("bootstrap/app.go", []byte("package bootstrap"), 0644)

	// With newer files and no executable found, should not trigger rebuild
	result := needsRebuild()
	if result {
		t.Error("needsRebuild() should return false when executable cannot be found")
	}
}

func TestNeedsRebuild_WithNewerFilesInMigrations(t *testing.T) {
	// Save and mock osExecutable to simulate no executable found
	originalOsExecutable := osExecutable
	defer func() { osExecutable = originalOsExecutable }()
	osExecutable = func() (string, error) {
		return "", errors.New("executable not found")
	}

	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create database/migrations directory with a Go file
	os.MkdirAll("database/migrations", 0755)
	os.WriteFile("database/migrations/001_initial.go", []byte("package migrations"), 0644)

	// With newer files and no executable found, should not trigger rebuild
	result := needsRebuild()
	if result {
		t.Error("needsRebuild() should return false when executable cannot be found")
	}
}

func TestNeedsRebuild_WithOlderGoMod(t *testing.T) {
	// Save and mock osExecutable to simulate no executable found
	originalOsExecutable := osExecutable
	defer func() { osExecutable = originalOsExecutable }()
	osExecutable = func() (string, error) {
		return "", errors.New("executable not found")
	}

	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create a go.mod file
	os.WriteFile("go.mod", []byte("module test"), 0644)

	// Set the modification time to the past
	pastTime := time.Now().Add(-24 * time.Hour)
	os.Chtimes("go.mod", pastTime, pastTime)

	// With older go.mod and no executable, should return false
	result := needsRebuild()
	if result {
		t.Error("needsRebuild() should return false when executable cannot be found")
	}
}

func TestNeedsRebuild_WithOlderGoSum(t *testing.T) {
	// Save and mock osExecutable to simulate no executable found
	originalOsExecutable := osExecutable
	defer func() { osExecutable = originalOsExecutable }()
	osExecutable = func() (string, error) {
		return "", errors.New("executable not found")
	}

	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create a go.sum file
	os.WriteFile("go.sum", []byte("github.com/example/pkg v1.0.0 h1:abc\n"), 0644)

	// Set the modification time to the past
	pastTime := time.Now().Add(-24 * time.Hour)
	os.Chtimes("go.sum", pastTime, pastTime)

	// With older go.sum and no executable, should return false
	result := needsRebuild()
	if result {
		t.Error("needsRebuild() should return false when executable cannot be found")
	}
}

func TestNeedsRebuild_ExecutableError(t *testing.T) {
	// Save original osExecutable and restore after test
	originalOsExecutable := osExecutable
	defer func() { osExecutable = originalOsExecutable }()

	// Mock osExecutable to return an error
	osExecutable = func() (string, error) {
		return "", errors.New("executable path unknown")
	}

	// When os.Executable() returns an error, needsRebuild should return false
	result := needsRebuild()
	if result {
		t.Error("needsRebuild() should return false when os.Executable() returns an error")
	}
}

func TestNeedsRebuild_StatError(t *testing.T) {
	// Save original osExecutable and restore after test
	originalOsExecutable := osExecutable
	defer func() { osExecutable = originalOsExecutable }()

	// Mock osExecutable to return a non-existent path
	osExecutable = func() (string, error) {
		return "/nonexistent/path/to/binary", nil
	}

	// When os.Stat() fails on the executable, needsRebuild should return false
	result := needsRebuild()
	if result {
		t.Error("needsRebuild() should return false when executable path cannot be stat'd")
	}
}

// Additional tests for Execute coverage

func TestExecute_WithInvalidCommand(t *testing.T) {
	rootCmd = nil

	// Save original args and restore after test
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Set args to an invalid command
	os.Args = []string{"vel", "nonexistent-command"}

	err := Execute()
	if err == nil {
		t.Error("Execute() should return error for invalid command")
	}
}

func TestExecute_WithHelpFlag(t *testing.T) {
	rootCmd = nil

	// Save original args and restore after test
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Set args to help flag
	os.Args = []string{"vel", "--help"}

	err := Execute()
	// Help flag should not return an error
	if err != nil {
		t.Errorf("Execute() with --help should not return error, got: %v", err)
	}
}

func TestExecute_WithVersionFlag(t *testing.T) {
	rootCmd = nil

	// Save original args and restore after test
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Set args to version flag
	os.Args = []string{"vel", "--version"}

	err := Execute()
	// Version flag should not return an error
	if err != nil {
		t.Errorf("Execute() with --version should not return error, got: %v", err)
	}
}

// Additional tests for hasNewerFiles edge cases

func TestHasNewerFiles_WithMixedFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create both Go and non-Go files
	os.WriteFile(tmpDir+"/main.go", []byte("package main"), 0644)
	os.WriteFile(tmpDir+"/readme.md", []byte("# Test"), 0644)
	os.WriteFile(tmpDir+"/config.json", []byte("{}"), 0644)

	// Use a time in the past
	result := hasNewerFiles(tmpDir, time.Now().Add(-time.Hour))
	if result != true {
		t.Error("hasNewerFiles() should return true when Go files are newer, ignoring non-Go files")
	}
}

func TestHasNewerFiles_WithSubdirectoriesNoGoFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create nested directories with only non-Go files
	os.MkdirAll(tmpDir+"/pkg/config", 0755)
	os.WriteFile(tmpDir+"/pkg/config/settings.json", []byte("{}"), 0644)
	os.WriteFile(tmpDir+"/readme.md", []byte("# Test"), 0644)

	result := hasNewerFiles(tmpDir, time.Now().Add(-time.Hour))
	if result {
		t.Error("hasNewerFiles() should return false when subdirectories contain no Go files")
	}
}

func TestHasNewerFiles_WithDeeplyNestedGoFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create deeply nested directory structure
	os.MkdirAll(tmpDir+"/a/b/c/d/e", 0755)
	os.WriteFile(tmpDir+"/a/b/c/d/e/deep.go", []byte("package deep"), 0644)

	result := hasNewerFiles(tmpDir, time.Now().Add(-time.Hour))
	if result != true {
		t.Error("hasNewerFiles() should find Go files in deeply nested directories")
	}
}
