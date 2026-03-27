# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Purpose

A linter for Slurm HPC workload manager job scripts. Analyzes `#SBATCH` directives and job script contents to catch errors, warn about misconfigurations, and enforce site-specific policies.

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

## Plans

Store plan files in `/Users/smacmillan/projects/claude/slurm-linter/plans/`.
