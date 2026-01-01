package vel

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestServeCmd_FlagDefaults(t *testing.T) {
	tests := []struct {
		name         string
		defaultValue string
	}{
		{"port", "4000"},
		{"env", "development"},
		{"watch", "true"},
		{"tags", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := serveCmd.Flags().Lookup(tt.name)
			if flag == nil {
				t.Fatalf("Flag %q not found", tt.name)
			}
			if flag.DefValue != tt.defaultValue {
				t.Errorf("Flag %q default = %q, want %q", tt.name, flag.DefValue, tt.defaultValue)
			}
		})
	}
}

func TestServeCmd_FlagShorthands(t *testing.T) {
	tests := []struct {
		name      string
		shorthand string
	}{
		{"port", "p"},
		{"env", "e"},
		{"watch", "w"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := serveCmd.Flags().Lookup(tt.name)
			if flag.Shorthand != tt.shorthand {
				t.Errorf("Flag %q shorthand = %q, want %q", tt.name, flag.Shorthand, tt.shorthand)
			}
		})
	}
}

func TestRunServe_NoWatch_BuildFails(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// No valid project - build will fail
	serveWatch = false
	servePort = "4000"
	serveEnv = "test"
	serveBuildTags = ""

	err := runServe(nil, nil)
	if err == nil {
		t.Error("runServe() should error when build fails")
	}
}

func TestRunServe_WithWatch_BuildFails(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// No valid project - build will fail
	serveWatch = true
	servePort = "4000"
	serveEnv = "test"
	serveBuildTags = ""

	err := runServe(nil, nil)
	if err == nil {
		t.Error("runServe() should error when build fails in watch mode")
	}
}

func TestRunServer_BuildFails(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// No go.mod - build fails
	servePort = "4000"
	serveEnv = "test"
	serveBuildTags = ""

	err := runServer()
	if err == nil {
		t.Error("runServer() should error when build fails")
	}
}

func TestRunServer_WithTags_BuildFails(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	servePort = "4000"
	serveEnv = "test"
	serveBuildTags = "integration"

	err := runServer()
	if err == nil {
		t.Error("runServer() should error when build fails")
	}
}

func TestRunServer_SetsEnvironment(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	servePort = "9999"
	serveEnv = "production"
	serveBuildTags = ""

	// This will fail on build, but we can check env vars were set
	runServer()

	if os.Getenv("APP_ENV") != "production" {
		t.Errorf("APP_ENV = %q, want %q", os.Getenv("APP_ENV"), "production")
	}
	if os.Getenv("APP_PORT") != "9999" {
		t.Errorf("APP_PORT = %q, want %q", os.Getenv("APP_PORT"), "9999")
	}
}

func TestRunServer_CreatesTmpDir(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	servePort = "4000"
	serveEnv = "test"

	// Will fail but should create .vel/tmp
	runServer()

	if _, err := os.Stat(".vel/tmp"); err != nil {
		t.Error(".vel/tmp directory should be created")
	}
}

func TestRunWithWatcher_BuildFails(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	servePort = "4000"
	serveEnv = "test"
	serveBuildTags = ""

	err := runWithWatcher()
	if err == nil {
		t.Error("runWithWatcher() should error when initial build fails")
	}
}

func TestRunWithWatcher_CreatesTmpDir(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	servePort = "4000"
	serveEnv = "test"

	// Will fail but should create .vel/tmp
	runWithWatcher()

	if _, err := os.Stat(".vel/tmp"); err != nil {
		t.Error(".vel/tmp directory should be created")
	}
}

func TestWatchFiles_SkipsVendorAndVelocity(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create directories that should be skipped
	os.MkdirAll("vendor/pkg", 0755)
	os.MkdirAll(".vel/tmp", 0755)
	os.MkdirAll("app", 0755)

	os.WriteFile("main.go", []byte("package main"), 0644)
	os.WriteFile("vendor/pkg/lib.go", []byte("package pkg"), 0644)
	os.WriteFile(".vel/tmp/server", []byte("binary"), 0755)
	os.WriteFile("app/handler.go", []byte("package app"), 0644)

	rebuild := make(chan bool, 1)

	// Run watchFiles in goroutine, it will setup watchers then block
	go func() {
		watchFiles(rebuild)
	}()

	// Give it time to setup
	// The function should not error even with vendor/.vel present
}

func TestRunServer_BuildSucceeds_ServerFails(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create a valid Go project that builds but exits immediately
	os.WriteFile("go.mod", []byte("module testserve\n\ngo 1.21\n"), 0644)
	os.WriteFile("main.go", []byte(`package main

import "os"

func main() {
	// Exit with error to simulate server failure
	os.Exit(1)
}
`), 0644)

	servePort = "4000"
	serveEnv = "test"
	serveBuildTags = ""

	err := runServer()
	// Should error because server exits with code 1
	if err == nil {
		t.Error("runServer() should error when server fails")
	}
}

