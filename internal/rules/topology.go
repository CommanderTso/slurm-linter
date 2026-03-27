package rules

import (
	"fmt"
	"strings"

	"github.com/CommanderTso/slurm-linter/internal/diagnostic"
	"github.com/CommanderTso/slurm-linter/internal/model"
	"github.com/CommanderTso/slurm-linter/internal/parser"
)

// TopologyRule validates the structure of topology.conf.
type TopologyRule struct{}

func (r TopologyRule) Check(input *LintInput) []diagnostic.Diagnostic {
	if input.Topology == nil {
		return nil
	}
	var diags []diagnostic.Diagnostic
	topo := input.Topology

	// Build a set of all defined switch names for O(1) lookup.
	defined := make(map[string]bool, len(topo.Switches))
	for _, sw := range topo.Switches {
		defined[sw.Name] = true
	}

	// Build adjacency list for cycle detection (switch → child switches).
	edges := make(map[string][]string)

	for _, sw := range topo.Switches {
		hasBoth := sw.Nodes != "" && sw.Switches != ""
		hasNeither := sw.Nodes == "" && sw.Switches == ""

		if hasBoth {
			diags = append(diags, diagnostic.Diagnostic{
				Severity: diagnostic.Error,
				File:     input.TopologyFile,
				Line:     sw.Line,
				Message:  fmt.Sprintf("switch %q defines both Nodes and Switches; only one is allowed", sw.Name),
				Rule:     "topology",
			})
			continue
		}
		if hasNeither {
			diags = append(diags, diagnostic.Diagnostic{
				Severity: diagnostic.Error,
				File:     input.TopologyFile,
				Line:     sw.Line,
				Message:  fmt.Sprintf("switch %q defines neither Nodes nor Switches", sw.Name),
				Rule:     "topology",
			})
			continue
		}

		if sw.Switches != "" {
			children, err := expandSwitchList(sw.Switches)
			if err != nil {
				diags = append(diags, diagnostic.Diagnostic{
					Severity: diagnostic.Error,
					File:     input.TopologyFile,
					Line:     sw.Line,
					Message:  fmt.Sprintf("switch %q has malformed Switches value: %v", sw.Name, err),
					Rule:     "topology",
				})
				continue
			}
			for _, child := range children {
				if !defined[child] {
					diags = append(diags, diagnostic.Diagnostic{
						Severity: diagnostic.Error,
						File:     input.TopologyFile,
						Line:     sw.Line,
						Message:  fmt.Sprintf("switch %q references undefined switch %q", sw.Name, child),
						Rule:     "topology",
					})
				}
			}
			edges[sw.Name] = children
		}
	}

	// Detect cycles using DFS.
	if cycles := findCycles(edges); len(cycles) > 0 {
		diags = append(diags, diagnostic.Diagnostic{
			Severity: diagnostic.Error,
			File:     input.TopologyFile,
			Line:     0,
			Message:  fmt.Sprintf("topology switch graph contains a cycle involving: %s", strings.Join(cycles, ", ")),
			Rule:     "topology",
		})
	}

	return diags
}

// expandSwitchList expands a comma/bracket switch list like "s[0-2]" or "s0,s1".
func expandSwitchList(expr string) ([]string, error) {
	var result []string
	for _, part := range strings.Split(expr, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		expanded, err := parser.ExpandNodeRange(part)
		if err != nil {
			return nil, err
		}
		result = append(result, expanded...)
	}
	return result, nil
}

// findCycles returns names of nodes involved in a cycle, or nil if none.
func findCycles(edges map[string][]string) []string {
	const (
		unvisited = 0
		inStack   = 1
		done      = 2
	)
	state := make(map[string]int)
	var cycleNodes []string

	var dfs func(node string) bool
	dfs = func(node string) bool {
		state[node] = inStack
		for _, child := range edges[node] {
			if state[child] == inStack {
				cycleNodes = append(cycleNodes, child)
				return true
			}
			if state[child] == unvisited {
				if dfs(child) {
					return true
				}
			}
		}
		state[node] = done
		return false
	}

	for node := range edges {
		if state[node] == unvisited {
			if dfs(node) {
				return cycleNodes
			}
		}
	}
	return nil
}

// ResolveNodeList expands a comma-separated list of node range expressions.
// Exported for use by the cross-reference rule.
//
// Uses SplitNodeList to correctly handle commas inside bracket groups
// (e.g. "node[01-02,11-12]") before expanding each individual expression.
func ResolveNodeList(expr string) ([]string, error) {
	var result []string
	for _, part := range parser.SplitNodeList(expr) {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		expanded, err := parser.ExpandNodeRange(part)
		if err != nil {
			return nil, err
		}
		result = append(result, expanded...)
	}
	return result, nil
}

// AllTopologyNodes returns the set of all node names referenced in topology.conf.
func AllTopologyNodes(topo *model.TopologyConfig) (map[string]bool, error) {
	nodes := make(map[string]bool)
	for _, sw := range topo.Switches {
		if sw.Nodes == "" {
			continue
		}
		expanded, err := ResolveNodeList(sw.Nodes)
		if err != nil {
			return nil, err
		}
		for _, n := range expanded {
			nodes[n] = true
		}
	}
	return nodes, nil
}
