package parser_test

import (
	"strings"
	"testing"

	"github.com/CommanderTso/slurm-linter/internal/parser"
)

const sampleSlurmConf = `
# Sample slurm.conf
ClusterName=testcluster
SlurmctldHost=controller
AuthType=auth/munge

NodeName=node[01-04] CPUs=4 RealMemory=8000 State=UNKNOWN
NodeName=gpu01 CPUs=32 RealMemory=256000 State=UNKNOWN

PartitionName=compute Nodes=node[01-04] Default=YES MaxTime=INFINITE State=UP
PartitionName=gpu Nodes=gpu01 MaxTime=24:00:00 State=UP
`

func TestParseSlurmConf_Globals(t *testing.T) {
	cfg, err := parser.ParseSlurmConf("slurm.conf", strings.NewReader(sampleSlurmConf))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := cfg.Globals["ClusterName"]; got != "testcluster" {
		t.Errorf("ClusterName = %q, want %q", got, "testcluster")
	}
	if got := cfg.Globals["AuthType"]; got != "auth/munge" {
		t.Errorf("AuthType = %q, want %q", got, "auth/munge")
	}
}

func TestParseSlurmConf_Nodes(t *testing.T) {
	cfg, err := parser.ParseSlurmConf("slurm.conf", strings.NewReader(sampleSlurmConf))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Nodes) != 2 {
		t.Fatalf("expected 2 NodeDefs, got %d", len(cfg.Nodes))
	}
	if cfg.Nodes[0].Name != "node[01-04]" {
		t.Errorf("Nodes[0].Name = %q, want %q", cfg.Nodes[0].Name, "node[01-04]")
	}
	if cfg.Nodes[0].Params["CPUs"] != "4" {
		t.Errorf("Nodes[0].CPUs = %q, want %q", cfg.Nodes[0].Params["CPUs"], "4")
	}
}

func TestParseSlurmConf_Partitions(t *testing.T) {
	cfg, err := parser.ParseSlurmConf("slurm.conf", strings.NewReader(sampleSlurmConf))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Partitions) != 2 {
		t.Fatalf("expected 2 Partitions, got %d", len(cfg.Partitions))
	}
	if cfg.Partitions[0].Name != "compute" {
		t.Errorf("Partitions[0].Name = %q, want %q", cfg.Partitions[0].Name, "compute")
	}
	if cfg.Partitions[0].Params["Nodes"] != "node[01-04]" {
		t.Errorf("Partitions[0].Nodes = %q", cfg.Partitions[0].Params["Nodes"])
	}
}

// TestParseSlurmConf_SpacesAroundEquals verifies that Slurm's allowed
// "Key = Value" syntax (spaces around =) is parsed correctly.
func TestParseSlurmConf_SpacesAroundEquals(t *testing.T) {
	input := "ClusterName = testcluster\nSlurmctldHost = controller\n" +
		"NodeName = node01 CPUs = 4\n" +
		"PartitionName = compute Nodes = node01\n"
	cfg, err := parser.ParseSlurmConf("slurm.conf", strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := cfg.Globals["ClusterName"]; got != "testcluster" {
		t.Errorf("ClusterName = %q, want %q", got, "testcluster")
	}
	if got := cfg.Globals["SlurmctldHost"]; got != "controller" {
		t.Errorf("SlurmctldHost = %q, want %q", got, "controller")
	}
	if len(cfg.Nodes) != 1 || cfg.Nodes[0].Name != "node01" {
		t.Errorf("expected 1 node named node01, got %v", cfg.Nodes)
	}
	if cfg.Nodes[0].Params["CPUs"] != "4" {
		t.Errorf("CPUs = %q, want %q", cfg.Nodes[0].Params["CPUs"], "4")
	}
	if len(cfg.Partitions) != 1 || cfg.Partitions[0].Name != "compute" {
		t.Errorf("expected 1 partition named compute, got %v", cfg.Partitions)
	}
}

func TestParseSlurmConf_LineContinuation(t *testing.T) {
	input := "ClusterName=my\\\ncluster\n"
	cfg, err := parser.ParseSlurmConf("slurm.conf", strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := cfg.Globals["ClusterName"]; got != "mycluster" {
		t.Errorf("ClusterName with continuation = %q, want %q", got, "mycluster")
	}
}
