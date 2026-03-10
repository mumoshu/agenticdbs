package cmd

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	embedpkg "github.com/mumoshu/agenticdbs/embed"
	"github.com/mumoshu/agenticdbs/internal"
	"github.com/spf13/cobra"
)

var seedBackend string

func init() {
	seedCmd.Flags().StringVarP(&seedBackend, "backend", "b", "all", "Backend to seed: surreal, postgres, or all")
	rootCmd.AddCommand(seedCmd)
}

var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Seed databases with demo data",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		switch seedBackend {
		case "surreal":
			return seedSurrealDB(ctx)
		case "postgres":
			return seedPostgres(ctx)
		case "all":
			if err := seedSurrealDB(ctx); err != nil {
				return fmt.Errorf("seeding SurrealDB: %w", err)
			}
			if err := seedPostgres(ctx); err != nil {
				return fmt.Errorf("seeding PostgreSQL: %w", err)
			}
			return nil
		default:
			return fmt.Errorf("unknown backend: %s (use surreal, postgres, or all)", seedBackend)
		}
	},
}

func seedSurrealDB(ctx context.Context) error {
	fmt.Println("Seeding SurrealDB...")
	client := internal.NewSurrealClient()

	// Execute schema
	fmt.Println("  Applying schema...")
	if _, err := client.Query(ctx, string(embedpkg.SurrealDBSchema)); err != nil {
		return fmt.Errorf("applying SurrealDB schema: %w", err)
	}

	// Execute data
	fmt.Println("  Inserting data...")
	if _, err := client.Query(ctx, string(embedpkg.SurrealDBData)); err != nil {
		return fmt.Errorf("inserting SurrealDB data: %w", err)
	}

	// Verify: check expected record counts
	fmt.Println("  Verifying seed data...")
	verifications := map[string]int{
		"users":            5,
		"products":         5,
		"orders":           5,
		"page_metrics":     12,
		"user_preferences": 5,
	}

	for table, expected := range verifications {
		query := fmt.Sprintf("SELECT count() AS count FROM %s GROUP ALL", table)
		result, err := client.Query(ctx, query)
		if err != nil {
			return fmt.Errorf("verifying SurrealDB table %s: %w", table, err)
		}
		if !strings.Contains(result, strconv.Itoa(expected)) {
			return fmt.Errorf("SurrealDB table %s: expected %d records, got: %s", table, expected, result)
		}
	}

	// Verify graph edges
	graphQuery := "SELECT count() AS count FROM knows GROUP ALL"
	result, err := client.Query(ctx, graphQuery)
	if err != nil {
		return fmt.Errorf("verifying SurrealDB graph edges: %w", err)
	}
	if !strings.Contains(result, "5") {
		return fmt.Errorf("SurrealDB knows edges: expected 5, got: %s", result)
	}

	fmt.Println("  SurrealDB seeded and verified successfully.")
	return nil
}

func newPostgresClient() *internal.PostgresClient {
	client := internal.NewPostgresClient()
	if dir, err := resolveWorkDir(); err == nil {
		client.ComposeDir = filepath.Join(dir, "docker")
	}
	return client
}

func seedPostgres(ctx context.Context) error {
	fmt.Println("Seeding PostgreSQL...")
	client := newPostgresClient()

	// Execute extensions
	fmt.Println("  Creating extensions...")
	if err := client.ExecFile(ctx, string(embedpkg.PostgresExtensions)); err != nil {
		return fmt.Errorf("creating PostgreSQL extensions: %w", err)
	}

	// Execute schema
	fmt.Println("  Applying schema...")
	if err := client.ExecFile(ctx, string(embedpkg.PostgresSchema)); err != nil {
		return fmt.Errorf("applying PostgreSQL schema: %w", err)
	}

	// Execute data
	fmt.Println("  Inserting data...")
	if err := client.ExecFile(ctx, string(embedpkg.PostgresData)); err != nil {
		return fmt.Errorf("inserting PostgreSQL data: %w", err)
	}

	// Verify: check expected record counts
	fmt.Println("  Verifying seed data...")
	verifications := map[string]int{
		"users":            5,
		"products":         5,
		"orders":           5,
		"page_metrics":     12,
		"user_preferences": 5,
	}

	for table, expected := range verifications {
		query := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
		result, err := client.Query(ctx, query)
		if err != nil {
			return fmt.Errorf("verifying PostgreSQL table %s: %w", table, err)
		}
		count := strings.TrimSpace(result)
		if count != strconv.Itoa(expected) {
			return fmt.Errorf("PostgreSQL table %s: expected %d records, got: %q", table, expected, count)
		}
	}

	// Verify graph: count user vertices
	graphQuery := `SELECT * FROM cypher('social', $$ MATCH (u:User) RETURN count(u) $$) as (count agtype)`
	result, err := client.Query(ctx, graphQuery)
	if err != nil {
		return fmt.Errorf("verifying PostgreSQL graph: %w", err)
	}
	if !strings.Contains(result, "5") {
		return fmt.Errorf("PostgreSQL graph users: expected 5, got: %s", result)
	}

	fmt.Println("  PostgreSQL seeded and verified successfully.")
	return nil
}
