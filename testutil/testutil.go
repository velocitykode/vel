// Package testutil provides reusable test utilities for the vel CLI test suite.
package testutil

import (
	"bytes"
	"io"
	"os"
	"regexp"
	"testing"
)

// CaptureStdout captures stdout during function execution and returns the output as a string.
// The function is run synchronously and stdout is restored after completion.
func CaptureStdout(t *testing.T, f func()) string {
	t.Helper()

	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("CaptureStdout: failed to create pipe: %v", err)
	}
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

// StripANSI removes ANSI escape codes from a string.
// This is useful for comparing terminal output that contains color codes.
func StripANSI(s string) string {
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return ansiRegex.ReplaceAllString(s, "")
}

// WithTempDir creates a temporary directory, changes to it, runs the provided function,
// then restores the original working directory. The temp directory is automatically
// cleaned up by t.TempDir().
func WithTempDir(t *testing.T, f func(dir string)) {
	t.Helper()

	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("WithTempDir: failed to get current directory: %v", err)
	}

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("WithTempDir: failed to change to temp directory: %v", err)
	}
	defer os.Chdir(originalDir)

	f(tmpDir)
}

// WithEnv runs a function with the specified environment variables set.
// After the function completes, all environment variables are restored
// to their original values (or unset if they didn't exist).
func WithEnv(t *testing.T, vars map[string]string, f func()) {
	t.Helper()

	// Save original values
	originalValues := make(map[string]string)
	originalExists := make(map[string]bool)

	for key := range vars {
		if val, exists := os.LookupEnv(key); exists {
			originalValues[key] = val
			originalExists[key] = true
		} else {
			originalExists[key] = false
		}
	}

	// Set new values
	for key, val := range vars {
		if err := os.Setenv(key, val); err != nil {
			t.Fatalf("WithEnv: failed to set env var %s: %v", key, err)
		}
	}

	// Ensure cleanup happens
	defer func() {
		for key := range vars {
			if originalExists[key] {
				os.Setenv(key, originalValues[key])
			} else {
				os.Unsetenv(key)
			}
		}
	}()

	f()
}
