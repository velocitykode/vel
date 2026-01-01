package vel

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/velocitykode/vel/internal/ui"
)

var keyGenerateCmd = &cobra.Command{
	Use:   "key:generate",
	Short: "Generate a new application key",
	Long:  `Generate a new random application key and update the .env file.`,
	RunE:  runKeyGenerate,
}

func runKeyGenerate(cmd *cobra.Command, args []string) error {
	ui.Header("key:generate")

	// Generate 32-byte key
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		ui.Error(fmt.Sprintf("Failed to generate key: %v", err))
		return err
	}

	// Encode to base64
	encodedKey := base64.StdEncoding.EncodeToString(key)

	// Read .env file
	envPath := ".env"
	content, err := os.ReadFile(envPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create .env with key
			content = []byte(fmt.Sprintf("APP_KEY=%s\n", encodedKey))
			if err := os.WriteFile(envPath, content, 0644); err != nil {
				ui.Error(fmt.Sprintf("Failed to create .env: %v", err))
				return err
			}
			ui.Success(fmt.Sprintf("Created .env with APP_KEY"))
			return nil
		}
		ui.Error(fmt.Sprintf("Failed to read .env: %v", err))
		return err
	}

	// Update APP_KEY in content
	lines := strings.Split(string(content), "\n")
	found := false
	for i, line := range lines {
		if strings.HasPrefix(line, "APP_KEY=") {
			lines[i] = fmt.Sprintf("APP_KEY=%s", encodedKey)
			found = true
			break
		}
	}

	if !found {
		// Add APP_KEY if not present
		lines = append([]string{fmt.Sprintf("APP_KEY=%s", encodedKey)}, lines...)
	}

	// Write back
	newContent := strings.Join(lines, "\n")
	if err := os.WriteFile(envPath, []byte(newContent), 0644); err != nil {
		ui.Error(fmt.Sprintf("Failed to update .env: %v", err))
		return err
	}

	ui.Success("Application key set successfully")
	return nil
}
