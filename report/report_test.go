package report

import (
	"testing"
	"time"

	"github.com/Camil-H/cli-agent-lint/checks"
)

func makeResult(id string, cat checks.Category, sev checks.Severity, status checks.Status) *checks.Result {
	return &checks.Result{
		CheckID:   id,
		CheckName: "check-" + id,
		Category:  cat,
		Severity:  sev,
		Status:    status,
	}
}

func TestNewReport_AllPasses(t *testing.T) {
	results := []*checks.Result{
		makeResult("A1", checks.CatTokenEfficiency, checks.Fail, checks.StatusPass),
		makeResult("A2", checks.CatTokenEfficiency, checks.Warn, checks.StatusPass),
		makeResult("A3", checks.CatFlowSafety, checks.Info, checks.StatusPass),
	}

	r := NewReport(results, "/bin/test", "1.0.0", 100*time.Millisecond)

	if r.Percent != 100 {
		t.Errorf("expected 100%%, got %.2f%%", r.Percent)
	}
	if r.Grade != GradeA {
		t.Errorf("expected grade A, got %s", r.Grade)
	}
	// Total earned should equal total possible.
	if r.TotalEarned != r.TotalPossible {
		t.Errorf("expected earned (%d) == possible (%d)", r.TotalEarned, r.TotalPossible)
	}
	// Verify expected total: Fail=3 + Warn=2 + Info=1 = 6
	if r.TotalPossible != 6 {
		t.Errorf("expected TotalPossible=6, got %d", r.TotalPossible)
	}
}

func TestNewReport_AllFails(t *testing.T) {
	results := []*checks.Result{
		makeResult("B1", checks.CatTokenEfficiency, checks.Fail, checks.StatusFail),
		makeResult("B2", checks.CatFlowSafety, checks.Warn, checks.StatusFail),
		makeResult("B3", checks.CatAutomationSafety, checks.Info, checks.StatusFail),
	}

	r := NewReport(results, "/bin/test", "", 50*time.Millisecond)

	if r.Percent != 0 {
		t.Errorf("expected 0%%, got %.2f%%", r.Percent)
	}
	if r.Grade != GradeF {
		t.Errorf("expected grade F, got %s", r.Grade)
	}
	if r.TotalEarned != 0 {
		t.Errorf("expected TotalEarned=0, got %d", r.TotalEarned)
	}
}

func TestNewReport_Mixed(t *testing.T) {
	results := []*checks.Result{
		// 2 pass (Fail sev = 3 pts each) → 6 earned
		makeResult("C1", checks.CatTokenEfficiency, checks.Fail, checks.StatusPass),
		makeResult("C2", checks.CatTokenEfficiency, checks.Fail, checks.StatusPass),
		// 1 fail (Fail sev = 3 pts possible) → 0 earned
		makeResult("C3", checks.CatFlowSafety, checks.Fail, checks.StatusFail),
	}

	r := NewReport(results, "/bin/test", "", time.Second)

	// 6 earned / 9 possible = 66.67%
	expectedPct := float64(6) / float64(9) * 100
	if r.Percent < expectedPct-0.01 || r.Percent > expectedPct+0.01 {
		t.Errorf("expected ~%.2f%%, got %.2f%%", expectedPct, r.Percent)
	}
	// 66.67% → Grade C
	if r.Grade != GradeC {
		t.Errorf("expected grade C, got %s", r.Grade)
	}
}

