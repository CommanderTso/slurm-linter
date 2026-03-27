package rules

import (
	"fmt"

	"github.com/CommanderTso/slurm-linter/internal/diagnostic"
	"github.com/CommanderTso/slurm-linter/internal/parser"
)

// CrossRefRule validates that node names referenced in topology.conf and
// partition stanzas are defined as NodeName= entries in slurm.conf.
type CrossRefRule struct{}

func (r CrossRefRule) Check(input *LintInput) []diagnostic.Diagnostic {
	var diags []diagnostic.Diagnostic

	// Build a set of all individual node names defined in slurm.conf NodeName stanzas.
	defined := make(map[string]bool)
	for _, node := range input.Slurm.Nodes {
		expanded, err := parser.ExpandNodeRange(node.Name)
		if err != nil {
			diags = append(diags, diagnostic.Diagnostic{
				Severity: diagnostic.Error,
				File:     input.SlurmFile,
				Line:     node.Line,
				Message:  fmt.Sprintf("NodeName %q has invalid range expression: %v", node.Name, err),
				Rule:     "cross-ref",
			})
			continue
		}
		for _, name := range expanded {
			defined[name] = true
		}
	}

	// Check topology.conf node references.
	if input.Topology != nil {
		for _, sw := range input.Topology.Switches {
			if sw.Nodes == "" {
				continue
			}
			nodes, err := ResolveNodeList(sw.Nodes)
			if err != nil {
				diags = append(diags, diagnostic.Diagnostic{
					Severity: diagnostic.Error,
					File:     input.TopologyFile,
					Line:     sw.Line,
					Message:  fmt.Sprintf("switch %q has malformed Nodes expression: %v", sw.Name, err),
					Rule:     "cross-ref",
				})
				continue
			}
			for _, name := range nodes {
				if !defined[name] {
					diags = append(diags, diagnostic.Diagnostic{
						Severity: diagnostic.Error,
						File:     input.TopologyFile,
						Line:     sw.Line,
						Message:  fmt.Sprintf("switch %q references node %q which is not defined in slurm.conf", sw.Name, name),
						Rule:     "cross-ref",
					})
				}
			}
		}
	}

	// Check partition Nodes= references.
	for _, part := range input.Slurm.Partitions {
		nodesExpr, ok := part.Params["Nodes"]
		if !ok {
			continue
		}
		nodes, err := ResolveNodeList(nodesExpr)
		if err != nil {
			diags = append(diags, diagnostic.Diagnostic{
				Severity: diagnostic.Error,
				File:     input.SlurmFile,
				Line:     part.Line,
				Message:  fmt.Sprintf("partition %q has malformed Nodes expression: %v", part.Name, err),
				Rule:     "cross-ref",
			})
			continue
		}
		for _, name := range nodes {
			if !defined[name] {
				diags = append(diags, diagnostic.Diagnostic{
					Severity: diagnostic.Error,
					File:     input.SlurmFile,
					Line:     part.Line,
					Message:  fmt.Sprintf("partition %q references node %q which is not defined in any NodeName stanza", part.Name, name),
					Rule:     "cross-ref",
				})
			}
		}
	}

	return diags
}
