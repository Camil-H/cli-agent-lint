package report

import (
	"time"

	"github.com/Camil-H/cli-agent-lint/checks"
)

type Grade string

const (
	GradeA Grade = "A"
	GradeB Grade = "B"
	GradeC Grade = "C"
	GradeD Grade = "D"
	GradeF Grade = "F"
)

func GradeFromPercent(pct float64) Grade {
	switch {
	case pct >= 90:
		return GradeA
	case pct >= 70:
		return GradeB
	case pct >= 50:
		return GradeC
	case pct >= 30:
		return GradeD
	default:
		return GradeF
	}
}

func GradeLabel(g Grade) string {
	switch g {
	case GradeA:
		return "Agent-ready"
	case GradeB:
		return "Mostly ready, some gaps"
	case GradeC:
		return "Significant gaps"
	case GradeD:
		return "Major work needed"
	case GradeF:
		return "Not agent-ready"
	default:
		return ""
	}
}

type CategoryScore struct {
	Category checks.Category
	Earned   int
	Possible int
	Percent  float64
	Results  []*checks.Result
}

type Report struct {
	TargetPath    string
	TargetVersion string
	Timestamp     time.Time
	Duration      time.Duration
	Results       []*checks.Result
	Categories    []*CategoryScore
	TotalEarned   int
	TotalPossible int
	Percent       float64
	Grade         Grade
}

func NewReport(results []*checks.Result, targetPath string, targetVersion string, duration time.Duration) *Report {
	checks.SortResultsByCategory(results)

	r := &Report{
		TargetPath:    targetPath,
		TargetVersion: targetVersion,
		Timestamp:     time.Now(),
		Duration:      duration,
		Results:       results,
	}

	catMap := make(map[checks.Category]*CategoryScore)
	var catOrder []checks.Category

	for _, res := range results {
		cs, ok := catMap[res.Category]
		if !ok {
			cs = &CategoryScore{Category: res.Category}
			catMap[res.Category] = cs
			catOrder = append(catOrder, res.Category)
		}
		cs.Results = append(cs.Results, res)
		earned := res.Points()
		possible := res.MaxPoints()
		cs.Earned += earned
		cs.Possible += possible
		r.TotalEarned += earned
		r.TotalPossible += possible
	}

	for _, cat := range catOrder {
		cs := catMap[cat]
		if cs.Possible > 0 {
			cs.Percent = float64(cs.Earned) / float64(cs.Possible) * 100
		}
		r.Categories = append(r.Categories, cs)
	}

	if r.TotalPossible > 0 {
		r.Percent = float64(r.TotalEarned) / float64(r.TotalPossible) * 100
	}
	r.Grade = GradeFromPercent(r.Percent)

	return r
}

func (r *Report) HasCriticalFailures() bool {
	for _, res := range r.Results {
		if res.Severity == checks.Fail && res.Status != checks.StatusPass && res.Status != checks.StatusSkip {
			return true
		}
	}
	return false
}

type Summary struct {
	Total int
	Pass  int
	Warn  int
	Fail  int
	Skip  int
	Error int
}

func (r *Report) GetSummary() Summary {
	var s Summary
	for _, res := range r.Results {
		s.Total++
		switch res.Status {
		case checks.StatusPass:
			s.Pass++
		case checks.StatusFail:
			if res.Severity == checks.Warn {
				s.Warn++
			} else {
				s.Fail++
			}
		case checks.StatusSkip:
			s.Skip++
		case checks.StatusError:
			s.Error++
		}
	}
	return s
}

// AttentionCount returns the number of non-pass, non-skip checks.
func (r *Report) AttentionCount() (total, failCount, warnCount int) {
	for _, res := range r.Results {
		if res.Status == checks.StatusFail {
			total++
			if res.Severity == checks.Fail {
				failCount++
			} else {
				warnCount++
			}
		}
	}
	return
}
