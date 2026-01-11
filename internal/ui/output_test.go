package ui

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestHighlight(t *testing.T) {
	result := Highlight("test")
	if result == "" {
		t.Error("Highlight should return non-empty string")
	}
	if !strings.Contains(result, "test") {
		t.Error("Highlight should contain the input text")
	}
}

func TestCommand(t *testing.T) {
	result := Command("go run main.go")
	if result == "" {
		t.Error("Command should return non-empty string")
	}
	if !strings.Contains(result, "go run main.go") {
		t.Error("Command should contain the input text")
	}
}

func TestTask(t *testing.T) {
	called := false
	err := Task("Testing step", "Test complete", func() error {
		called = true
		return nil
	})
	if err != nil {
		t.Errorf("Task should return nil for successful action, got: %v", err)
	}
	if !called {
		t.Error("Task action was not called")
	}
}

func TestTask_WithError(t *testing.T) {
	expectedErr := fmt.Errorf("test error")
	err := Task("Testing step", "Test complete", func() error {
		return expectedErr
	})
	if err != expectedErr {
		t.Errorf("Task should return error from action, got: %v", err)
	}
}

func TestTask_ErrorPreservesMessage(t *testing.T) {
	expectedMessage := "connection timeout: failed to reach server"
	originalErr := errors.New(expectedMessage)
	err := Task("Connecting", "Connected", func() error {
		return originalErr
	})
	if err == nil {
		t.Fatal("Task should return error when action fails")
	}
	if err.Error() != expectedMessage {
		t.Errorf("Task should preserve original error message, got: %q, want: %q", err.Error(), expectedMessage)
	}
}

func TestSpinner(t *testing.T) {
	called := false
	err := Spinner("Loading", func() error {
		called = true
		return nil
	})
	if err != nil {
		t.Errorf("Spinner should return nil for successful action, got: %v", err)
	}
	if !called {
		t.Error("Spinner action was not called")
	}
}

func TestSpinnerWithError(t *testing.T) {
	expectedErr := fmt.Errorf("test error")
	err := Spinner("Loading", func() error {
		return expectedErr
	})
	if err != expectedErr {
		t.Errorf("Spinner should return error from action, got: %v", err)
	}
}

func TestSpinner_ShowsDots(t *testing.T) {
	// Test that the spinner shows dots animation when action takes longer than ticker interval
	output := captureStdout(func() {
		err := Spinner("Processing", func() error {
			// Sleep longer than ticker interval (300ms) to trigger dot animation
			time.Sleep(350 * time.Millisecond)
			return nil
		})
		if err != nil {
			t.Errorf("Spinner should return nil for successful action, got: %v", err)
		}
	})

	// The output should contain the message and dots (at least one dot printed)
	if !strings.Contains(output, "Processing") {
		t.Errorf("Spinner output should contain the message, got: %q", output)
	}
	// Should contain at least one dot from the ticker
	if !strings.Contains(output, ".") {
		t.Errorf("Spinner output should contain dots from animation, got: %q", output)
	}
}

func TestInfo(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		wantText string
	}{
		{
			name:     "prints info message",
			message:  "Starting server",
			wantText: "Starting server",
		},
		{
			name:     "handles empty string",
			message:  "",
			wantText: "",
		},
		{
			name:     "handles message with special chars",
			message:  "Loading config from /etc/app.conf",
			wantText: "Loading config from /etc/app.conf",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureStdout(func() {
				Info(tt.message)
			})

			stripped := stripANSI(output)
			if !strings.Contains(stripped, tt.wantText) {
				t.Errorf("Info() output = %q, want to contain %q", stripped, tt.wantText)
			}
			// Should contain the arrow symbol
			if !strings.Contains(output, "→") {
				t.Errorf("Info() output should contain arrow symbol '→'")
			}
		})
	}
}

func TestSuccess(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		wantText string
	}{
		{
			name:     "prints success message",
			message:  "Build completed",
			wantText: "Build completed",
		},
		{
			name:     "handles empty string",
			message:  "",
			wantText: "",
		},
		{
			name:     "handles success with path",
			message:  "Created file: /tmp/output.txt",
			wantText: "Created file: /tmp/output.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureStdout(func() {
				Success(tt.message)
			})

			stripped := stripANSI(output)
			if !strings.Contains(stripped, tt.wantText) {
				t.Errorf("Success() output = %q, want to contain %q", stripped, tt.wantText)
			}
			// Should contain the checkmark symbol
			if !strings.Contains(output, "✓") {
				t.Errorf("Success() output should contain checkmark symbol '✓'")
			}
		})
	}
}

