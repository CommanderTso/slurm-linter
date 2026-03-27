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