func TestNewReport_SkippedExcludedFromPossible(t *testing.T) {
	results := []*checks.Result{
		makeResult("D1", checks.CatTokenEfficiency, checks.Fail, checks.StatusPass),  // 3/3
		makeResult("D2", checks.CatTokenEfficiency, checks.Fail, checks.StatusSkip),  // 0/0 (excluded)
		makeResult("D3", checks.CatFlowSafety, checks.Warn, checks.StatusPass),   // 2/2
	}

	r := NewReport(results, "/bin/test", "", time.Second)

	// Skipped check should not contribute to possible points.
	// Expected: 5 earned, 5 possible → 100%
	if r.TotalPossible != 5 {
		t.Errorf("expected TotalPossible=5 (skip excluded), got %d", r.TotalPossible)
	}
	if r.TotalEarned != 5 {
		t.Errorf("expected TotalEarned=5, got %d", r.TotalEarned)
	}
	if r.Percent != 100 {
		t.Errorf("expected 100%%, got %.2f%%", r.Percent)
	}
}

func TestHasCriticalFailures_True(t *testing.T) {
	results := []*checks.Result{
		makeResult("E1", checks.CatTokenEfficiency, checks.Fail, checks.StatusPass),
		makeResult("E2", checks.CatTokenEfficiency, checks.Fail, checks.StatusFail), // critical fail
	}

	r := NewReport(results, "/bin/test", "", time.Second)

	if !r.HasCriticalFailures() {
		t.Error("expected HasCriticalFailures() to return true")
	}
}

func TestHasCriticalFailures_False_AllPass(t *testing.T) {
	results := []*checks.Result{
		makeResult("F1", checks.CatTokenEfficiency, checks.Fail, checks.StatusPass),
		makeResult("F2", checks.CatTokenEfficiency, checks.Warn, checks.StatusPass),
	}

	r := NewReport(results, "/bin/test", "", time.Second)

	if r.HasCriticalFailures() {
		t.Error("expected HasCriticalFailures() to return false when all pass")
	}
}

func TestHasCriticalFailures_False_WarnOnly(t *testing.T) {
	results := []*checks.Result{
		makeResult("G1", checks.CatTokenEfficiency, checks.Warn, checks.StatusFail), // warn sev, not critical
		makeResult("G2", checks.CatTokenEfficiency, checks.Fail, checks.StatusPass),
	}

	r := NewReport(results, "/bin/test", "", time.Second)

	if r.HasCriticalFailures() {
		t.Error("expected HasCriticalFailures() to return false when only warn-severity checks fail")
	}
}

func TestHasCriticalFailures_False_SkippedCritical(t *testing.T) {
	results := []*checks.Result{
		makeResult("H1", checks.CatTokenEfficiency, checks.Fail, checks.StatusSkip), // skipped, not a failure
	}

	r := NewReport(results, "/bin/test", "", time.Second)

	if r.HasCriticalFailures() {
		t.Error("expected HasCriticalFailures() to return false when critical checks are skipped")
	}
}

func TestGetSummary(t *testing.T) {
	results := []*checks.Result{
		makeResult("S1", checks.CatTokenEfficiency, checks.Fail, checks.StatusPass),
		makeResult("S2", checks.CatTokenEfficiency, checks.Fail, checks.StatusFail),
		makeResult("S3", checks.CatFlowSafety, checks.Warn, checks.StatusFail),
		makeResult("S4", checks.CatFlowSafety, checks.Info, checks.StatusSkip),
		makeResult("S5", checks.CatAutomationSafety, checks.Fail, checks.StatusPass),
	}

	r := NewReport(results, "/bin/test", "", time.Second)
	s := r.GetSummary()

	if s.Total != 5 {
		t.Errorf("expected Total=5, got %d", s.Total)
	}
	if s.Pass != 2 {
		t.Errorf("expected Pass=2, got %d", s.Pass)
	}
	if s.Fail != 1 {
		t.Errorf("expected Fail=1, got %d", s.Fail)
	}
	if s.Warn != 1 {
		t.Errorf("expected Warn=1, got %d", s.Warn)
	}
	if s.Skip != 1 {
		t.Errorf("expected Skip=1, got %d", s.Skip)
	}
	if s.Error != 0 {
		t.Errorf("expected Error=0, got %d", s.Error)
	}
}

