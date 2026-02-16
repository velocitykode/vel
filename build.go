package vel

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/velocitykode/velocity-cli/internal/ui"
)

var (
	buildOutput string
	buildOS     string
	buildArch   string
	buildTags   string
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build the application for production",
	Long:  `Build the Velocity application for production deployment.`,
	RunE:  runBuild,
}

func init() {
	buildCmd.Flags().StringVarP(&buildOutput, "output", "o", "", "Output binary name")
	buildCmd.Flags().StringVar(&buildOS, "os", runtime.GOOS, "Target operating system")
	buildCmd.Flags().StringVar(&buildArch, "arch", runtime.GOARCH, "Target architecture")
	buildCmd.Flags().StringVar(&buildTags, "tags", "", "Build tags")
}

func runBuild(cmd *cobra.Command, args []string) error {
	ui.Header("build")

	// Determine output name
	output := buildOutput
	if output == "" {
		// Use current directory name
		cwd, _ := os.Getwd()
		output = filepath.Base(cwd)
		if buildOS == "windows" {
			output += ".exe"
		}
	}

	ui.Info(fmt.Sprintf("Building for %s/%s...", buildOS, buildArch))

	// Set environment for cross-compilation
	env := os.Environ()
	env = append(env, fmt.Sprintf("GOOS=%s", buildOS))
	env = append(env, fmt.Sprintf("GOARCH=%s", buildArch))
	env = append(env, "CGO_ENABLED=0")

	// Build command
	buildArgs := []string{"build", "-o", output, "-ldflags", "-s -w"}
	if buildTags != "" {
		buildArgs = append(buildArgs, "-tags", buildTags)
	}
	buildArgs = append(buildArgs, ".")

	buildCmd := exec.Command("go", buildArgs...)
	buildCmd.Env = env
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr

	if err := buildCmd.Run(); err != nil {
		ui.Error(fmt.Sprintf("Build failed: %v", err))
		return err
	}

	ui.Success(fmt.Sprintf("Built: %s", output))
	return nil
}
