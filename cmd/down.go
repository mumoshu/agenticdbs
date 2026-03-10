package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/mumoshu/agenticdbs/internal"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(downCmd)
}

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Stop and remove database containers and volumes",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		dir, err := resolveWorkDir()
		if err != nil {
			return err
		}

		dockerDir := filepath.Join(dir, "docker")

		fmt.Println("Stopping databases...")
		if err := internal.DockerComposeDown(ctx, dockerDir); err != nil {
			return err
		}

		// Verify containers are actually down
		if err := internal.VerifyContainersDown(ctx, dockerDir); err != nil {
			return err
		}

		fmt.Println("All containers stopped and volumes removed.")
		return nil
	},
}
