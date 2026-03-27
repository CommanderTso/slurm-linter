package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "slurm-linter",
		Short: "Validate Slurm configuration files",
	}

	checkCmd := &cobra.Command{
		Use:   "check",
		Short: "Lint slurm.conf and topology.conf",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("not yet implemented")
			return nil
		},
	}

	rootCmd.AddCommand(checkCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
