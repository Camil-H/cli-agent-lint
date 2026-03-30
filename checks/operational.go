package checks

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/cli-agent-lint/cli-agent-lint/discovery"
	"github.com/cli-agent-lint/cli-agent-lint/probe"
)

// ---------------------------------------------------------------------------
// OR-1: Exit codes
// ---------------------------------------------------------------------------

type checkOR1 struct {
	BaseCheck
}

func newCheckOR1() *checkOR1 {
	return &checkOR1{
		BaseCheck: BaseCheck{
			CheckID:             "OR-1",
			CheckName:           "Exit codes",
			CheckCategory:       CatOperational,
			CheckSeverity:       Fail,
			CheckMethod:         Active,
			CheckRecommendation: "Use distinct non-zero exit codes for different failure modes (bad input, auth failure, server error).",
		},
	}
}

func (c *checkOR1) Run(ctx context.Context, input *Input) *Result {
	if r := skipIfNoProber(c, input); r != nil {
		return r
	}

	helpResult, err := input.Prober.RunHelp(ctx)
	if err != nil {
		return ErrorResult(c, fmt.Errorf("running --help: %w", err))
	}
	if helpResult.ExitCode != 0 {
		return FailResult(c, fmt.Sprintf("--help returned non-zero exit code %d", helpResult.ExitCode))
	}

	badResult, err := input.Prober.Run(ctx, probe.Opts{
		Args: []string{"__nonexistent_subcommand__"},
	})
	if err != nil {
		return ErrorResult(c, fmt.Errorf("running nonexistent subcommand: %w", err))
	}
	if badResult.ExitCode == 0 {
		return FailResult(c, "nonexistent subcommand returned exit code 0; expected non-zero")
	}

	return PassResult(c, fmt.Sprintf("--help exits 0, bad subcommand exits %d", badResult.ExitCode))
}

// ---------------------------------------------------------------------------
// OR-2: Timeout flag
// ---------------------------------------------------------------------------

type checkOR2 struct {
	BaseCheck
}

func newCheckOR2() *checkOR2 {
	return &checkOR2{
		BaseCheck: BaseCheck{
			CheckID:             "OR-2",
			CheckName:           "Timeout flag",
			CheckCategory:       CatOperational,
			CheckSeverity:       Info,
			CheckMethod:         Passive,
			CheckRecommendation: "Add `--timeout` flag so agents can enforce time budgets.",
		},
	}
}

