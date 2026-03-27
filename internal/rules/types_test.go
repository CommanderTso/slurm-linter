package rules_test

import (
	"testing"

	"github.com/CommanderTso/slurm-linter/internal/diagnostic"
	"github.com/CommanderTso/slurm-linter/internal/model"
	"github.com/CommanderTso/slurm-linter/internal/rules"
)

func makeInputWithGlobals(globals map[string]string) *rules.LintInput {
	lines := make(map[string]int)
	for k := range globals {
		lines[k] = 1
	}
	return &rules.LintInput{
		SlurmFile: "slurm.conf",
		Slurm: &model.SlurmConfig{
			Globals:     globals,
			GlobalLines: lines,
		},
	}
}

func TestTypeValidation_ValidInteger(t *testing.T) {
	input := makeInputWithGlobals(map[string]string{"KillWait": "30"})
	rule := rules.TypeValidation{}
	diags := rule.Check(input)
	if len(diags) != 0 {
		t.Errorf("expected no diagnostics, got %v", diags)
	}
}

func TestTypeValidation_InvalidInteger(t *testing.T) {
	input := makeInputWithGlobals(map[string]string{"KillWait": "notanumber"})
	rule := rules.TypeValidation{}
	diags := rule.Check(input)
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}
	if diags[0].Severity != diagnostic.Error {
		t.Errorf("expected Error, got %v", diags[0].Severity)
	}
}

func TestTypeValidation_NegativeInteger(t *testing.T) {
	input := makeInputWithGlobals(map[string]string{"KillWait": "-5"})
	rule := rules.TypeValidation{}
	diags := rule.Check(input)
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}
}

func TestTypeValidation_ValidEnum(t *testing.T) {
	input := makeInputWithGlobals(map[string]string{"SchedulerType": "sched/backfill"})
	rule := rules.TypeValidation{}
	diags := rule.Check(input)
	if len(diags) != 0 {
		t.Errorf("expected no diagnostics, got %v", diags)
	}
}

func TestTypeValidation_InvalidEnum(t *testing.T) {
	input := makeInputWithGlobals(map[string]string{"SchedulerType": "sched/invalid"})
	rule := rules.TypeValidation{}
	diags := rule.Check(input)
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}
	if diags[0].Rule != "type-validation" {
		t.Errorf("expected rule=type-validation, got %q", diags[0].Rule)
	}
}

func TestTypeValidation_ValidTimeINFINITE(t *testing.T) {
	input := makeInputWithGlobals(map[string]string{"MaxTime": "INFINITE"})
	rule := rules.TypeValidation{}
	diags := rule.Check(input)
	if len(diags) != 0 {
		t.Errorf("expected no diagnostics, got %v", diags)
	}
}

func TestTypeValidation_ValidTimeHHMMSS(t *testing.T) {
	input := makeInputWithGlobals(map[string]string{"MaxTime": "24:00:00"})
	rule := rules.TypeValidation{}
	diags := rule.Check(input)
	if len(diags) != 0 {
		t.Errorf("expected no diagnostics, got %v", diags)
	}
}

func TestTypeValidation_InvalidTime(t *testing.T) {
	input := makeInputWithGlobals(map[string]string{"MaxTime": "notatime"})
	rule := rules.TypeValidation{}
	diags := rule.Check(input)
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}
}
