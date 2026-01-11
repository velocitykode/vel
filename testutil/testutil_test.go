package testutil

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCaptureStdout_BasicOutput(t *testing.T) {
	output := CaptureStdout(t, func() {
		fmt.Print("hello world")
	})

	if output != "hello world" {
		t.Errorf("CaptureStdout() = %q, want %q", output, "hello world")
	}
}

func TestCaptureStdout_MultilineOutput(t *testing.T) {
	output := CaptureStdout(t, func() {
		fmt.Println("line 1")
		fmt.Println("line 2")
		fmt.Print("line 3")
	})

	want := "line 1\nline 2\nline 3"
	if output != want {
		t.Errorf("CaptureStdout() = %q, want %q", output, want)
	}
}

func TestCaptureStdout_EmptyOutput(t *testing.T) {
	output := CaptureStdout(t, func() {
		// Do nothing
	})

	if output != "" {
		t.Errorf("CaptureStdout() = %q, want empty string", output)
	}
}

func TestCaptureStdout_RestoresStdout(t *testing.T) {
	originalStdout := os.Stdout

	_ = CaptureStdout(t, func() {
		fmt.Print("captured")
	})

	if os.Stdout != originalStdout {
		t.Error("CaptureStdout() should restore original stdout")
	}
}

func TestStripANSI_RemovesColorCodes(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "removes red color",
			input: "\x1b[31mred text\x1b[0m",
			want:  "red text",
		},
		{
			name:  "removes green color",
			input: "\x1b[32mgreen text\x1b[0m",
			want:  "green text",
		},
		{
			name:  "removes bold",
			input: "\x1b[1mbold text\x1b[0m",
			want:  "bold text",
		},
		{
			name:  "removes multiple codes",
			input: "\x1b[1;31mbold red\x1b[0m and \x1b[32mgreen\x1b[0m",
			want:  "bold red and green",
		},
		{
			name:  "handles no ANSI codes",
			input: "plain text",
			want:  "plain text",
		},
		{
			name:  "handles empty string",
			input: "",
			want:  "",
		},
		{
			name:  "handles complex escape sequences",
			input: "\x1b[38;5;196mextended color\x1b[0m",
			want:  "extended color",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StripANSI(tt.input)
			if got != tt.want {
				t.Errorf("StripANSI(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestWithTempDir_CreatesTempDirectory(t *testing.T) {
	var capturedDir string

	WithTempDir(t, func(dir string) {
		capturedDir = dir

		// Verify the directory exists
		info, err := os.Stat(dir)
		if err != nil {
			t.Fatalf("temp directory does not exist: %v", err)
		}
		if !info.IsDir() {
			t.Error("temp path is not a directory")
		}
	})

	if capturedDir == "" {
		t.Error("WithTempDir should provide a non-empty directory path")
	}
}

func TestWithTempDir_ChangesWorkingDirectory(t *testing.T) {
	originalDir, _ := os.Getwd()
	var insideDir string

	WithTempDir(t, func(dir string) {
		insideDir, _ = os.Getwd()
	})

	if insideDir == originalDir {
		t.Error("WithTempDir should change working directory to temp dir")
	}
}

func TestWithTempDir_RestoresWorkingDirectory(t *testing.T) {
	originalDir, _ := os.Getwd()

	WithTempDir(t, func(dir string) {
		// Do something in temp dir
		os.WriteFile("test.txt", []byte("test"), 0644)
	})

	currentDir, _ := os.Getwd()
	if currentDir != originalDir {
		t.Errorf("WithTempDir should restore original directory, got %q, want %q", currentDir, originalDir)
	}
}

func TestWithTempDir_AllowsFileOperations(t *testing.T) {
	WithTempDir(t, func(dir string) {
		// Create a file
		filename := "test.txt"
		content := []byte("hello world")
		err := os.WriteFile(filename, content, 0644)
		if err != nil {
			t.Fatalf("failed to write file: %v", err)
		}

		// Verify the file exists in the temp directory
		fullPath := filepath.Join(dir, filename)
		if _, err := os.Stat(fullPath); err != nil {
			t.Errorf("file should exist at %s", fullPath)
		}

		// Read it back
		readContent, err := os.ReadFile(filename)
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}
		if string(readContent) != string(content) {
			t.Errorf("file content = %q, want %q", readContent, content)
		}
	})
}

func TestWithEnv_SetsEnvironmentVariables(t *testing.T) {
	WithEnv(t, map[string]string{
		"TEST_VAR_1": "value1",
		"TEST_VAR_2": "value2",
	}, func() {
		if val := os.Getenv("TEST_VAR_1"); val != "value1" {
			t.Errorf("TEST_VAR_1 = %q, want %q", val, "value1")
		}
		if val := os.Getenv("TEST_VAR_2"); val != "value2" {
			t.Errorf("TEST_VAR_2 = %q, want %q", val, "value2")
		}
	})
}

func TestWithEnv_RestoresOriginalValues(t *testing.T) {
	// Set an initial value
	os.Setenv("TEST_RESTORE_VAR", "original")
	defer os.Unsetenv("TEST_RESTORE_VAR")

	WithEnv(t, map[string]string{
		"TEST_RESTORE_VAR": "modified",
	}, func() {
		if val := os.Getenv("TEST_RESTORE_VAR"); val != "modified" {
			t.Errorf("TEST_RESTORE_VAR inside = %q, want %q", val, "modified")
		}
	})

	// Should be restored to original
	if val := os.Getenv("TEST_RESTORE_VAR"); val != "original" {
		t.Errorf("TEST_RESTORE_VAR after = %q, want %q", val, "original")
	}
}

func TestWithEnv_UnsetsNewVariablesAfter(t *testing.T) {
	// Make sure the variable doesn't exist
	os.Unsetenv("TEST_NEW_VAR")

	WithEnv(t, map[string]string{
		"TEST_NEW_VAR": "temporary",
	}, func() {
		if val := os.Getenv("TEST_NEW_VAR"); val != "temporary" {
			t.Errorf("TEST_NEW_VAR inside = %q, want %q", val, "temporary")
		}
	})

	// Should be unset after
	if val, exists := os.LookupEnv("TEST_NEW_VAR"); exists {
		t.Errorf("TEST_NEW_VAR should be unset after, got %q", val)
	}
}

func TestWithEnv_HandlesEmptyMap(t *testing.T) {
	called := false

	WithEnv(t, map[string]string{}, func() {
		called = true
	})

	if !called {
		t.Error("WithEnv should call the function even with empty map")
	}
}

func TestWithEnv_HandlesEmptyValue(t *testing.T) {
	WithEnv(t, map[string]string{
		"TEST_EMPTY_VAR": "",
	}, func() {
		val, exists := os.LookupEnv("TEST_EMPTY_VAR")
		if !exists {
			t.Error("TEST_EMPTY_VAR should exist")
		}
		if val != "" {
			t.Errorf("TEST_EMPTY_VAR = %q, want empty string", val)
		}
	})
}

func TestWithEnv_MultipleVariables(t *testing.T) {
	// Set up some existing vars
	os.Setenv("EXISTING_VAR", "old_value")
	defer os.Unsetenv("EXISTING_VAR")
	os.Unsetenv("NEW_VAR")

	WithEnv(t, map[string]string{
		"EXISTING_VAR": "new_value",
		"NEW_VAR":      "created",
	}, func() {
		if val := os.Getenv("EXISTING_VAR"); val != "new_value" {
			t.Errorf("EXISTING_VAR = %q, want %q", val, "new_value")
		}
		if val := os.Getenv("NEW_VAR"); val != "created" {
			t.Errorf("NEW_VAR = %q, want %q", val, "created")
		}
	})

	// Check restoration
	if val := os.Getenv("EXISTING_VAR"); val != "old_value" {
		t.Errorf("EXISTING_VAR should be restored to %q, got %q", "old_value", val)
	}
	if _, exists := os.LookupEnv("NEW_VAR"); exists {
		t.Error("NEW_VAR should be unset after WithEnv")
	}
}

// Integration-style tests combining multiple utilities

func TestCombined_CaptureStdoutWithEnv(t *testing.T) {
	output := CaptureStdout(t, func() {
		WithEnv(t, map[string]string{
			"GREETING": "Hello from env",
		}, func() {
			fmt.Print(os.Getenv("GREETING"))
		})
	})

	if output != "Hello from env" {
		t.Errorf("Combined output = %q, want %q", output, "Hello from env")
	}
}

func TestCombined_WithTempDirAndEnv(t *testing.T) {
	WithTempDir(t, func(dir string) {
		WithEnv(t, map[string]string{
			"TEMP_PATH": dir,
		}, func() {
			envPath := os.Getenv("TEMP_PATH")
			if envPath != dir {
				t.Errorf("TEMP_PATH = %q, want %q", envPath, dir)
			}

			// Create a file using env var path
			filename := filepath.Join(envPath, "env_test.txt")
			err := os.WriteFile(filename, []byte("test"), 0644)
			if err != nil {
				t.Fatalf("failed to write file: %v", err)
			}

			if _, err := os.Stat(filename); err != nil {
				t.Errorf("file should exist at %s", filename)
			}
		})
	})
}

func TestCombined_StripANSIWithCapturedOutput(t *testing.T) {
	output := CaptureStdout(t, func() {
		// Simulate colored output
		fmt.Print("\x1b[32mSuccess:\x1b[0m test passed")
	})

	stripped := StripANSI(output)
	if !strings.Contains(stripped, "Success: test passed") {
		t.Errorf("stripped output = %q, should contain %q", stripped, "Success: test passed")
	}
}
