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
