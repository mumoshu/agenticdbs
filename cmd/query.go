package cmd

import (
	"fmt"
	"strings"

	"github.com/mumoshu/agenticdbs/internal"
	"github.com/spf13/cobra"
)

var queryBackend string

func init() {
	queryCmd.Flags().StringVarP(&queryBackend, "backend", "b", "", "Backend to query: surreal or postgres (required)")
	_ = queryCmd.MarkFlagRequired("backend")
	rootCmd.AddCommand(queryCmd)
}

var queryCmd = &cobra.Command{
	Use:   "query [query string]",
	Short: "Execute a query against a database backend",
	Long:  `Executes a query against the specified backend and prints the result.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		query := args[0]

		switch strings.ToLower(queryBackend) {
		case "surreal", "surrealdb":
			client := internal.NewSurrealClient()
			result, err := client.Query(ctx, query)
			if err != nil {
				return fmt.Errorf("SurrealDB query failed: %w", err)
			}
			fmt.Println(result)
			return nil

		case "postgres", "postgresql":
			client := newPostgresClient()
			result, err := client.Query(ctx, query)
			if err != nil {
				return fmt.Errorf("PostgreSQL query failed: %w", err)
			}
			fmt.Println(result)
			return nil

		default:
			return fmt.Errorf("unknown backend: %s (use surreal or postgres)", queryBackend)
		}
	},
}
