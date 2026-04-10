package report

import (
	"encoding/json"
	"io"

	"github.com/Camil-H/cli-agent-lint/checks"
)

type JSONFormatter struct{}

type jsonReport struct {
	Version    string         `json:"version"`
	Target     string         `json:"target"`
	Timestamp  string         `json:"timestamp"`
	DurationMs int64          `json:"duration_ms"`
	Score      jsonScore      `json:"score"`
	Summary    jsonSummary    `json:"summary"`
	Checks     []jsonCheck    `json:"checks"`
	Categories []jsonCategory `json:"categories"`
}

type jsonScore struct {
	Percentage float64 `json:"percentage"`
	Grade      string  `json:"grade"`
}

type jsonSummary struct {
	Total int `json:"total"`
	Pass  int `json:"pass"`
	Warn  int `json:"warn"`
	Fail  int `json:"fail"`
	Skip  int `json:"skip"`
	Error int `json:"error"`
}

type jsonCheck struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Category       string `json:"category"`
	Severity       string `json:"severity"`
	Status         string `json:"status"`
	Points         int    `json:"points"`
	MaxPoints      int    `json:"max_points"`
	Passive        bool   `json:"passive"`
	Recommendation string `json:"recommendation"`
	Detail         string `json:"detail,omitempty"`
}

type jsonCategory struct {
	Name       string  `json:"name"`
	Earned     int     `json:"earned"`
	Possible   int     `json:"possible"`
	Percentage float64 `json:"percentage"`
}

func (f *JSONFormatter) Format(w io.Writer, r *Report) error {
	summary := r.GetSummary()

	jr := jsonReport{
		Version:    "1.0.0",
		Target:     r.TargetPath,
		Timestamp:  r.Timestamp.UTC().Format("2006-01-02T15:04:05Z"),
		DurationMs: r.Duration.Milliseconds(),
		Score: jsonScore{
			Percentage: r.Percent,
			Grade:      string(r.Grade),
		},
		Summary: jsonSummary{
			Total: summary.Total,
			Pass:  summary.Pass,
			Warn:  summary.Warn,
			Fail:  summary.Fail,
			Skip:  summary.Skip,
			Error: summary.Error,
		},
	}

	for _, res := range r.Results {
		jc := jsonCheck{
			ID:             res.CheckID,
			Name:           res.CheckName,
			Category:       string(res.Category),
			Severity:       res.Severity.String(),
			Status:         resultStatus(res),
			Points:         res.Points(),
			MaxPoints:      res.MaxPoints(),
			Passive:        isPassive(res),
			Recommendation: res.Recommendation,
			Detail:         res.Detail,
		}
		if res.Error != nil {
			jc.Detail = res.Error.Error()
		}
		jr.Checks = append(jr.Checks, jc)
	}

	for _, cs := range r.Categories {
		jr.Categories = append(jr.Categories, jsonCategory{
			Name:       string(cs.Category),
			Earned:     cs.Earned,
			Possible:   cs.Possible,
			Percentage: cs.Percent,
		})
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(jr)
}

// resultStatus maps internal status to display status.
// Failed checks with warn/info severity show as "warn"/"info" instead of "fail".
func resultStatus(res *checks.Result) string {
	if res.Status == checks.StatusFail && res.Severity == checks.Warn {
		return "warn"
	}
	if res.Status == checks.StatusFail && res.Severity == checks.Info {
		return "info"
	}
	return res.Status.String()
}

func isPassive(res *checks.Result) bool {
	return res.Method == checks.Passive
}
