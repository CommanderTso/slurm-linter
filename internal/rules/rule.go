package rules

import (
	"github.com/CommanderTso/slurm-linter/internal/diagnostic"
	"github.com/CommanderTso/slurm-linter/internal/model"
)

// LintInput holds all parsed configs available to rules.
// Topology may be nil if no topology.conf was provided.
type LintInput struct {
	SlurmFile    string
	TopologyFile string
	Slurm        *model.SlurmConfig
	Topology     *model.TopologyConfig // may be nil
}

// Rule is a single lint check. Each rule examines LintInput and returns
// zero or more diagnostics.
type Rule interface {
	Check(input *LintInput) []diagnostic.Diagnostic
}
