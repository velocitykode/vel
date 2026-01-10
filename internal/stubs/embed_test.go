package stubs

import (
	"strings"
	"testing"
)

func TestGet_ExistingStub(t *testing.T) {
	tests := []struct {
		name     string
		stubPath string
		contains string
	}{
		{
			name:     "handler stub",
			stubPath: "internal/handlers/handler.go.stub",
			contains: "package",
		},
		{
			name:     "middleware stub",
			stubPath: "internal/middleware/middleware.go.stub",
			contains: "package",
		},
		{
			name:     "main.go stub",
			stubPath: "main.go.stub",
			contains: "package main",
		},
		{
			name:     "config stub",
			stubPath: "config/config.go.stub",
			contains: "package",
		},
		{
			name:     "routes web stub",
			stubPath: "routes/web.go.stub",
			contains: "package",
		},
		{
			name:     "routes api stub",
			stubPath: "routes/api.go.stub",
			contains: "package",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := Get(tt.stubPath)
			if err != nil {
				t.Fatalf("Get(%q) error = %v", tt.stubPath, err)
			}
			if len(content) == 0 {
				t.Errorf("Get(%q) returned empty content", tt.stubPath)
			}
			if !strings.Contains(string(content), tt.contains) {
				t.Errorf("Get(%q) content should contain %q", tt.stubPath, tt.contains)
			}
		})
	}
}

func TestGet_NonExistentStub(t *testing.T) {
	_, err := Get("nonexistent/file.stub")
	if err == nil {
		t.Error("Get() should error for non-existent stub")
	}
}

func TestGet_InvalidPath(t *testing.T) {
	_, err := Get("")
	if err == nil {
		t.Error("Get() should error for empty path")
	}
}
