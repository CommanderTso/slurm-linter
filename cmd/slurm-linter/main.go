package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/CommanderTso/slurm-linter/internal/diagnostic"
	"github.com/CommanderTso/slurm-linter/internal/parser"
	"github.com/CommanderTso/slurm-linter/internal/reporter"
	"github.com/CommanderTso/slurm-linter/internal/rules"
	"github.com/CommanderTso/slurm-linter/internal/validator"
)

func main() {
	var confFile string
	var topologyFile string
	var format string

	checkCmd := &cobra.Command{
		Use:   "check",
		Short: "Lint slurm.conf and optionally topology.conf",
		RunE: func(cmd *cobra.Command, args []string) error {
			if confFile == "" {
				return fmt.Errorf("--conf is required")
			}

			input, err := buildLintInput(confFile, topologyFile)
			if err != nil {
				return err
			}

			v := validator.New(validator.DefaultRules()...)
			diags := v.Run(input)

			switch format {
			case "json":
				reporter.JSON{Writer: os.Stdout}.Report(diags)
			default:
				reporter.Text{Writer: os.Stdout}.Report(diags)
			}

			os.Exit(exitCode(diags))
			return nil
		},
	}

	checkCmd.Flags().StringVar(&confFile, "conf", "", "path to slurm.conf (required)")
	checkCmd.Flags().StringVar(&topologyFile, "topology", "", "path to topology.conf (optional)")
	checkCmd.Flags().StringVar(&format, "format", "text", "output format: text or json")

	rootCmd := &cobra.Command{
		Use:          "slurm-linter",
		Short:        "Validate Slurm configuration files",
		SilenceUsage: true,
	}
	rootCmd.AddCommand(checkCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func buildLintInput(confFile, topologyFile string) (*rules.LintInput, error) {
	cf, err := os.Open(confFile)
	if err != nil {
		return nil, fmt.Errorf("cannot open %s: %w", confFile, err)
	}
	defer cf.Close()

	slurm, err := parser.ParseSlurmConf(confFile, cf)
	if err != nil {
		return nil, err
	}

	input := &rules.LintInput{
		SlurmFile: confFile,
		Slurm:     slurm,
	}

	if topologyFile != "" {
		tf, err := os.Open(topologyFile)
		if err != nil {
			return nil, fmt.Errorf("cannot open %s: %w", topologyFile, err)
		}
		defer tf.Close()

		topo, err := parser.ParseTopologyConf(topologyFile, tf)
		if err != nil {
			return nil, err
		}
		input.TopologyFile = topologyFile
		input.Topology = topo
	}

	return input, nil
}

// exitCode returns 0 (clean), 1 (warnings only), or 2 (errors present).
func exitCode(diags []diagnostic.Diagnostic) int {
	code := 0
	for _, d := range diags {
		if d.Severity == diagnostic.Error {
			return 2
		}
		if d.Severity == diagnostic.Warning {
			code = 1
		}
	}
	return code
}
