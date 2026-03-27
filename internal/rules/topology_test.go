package rules_test

import (
	"testing"

	"github.com/CommanderTso/slurm-linter/internal/model"
	"github.com/CommanderTso/slurm-linter/internal/rules"
)

func topoInput(switches []model.Switch) *rules.LintInput {
	return &rules.LintInput{
		SlurmFile:    "slurm.conf",
		TopologyFile: "topology.conf",
		Slurm: &model.SlurmConfig{
			Globals:     map[string]string{},
			GlobalLines: map[string]int{},
		},
		Topology: &model.TopologyConfig{Switches: switches},
	}
}

func TestTopologyRule_ValidTree(t *testing.T) {
	input := topoInput([]model.Switch{
		{Name: "s0", Nodes: "node[01-04]", Line: 1},
		{Name: "s1", Nodes: "node[05-08]", Line: 2},
		{Name: "s2", Switches: "s0,s1", Line: 3},
	})
	rule := rules.TopologyRule{}
	diags := rule.Check(input)
	if len(diags) != 0 {
		t.Errorf("expected no diagnostics, got %v", diags)
	}
}

func TestTopologyRule_BothNodesAndSwitches(t *testing.T) {
	input := topoInput([]model.Switch{
		{Name: "s0", Nodes: "node01", Switches: "s1", Line: 1},
	})
	rule := rules.TopologyRule{}
	diags := rule.Check(input)
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}
}

func TestTopologyRule_NeitherNodesNorSwitches(t *testing.T) {
	input := topoInput([]model.Switch{
		{Name: "s0", Line: 1},
	})
	rule := rules.TopologyRule{}
	diags := rule.Check(input)
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}
}

func TestTopologyRule_UndefinedSwitchReference(t *testing.T) {
	input := topoInput([]model.Switch{
		{Name: "s0", Nodes: "node[01-04]", Line: 1},
		{Name: "s2", Switches: "s0,s999", Line: 2}, // s999 undefined
	})
	rule := rules.TopologyRule{}
	diags := rule.Check(input)
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}
}

func TestTopologyRule_Cycle(t *testing.T) {
	input := topoInput([]model.Switch{
		{Name: "s0", Switches: "s1", Line: 1},
		{Name: "s1", Switches: "s0", Line: 2}, // cycle: s0→s1→s0
	})
	rule := rules.TopologyRule{}
	diags := rule.Check(input)
	if len(diags) == 0 {
		t.Error("expected cycle diagnostic, got none")
	}
}
