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
