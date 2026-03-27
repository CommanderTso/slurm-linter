# Slurm Config Linter — MVP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a CLI tool that validates `slurm.conf` and `topology.conf` files, reporting errors, warnings, and cross-file inconsistencies.

**Architecture:** A Go CLI (cobra) that parses both config files into an in-memory model, runs composable lint rules against that model, and reports diagnostics in text or JSON format. Rules are organized by domain: required params, type/enum checks, topology graph checks, and cross-file reference checks. All parsing and validation is strictly separated.

**Tech Stack:** Go 1.22+, `github.com/spf13/cobra` (CLI), standard library only for parsing and output.

---

## File Map

| File | Responsibility |
|------|---------------|
| `cmd/slurm-linter/main.go` | cobra CLI entry point, wires everything together |
| `internal/diagnostic/diagnostic.go` | `Severity` and `Diagnostic` types |
| `internal/model/model.go` | `SlurmConfig`, `NodeDef`, `Partition`, `TopologyConfig`, `Switch` types |
| `internal/parser/noderange.go` | Expand Slurm bracket range expressions → `[]string` |
| `internal/parser/noderange_test.go` | Tests for node range expansion |
| `internal/parser/slurmconf.go` | Parse `slurm.conf` → `*model.SlurmConfig` |
| `internal/parser/slurmconf_test.go` | Tests for slurm.conf parsing |
| `internal/parser/topology.go` | Parse `topology.conf` → `*model.TopologyConfig` |
| `internal/parser/topology_test.go` | Tests for topology.conf parsing |
| `internal/rules/rule.go` | `Rule` interface and `LintInput` type |
| `internal/rules/required.go` | Rule: required top-level parameters present |
| `internal/rules/required_test.go` | Tests |
| `internal/rules/types.go` | Rule: integer, enum, time, memory type validation |
| `internal/rules/types_test.go` | Tests |
| `internal/rules/topology.go` | Rule: topology graph structure (references, cycles) |
| `internal/rules/topology_test.go` | Tests |
| `internal/rules/crossref.go` | Rule: cross-file node/partition reference validation |
| `internal/rules/crossref_test.go` | Tests |
| `internal/validator/validator.go` | Orchestrate rules, collect diagnostics |
| `internal/validator/validator_test.go` | Integration tests via testdata fixtures |
| `internal/reporter/reporter.go` | `Reporter` interface |
| `internal/reporter/text.go` | Human-readable text output |
| `internal/reporter/json.go` | JSON output |
| `internal/reporter/reporter_test.go` | Tests for both reporters |
| `testdata/valid/slurm.conf` | Fixture: a complete valid config |
| `testdata/valid/topology.conf` | Fixture: a valid topology |
| `testdata/invalid/missing_required/slurm.conf` | Fixture: missing ClusterName |
| `testdata/invalid/bad_topology/topology.conf` | Fixture: undefined switch reference |
| `Makefile` | `build`, `test`, `lint`, `clean` targets |
| `go.mod` | Module definition |

---

## Task 1: Project Scaffold

**Files:**
- Create: `go.mod`
- Create: `Makefile`
- Create: `cmd/slurm-linter/main.go`

- [ ] **Step 1: Initialize Go module**

```bash
cd /Users/smacmillan/projects/claude/slurm-linter
go mod init github.com/CommanderTso/slurm-linter
```

- [ ] **Step 2: Install cobra**

```bash
go get github.com/spf13/cobra@latest
```

- [ ] **Step 3: Create stub main.go**

Create `cmd/slurm-linter/main.go`:

```go
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "slurm-linter",
		Short: "Validate Slurm configuration files",
	}

	checkCmd := &cobra.Command{
		Use:   "check",
		Short: "Lint slurm.conf and topology.conf",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("not yet implemented")
			return nil
		},
	}

	rootCmd.AddCommand(checkCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
```

- [ ] **Step 4: Create Makefile**

```makefile
.PHONY: build test lint clean

BIN := bin/slurm-linter

build:
	go build -o $(BIN) ./cmd/slurm-linter

test:
	go test ./...

test-v:
	go test -v ./...

lint:
	go vet ./...

clean:
	rm -rf bin/
```

- [ ] **Step 5: Verify it builds**

```bash
make build
```

Expected: `bin/slurm-linter` created, no errors.

- [ ] **Step 6: Commit**

```bash
git -C /Users/smacmillan/projects/claude/slurm-linter init
git -C /Users/smacmillan/projects/claude/slurm-linter add go.mod go.sum Makefile cmd/
git -C /Users/smacmillan/projects/claude/slurm-linter commit -m "chore: initialize Go module and CLI scaffold"
```

---

## Task 2: Diagnostic Model

**Files:**
- Create: `internal/diagnostic/diagnostic.go`
- Create: `internal/diagnostic/diagnostic_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/diagnostic/diagnostic_test.go`:

```go
package diagnostic_test

import (
	"testing"

	"github.com/CommanderTso/slurm-linter/internal/diagnostic"
)

func TestSeverityString(t *testing.T) {
	cases := []struct {
		s    diagnostic.Severity
		want string
	}{
		{diagnostic.Info, "info"},
		{diagnostic.Warning, "warning"},
		{diagnostic.Error, "error"},
	}
	for _, c := range cases {
		if got := c.s.String(); got != c.want {
			t.Errorf("Severity(%d).String() = %q, want %q", c.s, got, c.want)
		}
	}
}

func TestDiagnosticFields(t *testing.T) {
	d := diagnostic.Diagnostic{
		Severity: diagnostic.Error,
		File:     "slurm.conf",
		Line:     42,
		Message:  "ClusterName is required",
		Rule:     "required-params",
	}
	if d.File != "slurm.conf" {
		t.Errorf("expected File=slurm.conf, got %q", d.File)
	}
	if d.Line != 42 {
		t.Errorf("expected Line=42, got %d", d.Line)
	}
}
```

- [ ] **Step 2: Run to confirm failure**

```bash
go test ./internal/diagnostic/...
```

Expected: `cannot find package` or `no Go files` — confirms the type doesn't exist yet.

- [ ] **Step 3: Write implementation**

Create `internal/diagnostic/diagnostic.go`:

