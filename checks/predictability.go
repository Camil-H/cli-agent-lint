package checks

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/Camil-H/cli-agent-lint/discovery"
)

// PV-1: Timeout flag

type checkPV1 struct {
	BaseCheck
}

func newCheckPV1() *checkPV1 {
	return &checkPV1{
		BaseCheck: BaseCheck{
			CheckID:             "PV-1",
			CheckName:           "Timeout flag",
			CheckCategory:       CatPredictability,
			CheckSeverity:       Info,
			CheckMethod:         Passive,
			CheckRecommendation: "Add `--timeout` flag so agents can enforce time budgets.",
		},
	}
}

func (c *checkPV1) Run(ctx context.Context, input *Input) *Result {
	idx := input.GetIndex()
	if idx == nil {
		return SkipResult(c, "no command tree available")
	}

	if idx.HasFlag(timeoutFlagNames...) {
		return PassResult(c, "found --timeout or --request-timeout flag")
	}

	if _, ok := idx.HelpContains("--timeout"); ok {
		return PassResult(c, "found --timeout reference in help text")
	}

	return FailResult(c, "no --timeout or --request-timeout flag found")
}

// PV-2: Retry / rate-limit hints

type checkPV2 struct {
	BaseCheck
}

func newCheckPV2() *checkPV2 {
	return &checkPV2{
		BaseCheck: BaseCheck{
			CheckID:             "PV-2",
			CheckName:           "Retry / rate-limit hints",
			CheckCategory:       CatPredictability,
			CheckSeverity:       Info,
			CheckMethod:         Passive,
			CheckRecommendation: "Surface rate-limit and retry-after information in structured error output.",
		},
	}
}

func hasNetworkIndicators(idx *discovery.CommandIndex) bool {
	if idx.HasFlag(networkIndicatorFlags...) {
		return true
	}
	_, found := idx.HelpContainsAny(networkHelpTerms...)
	return found
}

func (c *checkPV2) Run(ctx context.Context, input *Input) *Result {
	idx := input.GetIndex()
	if idx == nil {
		return SkipResult(c, "no command tree available")
	}

	if idx.HasFlag(retryFlagNames...) {
		return PassResult(c, "found retry-related flag")
	}

	if _, ok := idx.HelpContainsAny(retryHelpTerms...); ok {
		return PassResult(c, "found retry/rate-limit reference in help text")
	}

	if !hasNetworkIndicators(idx) {
		return PassResult(c, "no network/API commands detected; retry not applicable")
	}

	return FailResult(c, "no retry or rate-limit flags or help text found")
}

// PV-3: Deterministic output

type checkPV3 struct {
	BaseCheck
}

func newCheckPV3() *checkPV3 {
	return &checkPV3{
		BaseCheck: BaseCheck{
			CheckID:             "PV-3",
			CheckName:           "Deterministic output",
			CheckCategory:       CatPredictability,
			CheckSeverity:       Info,
			CheckMethod:         Active,
			CheckRecommendation: "Ensure identical inputs produce identical outputs. Avoid injecting timestamps or random values unless requested.",
		},
	}
}

func (c *checkPV3) Run(ctx context.Context, input *Input) *Result {
	if r := skipIfNoProber(c, input); r != nil {
		return r
	}

	result1, err := input.Prober.RunHelp(ctx)
	if err != nil {
		return ErrorResult(c, fmt.Errorf("running --help (first): %w", err))
	}

	result2, err := input.Prober.RunHelp(ctx)
	if err != nil {
		return ErrorResult(c, fmt.Errorf("running --help (second): %w", err))
	}

	if !bytes.Equal(result1.Stdout, result2.Stdout) {
		return FailResult(c, "two consecutive --help invocations produced different stdout")
	}

	return PassResult(c, "--help output is deterministic across two runs")
}

// PV-4: Distinct exit codes for error classes

type checkPV4 struct {
	BaseCheck
}

func newCheckPV4() *checkPV4 {
	return &checkPV4{
		BaseCheck: BaseCheck{
			CheckID:             "PV-4",
			CheckName:           "Distinct exit codes for error classes",
			CheckCategory:       CatPredictability,
			CheckSeverity:       Info,
			CheckMethod:         Passive,
			CheckRecommendation: "Document and use distinct exit codes (e.g., 1=error, 2=usage, 4=auth) so agents can classify failures without parsing stderr.",
		},
	}
}

var exitCodeSectionRe = regexp.MustCompile(`(?m)^(?:EXIT STATUS|EXIT CODES|Exit [Cc]odes?)`)

func (c *checkPV4) Run(ctx context.Context, input *Input) *Result {
	idx := input.GetIndex()
	if idx == nil {
		return SkipResult(c, "no command tree available")
	}

	for _, cmd := range idx.All() {
		if exitCodeSectionRe.MatchString(cmd.RawHelp) {
			return PassResult(c, fmt.Sprintf("found exit code documentation section in %q help text", strings.Join(cmd.FullPath, " ")))
		}
	}

	if _, ok := idx.HelpContainsAny(exitCodeHelpTerms...); ok {
		return PassResult(c, "found exit code documentation in help text")
	}

	return FailResult(c, "no exit code documentation found in help text")
}

// PV-5: Reports actual effects in output

type checkPV5 struct {
	BaseCheck
}

func newCheckPV5() *checkPV5 {
	return &checkPV5{
		BaseCheck: BaseCheck{
			CheckID:             "PV-5",
			CheckName:           "Reports actual effects",
			CheckCategory:       CatPredictability,
			CheckSeverity:       Info,
			CheckMethod:         Passive,
			CheckRecommendation: "Mutating commands should report what actually happened (e.g., \"created 3 items, skipped 1 duplicate\") so agents can verify their work.",
		},
	}
}

var effectReportTerms = []string{
	"created", "updated", "deleted", "removed", "skipped",
	"unchanged", "modified", "applied", "succeeded", "failed",
	"added", "total", "count", "summary",
}

func (c *checkPV5) Run(ctx context.Context, input *Input) *Result {
	idx := input.GetIndex()
	if idx == nil {
		return SkipResult(c, "no command tree available")
	}

	mutating := idx.Mutating()
	if len(mutating) == 0 {
		return PassResult(c, "no mutating commands detected; effect reporting not applicable")
	}

	// Check if mutating commands mention effect reporting in help
	for _, cmd := range mutating {
		h := idx.LowerHelp(cmd)
		for _, term := range effectReportTerms {
			if strings.Contains(h, term) {
				return PassResult(c, fmt.Sprintf("found effect-reporting term %q in %q help text",
					term, strings.Join(cmd.FullPath, " ")))
			}
		}
	}

	// Check if JSON output mentions result/count fields
	if _, ok := idx.HelpContainsAny("result", "count", "summary", "affected", "rows"); ok {
		return PassResult(c, "found result/count reference in help text")
	}

	return FailResult(c, "mutating commands found but no effect-reporting terms in help text")
}
