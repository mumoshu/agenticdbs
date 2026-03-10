package cmd

import (
	"github.com/spf13/cobra"
)

var workDir string

var rootCmd = &cobra.Command{
	Use:   "agenticdbs",
	Short: "Agentic database system backed by SurrealDB and PostgreSQL",
	Long: `agenticdbs is a multi-backend agentic database system.
It runs SurrealDB and PostgreSQL side by side, each configured with
full multi-model capabilities (relational, graph, vector, timeseries,
JSON, full-text search). Use Claude Code CLI as the front agent with
the /db skill for interactive database operations.`,
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&workDir, "dir", "d", ".", "Working directory for agenticdbs files")
}

func Execute() error {
	return rootCmd.Execute()
}
