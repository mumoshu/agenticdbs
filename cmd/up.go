package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	embedpkg "github.com/mumoshu/agenticdbs/embed"
	"github.com/mumoshu/agenticdbs/internal"
	"github.com/spf13/cobra"
)

var upTimeout time.Duration

func init() {
	upCmd.Flags().DurationVar(&upTimeout, "timeout", 120*time.Second, "Timeout for health checks")
	rootCmd.AddCommand(upCmd)
}

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Start SurrealDB and PostgreSQL containers",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		dir, err := resolveWorkDir()
		if err != nil {
			return err
		}

		dockerDir := filepath.Join(dir, "docker")

		// Ensure docker files exist
		composeFile := filepath.Join(dockerDir, "docker-compose.yml")
		if _, err := statFile(composeFile); err != nil {
			fmt.Println("Docker files not found, running setup first...")
			if err := internal.WriteEmbeddedFS(embedpkg.DockerFS, "docker", dockerDir); err != nil {
				return fmt.Errorf("writing docker files: %w", err)
			}
		}

		fmt.Println("Starting databases...")
		if err := internal.DockerComposeUp(ctx, dockerDir); err != nil {
			return err
		}

		fmt.Println("Waiting for databases to become healthy...")
		if err := internal.WaitForHealth(ctx, dockerDir, upTimeout); err != nil {
			return err
		}

		// Final verification: check container status
		if err := verifyContainersRunning(ctx, dockerDir); err != nil {
			return err
		}

		fmt.Println("\nBoth databases are up and healthy.")
		fmt.Println("  SurrealDB: http://localhost:18000")
		fmt.Println("  PostgreSQL: localhost:15432 (user: agenticdbs, db: agenticdbs)")
		return nil
	},
}

func verifyContainersRunning(ctx context.Context, composeDir string) error {
	healthy, status := checkHealthDirect(ctx, composeDir)
	if !healthy {
		internal.DumpDiagnostics(ctx, composeDir)
		return fmt.Errorf("containers not healthy after startup: %s", status)
	}
	return nil
}

func checkHealthDirect(ctx context.Context, composeDir string) (bool, string) {
	// Check SurrealDB health endpoint
	surreal := internal.NewSurrealClient()
	if err := surreal.Health(ctx); err != nil {
		return false, fmt.Sprintf("SurrealDB: %v", err)
	}

	// Check PostgreSQL health
	pg := newPostgresClient()
	if err := pg.Health(ctx); err != nil {
		return false, fmt.Sprintf("PostgreSQL: %v", err)
	}

	return true, "all healthy"
}

func statFile(path string) (os.FileInfo, error) {
	return os.Stat(path)
}