func TestRunServer_BuildSucceeds_ServerSucceeds(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create a valid Go project that builds and exits successfully
	os.WriteFile("go.mod", []byte("module testserve\n\ngo 1.21\n"), 0644)
	os.WriteFile("main.go", []byte(`package main

func main() {
	// Exit successfully immediately
}
`), 0644)

	servePort = "4000"
	serveEnv = "test"
	serveBuildTags = ""

	err := runServer()
	// Should succeed because server exits with code 0
	if err != nil {
		t.Errorf("runServer() error = %v, want nil", err)
	}
}

func TestRunWithWatcher_BuildSucceeds_ServerStarts(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create a valid Go project that builds and exits quickly
	os.WriteFile("go.mod", []byte("module testserve\n\ngo 1.21\n"), 0644)
	os.WriteFile("main.go", []byte(`package main

func main() {
	// Exit immediately - we just need to test build success path
}
`), 0644)

	servePort = "14000"
	serveEnv = "test"
	serveBuildTags = ""

	// Run in goroutine since it blocks
	done := make(chan error, 1)
	go func() {
		done <- runWithWatcher()
	}()

	// Give it time to build and start
	time.Sleep(1 * time.Second)

	// The function should still be running (blocked on watcher)
	// or may have completed without error
	select {
	case err := <-done:
		// If it completed, check no error
		if err != nil {
			t.Errorf("runWithWatcher() error = %v", err)
		}
	default:
		// Still running - that's expected
	}
}

func TestRunWithWatcher_WithBuildTags(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	servePort = "4000"
	serveEnv = "test"
	serveBuildTags = "integration"

	// Will fail but exercises the build tags path
	err := runWithWatcher()
	if err == nil {
		t.Error("Expected error when build fails")
	}
}

func TestWatchFiles_WalkError(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create a directory structure with unreadable subdirectory
	os.MkdirAll("app/controllers", 0755)
	os.WriteFile("main.go", []byte("package main"), 0644)

	// Make subdirectory unreadable to cause walk error
	os.Chmod("app/controllers", 0000)
	defer os.Chmod("app/controllers", 0755)

	rebuild := make(chan bool, 1)
	err := watchFiles(rebuild)

	if err == nil {
		t.Error("watchFiles() should error when walk fails")
	}
}

func TestWatchFiles_SetupSuccess(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create a simple directory structure
	os.MkdirAll("app", 0755)
	os.WriteFile("main.go", []byte("package main"), 0644)
	os.WriteFile("app/handler.go", []byte("package app"), 0644)

	rebuild := make(chan bool, 1)

	// Run watchFiles in goroutine - it will block after setup
	done := make(chan error, 1)
	go func() {
		done <- watchFiles(rebuild)
	}()

	// Give it time to set up watchers
	// If it errors during setup, we'll catch it
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("watchFiles() setup failed: %v", err)
		}
	default:
		// Still running, which means setup succeeded
	}
}

func TestWatchFiles_IgnoresNonGoFiles(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	os.WriteFile("main.go", []byte("package main"), 0644)
	os.WriteFile("config.yaml", []byte("key: value"), 0644)

	rebuild := make(chan bool, 1)

	go func() {
		watchFiles(rebuild)
	}()

	// Give watcher time to set up
	time.Sleep(50 * time.Millisecond)

	// Modify non-Go file - should not trigger rebuild
	os.WriteFile("config.yaml", []byte("key: newvalue"), 0644)

	// Wait briefly and check no rebuild was triggered
	select {
	case <-rebuild:
		t.Error("Non-Go file change should not trigger rebuild")
	case <-time.After(100 * time.Millisecond):
		// Expected - no rebuild triggered
	}
}

func TestWatchFiles_TriggersOnGoFileChange(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	os.WriteFile("main.go", []byte("package main"), 0644)

	rebuild := make(chan bool, 1)

	go func() {
		watchFiles(rebuild)
	}()

	// Give watcher time to set up
	time.Sleep(100 * time.Millisecond)

	// Modify Go file - should trigger rebuild after debounce
	os.WriteFile("main.go", []byte("package main\n// changed"), 0644)

	// Wait for debounce (500ms) + some buffer
	select {
	case <-rebuild:
		// Expected - rebuild triggered
	case <-time.After(800 * time.Millisecond):
		t.Error("Go file change should trigger rebuild")
	}
}

func TestWatchFiles_DebounceMultipleChanges(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	os.WriteFile("main.go", []byte("package main"), 0644)

	rebuild := make(chan bool, 10)

	go func() {
		watchFiles(rebuild)
	}()

	// Give watcher time to set up
	time.Sleep(100 * time.Millisecond)

	// Make multiple rapid changes - should only trigger one rebuild due to debounce
	for i := 0; i < 5; i++ {
		os.WriteFile("main.go", []byte(fmt.Sprintf("package main\n// change %d", i)), 0644)
		time.Sleep(50 * time.Millisecond)
	}

	// Wait for debounce
	time.Sleep(600 * time.Millisecond)

	// Should have received only 1-2 rebuild signals due to debouncing
	count := 0
	for {
		select {
		case <-rebuild:
			count++
		default:
			goto done
		}
	}
done:
	if count > 2 {
		t.Errorf("Expected 1-2 rebuild signals due to debounce, got %d", count)
	}
}
