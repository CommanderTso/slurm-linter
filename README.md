# slurm-linter

[![CI](https://github.com/CommanderTso/slurm-linter/actions/workflows/ci.yml/badge.svg)](https://github.com/CommanderTso/slurm-linter/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/CommanderTso/slurm-linter/branch/main/graph/badge.svg)](https://codecov.io/gh/CommanderTso/slurm-linter)

A CLI linter for [Slurm](https://slurm.schedmd.com/) configuration files. Validates `slurm.conf` and `topology.conf`, catching common errors before they cause cluster problems.

## Features

- **Required parameter checks** — flags missing `ClusterName`, `SlurmctldHost`, node definitions, and partition definitions
- **Type validation** — verifies integers, port numbers/ranges (`SlurmctldPort`, `SlurmdPort`), enums (`SchedulerType`, `AuthType`, `SelectType`), and time formats (`INFINITE`, `HH:MM:SS`, `D-HH:MM:SS`)
- **Topology graph validation** — detects undefined switch references and cycles in the switch hierarchy
- **Cross-file validation** — ensures nodes referenced in `topology.conf` and partition stanzas are defined in `slurm.conf`
- **Text and JSON output** — human-readable by default, JSON for scripting and CI integration

## Installation

### Pre-built binaries

Download the latest release for your platform from the [Releases page](https://github.com/CommanderTso/slurm-linter/releases) and place the binary somewhere on your `PATH`.

### Build from source

Requires Go 1.26+.

```bash
git clone git@github.com:CommanderTso/slurm-linter.git
cd slurm-linter
make build
# binary at bin/slurm-linter
```

## Usage

```bash
# Lint slurm.conf only
slurm-linter check --conf /etc/slurm/slurm.conf

# Lint both config files (enables cross-file validation)
slurm-linter check --conf /etc/slurm/slurm.conf --topology /etc/slurm/topology.conf

# JSON output (for scripts or CI)
slurm-linter check --conf slurm.conf --topology topology.conf --format json
```

### Exit codes

| Code | Meaning |
|------|---------|
| `0` | No issues found |
| `1` | Warnings only |
| `2` | At least one error |

### Example output

```
slurm.conf: error: required parameter "ClusterName" is missing [required-params]
slurm.conf:14: error: SchedulerType has unknown value "sched/bogus" (valid: sched/backfill, sched/builtin, sched/hold) [type-validation]
topology.conf:6: error: switch "spine0" references undefined switch "s99" [topology]
```

## What gets validated

### slurm.conf

| Rule | ID |
|------|----|
| `ClusterName` and `SlurmctldHost` are present | `required-params` |
| At least one `NodeName` stanza exists | `required-params` |
| At least one `PartitionName` stanza exists | `required-params` |
| Integer parameters are valid non-negative integers | `type-validation` |
| Port parameters (`SlurmctldPort`, `SlurmdPort`) are a port number or range (`N-M`) | `type-validation` |
| Enum parameters (`SchedulerType`, `AuthType`, `SelectType`, `SwitchType`) have valid values | `type-validation` |
| Time parameters (`MaxTime`, `DefaultTime`) use valid Slurm time format | `type-validation` |
| Partition `Nodes=` references are defined in `NodeName` stanzas | `cross-ref` |

### topology.conf

| Rule | ID |
|------|----|
| Each switch defines exactly one of `Nodes` or `Switches` | `topology` |
| Switch names referenced in `Switches=` are defined | `topology` |
| Switch graph contains no cycles | `topology` |
| Nodes referenced in `Nodes=` are defined in slurm.conf | `cross-ref` |

## Development

```bash
make test      # run all tests
make test-v    # verbose
make lint      # go vet
```

See [CLAUDE.md](CLAUDE.md) for architecture details and how to add new lint rules.
