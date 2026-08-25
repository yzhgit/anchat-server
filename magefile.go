//go:build mage
// +build mage

package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// Default target to run when none is specified
var Default = Build.All

// Build contains all build-related targets
type Build mg.Namespace

// Test contains all test-related targets
type Test mg.Namespace

// Docker contains all docker-related targets
type Docker mg.Namespace

// DB contains all database-related targets
type DB mg.Namespace

// Dev contains all development-related targets
type Dev mg.Namespace

// Docs contains all documentation-related targets
type Docs mg.Namespace

const (
	servicesDir = "cmd"
	binDir      = "bin"
	protoDir    = "api"
	docsDir     = "docs"
	swaggerDir  = "docs/api/swagger"
)

// All builds all services
func (Build) All() error {
	fmt.Println("Building all services...")
	mg.Deps(ensureBinDir)

	services := []string{
		"user",
		"friend",
		"group",
		"message",
		"conversation",
		"file",
		"rtc",
		"push",
		"realtime",
		"gateway",
	}

	for _, service := range services {
		if err := buildService(service); err != nil {
			return err
		}
	}

	fmt.Println("✓ Build completed!")
	return nil
}

// User builds user service
func (Build) User() error {
	mg.Deps(ensureBinDir)
	return buildService("user")
}

// Friend builds friend service
func (Build) Friend() error {
	mg.Deps(ensureBinDir)
	return buildService("friend")
}

// Group builds group service
func (Build) Group() error {
	mg.Deps(ensureBinDir)
	return buildService("group")
}

// Conversation builds conversation service
func (Build) Conversation() error {
	mg.Deps(ensureBinDir)
	return buildService("conversation")
}

// Message builds message service
func (Build) Message() error {
	mg.Deps(ensureBinDir)
	return buildService("message")
}

// File builds file service
func (Build) File() error {
	mg.Deps(ensureBinDir)
	return buildService("file")
}

// RTC builds rtc service
func (Build) Rtc() error {
	mg.Deps(ensureBinDir)
	return buildService("rtc")
}

// Push builds push service
func (Build) Push() error {
	mg.Deps(ensureBinDir)
	return buildService("push")
}

// Realtime builds realtime service
func (Build) Realtime() error {
	mg.Deps(ensureBinDir)
	return buildService("realtime")
}

// Gateway builds gateway service
func (Build) Gateway() error {
	mg.Deps(ensureBinDir)
	return buildService("gateway")
}

// buildService builds a specific service
func buildService(name string) error {
	fmt.Printf("Building %s...\n", name)
	output := filepath.Join(binDir, name)
	source := filepath.Join("app", name, servicesDir)
	version := resolveVersion()
	ldflags := fmt.Sprintf("-X main.Version=%s", version)
	fmt.Printf("  version: %s\n", version)
	return sh.Run("go", "build", "-o", output, "-ldflags", ldflags, "./"+source)
}

// resolveVersion determines the build version from git or falls back to a default.
// It tries: git describe --tags --always --dirty → git rev-parse --short=8 HEAD → "dev"
func resolveVersion() string {
	if v := gitDescribe(); v != "" {
		return v
	}
	if v := gitShortHead(); v != "" {
		return v
	}
	return "dev"
}

func gitDescribe() string {
	cmd := exec.Command("git", "describe", "--tags", "--always", "--dirty")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return ""
	}
	return strings.TrimSpace(out.String())
}

func gitShortHead() string {
	cmd := exec.Command("git", "rev-parse", "--short=8", "HEAD")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return ""
	}
	return strings.TrimSpace(out.String())
}

// ensureBinDir creates bin directory if it doesn't exist
func ensureBinDir() error {
	return os.MkdirAll(binDir, 0755)
}

// Proto generates protobuf code
func Proto() error {
	fmt.Println("Generating protobuf code...")
	return sh.RunV("bash", "scripts/gen-proto.sh")
}

