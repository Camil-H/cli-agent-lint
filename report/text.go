package report

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/Camil-H/cli-agent-lint/checks"
)

var ansiRe = regexp.MustCompile(
	`\x1b\[[0-9;:]*[a-zA-Z]` + // CSI sequences (colors, cursor movement)
		`|\x1b\][^\x07]*\x07` + // OSC sequences (terminal title, hyperlinks, clipboard)
		`|\x1bP[^\x1b]*\x1b\\` + // DCS sequences (device control)
		`|\x1b[()][0-9A-B]` + // Character set selection
		`|\x1b[>=<~]` + // Keypad/cursor modes
		`|\x1b\[[\?]?[0-9;]*[hlmsuJKHf]`, // Private mode set/reset and other CSI
)

func stripANSI(s string) string {
	return ansiRe.ReplaceAllString(s, "")
}

type palette struct {
	green  string
	yellow string
	red    string
	bold   string
	dim    string
	reset  string
}

func newPalette(noColor bool) palette {
	if noColor {
		return palette{}
	}
	return palette{
		green:  "\x1b[32m",
		yellow: "\x1b[33m",
		red:    "\x1b[31m",
		bold:   "\x1b[1m",
		dim:    "\x1b[2m",
		reset:  "\x1b[0m",
	}
}

type TextFormatter struct {
	NoColor bool
	Quiet   bool
}

func (f *TextFormatter) Format(w io.Writer, r *Report) error {
	bw := bufio.NewWriter(w)
	p := newPalette(f.NoColor)

	var err error
	if f.Quiet {
		err = f.formatQuiet(bw, r, p)
	} else {
		err = f.formatFull(bw, r, p)
	}
	if err != nil {
		return err
	}
	return bw.Flush()
}

func (f *TextFormatter) formatFull(w io.Writer, r *Report, p palette) error {
	fmt.Fprintf(w, "%s%scli-agent-lint report%s\n", p.bold, p.green, p.reset)
	fmt.Fprintf(w, "Target: %s\n", r.TargetPath)
	if r.TargetVersion != "" {
		fmt.Fprintf(w, "Version: %s\n", r.TargetVersion)
	}
	fmt.Fprintln(w)

	for _, cs := range r.Categories {
		fmt.Fprintf(w, "%s%s── %s%s%s\n", p.bold, p.dim, categoryLabel(cs.Category), p.reset, "")
		for _, res := range cs.Results {
			f.writeResult(w, res, p)
		}
		fmt.Fprintln(w)
	}

	f.writeScore(w, r, p)

	attn, failCount, warnCount := r.AttentionCount()
	if attn > 0 {
		parts := []string{}
		if failCount > 0 {
			parts = append(parts, fmt.Sprintf("%d fail", failCount))
		}
		if warnCount > 0 {
			parts = append(parts, fmt.Sprintf("%d warn", warnCount))
		}
		fmt.Fprintf(w, "%s%d checks need attention (%s)%s\n", p.yellow, attn, strings.Join(parts, ", "), p.reset)
	}

	return nil
}

func (f *TextFormatter) formatQuiet(w io.Writer, r *Report, p palette) error {
	f.writeScore(w, r, p)

	// Show only failures.
	for _, res := range r.Results {
		if res.Status == checks.StatusFail && res.Severity == checks.Fail {
			f.writeResult(w, res, p)
		}
	}

	return nil
}

func (f *TextFormatter) writeResult(w io.Writer, res *checks.Result, p palette) {
	indicator := f.statusIndicator(res, p)
	fmt.Fprintf(w, "  %s %s%s%s: %s\n", indicator, p.bold, res.CheckID, p.reset, res.CheckName)

	if res.Status != checks.StatusPass && res.Status != checks.StatusSkip {
		if res.Detail != "" {
			fmt.Fprintf(w, "       %s%s%s\n", p.dim, stripANSI(res.Detail), p.reset)
		}
		if res.Recommendation != "" {
			fmt.Fprintf(w, "       %s→ %s%s\n", p.dim, res.Recommendation, p.reset)
		}
	}
}

func (f *TextFormatter) statusIndicator(res *checks.Result, p palette) string {
	switch res.Status {
	case checks.StatusPass:
		return fmt.Sprintf("%s[PASS]%s", p.green, p.reset)
	case checks.StatusFail:
		if res.Severity == checks.Fail {
			return fmt.Sprintf("%s[FAIL]%s", p.red, p.reset)
		}
		return fmt.Sprintf("%s[WARN]%s", p.yellow, p.reset)
	case checks.StatusSkip:
		return fmt.Sprintf("%s[SKIP]%s", p.dim, p.reset)
	case checks.StatusError:
		return fmt.Sprintf("%s[ERR ]%s", p.red, p.reset)
	default:
		return "[????]"
	}
}

func (f *TextFormatter) writeScore(w io.Writer, r *Report, p palette) {
	gradeColor := p.green
	switch r.Grade {
	case GradeC, GradeD:
		gradeColor = p.yellow
	case GradeF:
		gradeColor = p.red
	}

	fmt.Fprintf(w, "Score: %.0f%% %s%sGrade: %s — %s%s\n",
		r.Percent,
		gradeColor, p.bold, r.Grade, GradeLabel(r.Grade), p.reset)
}

func categoryLabel(cat checks.Category) string {
	switch cat {
	case checks.CatFlowSafety:
		return "Flow Safety"
	case checks.CatTokenEfficiency:
		return "Token Efficiency"
	case checks.CatSelfDescribing:
		return "Self-Describing"
	case checks.CatAutomationSafety:
		return "Automation Safety"
	case checks.CatPredictability:
		return "Predictable & Verifiable"
	default:
		return string(cat)
	}
}
