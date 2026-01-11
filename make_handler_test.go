package vel

import (
	"os"
	"strings"
	"testing"
)

func TestToHandlerName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"user", "User"},
		{"User", "User"},
		{"UserHandler", "User"},
		{"userHandler", "User"},
		{"user_profile", "UserProfile"},
		{"user-profile", "UserProfile"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := toHandlerName(tt.input)
			if got != tt.expected {
				t.Errorf("toHandlerName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestToPascalCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"user", "User"},
		{"user_profile", "UserProfile"},
		{"user-profile", "UserProfile"},
		{"USER", "USER"},
		{"userProfile", "UserProfile"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := toPascalCase(tt.input)
			if got != tt.expected {
				t.Errorf("toPascalCase(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"User", "user"},
		{"UserProfile", "user_profile"},
		{"userProfile", "user_profile"},
		{"user", "user"},
		{"API", "a_p_i"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := toSnakeCase(tt.input)
			if got != tt.expected {
				t.Errorf("toSnakeCase(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestSplitWords(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"user", []string{"user"}},
		{"user_profile", []string{"user", "profile"}},
		{"user-profile", []string{"user", "profile"}},
		{"UserProfile", []string{"User", "Profile"}},
		{"user profile", []string{"user", "profile"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := splitWords(tt.input)
			if len(got) != len(tt.expected) {
				t.Errorf("splitWords(%q) = %v, want %v", tt.input, got, tt.expected)
				return
			}
			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("splitWords(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.expected[i])
				}
			}
		})
	}
}

func TestRunMakeHandler_CreatesFile(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	err := runMakeHandler(nil, []string{"User"})
	if err != nil {
		t.Fatalf("runMakeHandler() error = %v", err)
	}

	expectedPath := "internal/handlers/user.go"
	if _, err := os.Stat(expectedPath); err != nil {
		t.Errorf("Handler file not created at %s", expectedPath)
	}

	// Check file contents - template generates a function
	content, _ := os.ReadFile(expectedPath)
	if !strings.Contains(string(content), "func User(") {
		t.Error("Generated file should contain User function")
	}
}

func TestRunMakeHandler_AlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	os.MkdirAll("internal/handlers", 0755)
	os.WriteFile("internal/handlers/user.go", []byte("existing"), 0644)

	err := runMakeHandler(nil, []string{"User"})
	if err == nil {
		t.Error("runMakeHandler() should error when file exists")
	}
}

func TestRunMakeHandler_WithPath(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	err := runMakeHandler(nil, []string{"Admin/User"})
	if err != nil {
		t.Fatalf("runMakeHandler() error = %v", err)
	}

	// Path is lowercased for Go convention
	expectedPath := "internal/handlers/admin/user.go"
	if _, err := os.Stat(expectedPath); err != nil {
		t.Errorf("Handler file not created at %s", expectedPath)
	}
}

func TestRunMakeHandler_NoArgs(t *testing.T) {
	// Test via command's Args validation
	err := makeHandlerCmd.Args(makeHandlerCmd, []string{})
	if err == nil {
		t.Error("Args validation should error with no args")
	}
	// Error is now empty (styled output printed separately)
}

func TestRunMakeHandler_TooManyArgs(t *testing.T) {
	// Test via command's Args validation - too many arguments
	err := makeHandlerCmd.Args(makeHandlerCmd, []string{"User", "Extra"})
	if err == nil {
		t.Error("Args validation should error with too many args")
	}
	if err != nil && !strings.Contains(err.Error(), "too many arguments") {
		t.Errorf("Error message should mention 'too many arguments', got: %v", err)
	}
}

func TestRunMakeHandler_DeepNestedPath(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	err := runMakeHandler(nil, []string{"Api/V1/Admin/User"})
	if err != nil {
		t.Fatalf("runMakeHandler() error = %v", err)
	}

	expectedPath := "internal/handlers/api/v1/admin/user.go"
	if _, err := os.Stat(expectedPath); err != nil {
		t.Errorf("Handler file not created at %s", expectedPath)
	}

	// Verify package name is the parent directory
	content, _ := os.ReadFile(expectedPath)
	if !strings.Contains(string(content), "package admin") {
		t.Error("Package should be 'admin' (parent directory)")
	}
}

func TestRunMakeHandler_SnakeCaseInput(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	err := runMakeHandler(nil, []string{"user_profile"})
	if err != nil {
		t.Fatalf("runMakeHandler() error = %v", err)
	}

	// Should create user_profile.go with UserProfile function
	expectedPath := "internal/handlers/user_profile.go"
	if _, err := os.Stat(expectedPath); err != nil {
		t.Errorf("Handler file not created at %s", expectedPath)
	}

	content, _ := os.ReadFile(expectedPath)
	if !strings.Contains(string(content), "func UserProfile(") {
		t.Error("Generated file should contain UserProfile function")
	}
}

func TestRunMakeHandler_MkdirError(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create internal as a file (not directory) to cause MkdirAll to fail
	os.WriteFile("internal", []byte("file"), 0644)

	err := runMakeHandler(nil, []string{"User"})
	if err == nil {
		t.Error("runMakeHandler() should error when cannot create directory")
	}
}

func TestRunMakeHandler_WriteError(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create directory structure
	os.MkdirAll("internal/handlers", 0755)

	// Make directory read-only
	os.Chmod("internal/handlers", 0555)
	defer os.Chmod("internal/handlers", 0755)

	err := runMakeHandler(nil, []string{"User"})
	if err == nil {
		t.Error("runMakeHandler() should error when cannot write file")
	}
}