```go
package diagnostic

// Severity indicates the seriousness of a diagnostic.
type Severity int

const (
	Info    Severity = iota // informational, no action required
	Warning                 // likely misconfiguration
	Error                   // definite error, job submission will fail
)

// String returns the lowercase name of the severity level.
func (s Severity) String() string {
	switch s {
	case Info:
		return "info"
	case Warning:
		return "warning"
	case Error:
		return "error"
	default:
		return "unknown"
	}
}

// Diagnostic is a single lint finding.
type Diagnostic struct {
	Severity Severity
	File     string // path to the file containing the issue
	Line     int    // 1-based line number; 0 if not applicable
	Message  string // human-readable description
	Rule     string // machine-readable rule identifier (e.g. "required-params")
}
```

- [ ] **Step 4: Run to confirm pass**

```bash
go test ./internal/diagnostic/...
```

Expected: `ok  github.com/CommanderTso/slurm-linter/internal/diagnostic`

- [ ] **Step 5: Commit**

```bash
git -C /Users/smacmillan/projects/claude/slurm-linter add internal/diagnostic/
git -C /Users/smacmillan/projects/claude/slurm-linter commit -m "feat: add Diagnostic and Severity types"
```

---

## Task 3: Config Model Types

**Files:**
- Create: `internal/model/model.go`

No tests here — this is a pure data type file. Types will be tested through the parsers and rules that use them.

- [ ] **Step 1: Create model.go**

Create `internal/model/model.go`:

```go
package model

// SlurmConfig is the parsed in-memory representation of slurm.conf.
type SlurmConfig struct {
	// Globals holds top-level key=value parameters (e.g. ClusterName, AuthType).
	// Key is the parameter name; value is the raw string value.
	Globals map[string]string

	// GlobalLines maps each global parameter name to its 1-based line number.
	GlobalLines map[string]int

	// Nodes holds all NodeName= stanza definitions.
	Nodes []NodeDef

	// Partitions holds all PartitionName= stanza definitions.
	Partitions []Partition
}

// NodeDef represents one NodeName= line in slurm.conf.
// Name may be a bracket range expression like "node[01-04]".
type NodeDef struct {
	Name   string            // raw name expression
	Params map[string]string // additional params (CPUs, RealMemory, State, etc.)
	Line   int               // 1-based line number
}

// Partition represents one PartitionName= line in slurm.conf.
type Partition struct {
	Name   string
	Params map[string]string // Nodes, MaxTime, Default, State, etc.
	Line   int
}

// TopologyConfig is the parsed in-memory representation of topology.conf.
type TopologyConfig struct {
	Switches []Switch
}

// Switch represents one SwitchName= line in topology.conf.
// Exactly one of Nodes or Switches will be non-empty.
type Switch struct {
	Name     string // switch name
	Nodes    string // raw node range expression (e.g. "node[01-04]"), or ""
	Switches string // raw switch range expression (e.g. "s[0-2]"), or ""
	Line     int    // 1-based line number
}
```

- [ ] **Step 2: Verify it compiles**

```bash
go build ./internal/model/...
```

Expected: no output, no errors.

- [ ] **Step 3: Commit**

```bash
git -C /Users/smacmillan/projects/claude/slurm-linter add internal/model/
git -C /Users/smacmillan/projects/claude/slurm-linter commit -m "feat: add SlurmConfig and TopologyConfig model types"
```

---

## Task 4: Node Range Parser

Slurm uses bracket expressions to refer to groups of nodes: `node[01-04,06,08-10]`.
This parser expands them to a flat `[]string`.

**Files:**
- Create: `internal/parser/noderange.go`
- Create: `internal/parser/noderange_test.go`

- [ ] **Step 1: Write the failing tests**

Create `internal/parser/noderange_test.go`:

```go
package parser_test

import (
	"reflect"
	"testing"

	"github.com/CommanderTso/slurm-linter/internal/parser"
)

func TestExpandNodeRange(t *testing.T) {
	cases := []struct {
		expr string
		want []string
		err  bool
	}{
		// plain name, no brackets
		{"node01", []string{"node01"}, false},
		// single number in brackets
		{"node[5]", []string{"node5"}, false},
		// simple range
		{"node[01-03]", []string{"node01", "node02", "node03"}, false},
		// zero-padded range
		{"node[001-003]", []string{"node001", "node002", "node003"}, false},
		// comma list
		{"node[1,3,5]", []string{"node1", "node3", "node5"}, false},
		// mixed range and singles
		{"node[01-03,05]", []string{"node01", "node02", "node03", "node05"}, false},
		// multi-range
		{"node[01-02,04-05]", []string{"node01", "node02", "node04", "node05"}, false},
		// missing close bracket → error
		{"node[01-03", nil, true},
		// start > end → error
		{"node[05-03]", nil, true},
	}

	for _, c := range cases {
		got, err := parser.ExpandNodeRange(c.expr)
		if c.err {
			if err == nil {
				t.Errorf("ExpandNodeRange(%q): expected error, got nil", c.expr)
			}
			continue
		}
		if err != nil {
			t.Errorf("ExpandNodeRange(%q): unexpected error: %v", c.expr, err)
			continue
		}
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("ExpandNodeRange(%q) = %v, want %v", c.expr, got, c.want)
		}
	}
}
```

- [ ] **Step 2: Run to confirm failure**

```bash
go test ./internal/parser/...
```

Expected: `undefined: parser.ExpandNodeRange`

- [ ] **Step 3: Write implementation**

Create `internal/parser/noderange.go`:

```go
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
```

- [ ] **Step 4: Run to confirm pass**

```bash
go test ./internal/parser/...
```

Expected: `ok  github.com/CommanderTso/slurm-linter/internal/parser`

- [ ] **Step 5: Commit**

```bash
git -C /Users/smacmillan/projects/claude/slurm-linter add internal/parser/
git -C /Users/smacmillan/projects/claude/slurm-linter commit -m "feat: add node range expression parser"
```

---

## Task 5: slurm.conf Parser

