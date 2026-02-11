// Package vel provides an importable CLI for Velocity projects.
// User projects import this package in cmd/vel/main.go to get
// access to commands like serve, migrate, make:handler, etc.
package vel

import (
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/velocitykode/vel/internal/ui"
)

// osExecutable is a variable for testing - allows mocking os.Executable
var osExecutable = os.Executable

// Version is the CLI version
var Version = "0.8.6"

var rootCmd *cobra.Command

func init() {
	initRootCmd()
}

func initRootCmd() {
	rootCmd = &cobra.Command{
		Use:           "vel",
		Short:         "Vel - Development tools for Velocity projects",
		Version:       Version,
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	rootCmd.CompletionOptions.DisableDefaultCmd = true

	// Register all commands
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(migrateCmd)
	rootCmd.AddCommand(migrateFreshCmd)
	rootCmd.AddCommand(migrateRollbackCmd)
	rootCmd.AddCommand(makeHandlerCmd)
	rootCmd.AddCommand(keyGenerateCmd)
}

// Execute runs the CLI
func Execute() error {
	if rootCmd == nil {
		initRootCmd()
	}

	// Check if rebuild is needed
	if needsRebuild() {
		rebuildSelf()
	}

	err := rootCmd.Execute()
	if err != nil && err.Error() != "" {
		ui.Error(err.Error())
	}
	return err
}

// needsRebuild checks if the vel binary needs to be rebuilt
func needsRebuild() bool {
	velBinary, err := osExecutable()
	if err != nil {
		return false
	}

	velInfo, err := os.Stat(velBinary)
	if err != nil {
		return false
	}

	binTime := velInfo.ModTime()

	// Check directories that affect the CLI
	dirs := []string{"cmd/vel", "bootstrap", "database/migrations"}
	for _, dir := range dirs {
		if hasNewerFiles(dir, binTime) {
			return true
		}
	}

	// Check go.mod/go.sum
	for _, f := range []string{"go.mod", "go.sum"} {
		if info, err := os.Stat(f); err == nil {
			if info.ModTime().After(binTime) {
				return true
			}
		}
	}

	return false
}

// hasNewerFiles checks if any .go files in dir are newer than t
func hasNewerFiles(dir string, t time.Time) bool {
	found := false
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && filepath.Ext(path) == ".go" {
			if info.ModTime().After(t) {
				found = true
				return filepath.SkipAll
			}
		}
		return nil
	})
	return found
}

// rebuildSelf rebuilds the vel binary
func rebuildSelf() {
	ui.Muted("Rebuilding vel...")
	cmd := exec.Command("go", "build", "-o", "vel", "./cmd/vel")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		ui.Warning("Rebuild failed: " + err.Error())
	}
}

// RootCmd returns the root command for testing
func RootCmd() *cobra.Command {
	return rootCmd
}

// trigger
