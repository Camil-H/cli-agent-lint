package checks

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/cli-agent-lint/cli-agent-lint/discovery"
	"github.com/cli-agent-lint/cli-agent-lint/probe"
)

// ---------------------------------------------------------------------------
// TH-1: Non-TTY detection (no ANSI in pipes)
// ---------------------------------------------------------------------------

type checkTH1 struct {
	BaseCheck
}

func newCheckTH1() *checkTH1 {
	return &checkTH1{
		BaseCheck: BaseCheck{
			CheckID:             "TH-1",
			CheckName:           "Non-TTY detection (no ANSI in pipes)",
			CheckCategory:       CatTerminalHygiene,
			CheckSeverity:       Fail,
			CheckMethod:         Active,
			CheckRecommendation: "Detect non-TTY stdout and disable color/formatting automatically.",
		},
	}
}

func (c *checkTH1) Run(ctx context.Context, input *Input) *Result {
	if r := skipIfNoProber(c, input); r != nil {
		return r
	}

	// No NO_COLOR set — tests whether the tool auto-detects non-TTY stdout.
	result, err := input.Prober.RunHelp(ctx)
	if err != nil {
		return ErrorResult(c, fmt.Errorf("running --help: %w", err))
	}

	combined := result.StdoutStr() + "\n" + result.StderrStr()

	if strings.Contains(combined, "\x1b[") {
		return FailResult(c, "ANSI escape sequences detected in non-TTY output")
	}

	return PassResult(c, "no ANSI escape sequences found in piped output")
}

// ---------------------------------------------------------------------------
// TH-2: --no-color flag
// ---------------------------------------------------------------------------

type checkTH2 struct {
	BaseCheck
}

func newCheckTH2() *checkTH2 {
	return &checkTH2{
		BaseCheck: BaseCheck{
			CheckID:             "TH-2",
			CheckName:           "--no-color flag",
			CheckCategory:       CatTerminalHygiene,
			CheckSeverity:       Warn,
			CheckMethod:         Passive,
			CheckRecommendation: "Support `--no-color` flag and/or the `NO_COLOR` env var (see https://no-color.org).",
		},
	}
}

var noColorHelpRe = regexp.MustCompile(`(?i)(--no-color|--color[= ]never|NO_COLOR)`)

func (c *checkTH2) Run(ctx context.Context, input *Input) *Result {
	if input.Tree == nil || input.Tree.Root == nil {
		return SkipResult(c, "no command tree available")
	}
	root := input.Tree.Root

	if root.HasFlag("no-color", "color") {
		return PassResult(c, "found --no-color or --color flag")
	}

	if noColorHelpRe.MatchString(root.RawHelp) {
		return PassResult(c, "found color-control reference in help text")
	}

	return FailResult(c, "no --no-color flag or NO_COLOR support detected")
}

// ---------------------------------------------------------------------------
// TH-3: --quiet / --silent flag
// ---------------------------------------------------------------------------

type checkTH3 struct {
	BaseCheck
}

func newCheckTH3() *checkTH3 {
	return &checkTH3{
		BaseCheck: BaseCheck{
			CheckID:             "TH-3",
			CheckName:           "--quiet / --silent flag",
			CheckCategory:       CatTerminalHygiene,
			CheckSeverity:       Info,
			CheckMethod:         Passive,
			CheckRecommendation: "Add `--quiet` flag to suppress informational output, leaving only essential data.",
		},
	}
}

var quietHelpRe = regexp.MustCompile(`(?i)(--quiet|--silent|-q\b)`)

func (c *checkTH3) Run(ctx context.Context, input *Input) *Result {
	if input.Tree == nil || input.Tree.Root == nil {
		return SkipResult(c, "no command tree available")
	}
	root := input.Tree.Root

	if root.HasFlag("quiet", "silent", "q") {
		return PassResult(c, "found --quiet/--silent/-q flag")
	}

	if quietHelpRe.MatchString(root.RawHelp) {
		return PassResult(c, "found quiet/silent reference in help text")
	}

	return FailResult(c, "no --quiet or --silent flag detected")
}

// ---------------------------------------------------------------------------
// TH-4: No interactive prompts in non-TTY
// ---------------------------------------------------------------------------

type checkTH4 struct {
	BaseCheck
}

