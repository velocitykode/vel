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

// TestRunKeyGenerate_TableDriven uses table-driven tests to cover additional edge cases
func TestRunKeyGenerate_TableDriven(t *testing.T) {
	tests := []struct {
		name           string
		setupEnv       func(t *testing.T) string // returns tmpDir path
		wantErr        bool
		validateResult func(t *testing.T, envPath string)
	}{
		{
			name: "creates env file with valid base64 key when no env exists",
			setupEnv: func(t *testing.T) string {
				tmpDir := t.TempDir()
				return tmpDir
			},
			wantErr: false,
			validateResult: func(t *testing.T, envPath string) {
				content, err := os.ReadFile(envPath)
				if err != nil {
					t.Fatalf("Failed to read .env: %v", err)
				}
				if !strings.HasPrefix(string(content), "APP_KEY=") {
					t.Errorf("Expected .env to start with APP_KEY=, got: %s", content)
				}
				key := strings.TrimPrefix(strings.TrimSpace(string(content)), "APP_KEY=")
				decoded, err := base64.StdEncoding.DecodeString(key)
				if err != nil {
					t.Errorf("Key is not valid base64: %v", err)
				}
				if len(decoded) != 32 {
					t.Errorf("Decoded key length = %d, want 32", len(decoded))
				}
			},
		},
		{
			name: "updates existing app key in middle of file",
			setupEnv: func(t *testing.T) string {
				tmpDir := t.TempDir()
				envContent := "DB_HOST=localhost\nAPP_KEY=oldkey123\nDB_PORT=5432\n"
				os.WriteFile(tmpDir+"/.env", []byte(envContent), 0644)
				return tmpDir
			},
			wantErr: false,
			validateResult: func(t *testing.T, envPath string) {
				content, err := os.ReadFile(envPath)
				if err != nil {
					t.Fatalf("Failed to read .env: %v", err)
				}
				if strings.Contains(string(content), "oldkey123") {
					t.Error("Old key should have been replaced")
				}
				if !strings.Contains(string(content), "DB_HOST=localhost") {
					t.Error("Other env vars should be preserved")
				}
			},
		},
		{
			name: "prepends app key when not present in existing file",
			setupEnv: func(t *testing.T) string {
				tmpDir := t.TempDir()
				envContent := "DB_HOST=localhost\nDB_PORT=5432"
				os.WriteFile(tmpDir+"/.env", []byte(envContent), 0644)
				return tmpDir
			},
			wantErr: false,
			validateResult: func(t *testing.T, envPath string) {
				content, err := os.ReadFile(envPath)
				if err != nil {
					t.Fatalf("Failed to read .env: %v", err)
				}
				lines := strings.Split(string(content), "\n")
				if !strings.HasPrefix(lines[0], "APP_KEY=") {
					t.Error("APP_KEY should be prepended as first line")
				}
				if !strings.Contains(string(content), "DB_HOST=localhost") {
					t.Error("Existing content should be preserved")
				}
			},
		},
		{
			name: "returns error when env is a directory",
			setupEnv: func(t *testing.T) string {
				tmpDir := t.TempDir()
				os.MkdirAll(tmpDir+"/.env", 0755)
				return tmpDir
			},
			wantErr: true,
			validateResult: func(t *testing.T, envPath string) {
				// Error expected, no validation needed
			},
		},
		{
			name: "returns error when env file is not writable",
			setupEnv: func(t *testing.T) string {
				tmpDir := t.TempDir()
				os.WriteFile(tmpDir+"/.env", []byte("APP_KEY=old\n"), 0644)
				os.Chmod(tmpDir+"/.env", 0444)
				t.Cleanup(func() {
					os.Chmod(tmpDir+"/.env", 0644)
				})
				return tmpDir
			},
			wantErr: true,
			validateResult: func(t *testing.T, envPath string) {
				// Error expected, no validation needed
			},
		},
		{
			name: "returns error when directory is read-only and env does not exist",
			setupEnv: func(t *testing.T) string {
				tmpDir := t.TempDir()
				readOnlyDir := tmpDir + "/readonly"
				os.MkdirAll(readOnlyDir, 0755)
				os.Chmod(readOnlyDir, 0555)
				t.Cleanup(func() {
					os.Chmod(readOnlyDir, 0755)
				})
				return readOnlyDir
			},
			wantErr: true,
			validateResult: func(t *testing.T, envPath string) {
				// Error expected, no validation needed
			},
		},
		{
			name: "handles empty env file",
			setupEnv: func(t *testing.T) string {
				tmpDir := t.TempDir()
				os.WriteFile(tmpDir+"/.env", []byte(""), 0644)
				return tmpDir
			},
			wantErr: false,
			validateResult: func(t *testing.T, envPath string) {
				content, err := os.ReadFile(envPath)
				if err != nil {
					t.Fatalf("Failed to read .env: %v", err)
				}
				if !strings.Contains(string(content), "APP_KEY=") {
					t.Error("APP_KEY should be added to empty file")
				}
			},
		},
		{
			name: "handles env file with only whitespace",
			setupEnv: func(t *testing.T) string {
				tmpDir := t.TempDir()
				os.WriteFile(tmpDir+"/.env", []byte("   \n\n  \n"), 0644)
				return tmpDir
			},
			wantErr: false,
			validateResult: func(t *testing.T, envPath string) {
				content, err := os.ReadFile(envPath)
				if err != nil {
					t.Fatalf("Failed to read .env: %v", err)
				}
				if !strings.Contains(string(content), "APP_KEY=") {
					t.Error("APP_KEY should be added")
				}
			},
		},
		{
			name: "handles env file with app key as first line",
			setupEnv: func(t *testing.T) string {
				tmpDir := t.TempDir()
				envContent := "APP_KEY=oldkey\nDB_HOST=localhost"
				os.WriteFile(tmpDir+"/.env", []byte(envContent), 0644)
				return tmpDir
			},
			wantErr: false,
			validateResult: func(t *testing.T, envPath string) {
				content, err := os.ReadFile(envPath)
				if err != nil {
					t.Fatalf("Failed to read .env: %v", err)
				}
				if strings.Contains(string(content), "oldkey") {
					t.Error("Old key should be replaced")
				}
				lines := strings.Split(string(content), "\n")
				if !strings.HasPrefix(lines[0], "APP_KEY=") {
					t.Error("APP_KEY should still be first line")
				}
			},
		},
		{
			name: "handles env file with app key as last line",
			setupEnv: func(t *testing.T) string {
				tmpDir := t.TempDir()
				envContent := "DB_HOST=localhost\nDB_PORT=5432\nAPP_KEY=oldkey"
				os.WriteFile(tmpDir+"/.env", []byte(envContent), 0644)
				return tmpDir
			},
			wantErr: false,
			validateResult: func(t *testing.T, envPath string) {
				content, err := os.ReadFile(envPath)
				if err != nil {
					t.Fatalf("Failed to read .env: %v", err)
				}
				if strings.Contains(string(content), "oldkey") {
					t.Error("Old key should be replaced")
				}
			},
		},
		{
			name: "handles env file with multiple lines containing app prefix",
			setupEnv: func(t *testing.T) string {
				tmpDir := t.TempDir()
				envContent := "APP_NAME=myapp\nAPP_KEY=oldkey\nAPP_ENV=production"
				os.WriteFile(tmpDir+"/.env", []byte(envContent), 0644)
				return tmpDir
			},
			wantErr: false,
			validateResult: func(t *testing.T, envPath string) {
				content, err := os.ReadFile(envPath)
				if err != nil {
					t.Fatalf("Failed to read .env: %v", err)
				}
				if strings.Contains(string(content), "oldkey") {
					t.Error("Old key should be replaced")
				}
				if !strings.Contains(string(content), "APP_NAME=myapp") {
					t.Error("APP_NAME should be preserved")
				}
				if !strings.Contains(string(content), "APP_ENV=production") {
					t.Error("APP_ENV should be preserved")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := tt.setupEnv(t)
			originalDir, _ := os.Getwd()
			os.Chdir(tmpDir)
			defer os.Chdir(originalDir)

			err := runKeyGenerate(nil, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("runKeyGenerate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				tt.validateResult(t, tmpDir+"/.env")
			}
		})
	}
}

// TestRunKeyGenerate_KeyProperties verifies the cryptographic properties of generated keys
func TestRunKeyGenerate_KeyProperties(t *testing.T) {
	tests := []struct {
		name     string
		validate func(t *testing.T, key string)
	}{
		{
			name: "generates key with correct length",
			validate: func(t *testing.T, key string) {
				decoded, err := base64.StdEncoding.DecodeString(key)
				if err != nil {
					t.Fatalf("Key is not valid base64: %v", err)
				}
				if len(decoded) != 32 {
					t.Errorf("Key length = %d bytes, want 32 bytes", len(decoded))
				}
			},
		},
		{
			name: "generates key with valid base64 encoding",
			validate: func(t *testing.T, key string) {
				if _, err := base64.StdEncoding.DecodeString(key); err != nil {
					t.Errorf("Key is not valid base64: %v", err)
				}
			},
		},
		{
			name: "generates non-empty key",
			validate: func(t *testing.T, key string) {
				if key == "" {
					t.Error("Generated key should not be empty")
				}
			},
		},
		{
			name: "generates key without whitespace",
			validate: func(t *testing.T, key string) {
				if strings.Contains(key, " ") || strings.Contains(key, "\n") || strings.Contains(key, "\t") {
					t.Error("Generated key should not contain whitespace")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

			key := strings.TrimPrefix(strings.TrimSpace(string(content)), "APP_KEY=")
			tt.validate(t, key)
		})
	}
}

// TestRunKeyGenerate_FileFormatPreservation tests that file formatting is preserved
func TestRunKeyGenerate_FileFormatPreservation(t *testing.T) {
	tests := []struct {
		name            string
		existingContent string
		checkFormat     func(t *testing.T, content string)
	}{
		{
			name:            "preserves line endings in existing content",
			existingContent: "DB_HOST=localhost\nDB_PORT=5432\n",
			checkFormat: func(t *testing.T, content string) {
				if !strings.Contains(content, "DB_HOST=localhost\n") {
					t.Error("Line endings should be preserved")
				}
			},
		},
		{
			name:            "preserves comments when adding key",
			existingContent: "# Database configuration\nDB_HOST=localhost\n",
			checkFormat: func(t *testing.T, content string) {
				if !strings.Contains(content, "# Database configuration") {
					t.Error("Comments should be preserved")
				}
			},
		},
		{
			name:            "preserves empty lines",
			existingContent: "DB_HOST=localhost\n\nDB_PORT=5432\n",
			checkFormat: func(t *testing.T, content string) {
				lines := strings.Split(content, "\n")
				foundEmpty := false
				for _, line := range lines {
					if line == "" {
						foundEmpty = true
						break
					}
				}
				if !foundEmpty {
					t.Error("Empty lines should be preserved")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			originalDir, _ := os.Getwd()
			os.Chdir(tmpDir)
			defer os.Chdir(originalDir)

			os.WriteFile(".env", []byte(tt.existingContent), 0644)

			err := runKeyGenerate(nil, nil)
			if err != nil {
				t.Fatalf("runKeyGenerate() error = %v", err)
			}

			content, err := os.ReadFile(".env")
			if err != nil {
				t.Fatalf("Failed to read .env: %v", err)
			}

			tt.checkFormat(t, string(content))
		})
	}
}

// TestRunKeyGenerate_Randomness verifies that multiple invocations produce different keys
func TestRunKeyGenerate_Randomness(t *testing.T) {
	const iterations = 10
	keys := make(map[string]bool)

	for i := 0; i < iterations; i++ {
		tmpDir := t.TempDir()
		originalDir, _ := os.Getwd()
		os.Chdir(tmpDir)

		err := runKeyGenerate(nil, nil)
		os.Chdir(originalDir)

		if err != nil {
			t.Fatalf("runKeyGenerate() error = %v", err)
		}

		content, err := os.ReadFile(tmpDir + "/.env")
		if err != nil {
			t.Fatalf("Failed to read .env: %v", err)
		}

		key := strings.TrimPrefix(strings.TrimSpace(string(content)), "APP_KEY=")
		if keys[key] {
			t.Errorf("Duplicate key generated: %s", key)
		}
		keys[key] = true
	}

	if len(keys) != iterations {
		t.Errorf("Expected %d unique keys, got %d", iterations, len(keys))
	}
}
