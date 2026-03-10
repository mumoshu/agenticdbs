package cmd

import (
	"fmt"
	"path/filepath"

	embedpkg "github.com/mumoshu/agenticdbs/embed"
	"github.com/mumoshu/agenticdbs/internal"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(setupCmd)
}

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Write embedded files to disk (Dockerfiles, compose, skills, docs)",
	Long:  `Extracts all embedded assets to the target directory. Idempotent — safe to re-run.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		absDir, err := resolveWorkDir()
		if err != nil {
			return err
		}

		fmt.Printf("Setting up agenticdbs in %s\n", absDir)

		// Write docker files
		dockerDir := filepath.Join(absDir, "docker")
		if err := internal.WriteEmbeddedFS(embedpkg.DockerFS, "docker", dockerDir); err != nil {
			return fmt.Errorf("writing docker files: %w", err)
		}
		fmt.Println("  docker/ (Dockerfiles, docker-compose.yml)")

		// Write seed files
		seedDir := filepath.Join(absDir, "seed")
		if err := internal.WriteEmbeddedFS(embedpkg.SeedFS, "seed", seedDir); err != nil {
			return fmt.Errorf("writing seed files: %w", err)
		}
		fmt.Println("  seed/ (SurrealDB and PostgreSQL seed data)")

		// Write docs
		docsDir := filepath.Join(absDir, "docs")
		if err := internal.WriteEmbeddedFS(embedpkg.DocsFS, "docs", docsDir); err != nil {
			return fmt.Errorf("writing docs: %w", err)
		}
		fmt.Println("  docs/ (feature matrix, query guides)")

		// Write skill
		skillPath := filepath.Join(absDir, ".claude", "skills", "db.md")
		if err := internal.WriteFile(skillPath, embedpkg.SkillDB); err != nil {
			return fmt.Errorf("writing skill: %w", err)
		}
		fmt.Println("  .claude/skills/db.md")

		// Write CLAUDE.md
		claudeMD := generateClaudeMD()
		claudePath := filepath.Join(absDir, "CLAUDE.md")
		if err := internal.WriteFile(claudePath, []byte(claudeMD)); err != nil {
			return fmt.Errorf("writing CLAUDE.md: %w", err)
		}
		fmt.Println("  CLAUDE.md")

		// Verify all expected files exist and are non-empty
		expectedFiles := []string{
			filepath.Join(dockerDir, "docker-compose.yml"),
			filepath.Join(dockerDir, "postgres", "Dockerfile"),
			filepath.Join(dockerDir, "surrealdb", "Dockerfile"),
			filepath.Join(seedDir, "surrealdb", "00-schema.surql"),
			filepath.Join(seedDir, "surrealdb", "01-data.surql"),
			filepath.Join(seedDir, "postgres", "00-extensions.sql"),
			filepath.Join(seedDir, "postgres", "01-schema.sql"),
			filepath.Join(seedDir, "postgres", "02-data.sql"),
			filepath.Join(docsDir, "feature-matrix.md"),
			filepath.Join(docsDir, "surrealdb-guide.md"),
			filepath.Join(docsDir, "postgres-guide.md"),
			skillPath,
			claudePath,
		}

		if err := internal.VerifyFilesExist(expectedFiles); err != nil {
			return fmt.Errorf("verification failed: %w", err)
		}

		fmt.Println("\nSetup complete. All files verified.")
		return nil
	},
}

func resolveWorkDir() (string, error) {
	return filepath.Abs(workDir)
}

func generateClaudeMD() string {
	return `# agenticdbs — Agentic Database System

This project is a multi-backend agentic database system running SurrealDB and PostgreSQL side by side.

## Quick Start

` + "```" + `bash
# Start databases
agenticdbs up

# Seed demo data
agenticdbs seed

# Run a query
agenticdbs query --backend surreal "SELECT * FROM users"
agenticdbs query --backend postgres "SELECT * FROM users"

# Stop databases
agenticdbs down
` + "```" + `

## Backends

| Backend | Port | Features |
|---------|------|----------|
| SurrealDB | 8000 | Native: relational, graph, vector, timeseries, JSON, full-text |
| PostgreSQL | 5432 | Extensions: pgvector, Apache AGE (graph), TimescaleDB, JSONB, tsvector |

## For the AI Agent

- Use the ` + "`/db`" + ` skill for all database interactions
- Read ` + "`docs/feature-matrix.md`" + ` for backend comparison and recommendations
- Read ` + "`docs/surrealdb-guide.md`" + ` for SurrealQL syntax reference
- Read ` + "`docs/postgres-guide.md`" + ` for SQL + extensions syntax reference
- Execute queries via ` + "`agenticdbs query --backend <surreal|postgres> \"QUERY\"`" + `
- For complex tasks, delegate to backend-specific sub-agents (see skill for templates)

## Architecture

The system supports three goals:
1. **Query DB** — Write and execute queries leveraging each backend's strengths
2. **Migrate** — Move schemas/data between backends and environments
3. **Develop Apps** — Build apps with agentic (flexible) or direct (performant) access patterns

See ` + "`docs/`" + ` directory for detailed guides.
`
}