func newCheckTH4() *checkTH4 {
	return &checkTH4{
		BaseCheck: BaseCheck{
			CheckID:             "TH-4",
			CheckName:           "No interactive prompts in non-TTY",
			CheckCategory:       CatTerminalHygiene,
			CheckSeverity:       Fail,
			CheckMethod:         Active,
			CheckRecommendation: "Never prompt for input when stdin is not a TTY. Fail fast with a clear error instead.",
		},
	}
}

var promptRe = regexp.MustCompile(`(?i)(enter\b|password:|press\b|continue\?|y/n|\(yes/no\))`)

func (c *checkTH4) Run(ctx context.Context, input *Input) *Result {
	if r := skipIfNoProber(c, input); r != nil {
		return r
	}

	idx := input.GetIndex()
	if idx == nil {
		return SkipResult(c, "no command index available")
	}

	// Pick a non-mutating (list-like) command to run bare — safe to execute
	// without side effects. Prefer a leaf, fall back to any list-like command.
	// Never run mutating commands bare (SEC-1: e.g. kubectl delete).
	var candidate *discovery.Command
	for _, cmd := range idx.ListLike() {
		if len(cmd.Subcommands) == 0 {
			candidate = cmd
			break
		}
	}
	if candidate == nil {
		listLike := idx.ListLike()
		if len(listLike) > 0 {
			candidate = listLike[0]
		}
	}
	if candidate == nil {
		return PassResult(c, "no non-mutating commands available to safely test")
	}

	args := make([]string, len(candidate.FullPath)-1)
	copy(args, candidate.FullPath[1:])

	result, err := input.Prober.Run(ctx, probe.Opts{
		Args: args,
	})
	if err != nil {
		return ErrorResult(c, fmt.Errorf("running %s: %w", strings.Join(candidate.FullPath, " "), err))
	}

	if result.TimedOut {
		return FailResult(c, fmt.Sprintf("command %q timed out (likely waiting for interactive input)", strings.Join(candidate.FullPath, " ")))
	}

	combined := result.StdoutStr() + "\n" + result.StderrStr()
	if promptRe.MatchString(combined) {
		return FailResult(c, fmt.Sprintf("command %q produced prompt-like output in non-TTY context", strings.Join(candidate.FullPath, " ")))
	}

	return PassResult(c, fmt.Sprintf("command %q exited without prompting", strings.Join(candidate.FullPath, " ")))
}

// ---------------------------------------------------------------------------
// TH-5: Confirmation bypass for destructive commands
// ---------------------------------------------------------------------------

type checkTH5 struct {
	BaseCheck
}

func newCheckTH5() *checkTH5 {
	return &checkTH5{
		BaseCheck: BaseCheck{
			CheckID:             "TH-5",
			CheckName:           "Confirmation bypass for destructive commands",
			CheckCategory:       CatTerminalHygiene,
			CheckSeverity:       Warn,
			CheckMethod:         Passive,
			CheckRecommendation: "Add a --yes or --force flag to destructive commands so agents can skip interactive confirmation.",
		},
	}
}

func (c *checkTH5) Run(ctx context.Context, input *Input) *Result {
	idx := input.GetIndex()
	if idx == nil {
		return SkipResult(c, "no command index available")
	}
	mutating := idx.Mutating()

	if len(mutating) == 0 {
		return PassResult(c, "no mutating commands detected")
	}

	var missing []string
	bypassNames := append(confirmBypassFlagNames, "y")
	for _, cmd := range mutating {
		if !idx.CmdHasFlag(cmd, bypassNames...) {
			missing = append(missing, strings.Join(cmd.FullPath, " "))
		}
	}

	if len(missing) == 0 {
		return PassResult(c, fmt.Sprintf("all %d mutating command(s) have a confirmation bypass flag", len(mutating)))
	}

	detail := fmt.Sprintf("%d of %d mutating command(s) missing confirmation bypass flag: %s",
		len(missing), len(mutating), strings.Join(missing, ", "))
	return FailResult(c, detail)
}

// ---------------------------------------------------------------------------
// Registration
// ---------------------------------------------------------------------------

func registerTerminalHygieneChecks(r *Registry) {
	r.Register(newCheckTH1())
	r.Register(newCheckTH2())
	r.Register(newCheckTH3())
	r.Register(newCheckTH4())
	r.Register(newCheckTH5())
}
