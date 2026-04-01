# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Purpose

A CLI linter for Slurm HPC workload manager configuration files. Validates `slurm.conf` and `topology.conf`, reporting errors and warnings about missing required parameters, type/enum mismatches, topology graph problems, and cross-file node reference inconsistencies.

## Module

`github.com/CommanderTso/slurm-linter`

## Common Commands

```bash
make build        # build → bin/slurm-linter
make test         # go test ./...
make test-v       # verbose test output
make lint         # go vet ./...
make coverage     # run tests with coverage → coverage.html

# Run the linter
./bin/slurm-linter check --conf slurm.conf --topology topology.conf
./bin/slurm-linter check --conf slurm.conf --format json

# Run a single test package
go test ./internal/rules/...

# Run a single test
go test ./internal/parser/... -run TestExpandNodeRange
```

## Architecture

Strict separation between parsing, validation, and reporting. Data flows in one direction: files → parsers → model → rules → diagnostics → reporter.

### Key packages

| Package | Role |
|---------|------|
| `internal/model` | Pure data types: `SlurmConfig`, `NodeDef`, `Partition`, `TopologyConfig`, `Switch` |
| `internal/parser` | Parses files into model types. `noderange.go` expands bracket expressions like `node[01-04,06]`. `slurmconf.go` handles globals, `NodeName=` and `PartitionName=` stanzas, and line continuation. `topology.go` parses `SwitchName=` stanzas. |
| `internal/rules` | One file per rule domain. Each rule implements `Rule.Check(*LintInput) []Diagnostic`. `LintInput` carries both parsed configs; `Topology` may be nil. |
| `internal/validator` | Orchestrates rules. `DefaultRules()` returns the standard rule set. `New(rules...)` for custom sets. |
| `internal/reporter` | `Text` and `JSON` reporters, both implement the `Reporter` interface. |
| `internal/diagnostic` | `Diagnostic` struct and `Severity` type (Info/Warning/Error). |
| `cmd/slurm-linter` | cobra CLI. `--conf` (required), `--topology` (optional), `--format text|json`. Exit codes: 0=clean, 1=warnings, 2=errors. |

### Adding a new lint rule

1. Create `internal/rules/myrule.go` with a struct implementing `Check(*LintInput) []diagnostic.Diagnostic`
2. Write `internal/rules/myrule_test.go` first (TDD)
3. Add the rule to `DefaultRules()` in `internal/validator/validator.go`

### slurm.conf parsing notes

- `NodeName=` and `PartitionName=` lines carry multiple `Key=Value` pairs on one line — they are parsed into `NodeDef` and `Partition` structs, not globals
- Line continuation with `\` is handled in `readLines()`
- `GlobalLines` map tracks line numbers for accurate diagnostic reporting

## Reference Documentation

Use context7 for Slurm documentation. Prefer these library IDs:
- `/websites/slurm_schedmd` — main SchedMD site, broadest coverage (172k snippets)
- `/schedmd/slurm` — GitHub source repo, useful for implementation details

## Test-Driven Development — MANDATORY

**TDD is non-negotiable on this project. No exceptions.**

Before writing any implementation code:

1. Use the `superpowers:test-driven-development` skill
2. Write the test(s) first
3. Run them and confirm they **fail** (red)
4. Write the minimum implementation to make them pass (green)
5. Refactor if needed, keeping tests green

Never write implementation code without a failing test already in place. Never skip the red phase — seeing the failure confirms the test is actually exercising the code.

**This applies to bug fixes too.** When fixing a bug, the first step is always a test that reproduces the bug and fails. Only then write the fix. No exceptions, even for small or obvious fixes.

## Committing Code Changes

Run `make build` before every commit that changes Go source files. This catches compile errors that tests alone may not surface and keeps `bin/slurm-linter` in sync with the committed source.

## Plans

Store plan files in `/Users/smacmillan/projects/claude/slurm-linter/plans/`.
