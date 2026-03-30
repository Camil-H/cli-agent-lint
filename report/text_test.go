package report

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/cli-agent-lint/cli-agent-lint/checks"
)

func buildTextReport() *Report {
	results := []*checks.Result{
		{
			CheckID:        "SO-1",
			CheckName:      "JSON stdout",
			Category:       checks.CatStructuredOutput,
			Severity:       checks.Fail,
			Status:         checks.StatusPass,
			Recommendation: "Use JSON output",
		},
		{
			CheckID:        "TH-1",
			CheckName:      "No interactive prompts",
			Category:       checks.CatTerminalHygiene,
			Severity:       checks.Fail,
			Status:         checks.StatusFail,
			Detail:         "Detected interactive prompt",
			Recommendation: "Remove prompts in non-TTY mode",
		},
		{
			CheckID:        "IV-1",
			CheckName:      "Input validation",
			Category:       checks.CatInputValidation,
			Severity:       checks.Warn,
			Status:         checks.StatusFail,
			Detail:         "Missing validation",
			Recommendation: "Validate all inputs",
		},
		{
			CheckID:        "SD-1",
			CheckName:      "Schema available",
			Category:       checks.CatSchemaDiscovery,
			Severity:       checks.Info,
			Status:         checks.StatusSkip,
			Detail:         "Skipped",
		},
	}

	return NewReport(results, "/usr/bin/example", "2.0.0", 200*time.Millisecond)
}

func TestTextFormat_NoColor_NoANSI(t *testing.T) {
	r := buildTextReport()

	var buf bytes.Buffer
	f := &TextFormatter{NoColor: true, Quiet: false}
	if err := f.Format(&buf, r); err != nil {
		t.Fatalf("Format error: %v", err)
	}

	output := buf.String()
	if strings.Contains(output, "\x1b[") {
		t.Error("expected no ANSI escape codes with NoColor=true, but found some")
	}
}

