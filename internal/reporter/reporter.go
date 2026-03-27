package reporter

import "github.com/CommanderTso/slurm-linter/internal/diagnostic"

// Reporter writes diagnostics to an output destination.
type Reporter interface {
	Report(diags []diagnostic.Diagnostic)
}
