package validator_test

import (
	"os"
	"strings"
	"testing"

	"github.com/CommanderTso/slurm-linter/internal/diagnostic"
	"github.com/CommanderTso/slurm-linter/internal/parser"
	"github.com/CommanderTso/slurm-linter/internal/rules"
	"github.com/CommanderTso/slurm-linter/internal/validator"
)

func TestIntegration_ValidConfig(t *testing.T) {
	sf, err := os.Open("../../testdata/valid/slurm.conf")
	if err != nil {
		t.Fatalf("open slurm.conf: %v", err)
	}
	defer sf.Close()

	tf, err := os.Open("../../testdata/valid/topology.conf")
	if err != nil {
		t.Fatalf("open topology.conf: %v", err)
	}
	defer tf.Close()

	slurm, err := parser.ParseSlurmConf("slurm.conf", sf)
	if err != nil {
		t.Fatalf("parse slurm.conf: %v", err)
	}
	topo, err := parser.ParseTopologyConf("topology.conf", tf)
	if err != nil {
		t.Fatalf("parse topology.conf: %v", err)
	}

	input := &rules.LintInput{
		SlurmFile:    "slurm.conf",
		TopologyFile: "topology.conf",
		Slurm:        slurm,
		Topology:     topo,
	}

	v := validator.New(validator.DefaultRules()...)
	diags := v.Run(input)
	for _, d := range diags {
		t.Errorf("unexpected diagnostic: [%s] %s:%d %s", d.Rule, d.File, d.Line, d.Message)
	}
}

func TestIntegration_MissingRequired(t *testing.T) {
	input := strings.NewReader(`AuthType=auth/munge`)
	slurm, err := parser.ParseSlurmConf("slurm.conf", input)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	lint := &rules.LintInput{
		SlurmFile: "slurm.conf",
		Slurm:     slurm,
	}

	v := validator.New(validator.DefaultRules()...)
	diags := v.Run(lint)

	var hasError bool
	for _, d := range diags {
		if d.Severity == diagnostic.Error {
			hasError = true
		}
	}
	if !hasError {
		t.Error("expected at least one error diagnostic")
	}
}
