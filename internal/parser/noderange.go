package parser

import (
	"fmt"
	"strconv"
	"strings"
)

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