`slurm.conf` line types:
- Blank lines and `#` comments → ignored
- `NodeName=<expr> Key=Val Key=Val ...` → `NodeDef`
- `PartitionName=<name> Key=Val Key=Val ...` → `Partition`
- `Key=Value` → global param
- Lines ending with `\` → continuation (joined to next line)

**Files:**
- Create: `internal/parser/slurmconf.go`
- Add to: `internal/parser/noderange_test.go` → actually a new file:
- Create: `internal/parser/slurmconf_test.go`

- [ ] **Step 1: Write the failing tests**

Create `internal/parser/slurmconf_test.go`:

```go
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
```

- [ ] **Step 2: Run to confirm failure**

```bash
go test ./internal/parser/...
```

Expected: `undefined: parser.ParseSlurmConf`

- [ ] **Step 3: Write implementation**

Create `internal/parser/slurmconf.go`:

```go
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

		switch firstKey {
		case "NodeName":
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

		case "PartitionName":
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
// Returns a slice of logical lines with 1-based index matching original line numbers
// (continuation lines are merged; the slice entry for a joined line uses the first line's index).
func readLines(r io.Reader) ([]string, error) {
	scanner := bufio.NewScanner(r)
	var logical []string
	var current strings.Builder

	for scanner.Scan() {
		raw := scanner.Text()
		// Strip inline comment (but not inside values — simple heuristic)
		if idx := strings.Index(raw, " #"); idx != -1 {
			raw = raw[:idx]
		}
		if strings.HasSuffix(raw, "\\") {
			current.WriteString(strings.TrimSuffix(raw, "\\"))
			logical = append(logical, "") // placeholder to preserve line numbering
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

// splitTokens splits a line into whitespace-separated tokens, respecting
// that values may contain no spaces (Slurm config values don't use quoting
// for multi-word values; they use comma separation).
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
```

- [ ] **Step 4: Run to confirm pass**

```bash
go test ./internal/parser/...
```

Expected: `ok  github.com/CommanderTso/slurm-linter/internal/parser`

- [ ] **Step 5: Commit**

```bash
git -C /Users/smacmillan/projects/claude/slurm-linter add internal/parser/slurmconf.go internal/parser/slurmconf_test.go
git -C /Users/smacmillan/projects/claude/slurm-linter commit -m "feat: add slurm.conf parser"
```

---

## Task 6: topology.conf Parser

`topology.conf` line format:
- `SwitchName=<name> Nodes=<noderange>` — leaf switch connected to nodes
- `SwitchName=<name> Switches=<switchrange>` — aggregation switch connected to other switches

**Files:**
- Create: `internal/parser/topology.go`
- Create: `internal/parser/topology_test.go`

- [ ] **Step 1: Write the failing tests**

Create `internal/parser/topology_test.go`:

```go
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
	// first non-blank/non-comment line is line 2
	if cfg.Switches[0].Line != 2 {
		t.Errorf("Switches[0].Line = %d, want 2", cfg.Switches[0].Line)
	}
}
```

- [ ] **Step 2: Run to confirm failure**

```bash
go test ./internal/parser/...
```

Expected: `undefined: parser.ParseTopologyConf`

- [ ] **Step 3: Write implementation**

Create `internal/parser/topology.go`:

```go
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
```

- [ ] **Step 4: Run to confirm pass**

```bash
go test ./internal/parser/...
```

Expected: `ok  github.com/CommanderTso/slurm-linter/internal/parser`

- [ ] **Step 5: Commit**

```bash
git -C /Users/smacmillan/projects/claude/slurm-linter add internal/parser/topology.go internal/parser/topology_test.go
git -C /Users/smacmillan/projects/claude/slurm-linter commit -m "feat: add topology.conf parser"
```

---

## Task 7: Rule Interface

**Files:**
- Create: `internal/rules/rule.go`

- [ ] **Step 1: Create rule.go**

Create `internal/rules/rule.go`:

```go
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
```

- [ ] **Step 2: Verify it compiles**

```bash
go build ./internal/rules/...
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git -C /Users/smacmillan/projects/claude/slurm-linter add internal/rules/rule.go
git -C /Users/smacmillan/projects/claude/slurm-linter commit -m "feat: add Rule interface and LintInput type"
```

---

## Task 8: Required Parameters Rule

Checks that the minimum required global parameters are present in `slurm.conf`.
Required: `ClusterName`, `SlurmctldHost`. Must also have at least one `NodeName` and one `PartitionName` stanza.

**Files:**
- Create: `internal/rules/required.go`
- Create: `internal/rules/required_test.go`

- [ ] **Step 1: Write the failing tests**

Create `internal/rules/required_test.go`:

```go
package rules_test

import (
	"testing"

	"github.com/CommanderTso/slurm-linter/internal/diagnostic"
	"github.com/CommanderTso/slurm-linter/internal/model"
	"github.com/CommanderTso/slurm-linter/internal/rules"
)

func makeInput(globals map[string]string, nodes []model.NodeDef, partitions []model.Partition) *rules.LintInput {
	return &rules.LintInput{
		SlurmFile: "slurm.conf",
		Slurm: &model.SlurmConfig{
			Globals:     globals,
			GlobalLines: make(map[string]int),
			Nodes:       nodes,
			Partitions:  partitions,
		},
	}
}

func TestRequiredParams_AllPresent(t *testing.T) {
	input := makeInput(
		map[string]string{"ClusterName": "test", "SlurmctldHost": "ctrl"},
		[]model.NodeDef{{Name: "node01", Params: map[string]string{}}},
		[]model.Partition{{Name: "compute", Params: map[string]string{}}},
	)
	rule := rules.RequiredParams{}
	diags := rule.Check(input)
	if len(diags) != 0 {
		t.Errorf("expected no diagnostics, got %d: %v", len(diags), diags)
	}
}

func TestRequiredParams_MissingClusterName(t *testing.T) {
	input := makeInput(
		map[string]string{"SlurmctldHost": "ctrl"},
		[]model.NodeDef{{Name: "node01", Params: map[string]string{}}},
		[]model.Partition{{Name: "compute", Params: map[string]string{}}},
	)
	rule := rules.RequiredParams{}
	diags := rule.Check(input)
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}
	if diags[0].Severity != diagnostic.Error {
		t.Errorf("expected Error severity, got %v", diags[0].Severity)
	}
	if diags[0].Rule != "required-params" {
		t.Errorf("expected rule=required-params, got %q", diags[0].Rule)
	}
}

func TestRequiredParams_NoNodes(t *testing.T) {
	input := makeInput(
		map[string]string{"ClusterName": "test", "SlurmctldHost": "ctrl"},
		nil,
		[]model.Partition{{Name: "compute", Params: map[string]string{}}},
	)
	rule := rules.RequiredParams{}
	diags := rule.Check(input)
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}
}

func TestRequiredParams_NoPartitions(t *testing.T) {
	input := makeInput(
		map[string]string{"ClusterName": "test", "SlurmctldHost": "ctrl"},
		[]model.NodeDef{{Name: "node01", Params: map[string]string{}}},
		nil,
	)
	rule := rules.RequiredParams{}
	diags := rule.Check(input)
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}
}
```

- [ ] **Step 2: Run to confirm failure**

```bash
go test ./internal/rules/...
```

Expected: `undefined: rules.RequiredParams`

- [ ] **Step 3: Write implementation**

Create `internal/rules/required.go`:

```go
package rules

