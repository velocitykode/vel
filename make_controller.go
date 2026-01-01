package vel

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"

	"github.com/spf13/cobra"
	"github.com/velocitykode/vel/internal/stubs"
	"github.com/velocitykode/vel/internal/ui"
)

var (
	makeControllerResource bool
	makeControllerAPI      bool
)

var makeControllerCmd = &cobra.Command{
	Use:     "make:controller [name]",
	Short:   "Create a new controller",
	Long:    `Create a new controller class in the app/http/controllers directory.`,
	Example: "  velocity make:controller User\n  velocity make:controller Admin/Dashboard --resource",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			ui.Error("controller name is required")
			ui.Newline()
			ui.Muted("Usage:")
			ui.Muted("  velocity make:controller [name]")
			ui.Newline()
			ui.Muted("Examples:")
			ui.Muted("  velocity make:controller User")
			ui.Muted("  velocity make:controller Admin/Dashboard --resource")
			return fmt.Errorf("") // Return empty error to exit with code 1
		}
		if len(args) > 1 {
			return fmt.Errorf("too many arguments, expected only controller name")
		}
		return nil
	},
	RunE: runMakeController,
}

func init() {
	makeControllerCmd.Flags().BoolVarP(&makeControllerResource, "resource", "r", false, "Generate a resource controller with CRUD methods")
	makeControllerCmd.Flags().BoolVar(&makeControllerAPI, "api", false, "Generate an API controller (JSON responses)")
}

func runMakeController(cmd *cobra.Command, args []string) error {
	name := args[0]

	ui.Header("make:controller")

	// Normalize name
	controllerName := toControllerName(name)

	// Determine package and path
	packageName := "controllers"
	outputDir := "app/http/controllers"

	// Check if name contains path separator
	if strings.Contains(name, "/") {
		parts := strings.Split(name, "/")
		controllerName = toControllerName(parts[len(parts)-1])
		packageName = strings.ToLower(parts[len(parts)-2])
		// Lowercase all path parts for conventional Go package directories
		for i := range parts[:len(parts)-1] {
			parts[i] = strings.ToLower(parts[i])
		}
		outputDir = filepath.Join("app/http/controllers", filepath.Join(parts[:len(parts)-1]...))
	}

	// Create directory if needed
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		ui.Error(fmt.Sprintf("Failed to create directory: %v", err))
		return err
	}

	// Generate filename
	filename := toSnakeCase(controllerName) + "_controller.go"
	outputPath := filepath.Join(outputDir, filename)

	// Check if file exists
	if _, err := os.Stat(outputPath); err == nil {
		ui.Error(fmt.Sprintf("Controller already exists: %s", outputPath))
		return fmt.Errorf("controller already exists")
	}

	// Load stub
	stubContent, err := stubs.Get("app/http/controllers/controller.go.stub")
	if err != nil {
		// Fallback to basic template if stub not found
		stubContent = []byte(`package {{ .Package }}

import "github.com/velocitykode/velocity/pkg/router"

func {{ .ControllerName }}(ctx *router.Context) error {
	return ctx.String(200, "{{ .ControllerName }}")
}
`)
	}

	// Parse and execute template
	tmpl, err := template.New("controller").Parse(string(stubContent))
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to parse template: %v", err))
		return err
	}

	data := map[string]interface{}{
		"Package":        packageName,
		"ControllerName": controllerName,
		"Resource":       makeControllerResource,
		"API":            makeControllerAPI,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		ui.Error(fmt.Sprintf("Failed to execute template: %v", err))
		return err
	}

	// Write file
	if err := os.WriteFile(outputPath, buf.Bytes(), 0644); err != nil {
		ui.Error(fmt.Sprintf("Failed to write file: %v", err))
		return err
	}

	ui.Success(fmt.Sprintf("Created: %s", outputPath))
	return nil
}

func toControllerName(name string) string {
	// Remove "Controller" suffix if present
	name = strings.TrimSuffix(name, "Controller")
	name = strings.TrimSuffix(name, "controller")

	// Convert to PascalCase
	return toPascalCase(name)
}

func toPascalCase(s string) string {
	words := splitWords(s)
	for i, word := range words {
		words[i] = strings.Title(strings.ToLower(word))
	}
	return strings.Join(words, "")
}

func toSnakeCase(s string) string {
	var result []rune
	for i, r := range s {
		if unicode.IsUpper(r) && i > 0 {
			result = append(result, '_')
		}
		result = append(result, unicode.ToLower(r))
	}
	return string(result)
}

func splitWords(s string) []string {
	var words []string
	var current []rune

	for _, r := range s {
		if r == '_' || r == '-' || r == ' ' {
			if len(current) > 0 {
				words = append(words, string(current))
				current = nil
			}
		} else if unicode.IsUpper(r) && len(current) > 0 {
			words = append(words, string(current))
			current = []rune{r}
		} else {
			current = append(current, r)
		}
	}

	if len(current) > 0 {
		words = append(words, string(current))
	}

	return words
}
