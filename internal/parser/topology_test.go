package parser_test

import (
	"strings"
	"testing"

	"github.com/CommanderTso/slurm-linter/internal/parser"
)

const sampleTopologyConf = `
# topology.conf
SwitchName=s0 Nodes=node[01-04]
SwitchName=s1 Nodes=node[05-08]
SwitchName=s2 Switches=s[0-1]
`

func TestParseTopologyConf_Switches(t *testing.T) {
	cfg, err := parser.ParseTopologyConf("topology.conf", strings.NewReader(sampleTopologyConf))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Switches) != 3 {
		t.Fatalf("expected 3 switches, got %d", len(cfg.Switches))
	}
	if cfg.Switches[0].Name != "s0" {
		t.Errorf("Switches[0].Name = %q, want %q", cfg.Switches[0].Name, "s0")
	}
	if cfg.Switches[0].Nodes != "node[01-04]" {
		t.Errorf("Switches[0].Nodes = %q, want %q", cfg.Switches[0].Nodes, "node[01-04]")
	}
	if cfg.Switches[0].Switches != "" {
		t.Errorf("Switches[0].Switches should be empty, got %q", cfg.Switches[0].Switches)
	}
	if cfg.Switches[2].Switches != "s[0-1]" {
		t.Errorf("Switches[2].Switches = %q, want %q", cfg.Switches[2].Switches, "s[0-1]")
	}
}

func TestParseTopologyConf_LineNumbers(t *testing.T) {
	cfg, err := parser.ParseTopologyConf("topology.conf", strings.NewReader(sampleTopologyConf))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// first non-blank/non-comment line is line 3 (blank, comment, then s0)
	if cfg.Switches[0].Line != 3 {
		t.Errorf("Switches[0].Line = %d, want 3", cfg.Switches[0].Line)
	}
}