func TestTextFormat_WithColor_HasANSI(t *testing.T) {
	r := buildTextReport()

	var buf bytes.Buffer
	f := &TextFormatter{NoColor: false, Quiet: false}
	if err := f.Format(&buf, r); err != nil {
		t.Fatalf("Format error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "\x1b[") {
		t.Error("expected ANSI escape codes with NoColor=false, but found none")
	}
}

func TestTextFormat_QuietMode(t *testing.T) {
	r := buildTextReport()

	var buf bytes.Buffer
	f := &TextFormatter{NoColor: true, Quiet: true}
	if err := f.Format(&buf, r); err != nil {
		t.Fatalf("Format error: %v", err)
	}

	output := buf.String()

	// Quiet mode should show the score line.
	if !strings.Contains(output, "Score:") {
		t.Error("quiet mode should contain 'Score:'")
	}

	// Quiet mode should show fail-severity failures (TH-1 is Fail severity + StatusFail).
	if !strings.Contains(output, "TH-1") {
		t.Error("quiet mode should show fail-severity failure TH-1")
	}

	// Quiet mode should NOT show passing checks.
	if strings.Contains(output, "SO-1") {
		t.Error("quiet mode should not show passing check SO-1")
	}

	// Quiet mode should NOT show warn-severity failures (only fail-severity).
	if strings.Contains(output, "IV-1") {
		t.Error("quiet mode should not show warn-severity failure IV-1")
	}

	// Quiet mode should not show category headers.
	if strings.Contains(output, "Structured Output") {
		t.Error("quiet mode should not contain category headers")
	}
}

func TestTextFormat_FullMode_CategoryHeaders(t *testing.T) {
	r := buildTextReport()

	var buf bytes.Buffer
	f := &TextFormatter{NoColor: true, Quiet: false}
	if err := f.Format(&buf, r); err != nil {
		t.Fatalf("Format error: %v", err)
	}

	output := buf.String()

	// Should contain category headers.
	if !strings.Contains(output, "Structured Output") {
		t.Error("full mode should contain 'Structured Output' category header")
	}
	if !strings.Contains(output, "Terminal Hygiene") {
		t.Error("full mode should contain 'Terminal Hygiene' category header")
	}
	if !strings.Contains(output, "Input Validation") {
		t.Error("full mode should contain 'Input Validation' category header")
	}
}

func TestTextFormat_StatusIndicators(t *testing.T) {
	r := buildTextReport()

	var buf bytes.Buffer
	f := &TextFormatter{NoColor: true, Quiet: false}
	if err := f.Format(&buf, r); err != nil {
		t.Fatalf("Format error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "[PASS]") {
		t.Error("expected [PASS] indicator in output")
	}
	if !strings.Contains(output, "[FAIL]") {
		t.Error("expected [FAIL] indicator in output")
	}
	if !strings.Contains(output, "[WARN]") {
		t.Error("expected [WARN] indicator in output")
	}
	if !strings.Contains(output, "[SKIP]") {
		t.Error("expected [SKIP] indicator in output")
	}
}

func TestTextFormat_TargetAndVersion(t *testing.T) {
	r := buildTextReport()

	var buf bytes.Buffer
	f := &TextFormatter{NoColor: true, Quiet: false}
	if err := f.Format(&buf, r); err != nil {
		t.Fatalf("Format error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "Target: /usr/bin/example") {
		t.Error("expected target path in output")
	}
	if !strings.Contains(output, "Version: 2.0.0") {
		t.Error("expected version in output")
	}
}

func TestTextFormat_FullMode_DetailsAndRecommendations(t *testing.T) {
	r := buildTextReport()

	var buf bytes.Buffer
	f := &TextFormatter{NoColor: true, Quiet: false}
	if err := f.Format(&buf, r); err != nil {
		t.Fatalf("Format error: %v", err)
	}

	output := buf.String()

	// Failing checks should show detail and recommendation.
	if !strings.Contains(output, "Detected interactive prompt") {
		t.Error("expected detail for failing check TH-1")
	}
	if !strings.Contains(output, "Remove prompts in non-TTY mode") {
		t.Error("expected recommendation for failing check TH-1")
	}
}

func TestTextFormat_GradeInOutput(t *testing.T) {
	r := buildTextReport()

	var buf bytes.Buffer
	f := &TextFormatter{NoColor: true, Quiet: false}
	if err := f.Format(&buf, r); err != nil {
		t.Fatalf("Format error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "Grade:") {
		t.Error("expected 'Grade:' in output")
	}
}

func TestStripANSI(t *testing.T) {
	input := "\x1b[31mred text\x1b[0m and \x1b[1;32mgreen bold\x1b[0m"
	got := stripANSI(input)
	want := "red text and green bold"
	if got != want {
		t.Errorf("stripANSI() = %q, want %q", got, want)
	}
}

func TestTextFormat_DetailStripsANSI(t *testing.T) {
	results := []*checks.Result{
		{
			CheckID:        "X1",
			CheckName:      "Test check",
			Category:       checks.CatStructuredOutput,
			Severity:       checks.Fail,
			Status:         checks.StatusFail,
			Detail:         "found \x1b[31mbad\x1b[0m thing",
			Recommendation: "fix it",
		},
	}
	r := NewReport(results, "/bin/tool", "", time.Second)

	var buf bytes.Buffer
	f := &TextFormatter{NoColor: true, Quiet: false}
	if err := f.Format(&buf, r); err != nil {
		t.Fatalf("Format error: %v", err)
	}

	output := buf.String()
	if strings.Contains(output, "\x1b[") {
		t.Error("expected no ANSI escape codes in rendered detail text")
	}
	if !strings.Contains(output, "found bad thing") {
		t.Error("expected stripped detail text in output")
	}
}

func TestTextFormat_NoVersion(t *testing.T) {
	results := []*checks.Result{
		makeResult("X1", checks.CatStructuredOutput, checks.Fail, checks.StatusPass),
	}
	r := NewReport(results, "/bin/tool", "", time.Second)

	var buf bytes.Buffer
	f := &TextFormatter{NoColor: true, Quiet: false}
	if err := f.Format(&buf, r); err != nil {
		t.Fatalf("Format error: %v", err)
	}

	output := buf.String()
	if strings.Contains(output, "Version:") {
		t.Error("should not show 'Version:' when TargetVersion is empty")
	}
}
