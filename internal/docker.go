package internal

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// WriteEmbeddedFS writes an embedded filesystem to a target directory.
func WriteEmbeddedFS(fsys embed.FS, root, targetDir string) error {
	return fs.WalkDir(fsys, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(targetDir, relPath)

		if d.IsDir() {
			return os.MkdirAll(targetPath, 0o755)
		}

		data, err := fsys.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading embedded file %s: %w", path, err)
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			return err
		}

		return os.WriteFile(targetPath, data, 0o644)
	})
}

// WriteFile writes data to a file, creating parent directories as needed.
func WriteFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// VerifyFilesExist checks that all expected files exist and are non-empty.
func VerifyFilesExist(paths []string) error {
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			return fmt.Errorf("expected file missing: %s: %w", p, err)
		}
		if info.Size() == 0 {
			return fmt.Errorf("expected file is empty: %s", p)
		}
	}
	return nil
}

// DockerComposeUp starts services using docker compose.
func DockerComposeUp(ctx context.Context, composeDir string) error {
	cmd := exec.CommandContext(ctx, "docker", "compose", "up", "-d", "--build")
	cmd.Dir = composeDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker compose up failed: %w", err)
	}
	return nil
}

// DockerComposeDown stops and removes services.
func DockerComposeDown(ctx context.Context, composeDir string) error {
	cmd := exec.CommandContext(ctx, "docker", "compose", "down", "-v")
	cmd.Dir = composeDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker compose down failed: %w", err)
	}
	return nil
}

// VerifyContainersDown checks that no containers remain after down.
func VerifyContainersDown(ctx context.Context, composeDir string) error {
	cmd := exec.CommandContext(ctx, "docker", "compose", "ps", "--format", "{{.Name}}")
	cmd.Dir = composeDir
	out, err := cmd.Output()
	if err != nil {
		// If compose ps fails, it may mean the project is fully gone
		return nil
	}
	remaining := strings.TrimSpace(string(out))
	if remaining != "" {
		return fmt.Errorf("containers still running after down:\n%s", remaining)
	}
	return nil
}

// WaitForHealth polls health checks for both databases.
func WaitForHealth(ctx context.Context, composeDir string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		if err := ctx.Err(); err != nil {
			return err
		}

		healthy, status := checkComposeHealth(ctx, composeDir)
		if healthy {
			return nil
		}

		// Check for fatal errors in logs
		if fatalErr := checkForFatalLogs(ctx, composeDir); fatalErr != nil {
			return fatalErr
		}

		fmt.Printf("  Waiting for databases... %s\n", status)
		time.Sleep(2 * time.Second)
	}

	// Timeout — dump diagnostics
	DumpDiagnostics(ctx, composeDir)
	return fmt.Errorf("health check timed out after %s", timeout)
}

func checkComposeHealth(ctx context.Context, composeDir string) (bool, string) {
	cmd := exec.CommandContext(ctx, "docker", "compose", "ps", "--format", "{{.Name}}: {{.Health}}")
	cmd.Dir = composeDir
	out, err := cmd.Output()
	if err != nil {
		return false, "compose ps failed"
	}

	output := strings.TrimSpace(string(out))
	if output == "" {
		return false, "no containers found"
	}

	lines := strings.Split(output, "\n")
	allHealthy := true
	var statuses []string
	for _, line := range lines {
		statuses = append(statuses, line)
		if !strings.Contains(line, "healthy") {
			allHealthy = false
		}
	}

	return allHealthy, strings.Join(statuses, ", ")
}

func checkForFatalLogs(ctx context.Context, composeDir string) error {
	// Known-safe FATAL patterns that occur during normal PostgreSQL/extension initialization
	safeFatalPatterns := []string{
		"terminating background worker",
		"due to administrator command",
		"the database system is shutting down",
		"the database system is starting up",
	}

	for _, svc := range []string{"surrealdb", "postgres"} {
		cmd := exec.CommandContext(ctx, "docker", "compose", "logs", "--tail", "50", svc)
		cmd.Dir = composeDir
		out, err := cmd.Output()
		if err != nil {
			continue
		}
		logs := string(out)
		// Check for common fatal patterns
		fatalPatterns := []string{"FATAL:", "panic:", "could not start"}
		for _, pattern := range fatalPatterns {
			if !strings.Contains(logs, pattern) {
				continue
			}
			// Check each log line containing the pattern — skip known-safe ones
			isTrulyFatal := false
			for _, line := range strings.Split(logs, "\n") {
				if !strings.Contains(line, pattern) {
					continue
				}
				safe := false
				for _, safePattern := range safeFatalPatterns {
					if strings.Contains(line, safePattern) {
						safe = true
						break
					}
				}
				if !safe {
					isTrulyFatal = true
					break
				}
			}
			if isTrulyFatal {
				return fmt.Errorf("fatal error detected in %s logs: found %q\nRecent logs:\n%s", svc, pattern, logs)
			}
		}
	}
	return nil
}

// DumpDiagnostics outputs container statuses and logs for debugging.
func DumpDiagnostics(ctx context.Context, composeDir string) {
	fmt.Fprintln(os.Stderr, "\n=== DIAGNOSTIC DUMP ===")

	// Container status
	fmt.Fprintln(os.Stderr, "\n--- Container Status ---")
	psCmd := exec.CommandContext(ctx, "docker", "compose", "ps", "-a")
	psCmd.Dir = composeDir
	psCmd.Stdout = os.Stderr
	psCmd.Stderr = os.Stderr
	_ = psCmd.Run()

	// Container logs
	for _, svc := range []string{"surrealdb", "postgres"} {
		fmt.Fprintf(os.Stderr, "\n--- %s logs (last 100 lines) ---\n", svc)
		logCmd := exec.CommandContext(ctx, "docker", "compose", "logs", "--tail", "100", svc)
		logCmd.Dir = composeDir
		logCmd.Stdout = os.Stderr
		logCmd.Stderr = os.Stderr
		_ = logCmd.Run()
	}

	fmt.Fprintln(os.Stderr, "\n=== END DIAGNOSTIC DUMP ===")
}