import (
	"fmt"

	"github.com/CommanderTso/slurm-linter/internal/diagnostic"
)

// RequiredParams checks that the minimum required global parameters are present.
type RequiredParams struct{}

var requiredGlobals = []string{"ClusterName", "SlurmctldHost"}

func (r RequiredParams) Check(input *LintInput) []diagnostic.Diagnostic {
	var diags []diagnostic.Diagnostic
	cfg := input.Slurm

	for _, key := range requiredGlobals {
		if _, ok := cfg.Globals[key]; !ok {
			diags = append(diags, diagnostic.Diagnostic{
				Severity: diagnostic.Error,
				File:     input.SlurmFile,
				Line:     0,
				Message:  fmt.Sprintf("required parameter %q is missing", key),
				Rule:     "required-params",
			})
		}
	}

	if len(cfg.Nodes) == 0 {
		diags = append(diags, diagnostic.Diagnostic{
			Severity: diagnostic.Error,
			File:     input.SlurmFile,
			Line:     0,
			Message:  "no NodeName stanzas defined",
			Rule:     "required-params",
		})
	}

	if len(cfg.Partitions) == 0 {
		diags = append(diags, diagnostic.Diagnostic{
			Severity: diagnostic.Error,
			File:     input.SlurmFile,
			Line:     0,
			Message:  "no PartitionName stanzas defined",
			Rule:     "required-params",
		})
	}

	return diags
}
```

- [ ] **Step 4: Run to confirm pass**

```bash
go test ./internal/rules/...
```

Expected: `ok  github.com/CommanderTso/slurm-linter/internal/rules`

- [ ] **Step 5: Commit**

```bash
git -C /Users/smacmillan/projects/claude/slurm-linter add internal/rules/
git -C /Users/smacmillan/projects/claude/slurm-linter commit -m "feat: add required-params lint rule"
```

---

## Task 9: Type Validation Rule

Validates that specific global parameters have values of the correct type or belong to a known set of valid values.

Checks:
- Integer params (`KillWait`, `MaxJobCount`, `MinJobAge`, `InactiveLimit`) are valid non-negative integers
- Enum params (`SchedulerType`, `AuthType`, `SelectType`) are one of the known valid values
- Time format params (`DefaultTime`, `MaxTime`) match Slurm time format: `INFINITE`, `UNLIMITED`, or `minutes`, `hours:minutes`, `hours:minutes:seconds`, `days-hours:minutes:seconds`

**Files:**
- Create: `internal/rules/types.go`
- Create: `internal/rules/types_test.go`

- [ ] **Step 1: Write the failing tests**

Create `internal/rules/types_test.go`:

```go
package rules_test

import (
	"testing"

	"github.com/CommanderTso/slurm-linter/internal/diagnostic"
	"github.com/CommanderTso/slurm-linter/internal/model"
	"github.com/CommanderTso/slurm-linter/internal/rules"
)

func makeInputWithGlobals(globals map[string]string) *rules.LintInput {
	return &rules.LintInput{
		SlurmFile: "slurm.conf",
		Slurm: &model.SlurmConfig{
			Globals:     globals,
			GlobalLines: map[string]int{"KillWait": 5, "SchedulerType": 6},
		},
	}
}

func TestTypeValidation_ValidInteger(t *testing.T) {
	input := makeInputWithGlobals(map[string]string{"KillWait": "30"})
	rule := rules.TypeValidation{}
	diags := rule.Check(input)
	if len(diags) != 0 {
		t.Errorf("expected no diagnostics, got %v", diags)
	}
}

func TestTypeValidation_InvalidInteger(t *testing.T) {
	input := makeInputWithGlobals(map[string]string{"KillWait": "notanumber"})
	rule := rules.TypeValidation{}
	diags := rule.Check(input)
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}
	if diags[0].Severity != diagnostic.Error {
		t.Errorf("expected Error, got %v", diags[0].Severity)
	}
}

func TestTypeValidation_NegativeInteger(t *testing.T) {
	input := makeInputWithGlobals(map[string]string{"KillWait": "-5"})
	rule := rules.TypeValidation{}
	diags := rule.Check(input)
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}
}

func TestTypeValidation_ValidEnum(t *testing.T) {
	input := makeInputWithGlobals(map[string]string{"SchedulerType": "sched/backfill"})
	rule := rules.TypeValidation{}
	diags := rule.Check(input)
	if len(diags) != 0 {
		t.Errorf("expected no diagnostics, got %v", diags)
	}
}

func TestTypeValidation_InvalidEnum(t *testing.T) {
	input := makeInputWithGlobals(map[string]string{"SchedulerType": "sched/invalid"})
	rule := rules.TypeValidation{}
	diags := rule.Check(input)
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}
	if diags[0].Rule != "type-validation" {
		t.Errorf("expected rule=type-validation, got %q", diags[0].Rule)
	}
}

func TestTypeValidation_ValidTimeINFINITE(t *testing.T) {
	input := makeInputWithGlobals(map[string]string{"MaxTime": "INFINITE"})
	rule := rules.TypeValidation{}
	diags := rule.Check(input)
	if len(diags) != 0 {
		t.Errorf("expected no diagnostics, got %v", diags)
	}
}

func TestTypeValidation_ValidTimeHHMMSS(t *testing.T) {
	input := makeInputWithGlobals(map[string]string{"MaxTime": "24:00:00"})
	rule := rules.TypeValidation{}
	diags := rule.Check(input)
	if len(diags) != 0 {
		t.Errorf("expected no diagnostics, got %v", diags)
	}
}

func TestTypeValidation_InvalidTime(t *testing.T) {
	input := makeInputWithGlobals(map[string]string{"MaxTime": "notatime"})
	rule := rules.TypeValidation{}
	diags := rule.Check(input)
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}
}
```

- [ ] **Step 2: Run to confirm failure**

```bash
go test ./internal/rules/...
```

Expected: `undefined: rules.TypeValidation`

- [ ] **Step 3: Write implementation**

Create `internal/rules/types.go`:

```go
package rules

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/CommanderTso/slurm-linter/internal/diagnostic"
)

// TypeValidation checks that global parameters have values of the correct type.
type TypeValidation struct{}

// integerParams lists global params that must be non-negative integers.
var integerParams = []string{
	"KillWait", "MaxJobCount", "MinJobAge", "InactiveLimit",
	"SlurmctldPort", "SlurmdPort", "FirstJobId", "WaitTime",
}

