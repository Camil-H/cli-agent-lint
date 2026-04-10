package report

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/cli-agent-lint/cli-agent-lint/checks"
)

func buildJSONReport() *Report {
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
		},
	}

	return NewReport(results, "/usr/bin/example", "2.0.0", 200*time.Millisecond)
}

func TestJSONFormat_ValidJSON(t *testing.T) {
	r := buildJSONReport()

	var buf bytes.Buffer
	f := &JSONFormatter{}
	if err := f.Format(&buf, r); err != nil {
		t.Fatalf("Format error: %v", err)
	}

	if !json.Valid(buf.Bytes()) {
		t.Fatal("output is not valid JSON")
	}
}

func TestJSONFormat_TopLevelFields(t *testing.T) {
	r := buildJSONReport()

	var buf bytes.Buffer
	f := &JSONFormatter{}
	if err := f.Format(&buf, r); err != nil {
		t.Fatalf("Format error: %v", err)
	}

	var parsed map[string]json.RawMessage
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	requiredFields := []string{"version", "target", "timestamp", "score", "summary", "checks", "categories"}
	for _, field := range requiredFields {
		if _, ok := parsed[field]; !ok {
			t.Errorf("missing required top-level field %q", field)
		}
	}
}

func TestJSONFormat_ScoreFields(t *testing.T) {
	r := buildJSONReport()

	var buf bytes.Buffer
	f := &JSONFormatter{}
	if err := f.Format(&buf, r); err != nil {
		t.Fatalf("Format error: %v", err)
	}

	var parsed struct {
		Score struct {
			Earned     int     `json:"earned"`
			Total      int     `json:"total"`
			Percentage float64 `json:"percentage"`
			Grade      string  `json:"grade"`
		} `json:"score"`
	}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if parsed.Score.Earned != r.TotalEarned {
		t.Errorf("score.earned: got %d, want %d", parsed.Score.Earned, r.TotalEarned)
	}
	if parsed.Score.Total != r.TotalPossible {
		t.Errorf("score.total: got %d, want %d", parsed.Score.Total, r.TotalPossible)
	}
	if parsed.Score.Percentage != r.Percent {
		t.Errorf("score.percentage: got %.2f, want %.2f", parsed.Score.Percentage, r.Percent)
	}
	if parsed.Score.Grade != string(r.Grade) {
		t.Errorf("score.grade: got %q, want %q", parsed.Score.Grade, string(r.Grade))
	}
}

func TestJSONFormat_ChecksArray(t *testing.T) {
	r := buildJSONReport()

	var buf bytes.Buffer
	f := &JSONFormatter{}
	if err := f.Format(&buf, r); err != nil {
		t.Fatalf("Format error: %v", err)
	}

	var parsed struct {
		Checks []struct {
			ID       string `json:"id"`
			Status   string `json:"status"`
			Severity string `json:"severity"`
		} `json:"checks"`
	}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if len(parsed.Checks) != len(r.Results) {
		t.Errorf("checks array length: got %d, want %d", len(parsed.Checks), len(r.Results))
	}

	// Build a map of check IDs for lookup.
	checkMap := make(map[string]struct {
		ID       string `json:"id"`
		Status   string `json:"status"`
		Severity string `json:"severity"`
	})
	for _, c := range parsed.Checks {
		checkMap[c.ID] = c
	}

	// Verify all results are present.
	for _, res := range r.Results {
		if _, ok := checkMap[res.CheckID]; !ok {
			t.Errorf("check %q missing from JSON checks array", res.CheckID)
		}
	}
}

func TestJSONFormat_WarnSeverityStatus(t *testing.T) {
	r := buildJSONReport()

	var buf bytes.Buffer
	f := &JSONFormatter{}
	if err := f.Format(&buf, r); err != nil {
		t.Fatalf("Format error: %v", err)
	}

	var parsed struct {
		Checks []struct {
			ID       string `json:"id"`
			Status   string `json:"status"`
			Severity string `json:"severity"`
		} `json:"checks"`
	}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	// SA-2 has Severity=Warn and Status=StatusFail. Its JSON status should be "warn", not "fail".
	for _, c := range parsed.Checks {
		if c.ID == "SA-2" {
			if c.Status != "warn" {
				t.Errorf("SA-2 (warn-severity, failed) should have status %q, got %q", "warn", c.Status)
			}
			return
		}
	}
	t.Error("SA-2 not found in checks array")
}

