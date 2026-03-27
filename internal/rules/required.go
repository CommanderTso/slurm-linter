package rules

import (
	"fmt"

	"github.com/CommanderTso/slurm-linter/internal/diagnostic"
)

// RequiredParams checks that the minimum required global parameters are present.
type RequiredParams struct{}

var requiredGlobals = []string{"ClusterName", "SlurmctldHost"}

func (r RequiredParams) Check(input *LintInput) []diagnostic.Diagnostic {
	var diags []diagnostic.Diagnostic
	cfg := input.Slurm

	for _, key := range requiredGlobals {
		if _, ok := cfg.Globals[key]; !ok {
			diags = append(diags, diagnostic.Diagnostic{
				Severity: diagnostic.Error,
				File:     input.SlurmFile,
				Line:     0,
				Message:  fmt.Sprintf("required parameter %q is missing", key),
				Rule:     "required-params",
			})
		}
	}

	if len(cfg.Nodes) == 0 {
		diags = append(diags, diagnostic.Diagnostic{
			Severity: diagnostic.Error,
			File:     input.SlurmFile,
			Line:     0,
			Message:  "no NodeName stanzas defined",
			Rule:     "required-params",
		})
	}

	if len(cfg.Partitions) == 0 {
		diags = append(diags, diagnostic.Diagnostic{
			Severity: diagnostic.Error,
			File:     input.SlurmFile,
			Line:     0,
			Message:  "no PartitionName stanzas defined",
			Rule:     "required-params",
		})
	}

	return diags
}