// enumParams maps param names to their valid values.
var enumParams = map[string][]string{
	"SchedulerType": {"sched/backfill", "sched/builtin", "sched/hold"},
	"AuthType":      {"auth/munge", "auth/none"},
	"SelectType":    {"select/linear", "select/cons_res", "select/cons_tres"},
	"SwitchType":    {"switch/none", "switch/nrt"},
}

// timeParams lists global params that must be in Slurm time format.
var timeParams = []string{"MaxTime", "DefaultTime", "MaxWallDurationPerJob"}

// slurmTimeRE matches: INFINITE | UNLIMITED | minutes | hh:mm | hh:mm:ss | d-hh:mm:ss
var slurmTimeRE = regexp.MustCompile(
	`^(INFINITE|UNLIMITED|\d+|\d+:\d{2}|\d+:\d{2}:\d{2}|\d+-\d+:\d{2}:\d{2})$`,
)

func (r TypeValidation) Check(input *LintInput) []diagnostic.Diagnostic {
	var diags []diagnostic.Diagnostic
	cfg := input.Slurm

	for _, key := range integerParams {
		val, ok := cfg.Globals[key]
		if !ok {
			continue
		}
		n, err := strconv.Atoi(val)
		if err != nil || n < 0 {
			diags = append(diags, diagnostic.Diagnostic{
				Severity: diagnostic.Error,
				File:     input.SlurmFile,
				Line:     cfg.GlobalLines[key],
				Message:  fmt.Sprintf("%s must be a non-negative integer, got %q", key, val),
				Rule:     "type-validation",
			})
		}
	}

	for key, valid := range enumParams {
		val, ok := cfg.Globals[key]
		if !ok {
			continue
		}
		if !contains(valid, val) {
			diags = append(diags, diagnostic.Diagnostic{
				Severity: diagnostic.Error,
				File:     input.SlurmFile,
				Line:     cfg.GlobalLines[key],
				Message:  fmt.Sprintf("%s has unknown value %q (valid: %s)", key, val, strings.Join(valid, ", ")),
				Rule:     "type-validation",
			})
		}
	}

	for _, key := range timeParams {
		val, ok := cfg.Globals[key]
		if !ok {
			continue
		}
		if !slurmTimeRE.MatchString(val) {
			diags = append(diags, diagnostic.Diagnostic{
				Severity: diagnostic.Error,
				File:     input.SlurmFile,
				Line:     cfg.GlobalLines[key],
				Message:  fmt.Sprintf("%s has invalid time format %q (expected INFINITE, minutes, HH:MM:SS, or D-HH:MM:SS)", key, val),
				Rule:     "type-validation",
			})
		}
	}

	return diags
}

func contains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}
```

- [ ] **Step 4: Run to confirm pass**

```bash
go test ./internal/rules/...
```

Expected: `ok  github.com/CommanderTso/slurm-linter/internal/rules`

- [ ] **Step 5: Commit**

```bash
git -C /Users/smacmillan/projects/claude/slurm-linter add internal/rules/types.go internal/rules/types_test.go
git -C /Users/smacmillan/projects/claude/slurm-linter commit -m "feat: add type/enum/time validation rule"
```

---

## Task 10: Topology Graph Validation Rule

Validates the structure of `topology.conf`:
- Each switch must define exactly one of `Nodes` or `Switches`
- All switch names referenced in `Switches=` must be defined
- No cycles in the switch graph (DFS cycle detection)

**Files:**
- Create: `internal/rules/topology.go`
- Create: `internal/rules/topology_test.go`

- [ ] **Step 1: Write the failing tests**

Create `internal/rules/topology_test.go`:

```go
package rules_test

import (
	"testing"

	"github.com/CommanderTso/slurm-linter/internal/model"
	"github.com/CommanderTso/slurm-linter/internal/rules"
)

func topoInput(switches []model.Switch) *rules.LintInput {
	return &rules.LintInput{
		SlurmFile:    "slurm.conf",
		TopologyFile: "topology.conf",
		Slurm: &model.SlurmConfig{
			Globals:     map[string]string{},
			GlobalLines: map[string]int{},
		},
		Topology: &model.TopologyConfig{Switches: switches},
	}
}

func TestTopologyRule_ValidTree(t *testing.T) {
	input := topoInput([]model.Switch{
		{Name: "s0", Nodes: "node[01-04]", Line: 1},
		{Name: "s1", Nodes: "node[05-08]", Line: 2},
		{Name: "s2", Switches: "s0,s1", Line: 3},
	})
	rule := rules.TopologyRule{}
	diags := rule.Check(input)
	if len(diags) != 0 {
		t.Errorf("expected no diagnostics, got %v", diags)
	}
}

func TestTopologyRule_BothNodesAndSwitches(t *testing.T) {
	input := topoInput([]model.Switch{
		{Name: "s0", Nodes: "node01", Switches: "s1", Line: 1},
	})
	rule := rules.TopologyRule{}
	diags := rule.Check(input)
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}
}

func TestTopologyRule_NeitherNodesNorSwitches(t *testing.T) {
	input := topoInput([]model.Switch{
		{Name: "s0", Line: 1},
	})
	rule := rules.TopologyRule{}
	diags := rule.Check(input)
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}
}

func TestTopologyRule_UndefinedSwitchReference(t *testing.T) {
	input := topoInput([]model.Switch{
		{Name: "s0", Nodes: "node[01-04]", Line: 1},
		{Name: "s2", Switches: "s0,s999", Line: 2}, // s999 undefined
	})
	rule := rules.TopologyRule{}
	diags := rule.Check(input)
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}
}

func TestTopologyRule_Cycle(t *testing.T) {
	input := topoInput([]model.Switch{
		{Name: "s0", Switches: "s1", Line: 1},
		{Name: "s1", Switches: "s0", Line: 2}, // cycle: s0→s1→s0
	})
	rule := rules.TopologyRule{}
	diags := rule.Check(input)
	if len(diags) == 0 {
		t.Error("expected cycle diagnostic, got none")
	}
}
```

- [ ] **Step 2: Run to confirm failure**

```bash
go test ./internal/rules/...
```

Expected: `undefined: rules.TopologyRule`

- [ ] **Step 3: Write implementation**

Create `internal/rules/topology.go`:

```go
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

// findCycles returns the names of nodes involved in a cycle, or nil if none.
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

