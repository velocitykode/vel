package vel

import (
	"testing"
)

func TestExecute_Initializes(t *testing.T) {
	rootCmd = nil

	err := Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if rootCmd == nil {
		t.Fatal("Execute() should initialize rootCmd")
	}
}

func TestExecute_Idempotent(t *testing.T) {
	rootCmd = nil

	// First call initializes
	Execute()
	firstCmd := rootCmd

	// Second call should reuse
	Execute()
	if rootCmd != firstCmd {
		t.Error("Execute() should reuse existing rootCmd")
	}
}

func TestRootCmd_RegistersSubcommands(t *testing.T) {
	rootCmd = nil
	Execute()

	cmd := RootCmd()
	commands := make(map[string]bool)
	for _, c := range cmd.Commands() {
		commands[c.Name()] = true
	}

	required := []string{"serve", "build", "migrate", "migrate:fresh", "make:controller", "key:generate"}
	for _, name := range required {
		if !commands[name] {
			t.Errorf("Missing required command: %s", name)
		}
	}
}

func TestVersion_IsSet(t *testing.T) {
	if Version == "" {
		t.Error("Version should not be empty")
	}
}
