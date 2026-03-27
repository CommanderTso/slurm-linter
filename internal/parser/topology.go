package parser

import (
	"fmt"
	"io"
	"strings"

	"github.com/CommanderTso/slurm-linter/internal/model"
)

// ParseTopologyConf reads a topology.conf from r and returns a parsed TopologyConfig.
func ParseTopologyConf(filename string, r io.Reader) (*model.TopologyConfig, error) {
	cfg := &model.TopologyConfig{}

	lines, err := readLines(r)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", filename, err)
	}

	for lineNum, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		tokens := splitTokens(line)
		if len(tokens) == 0 {
			continue
		}

		params := make(map[string]string)
		for _, tok := range tokens {
			k, v, ok := splitKeyValue(tok)
			if !ok {
				return nil, fmt.Errorf("%s:%d: expected Key=Value, got %q", filename, lineNum+1, tok)
			}
			params[k] = v
		}

		name, ok := params["SwitchName"]
		if !ok {
			return nil, fmt.Errorf("%s:%d: missing SwitchName", filename, lineNum+1)
		}

		sw := model.Switch{
			Name:     name,
			Nodes:    params["Nodes"],
			Switches: params["Switches"],
			Line:     lineNum + 1,
		}
		cfg.Switches = append(cfg.Switches, sw)
	}

	return cfg, nil
}
