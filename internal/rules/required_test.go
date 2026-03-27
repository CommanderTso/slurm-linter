package rules_test

import (
	"testing"

	"github.com/CommanderTso/slurm-linter/internal/diagnostic"
	"github.com/CommanderTso/slurm-linter/internal/model"
	"github.com/CommanderTso/slurm-linter/internal/rules"
)

func makeInput(globals map[string]string, nodes []model.NodeDef, partitions []model.Partition) *rules.LintInput {
	return &rules.LintInput{
		SlurmFile: "slurm.conf",
		Slurm: &model.SlurmConfig{
			Globals:     globals,
			GlobalLines: make(map[string]int),
			Nodes:       nodes,
			Partitions:  partitions,
		},
	}
}

func TestRequiredParams_AllPresent(t *testing.T) {
	input := makeInput(
		map[string]string{"ClusterName": "test", "SlurmctldHost": "ctrl"},
		[]model.NodeDef{{Name: "node01", Params: map[string]string{}}},
		[]model.Partition{{Name: "compute", Params: map[string]string{}}},
	)
	rule := rules.RequiredParams{}
	diags := rule.Check(input)
	if len(diags) != 0 {
		t.Errorf("expected no diagnostics, got %d: %v", len(diags), diags)
	}
}

func TestRequiredParams_MissingClusterName(t *testing.T) {
	input := makeInput(
		map[string]string{"SlurmctldHost": "ctrl"},
		[]model.NodeDef{{Name: "node01", Params: map[string]string{}}},
		[]model.Partition{{Name: "compute", Params: map[string]string{}}},
	)
	rule := rules.RequiredParams{}
	diags := rule.Check(input)
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}
	if diags[0].Severity != diagnostic.Error {
		t.Errorf("expected Error severity, got %v", diags[0].Severity)
	}
	if diags[0].Rule != "required-params" {
		t.Errorf("expected rule=required-params, got %q", diags[0].Rule)
	}
}

func TestRequiredParams_NoNodes(t *testing.T) {
	input := makeInput(
		map[string]string{"ClusterName": "test", "SlurmctldHost": "ctrl"},
		nil,
		[]model.Partition{{Name: "compute", Params: map[string]string{}}},
	)
	rule := rules.RequiredParams{}
	diags := rule.Check(input)
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}
}

func TestRequiredParams_NoPartitions(t *testing.T) {
	input := makeInput(
		map[string]string{"ClusterName": "test", "SlurmctldHost": "ctrl"},
		[]model.NodeDef{{Name: "node01", Params: map[string]string{}}},
		nil,
	)
	rule := rules.RequiredParams{}
	diags := rule.Check(input)
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}
}