// captureStdout captures stdout during function execution
func captureStdout(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

// stripANSI removes ANSI escape codes from string
func stripANSI(s string) string {
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return ansiRegex.ReplaceAllString(s, "")
}

func TestHeader(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		wantText string
	}{
		{
			name:     "prints uppercase command",
			command:  "make migration",
			wantText: "MAKE MIGRATION",
		},
		{
			name:     "handles empty string",
			command:  "",
			wantText: "",
		},
		{
			name:     "handles already uppercase",
			command:  "CREATE",
			wantText: "CREATE",
		},
		{
			name:     "handles mixed case",
			command:  "MaKe CoNtRoLlEr",
			wantText: "MAKE CONTROLLER",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureStdout(func() {
				Header(tt.command)
			})

			stripped := stripANSI(output)
			if !strings.Contains(stripped, tt.wantText) {
				t.Errorf("Header() output = %q, want to contain %q", stripped, tt.wantText)
			}
		})
	}
}

func TestWarning(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		wantText string
	}{
		{
			name:     "prints warning message",
			message:  "This is a warning",
			wantText: "This is a warning",
		},
		{
			name:     "handles empty string",
			message:  "",
			wantText: "",
		},
		{
			name:     "handles special characters",
			message:  "Warning: file not found!",
			wantText: "Warning: file not found!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureStdout(func() {
				Warning(tt.message)
			})

			stripped := stripANSI(output)
			if !strings.Contains(stripped, tt.wantText) {
				t.Errorf("Warning() output = %q, want to contain %q", stripped, tt.wantText)
			}
			// Should contain the warning symbol
			if !strings.Contains(output, "!") {
				t.Errorf("Warning() output should contain warning symbol '!'")
			}
		})
	}
}

func TestError(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		wantText string
	}{
		{
			name:     "prints error message",
			message:  "Something went wrong",
			wantText: "Something went wrong",
		},
		{
			name:     "handles empty string",
			message:  "",
			wantText: "",
		},
		{
			name:     "handles error with details",
			message:  "Error: connection refused",
			wantText: "Error: connection refused",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureStdout(func() {
				Error(tt.message)
			})

			stripped := stripANSI(output)
			if !strings.Contains(stripped, tt.wantText) {
				t.Errorf("Error() output = %q, want to contain %q", stripped, tt.wantText)
			}
			// Should contain the error symbol
			if !strings.Contains(output, "✗") {
				t.Errorf("Error() output should contain error symbol '✗'")
			}
		})
	}
}

func TestStep(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		wantText string
	}{
		{
			name:     "prints step message",
			message:  "Running tests",
			wantText: "Running tests",
		},
		{
			name:     "handles empty string",
			message:  "",
			wantText: "",
		},
		{
			name:     "handles long message",
			message:  "Compiling all source files and checking for errors",
			wantText: "Compiling all source files and checking for errors",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureStdout(func() {
				Step(tt.message)
			})

			stripped := stripANSI(output)
			if !strings.Contains(stripped, tt.wantText) {
				t.Errorf("Step() output = %q, want to contain %q", stripped, tt.wantText)
			}
			// Should be indented
			if !strings.HasPrefix(stripped, "  ") {
				t.Errorf("Step() output should be indented with two spaces")
			}
		})
	}
}

func TestMuted(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		wantText string
	}{
		{
			name:     "prints muted message",
			message:  "Additional info",
			wantText: "Additional info",
		},
		{
			name:     "handles empty string",
			message:  "",
			wantText: "",
		},
		{
			name:     "handles multiline text",
			message:  "First line\nSecond line",
			wantText: "First line",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureStdout(func() {
				Muted(tt.message)
			})

			stripped := stripANSI(output)
			if !strings.Contains(stripped, tt.wantText) {
				t.Errorf("Muted() output = %q, want to contain %q", stripped, tt.wantText)
			}
			// Should be indented
			if !strings.HasPrefix(stripped, "  ") {
				t.Errorf("Muted() output should be indented with two spaces")
			}
		})
	}
}