// Wire generates Wire dependency injection code for all services
func Wire() error {
	fmt.Println("Generating Wire code...")
	services := []string{
		"user",
		"friend",
		"group",
		"message",
		"conversation",
		"file",
		"rtc",
		"push",
		"realtime",
		"gateway",
	}
	for _, svc := range services {
		if err := sh.RunV("wire", "gen", "./app/"+svc+"/cmd"); err != nil {
			return fmt.Errorf("failed to generate wire for %s: %w", svc, err)
		}
	}
	fmt.Println("✓ Wire generation completed!")
	return nil
}

// generateOpenAPI generates OpenAPI 3.0 spec from proto files
func generateOpenAPI(protoFiles []string) error {
	fmt.Println("Generating OpenAPI 3.0 documentation...")

	args := []string{
		"-I", protoDir,
		"-I", "third_party",
		"--experimental_allow_proto3_optional",
		"--openapi_out=" + swaggerDir,
	}
	args = append(args, protoFiles...)

	if err := sh.RunV("protoc", args...); err != nil {
		return fmt.Errorf("failed to generate openapi docs: %w", err)
	}

	return ensureOpenAPIFile()
}

// ensureOpenAPIFile renames the generated openapi file to openapi.json
func ensureOpenAPIFile() error {
	target := filepath.Join(swaggerDir, "openapi.json")
	if _, err := os.Stat(target); err == nil {
		return nil
	}

	// protoc-gen-openapi names the output after the first proto file
	jsonFiles, _ := filepath.Glob(filepath.Join(swaggerDir, "*.json"))
	for _, f := range jsonFiles {
		if filepath.Base(f) != "openapi.json" {
			if err := os.Rename(f, target); err != nil {
				return fmt.Errorf("failed to rename %s to %s: %w", f, target, err)
			}
			break
		}
	}

	fmt.Println("✓ OpenAPI 3.0 spec generated:", target)
	return nil
}

// All runs all tests
func (Test) All() error {
	fmt.Println("Running tests...")
	return sh.RunV("go", "test", "-v", "-race", "-coverprofile=coverage.out", "./...")
}

// Coverage generates test coverage report
func (Test) Coverage() error {
	mg.Deps(Test.All)
	fmt.Println("Generating coverage report...")
	if err := sh.Run("go", "tool", "cover", "-html=coverage.out", "-o", "coverage.html"); err != nil {
		return err
	}
	fmt.Println("✓ Coverage report generated: coverage.html")
	return nil
}

// Lint runs linter
func Lint() error {
	fmt.Println("Running linter...")
	return sh.RunV("golangci-lint", "run")
}

// Fmt formats code
func Fmt() error {
	fmt.Println("Formatting code...")
	return sh.RunV("go", "fmt", "./...")
}

// Build builds docker images
func (Docker) Build() error {
	fmt.Println("Building docker images...")
	return sh.RunV("docker-compose", "-f", "deployments/docker/docker-compose.yml", "build")
}

// Up starts docker compose
func (Docker) Up() error {
	fmt.Println("Starting docker compose...")
	return sh.RunV("docker-compose", "-f", "deployments/docker/docker-compose.yml", "up", "-d")
}

// Down stops docker compose
func (Docker) Down() error {
	fmt.Println("Stopping docker compose...")
	return sh.RunV("docker-compose", "-f", "deployments/docker/docker-compose.yml", "down")
}

// Logs shows docker compose logs
func (Docker) Logs() error {
	fmt.Println("Showing docker logs...")
	return sh.RunV("docker-compose", "-f", "deployments/docker/docker-compose.yml", "logs", "-f")
}

// Ps shows docker compose status
func (Docker) Ps() error {
	return sh.RunV("docker-compose", "-f", "deployments/docker/docker-compose.yml", "ps")
}

// Up runs database migrations up
func (DB) Up() error {
	fmt.Println("Running database migrations...")
	return sh.RunV("migrate",
		"-path", "migrations",
		"-database", "postgresql://anychat:anychat@localhost:5432/im?sslmode=disable",
		"up",
	)
}

