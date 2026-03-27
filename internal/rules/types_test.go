package rules_test

import (
	"testing"

	"github.com/CommanderTso/slurm-linter/internal/diagnostic"
	"github.com/CommanderTso/slurm-linter/internal/model"
	"github.com/CommanderTso/slurm-linter/internal/rules"
)

func makeInputWithGlobals(globals map[string]string) *rules.LintInput {
	lines := make(map[string]int)
	for k := range globals {
		lines[k] = 1
	}
	return &rules.LintInput{
		SlurmFile: "slurm.conf",
		Slurm: &model.SlurmConfig{
			Globals:     globals,
			GlobalLines: lines,
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

// TestTypeValidation_PortRange verifies that SlurmctldPort and SlurmdPort
// accept a port range (N-M) in addition to a single integer.
func TestTypeValidation_PortRange(t *testing.T) {
	cases := []struct {
		key string
		val string
	}{
		{"SlurmctldPort", "6820"},
		{"SlurmctldPort", "6820-7824"},
		{"SlurmdPort", "6818"},
		{"SlurmdPort", "6818-6820"},
	}
	rule := rules.TypeValidation{}
	for _, c := range cases {
		input := makeInputWithGlobals(map[string]string{c.key: c.val})
		diags := rule.Check(input)
		if len(diags) != 0 {
			t.Errorf("%s=%q: expected no diagnostics, got %v", c.key, c.val, diags)
		}
	}
}

func TestTypeValidation_PortRange_Invalid(t *testing.T) {
	cases := []string{"notaport", "-5", "6820-", "-6820"}
	rule := rules.TypeValidation{}
	for _, val := range cases {
		input := makeInputWithGlobals(map[string]string{"SlurmctldPort": val})
		diags := rule.Check(input)
		if len(diags) != 1 {
			t.Errorf("SlurmctldPort=%q: expected 1 diagnostic, got %d", val, len(diags))
		}
	}
}

// TestTypeValidation_AuthType verifies that auth/munge and auth/slurm are valid,
// and that the removed auth/none is rejected.
func TestTypeValidation_AuthType(t *testing.T) {
	rule := rules.TypeValidation{}

	for _, val := range []string{"auth/munge", "auth/slurm"} {
		input := makeInputWithGlobals(map[string]string{"AuthType": val})
		if diags := rule.Check(input); len(diags) != 0 {
			t.Errorf("AuthType=%q: expected no diagnostics, got %v", val, diags)
		}
	}

	// auth/none was removed from Slurm and must be flagged
	input := makeInputWithGlobals(map[string]string{"AuthType": "auth/none"})
	if diags := rule.Check(input); len(diags) != 1 {
		t.Errorf("AuthType=auth/none: expected 1 diagnostic, got %d", len(diags))
	}
}

// TestTypeValidation_SchedulerType verifies sched/backfill and sched/builtin are valid,
// and that the removed sched/hold is rejected.
func TestTypeValidation_SchedulerType(t *testing.T) {
	rule := rules.TypeValidation{}

	for _, val := range []string{"sched/backfill", "sched/builtin"} {
		input := makeInputWithGlobals(map[string]string{"SchedulerType": val})
		if diags := rule.Check(input); len(diags) != 0 {
			t.Errorf("SchedulerType=%q: expected no diagnostics, got %v", val, diags)
		}
	}

	// sched/hold is not listed in current Slurm man page
	input := makeInputWithGlobals(map[string]string{"SchedulerType": "sched/hold"})
	if diags := rule.Check(input); len(diags) != 1 {
		t.Errorf("SchedulerType=sched/hold: expected 1 diagnostic, got %d", len(diags))
	}
}

// TestTypeValidation_SelectType verifies select/cons_tres and select/linear are valid,
// and that the removed legacy select/cons_res is rejected.
func TestTypeValidation_SelectType(t *testing.T) {
	rule := rules.TypeValidation{}

	for _, val := range []string{"select/cons_tres", "select/linear"} {
		input := makeInputWithGlobals(map[string]string{"SelectType": val})
		if diags := rule.Check(input); len(diags) != 0 {
			t.Errorf("SelectType=%q: expected no diagnostics, got %v", val, diags)
		}
	}

	// select/cons_res is explicitly called legacy and removed in current Slurm
	input := makeInputWithGlobals(map[string]string{"SelectType": "select/cons_res"})
	if diags := rule.Check(input); len(diags) != 1 {
		t.Errorf("SelectType=select/cons_res: expected 1 diagnostic, got %d", len(diags))
	}
}

// TestTypeValidation_SwitchType verifies switch/hpe_slingshot and switch/nvidia_imex
// are valid, and that the removed switch/none and switch/nrt are rejected.
func TestTypeValidation_SwitchType(t *testing.T) {
	rule := rules.TypeValidation{}

	for _, val := range []string{"switch/hpe_slingshot", "switch/nvidia_imex"} {
		input := makeInputWithGlobals(map[string]string{"SwitchType": val})
		if diags := rule.Check(input); len(diags) != 0 {
			t.Errorf("SwitchType=%q: expected no diagnostics, got %v", val, diags)
		}
	}

	for _, val := range []string{"switch/none", "switch/nrt"} {
		input := makeInputWithGlobals(map[string]string{"SwitchType": val})
		if diags := rule.Check(input); len(diags) != 1 {
			t.Errorf("SwitchType=%q: expected 1 diagnostic, got %d", val, len(diags))
		}
	}
}

// TestTypeValidation_TimeFormat_DaysHours verifies the days-hours and
// days-hours:minutes time formats are accepted (previously missing from regex).
func TestTypeValidation_TimeFormat_DaysHours(t *testing.T) {
	rule := rules.TypeValidation{}

	valid := []string{
		"INFINITE", "UNLIMITED",
		"60",          // minutes
		"1:30",        // minutes:seconds
		"24:00:00",    // hours:minutes:seconds
		"7-0",         // days-hours
		"7-00",        // days-hours (zero-padded)
		"7-12:30",     // days-hours:minutes
		"7-12:30:00",  // days-hours:minutes:seconds
	}
	for _, val := range valid {
		input := makeInputWithGlobals(map[string]string{"MaxTime": val})
		if diags := rule.Check(input); len(diags) != 0 {
			t.Errorf("MaxTime=%q: expected no diagnostics, got %v", val, diags)
		}
	}
}
