package parser

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/CommanderTso/slurm-linter/internal/model"
)

// ParseSlurmConf reads a slurm.conf from r and returns a parsed SlurmConfig.
// filename is used only for error messages.
func ParseSlurmConf(filename string, r io.Reader) (*model.SlurmConfig, error) {
	cfg := &model.SlurmConfig{
		Globals:     make(map[string]string),
		GlobalLines: make(map[string]int),
	}

	lines, err := readLines(r)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", filename, err)
	}

	for lineNum, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split into space-separated tokens: each token is Key=Value
		tokens := splitTokens(line)
		if len(tokens) == 0 {
			continue
		}

		firstKey, firstVal, ok := splitKeyValue(tokens[0])
		if !ok {
			return nil, fmt.Errorf("%s:%d: expected Key=Value, got %q", filename, lineNum+1, tokens[0])
		}

		switch strings.ToLower(firstKey) {
		case "nodename":
			node := model.NodeDef{
				Name:   firstVal,
				Params: make(map[string]string),
				Line:   lineNum + 1,
			}
			for _, tok := range tokens[1:] {
				k, v, ok := splitKeyValue(tok)
				if !ok {
					return nil, fmt.Errorf("%s:%d: expected Key=Value in NodeName stanza, got %q", filename, lineNum+1, tok)
				}
				node.Params[k] = v
			}
			cfg.Nodes = append(cfg.Nodes, node)

		case "partitionname":
			part := model.Partition{
				Name:   firstVal,
				Params: make(map[string]string),
				Line:   lineNum + 1,
			}
			for _, tok := range tokens[1:] {
				k, v, ok := splitKeyValue(tok)
				if !ok {
					return nil, fmt.Errorf("%s:%d: expected Key=Value in PartitionName stanza, got %q", filename, lineNum+1, tok)
				}
				part.Params[k] = v
			}
			cfg.Partitions = append(cfg.Partitions, part)

		default:
			if len(tokens) > 1 {
				return nil, fmt.Errorf("%s:%d: unexpected tokens after %s=%s", filename, lineNum+1, firstKey, firstVal)
			}
			cfg.Globals[firstKey] = firstVal
			cfg.GlobalLines[firstKey] = lineNum + 1
		}
	}

	return cfg, nil
}

// readLines reads all lines from r, joining continuation lines (ending in \).
// Returns a slice where continuation lines are merged; placeholder empty strings
// preserve the original line count for accurate line number reporting.
func readLines(r io.Reader) ([]string, error) {
	scanner := bufio.NewScanner(r)
	var logical []string
	var current strings.Builder

	for scanner.Scan() {
		raw := scanner.Text()
		if strings.HasSuffix(raw, "\\") {
			current.WriteString(strings.TrimSuffix(raw, "\\"))
			logical = append(logical, "") // placeholder preserves line numbering
		} else {
			current.WriteString(raw)
			logical = append(logical, current.String())
			current.Reset()
		}
	}
	if current.Len() > 0 {
		logical = append(logical, current.String())
	}
	return logical, scanner.Err()
}

// splitTokens splits a line into whitespace-separated tokens.
func splitTokens(line string) []string {
	return strings.Fields(line)
}

// splitKeyValue splits "Key=Value" into ("Key", "Value", true).
// Returns ("", "", false) if the token contains no "=".
func splitKeyValue(token string) (key, value string, ok bool) {
	idx := strings.IndexByte(token, '=')
	if idx == -1 {
		return "", "", false
	}
	return token[:idx], token[idx+1:], true
}
