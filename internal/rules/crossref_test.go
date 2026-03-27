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

// TestCrossRef_NodeNameCommaList verifies that a NodeName= value containing a
// comma-separated list of bracket groups (e.g. "holy8a244[01-06],holy8a245[01-12]")
// is fully expanded — not truncated at the first bracket group.
func TestCrossRef_NodeNameCommaList(t *testing.T) {
	input := &rules.LintInput{
		SlurmFile: "slurm.conf",
		Slurm: &model.SlurmConfig{
			Globals:     map[string]string{},
			GlobalLines: map[string]int{},
			Nodes: []model.NodeDef{
				// Simulates: NodeName=holy8a244[01-06],holy8a245[01-06] ...
				{Name: "holy8a244[01-06],holy8a245[01-06]", Params: map[string]string{}, Line: 10},
			},
			Partitions: []model.Partition{
				// References nodes from the second group — would fail if expansion truncates after first ]
				{Name: "yao", Params: map[string]string{"Nodes": "holy8a245[03-06]"}, Line: 20},
			},
		},
	}
	rule := rules.CrossRefRule{}
	diags := rule.Check(input)
	if len(diags) != 0 {
		t.Errorf("expected no diagnostics, got %v", diags)
	}
}

// TestCrossRef_NodeListCommaInsideBrackets verifies that commas inside bracket
// expressions in Nodes= values (e.g. "holy8a243[01-02,11-12]") are not treated
// as list separators.
func TestCrossRef_NodeListCommaInsideBrackets(t *testing.T) {
	input := &rules.LintInput{
		SlurmFile: "slurm.conf",
		Slurm: &model.SlurmConfig{
			Globals:     map[string]string{},
			GlobalLines: map[string]int{},
			Nodes: []model.NodeDef{
				{Name: "holy8a243[01-12]", Params: map[string]string{}, Line: 10},
			},
			Partitions: []model.Partition{
				// Nodes= uses comma inside brackets — must expand to holy8a24301,holy8a24302,holy8a24311,holy8a24312
				{Name: "compute", Params: map[string]string{"Nodes": "holy8a243[01-02,11-12]"}, Line: 20},
			},
		},
	}
	rule := rules.CrossRefRule{}
	diags := rule.Check(input)
	if len(diags) != 0 {
		t.Errorf("expected no diagnostics, got %v", diags)
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