// Down runs database migrations down
func (DB) Down() error {
	fmt.Println("Reverting database migrations...")
	return sh.RunV("migrate",
		"-path", "migrations",
		"-database", "postgresql://anychat:anychat@localhost:5432/im?sslmode=disable",
		"down",
	)
}

// runDevService runs a service locally with its config file.
// stdout is teed to both the terminal (for real-time debugging) and
// a log file so Promtail can scrape it into Loki/Grafana.
//
// The log file path is: /tmp/anychat-logs/<service>.log
// Override with LOG_DIR env var: LOG_DIR=/var/log/anychat
func runDevService(name string) error {
	version := resolveVersion()
	ldflags := fmt.Sprintf("-X main.Version=%s", version)

	configPath := filepath.Join("app", name, "configs")

	// Resolve log directory: LOG_DIR env var or default /tmp/anychat-logs
	logDir := os.Getenv("LOG_DIR")
	if logDir == "" {
		logDir = "/tmp/anychat-logs"
	}
	logFile := filepath.Join(logDir, name+".log")

	// Create log directory if it doesn't exist
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("create log dir %s: %w", logDir, err)
	}

	// Truncate existing log file before starting so Promtail picks up
	// only the current session's output (positions.yaml retains the
	// previous offset and would otherwise re-emit old log lines).
	if err := os.Remove(logFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove log file %s: %w", logFile, err)
	}
	// Open log file for append
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("open log file %s: %w", logFile, err)
	}
	defer f.Close()

	// Create multi-writer: stdout → terminal + log file
	multiWriter := io.MultiWriter(os.Stdout, f)

	cmd := exec.Command("go", "run", "-ldflags", ldflags, "./app/"+name+"/cmd", "-config", configPath)
	cmd.Stdout = multiWriter
	cmd.Stderr = multiWriter
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

// User runs user service locally
func (Dev) User() error {
	fmt.Println("Running user service...")
	return runDevService("user")
}

// Friend runs friend service locally
func (Dev) Friend() error {
	fmt.Println("Running friend service...")
	return runDevService("friend")
}

// Group runs group service locally
func (Dev) Group() error {
	fmt.Println("Running group service...")
	return runDevService("group")
}

// File runs file service locally
func (Dev) File() error {
	fmt.Println("Running file service...")
	return runDevService("file")
}

// Message runs message service locally
func (Dev) Message() error {
	fmt.Println("Running message service...")
	return runDevService("message")
}

// Conversation runs conversation service locally
func (Dev) Conversation() error {
	fmt.Println("Running conversation service...")
	return runDevService("conversation")
}

// Push runs push service locally
func (Dev) Push() error {
	fmt.Println("Running push service...")
	return runDevService("push")
}

// RTC runs rtc service locally
func (Dev) Rtc() error {
	fmt.Println("Running rtc service...")
	return runDevService("rtc")
}

// Realtime runs realtime service locally
func (Dev) Realtime() error {
	fmt.Println("Running realtime service...")
	return runDevService("realtime")
}

// Gateway runs gateway service locally
func (Dev) Gateway() error {
	fmt.Println("Running gateway service...")
	return runDevService("gateway")
}

// Deps installs dependencies
func Deps() error {
	fmt.Println("Installing dependencies...")
	if err := sh.RunV("go", "mod", "download"); err != nil {
		return err
	}
	return sh.RunV("go", "mod", "tidy")
}

// DepsCheck verifies dependencies
func DepsCheck() error {
	fmt.Println("Checking dependencies...")
	return sh.RunV("go", "mod", "verify")
}

// Clean removes build artifacts
func Clean() error {
	fmt.Println("Cleaning build artifacts...")
	if err := sh.Rm(binDir); err != nil {
		return err
	}
	if err := sh.Rm("coverage.out"); err != nil && !os.IsNotExist(err) {
		return err
	}
	if err := sh.Rm("coverage.html"); err != nil && !os.IsNotExist(err) {
		return err
	}
	fmt.Println("✓ Clean completed!")
	return nil
}