// resolveNodeList expands node range expressions from topology into individual node names.
// Exported for use by the cross-reference rule.
func ResolveNodeList(expr string) ([]string, error) {
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
```

- [ ] **Step 4: Run to confirm pass**

```bash
go test ./internal/rules/...
```

Expected: `ok  github.com/CommanderTso/slurm-linter/internal/rules`

- [ ] **Step 5: Commit**

```bash
git -C /Users/smacmillan/projects/claude/slurm-linter add internal/rules/topology.go internal/rules/topology_test.go
git -C /Users/smacmillan/projects/claude/slurm-linter commit -m "feat: add topology graph validation rule"
```

---

## Task 11: Cross-Reference Validation Rule

Validates that:
1. All node names in `topology.conf` `Nodes=` lists match a `NodeName=` stanza in `slurm.conf`
2. All node names in `PartitionName` `Nodes=` params match a `NodeName=` stanza in `slurm.conf`

**Files:**
- Create: `internal/rules/crossref.go`
- Create: `internal/rules/crossref_test.go`

- [ ] **Step 1: Write the failing tests**

Create `internal/rules/crossref_test.go`:

```go
package rules_test

import (
	"testing"

	"github.com/CommanderTso/slurm-linter/internal/model"
	"github.com/CommanderTso/slurm-linter/internal/rules"
)

func crossRefInput() *rules.LintInput {
	return &rules.LintInput{
		SlurmFile:    "slurm.conf",
		TopologyFile: "topology.conf",
		Slurm: &model.SlurmConfig{
			Globals:     map[string]string{},
			GlobalLines: map[string]int{},
			Nodes: []model.NodeDef{
				{Name: "node[01-04]", Params: map[string]string{}, Line: 5},
			},
			Partitions: []model.Partition{
				{Name: "compute", Params: map[string]string{"Nodes": "node[01-04]"}, Line: 8},
			},
		},
		Topology: &model.TopologyConfig{
			Switches: []model.Switch{
				{Name: "s0", Nodes: "node[01-04]", Line: 1},
			},
		},
	}
}

func TestCrossRef_AllValid(t *testing.T) {
	input := crossRefInput()
	rule := rules.CrossRefRule{}
	diags := rule.Check(input)
	if len(diags) != 0 {
		t.Errorf("expected no diagnostics, got %v", diags)
	}
}

func TestCrossRef_TopologyNodeNotInSlurm(t *testing.T) {
	input := crossRefInput()
	// Add a switch that references a node not in slurm.conf
	input.Topology.Switches = append(input.Topology.Switches, model.Switch{
		Name: "s1", Nodes: "gpu[01-02]", Line: 2,
	})
	rule := rules.CrossRefRule{}
	diags := rule.Check(input)
	if len(diags) == 0 {
		t.Error("expected diagnostics for undefined nodes, got none")
	}
}

func TestCrossRef_PartitionNodeNotInSlurm(t *testing.T) {
	input := crossRefInput()
	input.Slurm.Partitions = append(input.Slurm.Partitions, model.Partition{
		Name:   "gpu",
		Params: map[string]string{"Nodes": "gpu01"},
		Line:   9,
	})
	rule := rules.CrossRefRule{}
	diags := rule.Check(input)
	if len(diags) == 0 {
		t.Error("expected diagnostic for partition referencing undefined node")
	}
}

func TestCrossRef_NoTopology(t *testing.T) {
	input := crossRefInput()
	input.Topology = nil
	rule := rules.CrossRefRule{}
	diags := rule.Check(input)
	// only partition cross-ref check should run
	if len(diags) != 0 {
		t.Errorf("expected no diagnostics with no topology, got %v", diags)
	}
}
```

- [ ] **Step 2: Run to confirm failure**

```bash
go test ./internal/rules/...
```

Expected: `undefined: rules.CrossRefRule`

- [ ] **Step 3: Write implementation**

Create `internal/rules/crossref.go`:

```go
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
```

- [ ] **Step 4: Run to confirm pass**

```bash
go test ./internal/rules/...
```

Expected: `ok  github.com/CommanderTso/slurm-linter/internal/rules`

- [ ] **Step 5: Commit**

```bash
git -C /Users/smacmillan/projects/claude/slurm-linter add internal/rules/crossref.go internal/rules/crossref_test.go
git -C /Users/smacmillan/projects/claude/slurm-linter commit -m "feat: add cross-reference validation rule"
```

---

## Task 12: Validator Orchestrator

Wires all rules together, runs them against `LintInput`, and returns the combined diagnostic list.

**Files:**
- Create: `internal/validator/validator.go`
- Create: `internal/validator/validator_test.go`

- [ ] **Step 1: Write the failing tests**

Create `internal/validator/validator_test.go`:

```go
package validator_test

import (
	"testing"

	"github.com/CommanderTso/slurm-linter/internal/model"
	"github.com/CommanderTso/slurm-linter/internal/rules"
	"github.com/CommanderTso/slurm-linter/internal/validator"
)

func TestValidator_NoRules(t *testing.T) {
	input := &rules.LintInput{
		SlurmFile: "slurm.conf",
		Slurm:     &model.SlurmConfig{Globals: map[string]string{}, GlobalLines: map[string]int{}},
	}
	v := validator.New()
	diags := v.Run(input)
	if len(diags) != 0 {
		t.Errorf("expected 0 diagnostics with no rules, got %d", len(diags))
	}
}

func TestValidator_CollectsFromAllRules(t *testing.T) {
	input := &rules.LintInput{
		SlurmFile: "slurm.conf",
		Slurm: &model.SlurmConfig{
			Globals:     map[string]string{}, // missing ClusterName, SlurmctldHost
			GlobalLines: map[string]int{},
			// no nodes or partitions
		},
	}
	v := validator.New(rules.RequiredParams{})
	diags := v.Run(input)
	// Expect: missing ClusterName + missing SlurmctldHost + no nodes + no partitions = 4
	if len(diags) != 4 {
		t.Errorf("expected 4 diagnostics, got %d: %v", len(diags), diags)
	}
}
```

- [ ] **Step 2: Run to confirm failure**

```bash
go test ./internal/validator/...
```

Expected: `cannot find package`

- [ ] **Step 3: Write implementation**

Create `internal/validator/validator.go`:

```go
package validator

import (
	"github.com/CommanderTso/slurm-linter/internal/diagnostic"
	"github.com/CommanderTso/slurm-linter/internal/rules"
)

// Validator runs a set of rules against a LintInput and collects all diagnostics.
type Validator struct {
	rules []rules.Rule
}

// New creates a Validator with the provided rules.
func New(r ...rules.Rule) *Validator {
	return &Validator{rules: r}
}

// Run executes all rules and returns the combined diagnostics.
func (v *Validator) Run(input *rules.LintInput) []diagnostic.Diagnostic {
	var all []diagnostic.Diagnostic
	for _, rule := range v.rules {
		all = append(all, rule.Check(input)...)
	}
	return all
}

// DefaultRules returns the standard set of rules for production use.
func DefaultRules() []rules.Rule {
	return []rules.Rule{
		rules.RequiredParams{},
		rules.TypeValidation{},
		rules.TopologyRule{},
		rules.CrossRefRule{},
	}
}
```

- [ ] **Step 4: Run to confirm pass**

```bash
go test ./internal/validator/...
```

Expected: `ok  github.com/CommanderTso/slurm-linter/internal/validator`

- [ ] **Step 5: Commit**

```bash
git -C /Users/smacmillan/projects/claude/slurm-linter add internal/validator/
git -C /Users/smacmillan/projects/claude/slurm-linter commit -m "feat: add validator orchestrator"
```

---

## Task 13: Reporters

**Files:**
- Create: `internal/reporter/reporter.go`
- Create: `internal/reporter/text.go`
- Create: `internal/reporter/json.go`
- Create: `internal/reporter/reporter_test.go`

- [ ] **Step 1: Write the failing tests**

Create `internal/reporter/reporter_test.go`:

```go
package reporter_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/CommanderTso/slurm-linter/internal/diagnostic"
	"github.com/CommanderTso/slurm-linter/internal/reporter"
)