func TestGradeFromPercent_Boundaries(t *testing.T) {
	tests := []struct {
		pct  float64
		want Grade
	}{
		{100, GradeA},
		{90, GradeA},
		{89, GradeB},
		{70, GradeB},
		{69, GradeC},
		{50, GradeC},
		{49, GradeD},
		{30, GradeD},
		{29, GradeF},
		{0, GradeF},
	}

	for _, tt := range tests {
		got := GradeFromPercent(tt.pct)
		if got != tt.want {
			t.Errorf("GradeFromPercent(%.0f) = %s, want %s", tt.pct, got, tt.want)
		}
	}
}

func TestNewReport_NoResults(t *testing.T) {
	r := NewReport(nil, "/bin/test", "", time.Second)

	if r.Percent != 0 {
		t.Errorf("expected 0%% for no results, got %.2f%%", r.Percent)
	}
	if r.Grade != GradeF {
		t.Errorf("expected grade F for no results, got %s", r.Grade)
	}
	if r.TotalPossible != 0 {
		t.Errorf("expected TotalPossible=0, got %d", r.TotalPossible)
	}
}

func TestNewReport_Categories(t *testing.T) {
	results := []*checks.Result{
		makeResult("K1", checks.CatTokenEfficiency, checks.Fail, checks.StatusPass),
		makeResult("K2", checks.CatFlowSafety, checks.Fail, checks.StatusFail),
		makeResult("K3", checks.CatFlowSafety, checks.Warn, checks.StatusPass),
	}

	r := NewReport(results, "/bin/test", "", time.Second)

	if len(r.Categories) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(r.Categories))
	}

	// First category should be flow-safety (sorted by AllCategories order).
	if r.Categories[0].Category != checks.CatFlowSafety {
		t.Errorf("expected first category to be flow-safety, got %s", r.Categories[0].Category)
	}
	if r.Categories[0].Earned != 2 {
		t.Errorf("expected flow-safety earned=2, got %d", r.Categories[0].Earned)
	}
	if r.Categories[0].Possible != 5 {
		t.Errorf("expected flow-safety possible=5, got %d", r.Categories[0].Possible)
	}

	// Second category: token-efficiency. 3 earned / 3 possible = 100%
	if r.Categories[1].Category != checks.CatTokenEfficiency {
		t.Errorf("expected second category to be token-efficiency, got %s", r.Categories[1].Category)
	}
	if r.Categories[1].Percent != 100 {
		t.Errorf("expected token-efficiency at 100%%, got %.2f%%", r.Categories[1].Percent)
	}
}

func TestGradeLabel(t *testing.T) {
	tests := []struct {
		grade Grade
		want  string
	}{
		{GradeA, "Agent-ready"},
		{GradeB, "Mostly ready, some gaps"},
		{GradeC, "Significant gaps"},
		{GradeD, "Major work needed"},
		{GradeF, "Not agent-ready"},
	}
	for _, tt := range tests {
		got := GradeLabel(tt.grade)
		if got != tt.want {
			t.Errorf("GradeLabel(%s) = %q, want %q", tt.grade, got, tt.want)
		}
	}
}

func TestAttentionCount(t *testing.T) {
	results := []*checks.Result{
		makeResult("AC1", checks.CatTokenEfficiency, checks.Fail, checks.StatusPass),
		makeResult("AC2", checks.CatTokenEfficiency, checks.Fail, checks.StatusFail),
		makeResult("AC3", checks.CatFlowSafety, checks.Warn, checks.StatusFail),
		makeResult("AC4", checks.CatFlowSafety, checks.Info, checks.StatusSkip),
	}

	r := NewReport(results, "/bin/test", "", time.Second)
	total, failCount, warnCount := r.AttentionCount()

	if total != 2 {
		t.Errorf("expected total attention=2, got %d", total)
	}
	if failCount != 1 {
		t.Errorf("expected failCount=1, got %d", failCount)
	}
	if warnCount != 1 {
		t.Errorf("expected warnCount=1, got %d", warnCount)
	}
}
