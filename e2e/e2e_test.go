package e2e

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var (
	testDir    string
	binaryPath string
)

func TestMain(m *testing.M) {
	// Check that required tools are available
	if _, err := exec.LookPath("claude"); err != nil {
		fmt.Println("Skipping E2E tests: 'claude' CLI not found in PATH")
		os.Exit(0)
	}
	if _, err := exec.LookPath("docker"); err != nil {
		fmt.Println("Skipping E2E tests: 'docker' not found in PATH")
		os.Exit(0)
	}

	// Build the binary
	projectRoot, err := findProjectRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to find project root: %v\n", err)
		os.Exit(1)
	}

	binaryPath = filepath.Join(projectRoot, "agenticdbs")
	buildCmd := exec.Command("go", "build", "-o", binaryPath, ".")
	buildCmd.Dir = projectRoot
	if out, err := buildCmd.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to build binary: %v\n%s\n", err, out)
		os.Exit(1)
	}

	// Create temp directory for test isolation
	testDir, err = os.MkdirTemp("", "agenticdbs-e2e-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create temp dir: %v\n", err)
		os.Exit(1)
	}

	// Setup: write files, start containers, seed data
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)

	setupErr := func() error {
		// Run setup in temp dir
		if err := runBinary(ctx, testDir, "setup", "--dir", testDir); err != nil {
			return fmt.Errorf("setup failed: %w", err)
		}

		// Start containers
		if err := runBinary(ctx, testDir, "up", "--dir", testDir, "--timeout", "180s"); err != nil {
			return fmt.Errorf("up failed: %w", err)
		}

		// Seed data
		if err := runBinary(ctx, testDir, "seed", "--dir", testDir); err != nil {
			return fmt.Errorf("seed failed: %w", err)
		}

		return nil
	}()
	cancel()

	if setupErr != nil {
		fmt.Fprintf(os.Stderr, "E2E setup failed: %v\n", setupErr)
		dumpDiagnosticsAndCleanup()
		os.Exit(1)
	}

	// Run tests
	code := m.Run()

	// Cleanup: always tear down, even if tests failed
	dumpDiagnosticsAndCleanup()

	os.Exit(code)
}

func dumpDiagnosticsAndCleanup() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	dockerDir := filepath.Join(testDir, "docker")

	// Dump diagnostics on failure
	fmt.Fprintln(os.Stderr, "\n=== E2E DIAGNOSTIC DUMP ===")

	// Container status
	fmt.Fprintln(os.Stderr, "\n--- Container Status ---")
	psCmd := exec.CommandContext(ctx, "docker", "compose", "ps", "-a")
	psCmd.Dir = dockerDir
	psCmd.Stdout = os.Stderr
	psCmd.Stderr = os.Stderr
	_ = psCmd.Run()

	// Container logs
	for _, svc := range []string{"surrealdb", "postgres"} {
		fmt.Fprintf(os.Stderr, "\n--- %s logs (last 50 lines) ---\n", svc)
		logCmd := exec.CommandContext(ctx, "docker", "compose", "logs", "--tail", "50", svc)
		logCmd.Dir = dockerDir
		logCmd.Stdout = os.Stderr
		logCmd.Stderr = os.Stderr
		_ = logCmd.Run()
	}

	// List files in temp dir
	fmt.Fprintf(os.Stderr, "\n--- Files in temp dir: %s ---\n", testDir)
	_ = filepath.Walk(testDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(testDir, path)
		if info.IsDir() {
			fmt.Fprintf(os.Stderr, "  %s/\n", rel)
		} else {
			fmt.Fprintf(os.Stderr, "  %s (%d bytes)\n", rel, info.Size())
		}
		return nil
	})

	// Print Claude log file contents
	claudeLogDir := filepath.Join(testDir, "claude-logs")
	if entries, err := os.ReadDir(claudeLogDir); err == nil {
		for _, entry := range entries {
			logPath := filepath.Join(claudeLogDir, entry.Name())
			fmt.Fprintf(os.Stderr, "\n--- Claude log: %s ---\n", entry.Name())
			data, readErr := os.ReadFile(logPath)
			if readErr == nil {
				content := string(data)
				if len(content) > 3000 {
					content = content[:3000] + "\n... [truncated]"
				}
				fmt.Fprintln(os.Stderr, content)
			}
		}
	}

	fmt.Fprintln(os.Stderr, "\n=== END DIAGNOSTIC DUMP ===")

	// Tear down containers
	fmt.Fprintln(os.Stderr, "\nTearing down containers...")
	_ = runBinary(ctx, testDir, "down", "--dir", testDir)

	// Verify containers are gone
	verifyCmd := exec.CommandContext(ctx, "docker", "compose", "ps", "--format", "{{.Name}}")
	verifyCmd.Dir = dockerDir
	if out, err := verifyCmd.Output(); err == nil {
		remaining := strings.TrimSpace(string(out))
		if remaining != "" {
			fmt.Fprintf(os.Stderr, "WARNING: containers still running after down: %s\n", remaining)
			// Force remove
			forceCmd := exec.CommandContext(ctx, "docker", "compose", "kill")
			forceCmd.Dir = dockerDir
			_ = forceCmd.Run()
			rmCmd := exec.CommandContext(ctx, "docker", "compose", "rm", "-f")
			rmCmd.Dir = dockerDir
			_ = rmCmd.Run()
		}
	}

	// Remove temp dir
	if testDir != "" {
		_ = os.RemoveAll(testDir)
	}
}