var sampleDiags = []diagnostic.Diagnostic{
	{Severity: diagnostic.Error, File: "slurm.conf", Line: 5, Message: "ClusterName is required", Rule: "required-params"},
	{Severity: diagnostic.Warning, File: "slurm.conf", Line: 10, Message: "deprecated parameter", Rule: "deprecated"},
}

func TestTextReporter_ContainsFileAndMessage(t *testing.T) {
	var buf bytes.Buffer
	r := reporter.Text{Writer: &buf}
	r.Report(sampleDiags)
	out := buf.String()
	if !strings.Contains(out, "slurm.conf") {
		t.Errorf("text output missing filename, got:\n%s", out)
	}
	if !strings.Contains(out, "ClusterName is required") {
		t.Errorf("text output missing message, got:\n%s", out)
	}
	if !strings.Contains(out, "error") {
		t.Errorf("text output missing severity, got:\n%s", out)
	}
}

func TestJSONReporter_ValidJSON(t *testing.T) {
	var buf bytes.Buffer
	r := reporter.JSON{Writer: &buf}
	r.Report(sampleDiags)

	var out []map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("JSON output is not valid: %v\noutput: %s", err, buf.String())
	}
	if len(out) != 2 {
		t.Errorf("expected 2 JSON entries, got %d", len(out))
	}
	if out[0]["severity"] != "error" {
		t.Errorf("expected severity=error, got %v", out[0]["severity"])
	}
	if out[0]["message"] != "ClusterName is required" {
		t.Errorf("expected message, got %v", out[0]["message"])
	}
}
```

- [ ] **Step 2: Run to confirm failure**

```bash
go test ./internal/reporter/...
```

Expected: `cannot find package`

- [ ] **Step 3: Write implementation**

Create `internal/reporter/reporter.go`:

```go
package reporter

import "github.com/CommanderTso/slurm-linter/internal/diagnostic"

// Reporter writes diagnostics to an output destination.
type Reporter interface {
	Report(diags []diagnostic.Diagnostic)
}
```

Create `internal/reporter/text.go`:

```go
package reporter

import (
	"fmt"
	"io"

	"github.com/CommanderTso/slurm-linter/internal/diagnostic"
)

// Text writes diagnostics as human-readable lines.
// Format: file:line: severity: message [rule]
type Text struct {
	Writer io.Writer
}

func (t Text) Report(diags []diagnostic.Diagnostic) {
	for _, d := range diags {
		loc := d.File
		if d.Line > 0 {
			loc = fmt.Sprintf("%s:%d", d.File, d.Line)
		}
		fmt.Fprintf(t.Writer, "%s: %s: %s [%s]\n", loc, d.Severity, d.Message, d.Rule)
	}
}
```

Create `internal/reporter/json.go`:

```go
package reporter

import (
	"encoding/json"
	"io"

	"github.com/CommanderTso/slurm-linter/internal/diagnostic"
)

// JSON writes diagnostics as a JSON array.
type JSON struct {
	Writer io.Writer
}

type jsonDiag struct {
	Severity string `json:"severity"`
	File     string `json:"file"`
	Line     int    `json:"line"`
	Message  string `json:"message"`
	Rule     string `json:"rule"`
}

func (j JSON) Report(diags []diagnostic.Diagnostic) {
	out := make([]jsonDiag, len(diags))
	for i, d := range diags {
		out[i] = jsonDiag{
			Severity: d.Severity.String(),
			File:     d.File,
			Line:     d.Line,
			Message:  d.Message,
			Rule:     d.Rule,
		}
	}
	enc := json.NewEncoder(j.Writer)
	enc.SetIndent("", "  ")
	_ = enc.Encode(out)
}
```

- [ ] **Step 4: Run to confirm pass**

```bash
go test ./internal/reporter/...
```

Expected: `ok  github.com/CommanderTso/slurm-linter/internal/reporter`

- [ ] **Step 5: Commit**

```bash
git -C /Users/smacmillan/projects/claude/slurm-linter add internal/reporter/
git -C /Users/smacmillan/projects/claude/slurm-linter commit -m "feat: add text and JSON reporters"
```

---

## Task 14: Testdata Fixtures

**Files:**
- Create: `testdata/valid/slurm.conf`
- Create: `testdata/valid/topology.conf`
- Create: `testdata/invalid/missing_required/slurm.conf`
- Create: `testdata/invalid/bad_topology/topology.conf`

- [ ] **Step 1: Create valid fixtures**

Create `testdata/valid/slurm.conf`:

```
# Valid slurm.conf for integration testing
ClusterName=testcluster
SlurmctldHost=controller
AuthType=auth/munge
SchedulerType=sched/backfill
SelectType=select/cons_res
KillWait=30
MaxJobCount=10000

NodeName=node[01-04] CPUs=4 RealMemory=8000 State=UNKNOWN
NodeName=gpu01 CPUs=32 RealMemory=256000 State=UNKNOWN

