package vel

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
	"github.com/velocitykode/velocity-cli/internal/ui"
)

var (
	servePort      string
	serveEnv       string
	serveWatch     bool
	serveBuildTags string
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the development server",
	Long: `Start the Velocity development server with optional hot reload.

The server will automatically reload when Go files change if --watch is enabled.`,
	RunE: runServe,
}

func init() {
	serveCmd.Flags().StringVarP(&servePort, "port", "p", "4000", "Port to run the server on")
	serveCmd.Flags().StringVarP(&serveEnv, "env", "e", "development", "Environment to run in")
	serveCmd.Flags().BoolVarP(&serveWatch, "watch", "w", true, "Enable hot reload")
	serveCmd.Flags().StringVar(&serveBuildTags, "tags", "", "Build tags to pass to go build")
}

var viteCmd *exec.Cmd

func runServe(cmd *cobra.Command, args []string) error {
	ui.Header("serve")
	if serveWatch {
		ui.Info(fmt.Sprintf("Starting on port %s with hot reload...", ui.Highlight(servePort)))
	} else {
		ui.Info(fmt.Sprintf("Starting on port %s...", ui.Highlight(servePort)))
	}
	ui.Newline()

	// Start Vite dev server if package.json exists
	if _, err := os.Stat("package.json"); err == nil {
		startVite()
	}

	// Handle graceful shutdown
	setupGracefulShutdown()

	if serveWatch {
		return runWithWatcher()
	}
	return runServer()
}

func startVite() {
	// Detect package manager
	runner := "npm"
	if _, err := exec.LookPath("bun"); err == nil {
		if _, err := os.Stat("bun.lock"); err == nil {
			runner = "bun"
		}
	}

	ui.Step(fmt.Sprintf("Starting Vite (%s run dev)...", runner))
	viteCmd = exec.Command(runner, "run", "dev")
	viteCmd.Stdout = os.Stdout
	viteCmd.Stderr = os.Stderr
	viteCmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := viteCmd.Start(); err != nil {
		ui.Warning(fmt.Sprintf("Failed to start Vite: %v", err))
	}
}

func setupGracefulShutdown() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		ui.Newline()
		ui.Step("Shutting down...")

		if viteCmd != nil && viteCmd.Process != nil {
			syscall.Kill(-viteCmd.Process.Pid, syscall.SIGTERM)
		}

		os.Exit(0)
	}()
}

func runServer() error {
	os.Setenv("APP_ENV", serveEnv)
	os.Setenv("APP_PORT", servePort)

	// Ensure build directory exists
	os.MkdirAll(".vel/tmp", 0755)

	buildArgs := []string{"build", "-o", ".vel/tmp/server"}
	if serveBuildTags != "" {
		buildArgs = append(buildArgs, "-tags", serveBuildTags)
	}
	buildArgs = append(buildArgs, ".")

	buildCmd := exec.Command("go", buildArgs...)
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr

	if err := buildCmd.Run(); err != nil {
		ui.Error(fmt.Sprintf("Build failed: %v", err))
		return err
	}

	serverCmd := exec.Command(".vel/tmp/server")
	serverCmd.Stdout = os.Stdout
	serverCmd.Stderr = os.Stderr
	serverCmd.Env = os.Environ()

	if err := serverCmd.Run(); err != nil {
		ui.Error(fmt.Sprintf("Server failed: %v", err))
		return err
	}
	return nil
}

func runWithWatcher() error {
	os.MkdirAll(".vel/tmp", 0755)

	rebuild := make(chan bool, 1)
	errChan := make(chan error, 1)

	go func() {
		if err := watchFiles(rebuild); err != nil {
			errChan <- err
		}
	}()

	var serverCmd *exec.Cmd
	var mu sync.Mutex

	startServer := func() error {
		mu.Lock()
		defer mu.Unlock()

		if serverCmd != nil && serverCmd.Process != nil {
			ui.Step("Stopping server...")
			serverCmd.Process.Kill()
			serverCmd.Wait()
		}

		ui.Step("Building...")
		buildArgs := []string{"build", "-o", ".vel/tmp/server"}
		if serveBuildTags != "" {
			buildArgs = append(buildArgs, "-tags", serveBuildTags)
		}
		buildArgs = append(buildArgs, ".")

		buildCmd := exec.Command("go", buildArgs...)
		if output, err := buildCmd.CombinedOutput(); err != nil {
			ui.Error("Build failed:")
			ui.Muted(string(output))
			return err
		}

		ui.Success(fmt.Sprintf("Starting server on port %s...", servePort))
		serverCmd = exec.Command(".vel/tmp/server")
		serverCmd.Stdout = os.Stdout
		serverCmd.Stderr = os.Stderr
		serverCmd.Env = append(os.Environ(),
			fmt.Sprintf("APP_ENV=%s", serveEnv),
			fmt.Sprintf("APP_PORT=%s", servePort),
		)

		if err := serverCmd.Start(); err != nil {
			ui.Error(fmt.Sprintf("Failed to start server: %v", err))
			return err
		}
		return nil
	}

	// Initial build and start
	if err := startServer(); err != nil {
		return err
	}

	// Watch for rebuilds
	for {
		select {
		case err := <-errChan:
			return err
		case <-rebuild:
			ui.Newline()
			ui.Warning("File changed, reloading...")
			time.Sleep(100 * time.Millisecond)
			startServer() // Ignore error on reload, keep watching
		}
	}
}

func watchFiles(rebuild chan bool) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}
	defer watcher.Close()

	err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.Contains(path, "vendor") || strings.Contains(path, ".vel") {
			return filepath.SkipDir
		}

		if info.IsDir() {
			return watcher.Add(path)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to setup watcher: %w", err)
	}

	var debounce *time.Timer

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}

			if !strings.HasSuffix(event.Name, ".go") {
				continue
			}

			if debounce != nil {
				debounce.Stop()
			}
			debounce = time.AfterFunc(500*time.Millisecond, func() {
				select {
				case rebuild <- true:
				default:
				}
			})

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			ui.Error(fmt.Sprintf("Watcher error: %v", err))
		}
	}
}