func TestBold(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		wantText string
	}{
		{
			name:     "prints bold message",
			message:  "Important message",
			wantText: "Important message",
		},
		{
			name:     "handles empty string",
			message:  "",
			wantText: "",
		},
		{
			name:     "handles numeric text",
			message:  "Version 1.0.0",
			wantText: "Version 1.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureStdout(func() {
				Bold(tt.message)
			})

			stripped := stripANSI(output)
			if !strings.Contains(stripped, tt.wantText) {
				t.Errorf("Bold() output = %q, want to contain %q", stripped, tt.wantText)
			}
		})
	}
}

func TestKeyValue(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		value     string
		wantKey   string
		wantValue string
	}{
		{
			name:      "prints key-value pair",
			key:       "Name",
			value:     "John Doe",
			wantKey:   "Name:",
			wantValue: "John Doe",
		},
		{
			name:      "handles empty key",
			key:       "",
			value:     "value",
			wantKey:   ":",
			wantValue: "value",
		},
		{
			name:      "handles empty value",
			key:       "Status",
			value:     "",
			wantKey:   "Status:",
			wantValue: "",
		},
		{
			name:      "handles both empty",
			key:       "",
			value:     "",
			wantKey:   ":",
			wantValue: "",
		},
		{
			name:      "handles numeric value",
			key:       "Port",
			value:     "8080",
			wantKey:   "Port:",
			wantValue: "8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureStdout(func() {
				KeyValue(tt.key, tt.value)
			})

			stripped := stripANSI(output)
			if !strings.Contains(stripped, tt.wantKey) {
				t.Errorf("KeyValue() output = %q, want to contain key %q", stripped, tt.wantKey)
			}
			if !strings.Contains(stripped, tt.wantValue) {
				t.Errorf("KeyValue() output = %q, want to contain value %q", stripped, tt.wantValue)
			}
			// Should be indented
			if !strings.HasPrefix(stripped, "  ") {
				t.Errorf("KeyValue() output should be indented with two spaces")
			}
		})
	}
}

func TestNewline(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "prints newline",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureStdout(func() {
				Newline()
			})

			if output != "\n" {
				t.Errorf("Newline() output = %q, want %q", output, "\n")
			}
		})
	}
}

func TestNextSteps(t *testing.T) {
	tests := []struct {
		name      string
		steps     []string
		wantCount int
		wantSteps []string
	}{
		{
			name:      "prints single step",
			steps:     []string{"Run the server"},
			wantCount: 1,
			wantSteps: []string{"1.", "Run the server"},
		},
		{
			name:      "prints multiple steps",
			steps:     []string{"Build the project", "Run tests", "Deploy"},
			wantCount: 3,
			wantSteps: []string{"1.", "Build the project", "2.", "Run tests", "3.", "Deploy"},
		},
		{
			name:      "handles empty slice",
			steps:     []string{},
			wantCount: 0,
			wantSteps: []string{"Next steps:"},
		},
		{
			name:      "handles nil slice",
			steps:     nil,
			wantCount: 0,
			wantSteps: []string{"Next steps:"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureStdout(func() {
				NextSteps(tt.steps)
			})

			stripped := stripANSI(output)
			for _, want := range tt.wantSteps {
				if !strings.Contains(stripped, want) {
					t.Errorf("NextSteps() output = %q, want to contain %q", stripped, want)
				}
			}
		})
	}
}