PartitionName=compute Nodes=node[01-04] Default=YES MaxTime=INFINITE State=UP
PartitionName=gpu Nodes=gpu01 MaxTime=24:00:00 State=UP
```

Create `testdata/valid/topology.conf`:

```
# Valid topology.conf for integration testing
SwitchName=leaf0 Nodes=node[01-04]
SwitchName=leaf1 Nodes=gpu01
SwitchName=spine0 Switches=leaf0,leaf1
```

Create `testdata/invalid/missing_required/slurm.conf`:

```
# Missing ClusterName and SlurmctldHost
AuthType=auth/munge
```

Create `testdata/invalid/bad_topology/topology.conf`:

```
# References undefined switch s99
SwitchName=leaf0 Nodes=node[01-04]
SwitchName=spine0 Switches=leaf0,s99
```

- [ ] **Step 2: Commit fixtures**

```bash
git -C /Users/smacmillan/projects/claude/slurm-linter add testdata/
git -C /Users/smacmillan/projects/claude/slurm-linter commit -m "test: add testdata fixtures for integration tests"
```

---

## Task 15: CLI with cobra

Wires the parsers, validator, and reporters into the final `check` command.

**Files:**
- Replace stub: `cmd/slurm-linter/main.go`

Exit codes:
- `0` — no issues found
- `1` — warnings only
- `2` — at least one error

- [ ] **Step 1: Write an end-to-end integration test first**

Create `internal/validator/integration_test.go`:

```go
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
		t.Errorf("unexpected diagnostic: %v", d)
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
```

- [ ] **Step 2: Run to confirm the integration test passes with current code**

```bash
go test ./internal/validator/...
```

Expected: `ok` — these tests exercise the full pipeline end to end.

- [ ] **Step 3: Write the full CLI**

Replace `cmd/slurm-linter/main.go`:

```go
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/CommanderTso/slurm-linter/internal/diagnostic"
	"github.com/CommanderTso/slurm-linter/internal/parser"
	"github.com/CommanderTso/slurm-linter/internal/reporter"
	"github.com/CommanderTso/slurm-linter/internal/rules"
	"github.com/CommanderTso/slurm-linter/internal/validator"
)

func main() {
	var confFile string
	var topologyFile string
	var format string

	checkCmd := &cobra.Command{
		Use:   "check",
		Short: "Lint slurm.conf and optionally topology.conf",
		RunE: func(cmd *cobra.Command, args []string) error {
			if confFile == "" {
				return fmt.Errorf("--conf is required")
			}

			input, err := buildLintInput(confFile, topologyFile)
			if err != nil {
				return err
			}

			v := validator.New(validator.DefaultRules()...)
			diags := v.Run(input)

			switch format {
			case "json":
				reporter.JSON{Writer: os.Stdout}.Report(diags)
			default:
				reporter.Text{Writer: os.Stdout}.Report(diags)
			}

			os.Exit(exitCode(diags))
			return nil
		},
	}

	checkCmd.Flags().StringVar(&confFile, "conf", "", "path to slurm.conf (required)")
	checkCmd.Flags().StringVar(&topologyFile, "topology", "", "path to topology.conf (optional)")
	checkCmd.Flags().StringVar(&format, "format", "text", "output format: text or json")

	rootCmd := &cobra.Command{
		Use:   "slurm-linter",
		Short: "Validate Slurm configuration files",
	}
	rootCmd.AddCommand(checkCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func buildLintInput(confFile, topologyFile string) (*rules.LintInput, error) {
	cf, err := os.Open(confFile)
	if err != nil {
		return nil, fmt.Errorf("cannot open %s: %w", confFile, err)
	}
	defer cf.Close()

	slurm, err := parser.ParseSlurmConf(confFile, cf)
	if err != nil {
		return nil, err
	}

	input := &rules.LintInput{
		SlurmFile: confFile,
		Slurm:     slurm,
	}

	if topologyFile != "" {
		tf, err := os.Open(topologyFile)
		if err != nil {
			return nil, fmt.Errorf("cannot open %s: %w", topologyFile, err)
		}
		defer tf.Close()

		topo, err := parser.ParseTopologyConf(topologyFile, tf)
		if err != nil {
			return nil, err
		}
		input.TopologyFile = topologyFile
		input.Topology = topo
	}

	return input, nil
}

// exitCode returns 0 (clean), 1 (warnings only), or 2 (errors present).
func exitCode(diags []diagnostic.Diagnostic) int {
	code := 0
	for _, d := range diags {
		if d.Severity == diagnostic.Error {
			return 2
		}
		if d.Severity == diagnostic.Warning {
			code = 1
		}
	}
	return code
}
```

- [ ] **Step 4: Build and smoke test**

```bash
make build
```

```bash
./bin/slurm-linter check --conf testdata/valid/slurm.conf --topology testdata/valid/topology.conf
```

Expected: no output, exit code 0.

```bash
./bin/slurm-linter check --conf testdata/invalid/missing_required/slurm.conf
```

Expected: lines like `slurm.conf: error: required parameter "ClusterName" is missing [required-params]`

- [ ] **Step 5: Run full test suite**

```bash
make test
```

Expected: all packages pass.

- [ ] **Step 6: Commit**

```bash
git -C /Users/smacmillan/projects/claude/slurm-linter add cmd/ internal/validator/integration_test.go
git -C /Users/smacmillan/projects/claude/slurm-linter commit -m "feat: wire CLI with cobra, complete MVP"
```

---

## Self-Review Checklist

- [x] **slurm.conf parsing** — globals, NodeName stanzas, PartitionName stanzas, line continuation, comments ✓
- [x] **topology.conf parsing** — SwitchName with Nodes or Switches ✓
- [x] **Node range expansion** — `node[01-04,06]` → individual names ✓
- [x] **Required params rule** — ClusterName, SlurmctldHost, at least one NodeName + Partition ✓
- [x] **Type validation** — integers, enums (SchedulerType, AuthType, SelectType), time formats ✓
- [x] **Topology graph** — both/neither Nodes+Switches, undefined switch refs, cycle detection ✓
- [x] **Cross-reference** — topology nodes ↔ slurm.conf NodeName, partition Nodes ↔ NodeName ✓
- [x] **Reporters** — text and JSON ✓
- [x] **CLI** — `--conf`, `--topology`, `--format`, exit codes 0/1/2 ✓
- [x] **Integration tests** — valid and invalid fixture-based tests ✓
- [x] **TDD** — every rule/parser/reporter has failing test written before implementation ✓