// runBinary executes the agenticdbs binary with the given arguments.
func runBinary(ctx context.Context, workDir string, args ...string) error {
	cmd := exec.CommandContext(ctx, binaryPath, args...)
	cmd.Dir = workDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// logDir returns the directory for Claude prompt/response logs.
// It creates the directory if it doesn't exist.
func logDir() string {
	dir := filepath.Join(testDir, "claude-logs")
	_ = os.MkdirAll(dir, 0o755)
	return dir
}

// runClaude invokes the Claude CLI with a prompt and returns its output.
// It sets the working directory to testDir so that Claude can find the skill
// and docs files written by setup. All prompts and responses are logged to files.
func runClaude(t *testing.T, prompt string) string {
	t.Helper()

	// Create a sanitized filename from the test name
	testName := strings.ReplaceAll(t.Name(), "/", "_")
	logPrefix := filepath.Join(logDir(), testName)
	promptLogFile := logPrefix + ".prompt.txt"
	responseLogFile := logPrefix + ".response.txt"
	stderrLogFile := logPrefix + ".stderr.txt"

	// Log the prompt
	_ = os.WriteFile(promptLogFile, []byte(prompt), 0o644)
	t.Logf("Claude prompt logged to: %s", promptLogFile)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "claude",
		"-p", prompt,
		"--allowedTools", "Bash,Read,Glob,Grep",
		"--output-format", "text",
	)
	cmd.Dir = testDir

	// Build environment: ensure agenticdbs binary is in PATH and
	// unset CLAUDECODE to allow nested Claude Code invocations
	env := os.Environ()
	filteredEnv := make([]string, 0, len(env))
	for _, e := range env {
		if !strings.HasPrefix(e, "CLAUDECODE=") {
			filteredEnv = append(filteredEnv, e)
		}
	}
	cmd.Env = append(filteredEnv,
		fmt.Sprintf("PATH=%s:%s", filepath.Dir(binaryPath), os.Getenv("PATH")),
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	output := stdout.String()

	// Log response and stderr
	_ = os.WriteFile(responseLogFile, []byte(output), 0o644)
	_ = os.WriteFile(stderrLogFile, []byte(stderr.String()), 0o644)
	t.Logf("Claude response logged to: %s", responseLogFile)
	t.Logf("Claude stderr logged to: %s", stderrLogFile)

	if err != nil {
		t.Logf("Claude CLI stderr:\n%s", stderr.String())
		t.Logf("Claude CLI stdout:\n%s", output)
		t.Fatalf("Claude CLI failed: %v", err)
	}

	if output == "" {
		t.Logf("Claude CLI stderr:\n%s", stderr.String())
		t.Fatal("Claude CLI returned empty output")
	}

	return output
}

// assertContains checks that output contains the expected substring.
// It provides detailed failure output including the full Claude response.
func assertContains(t *testing.T, output, expected, context string) {
	t.Helper()
	if !strings.Contains(output, expected) {
		t.Errorf("%s: expected output to contain %q\nFull output:\n%s", context, expected, truncate(output, 2000))
	}
}

// assertContainsAny checks that output contains at least one of the expected substrings.
func assertContainsAny(t *testing.T, output string, expected []string, context string) {
	t.Helper()
	for _, e := range expected {
		if strings.Contains(output, e) {
			return
		}
	}
	t.Errorf("%s: expected output to contain one of %v\nFull output:\n%s", context, expected, truncate(output, 2000))
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "\n... [truncated]"
}

func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found")
		}
		dir = parent
	}
}