func (c *checkOR2) Run(ctx context.Context, input *Input) *Result {
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

// ---------------------------------------------------------------------------
// OR-3: Pagination support
// ---------------------------------------------------------------------------

type checkOR3 struct {
	BaseCheck
}

func newCheckOR3() *checkOR3 {
	return &checkOR3{
		BaseCheck: BaseCheck{
			CheckID:             "OR-3",
			CheckName:           "Pagination support",
			CheckCategory:       CatOperational,
			CheckSeverity:       Warn,
			CheckMethod:         Passive,
			CheckRecommendation: "Support `--page-all` or NDJSON streaming for list commands to avoid silent truncation.",
		},
	}
}

func (c *checkOR3) Run(ctx context.Context, input *Input) *Result {
	idx := input.GetIndex()
	if idx == nil {
		return SkipResult(c, "no command tree available")
	}

	listCmds := idx.ListLike()
	if len(listCmds) == 0 {
		return PassResult(c, "no list-like commands found")
	}

	var listCommands []string
	var missing []string
	for _, cmd := range listCmds {
		fullPath := strings.Join(cmd.FullPath, " ")
		listCommands = append(listCommands, fullPath)
		if !idx.CmdHasFlag(cmd, paginationFlagNames...) {
			missing = append(missing, fullPath)
		}
	}

	if len(missing) == 0 {
		return PassResult(c, fmt.Sprintf("all %d list-like command(s) have pagination flags", len(listCommands)))
	}

	return FailResult(c, fmt.Sprintf("list-like commands missing pagination flags: %s", strings.Join(missing, ", ")))
}

// ---------------------------------------------------------------------------
// OR-4: Retry / rate-limit hints
// ---------------------------------------------------------------------------

type checkOR4 struct {
	BaseCheck
}

func newCheckOR4() *checkOR4 {
	return &checkOR4{
		BaseCheck: BaseCheck{
			CheckID:             "OR-4",
			CheckName:           "Retry / rate-limit hints",
			CheckCategory:       CatOperational,
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

func (c *checkOR4) Run(ctx context.Context, input *Input) *Result {
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

// ---------------------------------------------------------------------------
// OR-5: Deterministic output
// ---------------------------------------------------------------------------

type checkOR5 struct {
	BaseCheck
}

func newCheckOR5() *checkOR5 {
	return &checkOR5{
		BaseCheck: BaseCheck{
			CheckID:             "OR-5",
			CheckName:           "Deterministic output",
			CheckCategory:       CatOperational,
			CheckSeverity:       Info,
			CheckMethod:         Active,
			CheckRecommendation: "Ensure identical inputs produce identical outputs. Avoid injecting timestamps or random values unless requested.",
		},
	}
}

func (c *checkOR5) Run(ctx context.Context, input *Input) *Result {
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

// ---------------------------------------------------------------------------
// OR-6: Field masks / response filtering
// ---------------------------------------------------------------------------

type checkOR6 struct {
	BaseCheck
}

func newCheckOR6() *checkOR6 {
	return &checkOR6{
		BaseCheck: BaseCheck{
			CheckID:             "OR-6",
			CheckName:           "Field masks / response filtering",
			CheckCategory:       CatOperational,
			CheckSeverity:       Info,
			CheckMethod:         Passive,
			CheckRecommendation: "Support field masks or response filtering to limit output size and protect agent context windows.",
		},
	}
}

func (c *checkOR6) Run(ctx context.Context, input *Input) *Result {
	idx := input.GetIndex()
	if idx == nil {
		return SkipResult(c, "no command tree available")
	}

	if idx.HasFlag(filterFlagNames...) {
		return PassResult(c, "found field-filtering flag (e.g. --fields, --jq, --filter)")
	}

	prefixed := make([]string, len(filterFlagNames))
	for i, name := range filterFlagNames {
		prefixed[i] = "--" + name
	}
	if _, ok := idx.HelpContainsAny(prefixed...); ok {
		return PassResult(c, "found field-filtering reference in help text")
	}

	if len(idx.ListLike()) == 0 {
		return PassResult(c, "no data-listing commands detected; field filtering not applicable")
	}

	return FailResult(c, "no field-mask or response-filtering flags found")
}

// ---------------------------------------------------------------------------
// OR-7: Distinct exit codes for error classes
// ---------------------------------------------------------------------------

type checkOR7 struct {
	BaseCheck
}

func newCheckOR7() *checkOR7 {
	return &checkOR7{
		BaseCheck: BaseCheck{
			CheckID:             "OR-7",
			CheckName:           "Distinct exit codes for error classes",
			CheckCategory:       CatOperational,
			CheckSeverity:       Info,
			CheckMethod:         Passive,
			CheckRecommendation: "Document and use distinct exit codes (e.g., 1=error, 2=usage, 4=auth) so agents can classify failures without parsing stderr.",
		},
	}
}

var exitCodeSectionRe = regexp.MustCompile(`(?m)^(?:EXIT STATUS|EXIT CODES|Exit [Cc]odes?)`)

func (c *checkOR7) Run(ctx context.Context, input *Input) *Result {
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

// ---------------------------------------------------------------------------
// Registration
// ---------------------------------------------------------------------------

func registerOperationalChecks(r *Registry) {
	r.Register(newCheckOR1())
	r.Register(newCheckOR2())
	r.Register(newCheckOR3())
	r.Register(newCheckOR4())
	r.Register(newCheckOR5())
	r.Register(newCheckOR6())
	r.Register(newCheckOR7())
}
