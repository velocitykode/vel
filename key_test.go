package vel

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunKeyGenerate_CreatesEnvFile(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	err := runKeyGenerate(nil, nil)
	if err != nil {
		t.Fatalf("runKeyGenerate() error = %v", err)
	}

	content, err := os.ReadFile(".env")
	if err != nil {
		t.Fatalf("Failed to read .env: %v", err)
	}

	if !strings.HasPrefix(string(content), "APP_KEY=") {
		t.Errorf(".env should start with APP_KEY=, got: %s", content)
	}

	// Verify key is valid base64-encoded 32 bytes
	key := strings.TrimPrefix(strings.TrimSpace(string(content)), "APP_KEY=")
	decoded, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		t.Fatalf("Key is not valid base64: %v", err)
	}
	if len(decoded) != 32 {
		t.Errorf("Decoded key length = %d, want 32", len(decoded))
	}
}

func TestRunKeyGenerate_UpdatesExistingKey(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create existing .env with old key
	existingContent := "DB_HOST=localhost\nAPP_KEY=old_key_value\nDB_PORT=5432\n"
	os.WriteFile(".env", []byte(existingContent), 0644)

	err := runKeyGenerate(nil, nil)
	if err != nil {
		t.Fatalf("runKeyGenerate() error = %v", err)
	}

	content, _ := os.ReadFile(".env")

	// Key should be updated
	if strings.Contains(string(content), "old_key_value") {
		t.Error("Old key should have been replaced")
	}

	// Other values should be preserved
	if !strings.Contains(string(content), "DB_HOST=localhost") {
		t.Error("DB_HOST should be preserved")
	}
	if !strings.Contains(string(content), "DB_PORT=5432") {
		t.Error("DB_PORT should be preserved")
	}

	// New key should be valid
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "APP_KEY=") {
			key := strings.TrimPrefix(line, "APP_KEY=")
			if _, err := base64.StdEncoding.DecodeString(key); err != nil {
				t.Errorf("New key is not valid base64: %v", err)
			}
		}
	}
}

func TestRunKeyGenerate_AddsKeyWhenMissing(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create .env without APP_KEY
	os.WriteFile(".env", []byte("DB_HOST=localhost\nDB_PORT=5432\n"), 0644)

	err := runKeyGenerate(nil, nil)
	if err != nil {
		t.Fatalf("runKeyGenerate() error = %v", err)
	}

	content, _ := os.ReadFile(".env")

	if !strings.Contains(string(content), "APP_KEY=") {
		t.Error("APP_KEY should be added")
	}
	if !strings.Contains(string(content), "DB_HOST=localhost") {
		t.Error("Existing content should be preserved")
	}
}

func TestRunKeyGenerate_GeneratesUniqueKeys(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Generate first key
	runKeyGenerate(nil, nil)
	content1, _ := os.ReadFile(".env")
	key1 := strings.TrimPrefix(strings.TrimSpace(string(content1)), "APP_KEY=")

	// Generate second key
	os.Remove(".env")
	runKeyGenerate(nil, nil)
	content2, _ := os.ReadFile(".env")
	key2 := strings.TrimPrefix(strings.TrimSpace(string(content2)), "APP_KEY=")

	if key1 == key2 {
		t.Error("Each call should generate a unique key")
	}
}

func TestRunKeyGenerate_CreatesInSubdirectory(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "myproject")
	os.MkdirAll(subDir, 0755)

	originalDir, _ := os.Getwd()
	os.Chdir(subDir)
	defer os.Chdir(originalDir)

	err := runKeyGenerate(nil, nil)
	if err != nil {
		t.Fatalf("runKeyGenerate() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(subDir, ".env")); err != nil {
		t.Error(".env should be created in current directory")
	}
}

func TestRunKeyGenerate_EnvIsDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create .env as a directory - causes read error that's not "not exist"
	os.MkdirAll(".env", 0755)

	err := runKeyGenerate(nil, nil)
	if err == nil {
		t.Error("runKeyGenerate() should error when .env is a directory")
	}
}

func TestRunKeyGenerate_WriteError(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create .env with existing content
	os.WriteFile(".env", []byte("DB_HOST=localhost\n"), 0644)

	// Make .env read-only
	os.Chmod(".env", 0444)
	defer os.Chmod(".env", 0644) // Cleanup

	err := runKeyGenerate(nil, nil)
	if err == nil {
		t.Error("runKeyGenerate() should error when .env is not writable")
	}
}

func TestRunKeyGenerate_CreateEnvError(t *testing.T) {
	tmpDir := t.TempDir()
	readOnlyDir := filepath.Join(tmpDir, "readonly")
	os.MkdirAll(readOnlyDir, 0755)

	originalDir, _ := os.Getwd()
	os.Chdir(readOnlyDir)
	defer os.Chdir(originalDir)

	// Make directory read-only (no .env exists)
	os.Chmod(readOnlyDir, 0555)
	defer os.Chmod(readOnlyDir, 0755) // Cleanup

	err := runKeyGenerate(nil, nil)
	if err == nil {
		t.Error("runKeyGenerate() should error when directory is not writable")
	}
}
