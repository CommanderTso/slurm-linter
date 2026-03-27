package parser

import (
	"fmt"
	"strconv"
	"strings"
)

// SplitNodeList splits a comma-separated list of node expressions into
// individual expressions, respecting that commas inside bracket groups are
// part of a range and must not be treated as separators.
//
// Examples:
//
//	"node01"                              → ["node01"]
//	"node[01-02,05-06]"                   → ["node[01-02,05-06]"]
//	"holy8a244[05-06],holy8a245[03-06]"   → ["holy8a244[05-06]", "holy8a245[03-06]"]
func SplitNodeList(expr string) []string {
	var parts []string
	depth := 0
	start := 0
	for i := 0; i < len(expr); i++ {
		switch expr[i] {
		case '[':
			depth++
		case ']':
			depth--
		case ',':
			if depth == 0 {
				parts = append(parts, expr[start:i])
				start = i + 1
			}
		}
	}
	if start < len(expr) {
		parts = append(parts, expr[start:])
	}
	return parts
}

// ExpandNodeRange expands a Slurm bracket range expression into a list of
// individual node names.
//
// Examples:
//
//	"node01"         → ["node01"]
//	"node[01-03]"    → ["node01","node02","node03"]
//	"node[01-03,05]" → ["node01","node02","node03","node05"]
func ExpandNodeRange(expr string) ([]string, error) {
	openIdx := strings.Index(expr, "[")
	if openIdx == -1 {
		return []string{expr}, nil
	}
	closeIdx := strings.Index(expr, "]")
	if closeIdx == -1 || closeIdx < openIdx {
		return nil, fmt.Errorf("node range %q is missing closing ']'", expr)
	}

	prefix := expr[:openIdx]
	inner := expr[openIdx+1 : closeIdx]

	var names []string
	for _, part := range strings.Split(inner, ",") {
		dashIdx := strings.Index(part, "-")
		if dashIdx == -1 {
			names = append(names, prefix+part)
			continue
		}
		startStr := part[:dashIdx]
		endStr := part[dashIdx+1:]
		start, err := strconv.Atoi(startStr)
		if err != nil {
			return nil, fmt.Errorf("invalid range start %q in %q", startStr, expr)
		}
		end, err := strconv.Atoi(endStr)
		if err != nil {
			return nil, fmt.Errorf("invalid range end %q in %q", endStr, expr)
		}
		if start > end {
			return nil, fmt.Errorf("range start %d > end %d in %q", start, end, expr)
		}
		width := len(startStr) // preserve zero-padding from original
		for i := start; i <= end; i++ {
			names = append(names, fmt.Sprintf("%s%0*d", prefix, width, i))
		}
	}
	return names, nil
}
