package rules_test

import (
	"testing"

	"github.com/CommanderTso/slurm-linter/internal/model"
	"github.com/CommanderTso/slurm-linter/internal/rules"
)

func crossRefInput() *rules.LintInput {
	return &rules.LintInput{
		SlurmFile:    "slurm.conf",
		TopologyFile: "topology.conf",
		Slurm: &model.SlurmConfig{
			Globals:     map[string]string{},
			GlobalLines: map[string]int{},
			Nodes: []model.NodeDef{
				{Name: "node[01-04]", Params: map[string]string{}, Line: 5},
			},
			Partitions: []model.Partition{
				{Name: "compute", Params: map[string]string{"Nodes": "node[01-04]"}, Line: 8},
			},
		},
		Topology: &model.TopologyConfig{
			Switches: []model.Switch{
				{Name: "s0", Nodes: "node[01-04]", Line: 1},
			},
		},
	}
}

func TestCrossRef_AllValid(t *testing.T) {
	input := crossRefInput()
	rule := rules.CrossRefRule{}
	diags := rule.Check(input)
	if len(diags) != 0 {
		t.Errorf("expected no diagnostics, got %v", diags)
	}
}

func TestCrossRef_TopologyNodeNotInSlurm(t *testing.T) {
	input := crossRefInput()
	input.Topology.Switches = append(input.Topology.Switches, model.Switch{
		Name: "s1", Nodes: "gpu[01-02]", Line: 2,
	})
	rule := rules.CrossRefRule{}
	diags := rule.Check(input)
	if len(diags) == 0 {
		t.Error("expected diagnostics for undefined nodes, got none")
	}
}

func TestCrossRef_PartitionNodeNotInSlurm(t *testing.T) {
	input := crossRefInput()
	input.Slurm.Partitions = append(input.Slurm.Partitions, model.Partition{
		Name:   "gpu",
		Params: map[string]string{"Nodes": "gpu01"},
		Line:   9,
	})
	rule := rules.CrossRefRule{}
	diags := rule.Check(input)
	if len(diags) == 0 {
		t.Error("expected diagnostic for partition referencing undefined node")
	}
}

func TestCrossRef_NoTopology(t *testing.T) {
	input := crossRefInput()
	input.Topology = nil
	rule := rules.CrossRefRule{}
	diags := rule.Check(input)
	if len(diags) != 0 {
		t.Errorf("expected no diagnostics with no topology, got %v", diags)
	}
}