func TestTreeItem(t *testing.T) {
	tests := []struct {
		name       string
		prefix     string
		label      string
		status     string
		done       bool
		wantPrefix string
		wantLabel  string
		wantStatus string
	}{
		{
			name:       "prints completed tree item",
			prefix:     "├─",
			label:      "Install dependencies",
			status:     "complete",
			done:       true,
			wantPrefix: "├─",
			wantLabel:  "Install dependencies",
			wantStatus: "complete",
		},
		{
			name:       "prints incomplete tree item",
			prefix:     "├─",
			label:      "Build project",
			status:     "pending",
			done:       false,
			wantPrefix: "├─",
			wantLabel:  "Build project",
			wantStatus: "pending",
		},
		{
			name:       "prints last tree item with done",
			prefix:     "└─",
			label:      "Deploy",
			status:     "done",
			done:       true,
			wantPrefix: "└─",
			wantLabel:  "Deploy",
			wantStatus: "done",
		},
		{
			name:       "handles empty label",
			prefix:     "├─",
			label:      "",
			status:     "ok",
			done:       true,
			wantPrefix: "├─",
			wantLabel:  "",
			wantStatus: "ok",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureStdout(func() {
				TreeItem(tt.prefix, tt.label, tt.status, tt.done)
			})

			stripped := stripANSI(output)
			if !strings.Contains(stripped, tt.wantPrefix) {
				t.Errorf("TreeItem() output = %q, want to contain prefix %q", stripped, tt.wantPrefix)
			}
			if !strings.Contains(stripped, tt.wantLabel) {
				t.Errorf("TreeItem() output = %q, want to contain label %q", stripped, tt.wantLabel)
			}
			if !strings.Contains(stripped, tt.wantStatus) {
				t.Errorf("TreeItem() output = %q, want to contain status %q", stripped, tt.wantStatus)
			}
			if tt.done && !strings.Contains(output, "✓") {
				t.Errorf("TreeItem() with done=true should contain checkmark symbol '✓'")
			}
		})
	}
}

func TestTreeItemSkipped(t *testing.T) {
	tests := []struct {
		name       string
		prefix     string
		label      string
		reason     string
		wantPrefix string
		wantLabel  string
		wantReason string
	}{
		{
			name:       "prints skipped item with reason",
			prefix:     "├─",
			label:      "Optional step",
			reason:     "already exists",
			wantPrefix: "├─",
			wantLabel:  "Optional step",
			wantReason: "already exists",
		},
		{
			name:       "handles empty reason",
			prefix:     "└─",
			label:      "Cleanup",
			reason:     "",
			wantPrefix: "└─",
			wantLabel:  "Cleanup",
			wantReason: "skipped",
		},
		{
			name:       "handles all empty strings",
			prefix:     "",
			label:      "",
			reason:     "",
			wantPrefix: "",
			wantLabel:  "",
			wantReason: "skipped",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureStdout(func() {
				TreeItemSkipped(tt.prefix, tt.label, tt.reason)
			})

			stripped := stripANSI(output)
			if !strings.Contains(stripped, tt.wantPrefix) {
				t.Errorf("TreeItemSkipped() output = %q, want to contain prefix %q", stripped, tt.wantPrefix)
			}
			if !strings.Contains(stripped, tt.wantLabel) {
				t.Errorf("TreeItemSkipped() output = %q, want to contain label %q", stripped, tt.wantLabel)
			}
			if !strings.Contains(stripped, "skipped") {
				t.Errorf("TreeItemSkipped() output = %q, want to contain 'skipped'", stripped)
			}
			if tt.reason != "" && !strings.Contains(stripped, tt.wantReason) {
				t.Errorf("TreeItemSkipped() output = %q, want to contain reason %q", stripped, tt.wantReason)
			}
		})
	}
}

func TestClearLines(t *testing.T) {
	tests := []struct {
		name      string
		n         int
		wantCount int
	}{
		{
			name:      "clears single line",
			n:         1,
			wantCount: 1,
		},
		{
			name:      "clears multiple lines",
			n:         5,
			wantCount: 5,
		},
		{
			name:      "clears zero lines",
			n:         0,
			wantCount: 0,
		},
		{
			name:      "handles negative input",
			n:         -1,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureStdout(func() {
				ClearLines(tt.n)
			})

			// Count ANSI escape sequences for moving up and clearing
			upCount := strings.Count(output, "\033[A")
			clearCount := strings.Count(output, "\033[K")

			if upCount != tt.wantCount {
				t.Errorf("ClearLines(%d) generated %d up sequences, want %d", tt.n, upCount, tt.wantCount)
			}
			if clearCount != tt.wantCount {
				t.Errorf("ClearLines(%d) generated %d clear sequences, want %d", tt.n, clearCount, tt.wantCount)
			}
		})
	}
}
