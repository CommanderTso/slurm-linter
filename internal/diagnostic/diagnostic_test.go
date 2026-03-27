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
