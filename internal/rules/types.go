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
	"FirstJobId", "WaitTime",
}

// portParams lists global params that accept a single port or a range (N-M).
var portParams = []string{"SlurmctldPort", "SlurmdPort"}

// portRE matches a single non-negative integer or a N-M port range.
var portRE = regexp.MustCompile(`^\d+(-\d+)?$`)

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

	for _, key := range portParams {
		val, ok := cfg.Globals[key]
		if !ok {
			continue
		}
		if !portRE.MatchString(val) {
			diags = append(diags, diagnostic.Diagnostic{
				Severity: diagnostic.Error,
				File:     input.SlurmFile,
				Line:     cfg.GlobalLines[key],
				Message:  fmt.Sprintf("%s must be a port number or range (e.g. 6820 or 6820-6830), got %q", key, val),
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
