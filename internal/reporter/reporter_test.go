package reporter_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/CommanderTso/slurm-linter/internal/diagnostic"
	"github.com/CommanderTso/slurm-linter/internal/reporter"
)

var sampleDiags = []diagnostic.Diagnostic{
	{Severity: diagnostic.Error, File: "slurm.conf", Line: 5, Message: "ClusterName is required", Rule: "required-params"},
	{Severity: diagnostic.Warning, File: "slurm.conf", Line: 10, Message: "deprecated parameter", Rule: "deprecated"},
}

func TestTextReporter_ContainsFileAndMessage(t *testing.T) {
	var buf bytes.Buffer
	r := reporter.Text{Writer: &buf}
	r.Report(sampleDiags)
	out := buf.String()
	if !strings.Contains(out, "slurm.conf") {
		t.Errorf("text output missing filename, got:\n%s", out)
	}
	if !strings.Contains(out, "ClusterName is required") {
		t.Errorf("text output missing message, got:\n%s", out)
	}
	if !strings.Contains(out, "error") {
		t.Errorf("text output missing severity, got:\n%s", out)
	}
}

func TestJSONReporter_ValidJSON(t *testing.T) {
	var buf bytes.Buffer
	r := reporter.JSON{Writer: &buf}
	r.Report(sampleDiags)

	var out []map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("JSON output is not valid: %v\noutput: %s", err, buf.String())
	}
	if len(out) != 2 {
		t.Errorf("expected 2 JSON entries, got %d", len(out))
	}
	if out[0]["severity"] != "error" {
		t.Errorf("expected severity=error, got %v", out[0]["severity"])
	}
	if out[0]["message"] != "ClusterName is required" {
		t.Errorf("expected message, got %v", out[0]["message"])
	}
}