func TestJSONFormat_FailSeverityStatusIsFail(t *testing.T) {
	r := buildJSONReport()

	var buf bytes.Buffer
	f := &JSONFormatter{}
	if err := f.Format(&buf, r); err != nil {
		t.Fatalf("Format error: %v", err)
	}

	var parsed struct {
		Checks []struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"checks"`
	}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	// FS-2 has Severity=Fail and Status=StatusFail. Its JSON status should be "fail".
	for _, c := range parsed.Checks {
		if c.ID == "FS-2" {
			if c.Status != "fail" {
				t.Errorf("FS-2 (fail-severity, failed) should have status %q, got %q", "fail", c.Status)
			}
			return
		}
	}
	t.Error("FS-2 not found in checks array")
}

func TestJSONFormat_PassStatus(t *testing.T) {
	r := buildJSONReport()

	var buf bytes.Buffer
	f := &JSONFormatter{}
	if err := f.Format(&buf, r); err != nil {
		t.Fatalf("Format error: %v", err)
	}

	var parsed struct {
		Checks []struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"checks"`
	}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	for _, c := range parsed.Checks {
		if c.ID == "TE-1" {
			if c.Status != "pass" {
				t.Errorf("TE-1 (pass) should have status %q, got %q", "pass", c.Status)
			}
			return
		}
	}
	t.Error("TE-1 not found in checks array")
}

