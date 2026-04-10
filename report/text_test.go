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
			CheckID:        "TE-1",
			CheckName:      "JSON stdout",
			Category:       checks.CatTokenEfficiency,
			Severity:       checks.Fail,
			Status:         checks.StatusPass,
			Recommendation: "Use JSON output",
		},
		{
			CheckID:        "FS-2",
			CheckName:      "No interactive prompts",
			Category:       checks.CatFlowSafety,
			Severity:       checks.Fail,
			Status:         checks.StatusFail,
			Detail:         "Detected interactive prompt",
			Recommendation: "Remove prompts in non-TTY mode",
		},
		{
			CheckID:        "SA-2",
			CheckName:      "Input validation",
			Category:       checks.CatAutomationSafety,
			Severity:       checks.Warn,
			Status:         checks.StatusFail,
			Detail:         "Missing validation",
			Recommendation: "Validate all inputs",
		},
		{
			CheckID:        "SD-3",
			CheckName:      "Schema available",
			Category:       checks.CatSelfDescribing,
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

	// Quiet mode should show fail-severity failures (FS-2 is Fail severity + StatusFail).
	if !strings.Contains(output, "FS-2") {
		t.Error("quiet mode should show fail-severity failure FS-2")
	}

	// Quiet mode should NOT show passing checks.
	if strings.Contains(output, "TE-1") {
		t.Error("quiet mode should not show passing check TE-1")
	}

	// Quiet mode should NOT show warn-severity failures (only fail-severity).
	if strings.Contains(output, "SA-2") {
		t.Error("quiet mode should not show warn-severity failure SA-2")
	}

	// Quiet mode should not show category headers.
	if strings.Contains(output, "Token Efficiency") {
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
	if !strings.Contains(output, "Token Efficiency") {
		t.Error("full mode should contain 'Token Efficiency' category header")
	}
	if !strings.Contains(output, "Flow Safety") {
		t.Error("full mode should contain 'Flow Safety' category header")
	}
	if !strings.Contains(output, "Automation Safety") {
		t.Error("full mode should contain 'Automation Safety' category header")
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
		t.Error("expected detail for failing check FS-2")
	}
	if !strings.Contains(output, "Remove prompts in non-TTY mode") {
		t.Error("expected recommendation for failing check FS-2")
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
			Category:       checks.CatTokenEfficiency,
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
		makeResult("X1", checks.CatTokenEfficiency, checks.Fail, checks.StatusPass),
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