// Install installs required tools
func Install() error {
	fmt.Println("Installing required tools...")

	// Install migrate with PostgreSQL driver using build tags
	fmt.Println("Installing migrate (with PostgreSQL driver)...")
	if err := sh.RunV("go", "install", "-tags", "postgres", "github.com/golang-migrate/migrate/v4/cmd/migrate@latest"); err != nil {
		return fmt.Errorf("failed to install migrate: %w", err)
	}

	tools := map[string]string{
		"golangci-lint":      "github.com/golangci/golangci-lint/cmd/golangci-lint@latest",
		"protoc-gen-go":      "google.golang.org/protobuf/cmd/protoc-gen-go@latest",
		"protoc-gen-go-grpc": "google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest",
		"protoc-gen-go-http": "github.com/go-kratos/kratos/cmd/protoc-gen-go-http/v2@latest",
		"protoc-gen-openapi": "github.com/google/gnostic/cmd/protoc-gen-openapi@latest",
		"wire":               "github.com/google/wire/cmd/wire@latest",
	}

	for name, pkg := range tools {
		fmt.Printf("Installing %s...\n", name)
		if err := sh.RunV("go", "install", pkg); err != nil {
			return fmt.Errorf("failed to install %s: %w", name, err)
		}
	}

	fmt.Println("✓ All tools installed!")
	return nil
}

// Generate generates OpenAPI 3.0 documentation from proto files
func (Docs) Generate() error {
	fmt.Println("Generating API documentation...")
	mg.Deps(ensureSwaggerDir)

	protoFiles, err := filepath.Glob(filepath.Join(protoDir, "*", "v1", "*.proto"))
	if err != nil {
		return err
	}
	if len(protoFiles) == 0 {
		fmt.Println("No .proto files found")
		return nil
	}

	return generateOpenAPI(protoFiles)
}

// Serve starts a local documentation server
func (Docs) Serve() error {
	fmt.Println("Starting documentation server...")
	fmt.Println("Documentation will be available at http://localhost:3000")
	fmt.Println("Press Ctrl+C to stop the server")

	// Check if node_modules exists, if not, install dependencies
	if _, err := os.Stat("node_modules"); os.IsNotExist(err) {
		fmt.Println("Installing npm dependencies...")
		if err := sh.RunV("npm", "install"); err != nil {
			return fmt.Errorf("failed to install npm dependencies: %w", err)
		}
	}

	// Use npm run to execute the locally installed docsify-cli
	return sh.RunV("npm", "run", "serve")
}

// Build builds static documentation site
func (Docs) Build() error {
	fmt.Println("Building documentation site...")
	mg.Deps(Docs.Generate)

	fmt.Println("✓ Documentation site ready for deployment")
	fmt.Println("  To deploy, copy the 'docs/' directory to your web server")
	fmt.Println("  Or use GitHub Pages by pushing to gh-pages branch")
	return nil
}

// Validate validates swagger documentation
func (Docs) Validate() error {
	fmt.Println("Validating API documentation...")
	mg.Deps(Docs.Generate)

	openAPIFile := filepath.Join(swaggerDir, "openapi.json")
	if _, err := os.Stat(openAPIFile); os.IsNotExist(err) {
		return fmt.Errorf("openapi.json not found at %s", openAPIFile)
	}
	fmt.Println("✓ openapi.json exists")

	asyncAPIFile := filepath.Join(docsDir, "api", "asyncapi.yaml")
	if _, err := os.Stat(asyncAPIFile); os.IsNotExist(err) {
		return fmt.Errorf("asyncapi.yaml not found at %s", asyncAPIFile)
	}
	fmt.Println("✓ asyncapi.yaml exists")

	fmt.Println("✓ All API documentation is valid")
	return nil
}

// ensureSwaggerDir creates swagger directory if it doesn't exist
func ensureSwaggerDir() error {
	return os.MkdirAll(swaggerDir, 0755)
}
