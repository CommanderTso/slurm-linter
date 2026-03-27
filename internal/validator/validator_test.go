package validator_test

import (
	"testing"

	"github.com/CommanderTso/slurm-linter/internal/model"
	"github.com/CommanderTso/slurm-linter/internal/rules"
	"github.com/CommanderTso/slurm-linter/internal/validator"
)

func TestValidator_NoRules(t *testing.T) {
	input := &rules.LintInput{
		SlurmFile: "slurm.conf",
		Slurm:     &model.SlurmConfig{Globals: map[string]string{}, GlobalLines: map[string]int{}},
	}
	v := validator.New()
	diags := v.Run(input)
	if len(diags) != 0 {
		t.Errorf("expected 0 diagnostics with no rules, got %d", len(diags))
	}
}

func TestValidator_CollectsFromAllRules(t *testing.T) {
	input := &rules.LintInput{
		SlurmFile: "slurm.conf",
		Slurm: &model.SlurmConfig{
			Globals:     map[string]string{}, // missing ClusterName, SlurmctldHost
			GlobalLines: map[string]int{},
			// no nodes or partitions
		},
	}
	v := validator.New(rules.RequiredParams{})
	diags := v.Run(input)
	// Expect: missing ClusterName + missing SlurmctldHost + no nodes + no partitions = 4
	if len(diags) != 4 {
		t.Errorf("expected 4 diagnostics, got %d: %v", len(diags), diags)
	}
}
