package vel

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestBuildCmd_FlagDefaults(t *testing.T) {
	tests := []struct {
		name         string
		defaultValue string
	}{
		{"output", ""},
		{"os", runtime.GOOS},
		{"arch", runtime.GOARCH},
		{"tags", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := buildCmd.Flags().Lookup(tt.name)
			if flag == nil {
				t.Fatalf("Flag %q not found", tt.name)
			}
			if flag.DefValue != tt.defaultValue {
				t.Errorf("Flag %q default = %q, want %q", tt.name, flag.DefValue, tt.defaultValue)
			}
		})
	}
}

func TestRunBuild_SimpleProject(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create a minimal Go project
	os.WriteFile("go.mod", []byte("module testbuild\n\ngo 1.21\n"), 0644)
	os.WriteFile("main.go", []byte("package main\n\nfunc main() {}\n"), 0644)

	buildOutput = "testbinary"
	buildOS = runtime.GOOS
	buildArch = runtime.GOARCH
	buildTags = ""

	err := runBuild(nil, nil)
	if err != nil {
		t.Fatalf("runBuild() error = %v", err)
	}

	binaryName := "testbinary"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	if _, err := os.Stat(binaryName); err != nil {
		t.Errorf("Binary not created: %s", binaryName)
	}
}

func TestRunBuild_WithTags(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	os.WriteFile("go.mod", []byte("module testbuild\n\ngo 1.21\n"), 0644)
	os.WriteFile("main.go", []byte("package main\n\nfunc main() {}\n"), 0644)

	buildOutput = "testbinary_tags"
	buildOS = runtime.GOOS
	buildArch = runtime.GOARCH
	buildTags = "test"

	err := runBuild(nil, nil)
	if err != nil {
		t.Fatalf("runBuild() error = %v", err)
	}

	binaryName := "testbinary_tags"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	if _, err := os.Stat(binaryName); err != nil {
		t.Errorf("Binary not created: %s", binaryName)
	}
}

func TestRunBuild_CrossCompile(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	os.WriteFile("go.mod", []byte("module testbuild\n\ngo 1.21\n"), 0644)
	os.WriteFile("main.go", []byte("package main\n\nfunc main() {}\n"), 0644)

	buildOutput = "testbinary_linux"
	buildOS = "linux"
	buildArch = "amd64"
	buildTags = ""

	err := runBuild(nil, nil)
	if err != nil {
		t.Fatalf("runBuild() error = %v", err)
	}

	if _, err := os.Stat("testbinary_linux"); err != nil {
		t.Error("Cross-compiled binary not created")
	}
}

func TestRunBuild_InvalidProject(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// No go.mod - invalid project
	os.WriteFile("main.go", []byte("package main\n\nfunc main() {}\n"), 0644)

	buildOutput = "testbinary_invalid"
	buildOS = runtime.GOOS
	buildArch = runtime.GOARCH
	buildTags = ""

	err := runBuild(nil, nil)
	if err == nil {
		t.Error("runBuild() should error for invalid project")
	}
}

func TestRunBuild_DefaultOutputUsesModuleName(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "myproject")
	os.MkdirAll(projectDir, 0755)

	originalDir, _ := os.Getwd()
	os.Chdir(projectDir)
	defer os.Chdir(originalDir)

	os.WriteFile("go.mod", []byte("module myproject\n\ngo 1.21\n"), 0644)
	os.WriteFile("main.go", []byte("package main\n\nfunc main() {}\n"), 0644)

	buildOutput = ""
	buildOS = runtime.GOOS
	buildArch = runtime.GOARCH
	buildTags = ""

	err := runBuild(nil, nil)
	if err != nil {
		t.Fatalf("runBuild() error = %v", err)
	}

	// When output is empty, should use directory name "myproject"
	expectedBinary := "myproject"
	if runtime.GOOS == "windows" {
		expectedBinary += ".exe"
	}
	if _, err := os.Stat(expectedBinary); err != nil {
		t.Errorf("Binary not created with default name: %s", expectedBinary)
	}
}

func TestRunBuild_WindowsExeSuffix(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "winapp")
	os.MkdirAll(projectDir, 0755)

	originalDir, _ := os.Getwd()
	os.Chdir(projectDir)
	defer os.Chdir(originalDir)

	os.WriteFile("go.mod", []byte("module winapp\n\ngo 1.21\n"), 0644)
	os.WriteFile("main.go", []byte("package main\n\nfunc main() {}\n"), 0644)

	buildOutput = ""
	buildOS = "windows" // Target windows to trigger .exe suffix
	buildArch = "amd64"
	buildTags = ""

	err := runBuild(nil, nil)
	if err != nil {
		t.Fatalf("runBuild() error = %v", err)
	}

	// Should add .exe suffix for windows target
	if _, err := os.Stat("winapp.exe"); err != nil {
		t.Error("Binary should have .exe suffix when targeting windows")
	}
}