func TestJSONFormat_SummaryFields(t *testing.T) {
	r := buildJSONReport()

	var buf bytes.Buffer
	f := &JSONFormatter{}
	if err := f.Format(&buf, r); err != nil {
		t.Fatalf("Format error: %v", err)
	}

	var parsed struct {
		Summary struct {
			Total int `json:"total"`
			Pass  int `json:"pass"`
			Warn  int `json:"warn"`
			Fail  int `json:"fail"`
			Skip  int `json:"skip"`
			Error int `json:"error"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	summary := r.GetSummary()
	if parsed.Summary.Total != summary.Total {
		t.Errorf("summary.total: got %d, want %d", parsed.Summary.Total, summary.Total)
	}
	if parsed.Summary.Pass != summary.Pass {
		t.Errorf("summary.pass: got %d, want %d", parsed.Summary.Pass, summary.Pass)
	}
	if parsed.Summary.Warn != summary.Warn {
		t.Errorf("summary.warn: got %d, want %d", parsed.Summary.Warn, summary.Warn)
	}
	if parsed.Summary.Fail != summary.Fail {
		t.Errorf("summary.fail: got %d, want %d", parsed.Summary.Fail, summary.Fail)
	}
	if parsed.Summary.Skip != summary.Skip {
		t.Errorf("summary.skip: got %d, want %d", parsed.Summary.Skip, summary.Skip)
	}
	if parsed.Summary.Error != summary.Error {
		t.Errorf("summary.error: got %d, want %d", parsed.Summary.Error, summary.Error)
	}
}

func TestJSONFormat_CategoriesArray(t *testing.T) {
	r := buildJSONReport()

	var buf bytes.Buffer
	f := &JSONFormatter{}
	if err := f.Format(&buf, r); err != nil {
		t.Fatalf("Format error: %v", err)
	}

	var parsed struct {
		Categories []struct {
			Name       string  `json:"name"`
			Earned     int     `json:"earned"`
			Possible   int     `json:"possible"`
			Percentage float64 `json:"percentage"`
		} `json:"categories"`
	}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if len(parsed.Categories) != len(r.Categories) {
		t.Errorf("categories length: got %d, want %d", len(parsed.Categories), len(r.Categories))
	}

	// Verify each category matches the report data.
	catMap := make(map[string]struct {
		Name       string  `json:"name"`
		Earned     int     `json:"earned"`
		Possible   int     `json:"possible"`
		Percentage float64 `json:"percentage"`
	})
	for _, c := range parsed.Categories {
		catMap[c.Name] = c
	}

	for _, cs := range r.Categories {
		jc, ok := catMap[string(cs.Category)]
		if !ok {
			t.Errorf("category %q missing from JSON categories array", cs.Category)
			continue
		}
		if jc.Earned != cs.Earned {
			t.Errorf("category %q earned: got %d, want %d", cs.Category, jc.Earned, cs.Earned)
		}
		if jc.Possible != cs.Possible {
			t.Errorf("category %q possible: got %d, want %d", cs.Category, jc.Possible, cs.Possible)
		}
	}
}

func TestJSONFormat_VersionField(t *testing.T) {
	r := buildJSONReport()

	var buf bytes.Buffer
	f := &JSONFormatter{}
	if err := f.Format(&buf, r); err != nil {
		t.Fatalf("Format error: %v", err)
	}

	var parsed struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if parsed.Version != "1.0.0" {
		t.Errorf("version: got %q, want %q", parsed.Version, "1.0.0")
	}
}

func TestJSONFormat_TargetField(t *testing.T) {
	r := buildJSONReport()

	var buf bytes.Buffer
	f := &JSONFormatter{}
	if err := f.Format(&buf, r); err != nil {
		t.Fatalf("Format error: %v", err)
	}

	var parsed struct {
		Target string `json:"target"`
	}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if parsed.Target != "/usr/bin/example" {
		t.Errorf("target: got %q, want %q", parsed.Target, "/usr/bin/example")
	}
}

func TestJSONFormat_CheckDetails(t *testing.T) {
	r := buildJSONReport()

	var buf bytes.Buffer
	f := &JSONFormatter{}
	if err := f.Format(&buf, r); err != nil {
		t.Fatalf("Format error: %v", err)
	}

	var parsed struct {
		Checks []struct {
			ID             string `json:"id"`
			Name           string `json:"name"`
			Category       string `json:"category"`
			Severity       string `json:"severity"`
			Points         int    `json:"points"`
			MaxPoints      int    `json:"max_points"`
			Recommendation string `json:"recommendation"`
			Detail         string `json:"detail"`
		} `json:"checks"`
	}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	for _, c := range parsed.Checks {
		if c.ID == "FS-2" {
			if c.Name != "No interactive prompts" {
				t.Errorf("FS-2 name: got %q", c.Name)
			}
			if c.Category != string(checks.CatFlowSafety) {
				t.Errorf("FS-2 category: got %q", c.Category)
			}
			if c.Severity != "fail" {
				t.Errorf("FS-2 severity: got %q", c.Severity)
			}
			if c.Points != 0 {
				t.Errorf("FS-2 points: got %d, want 0 (failed check)", c.Points)
			}
			if c.MaxPoints != 3 {
				t.Errorf("FS-2 max_points: got %d, want 3", c.MaxPoints)
			}
			if c.Detail != "Detected interactive prompt" {
				t.Errorf("FS-2 detail: got %q", c.Detail)
			}
			if c.Recommendation != "Remove prompts in non-TTY mode" {
				t.Errorf("FS-2 recommendation: got %q", c.Recommendation)
			}
			return
		}
	}
	t.Error("FS-2 not found in checks array")
}

func TestJSONFormat_SkipStatus(t *testing.T) {
	r := buildJSONReport()

	var buf bytes.Buffer
	f := &JSONFormatter{}
	if err := f.Format(&buf, r); err != nil {
		t.Fatalf("Format error: %v", err)
	}

	var parsed struct {
		Checks []struct {
			ID        string `json:"id"`
			Status    string `json:"status"`
			Points    int    `json:"points"`
			MaxPoints int    `json:"max_points"`
		} `json:"checks"`
	}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	for _, c := range parsed.Checks {
		if c.ID == "SD-3" {
			if c.Status != "skip" {
				t.Errorf("SD-3 status: got %q, want %q", c.Status, "skip")
			}
			if c.Points != 0 {
				t.Errorf("SD-3 points: got %d, want 0", c.Points)
			}
			if c.MaxPoints != 0 {
				t.Errorf("SD-3 max_points: got %d, want 0", c.MaxPoints)
			}
			return
		}
	}
	t.Error("SD-3 not found in checks array")
}
