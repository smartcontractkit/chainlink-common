// Command migrate generates a goose migration that creates a discoverer
// announcements table. It is a thin CLI wrapper around the logic in
// ../internal.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	if err := rootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func rootCmd() *cobra.Command {
	var (
		table string
		path  string
	)

	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Generate a goose migration that creates a discoverer announcements table",
		RunE: func(cmd *cobra.Command, args []string) error {
			out, err := CreateMigration(path, table)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "created migration: %s\n", out)
			return nil
		},
	}

	cmd.Flags().StringVar(&table, "table", "", "name of the table to create (required)")
	cmd.Flags().StringVar(&path, "path", "", "directory to write the migration into (required)")
	_ = cmd.MarkFlagRequired("table")
	_ = cmd.MarkFlagRequired("path")

	return cmd
}
