package checks

import (
	"context"
	"fmt"
	"strings"

	"github.com/cli-agent-lint/cli-agent-lint/discovery"
	"github.com/cli-agent-lint/cli-agent-lint/probe"
)

// SA-2: Rejects path traversal

type checkSA2 struct{ BaseCheck }

func newCheckSA2() *checkSA2 {
	return &checkSA2{BaseCheck{
		CheckID:             "SA-2",
		CheckName:           "Rejects path traversal",
		CheckCategory:       CatAutomationSafety,
		CheckSeverity:       Warn,
		CheckMethod:         Active,
		CheckRecommendation: "Canonicalize and sandbox all file path inputs. Reject path traversal sequences.",
	}}
}

func (c *checkSA2) Run(ctx context.Context, input *Input) *Result {
	if r := skipIfNoProber(c, input); r != nil {
		return r
	}

	idx := input.GetIndex()
	fileCmds := idx.FileAccepting()
	if len(fileCmds) == 0 {
		return SkipResult(c, "no file-path accepting commands detected")
	}

	traversalPayload := "../../tmp/.cli-agent-lint-test-NONEXISTENT"
	var probeErrors []string

	for _, cmd := range fileCmds {
		args := make([]string, 0, len(cmd.FullPath)+2)
		// FullPath[0] is the binary name — skip it since the prober uses the target binary.
		args = append(args, cmd.FullPath[1:]...)

		flagName := idx.FileArgFlag(cmd)
		if flagName != "" {
			args = append(args, "--"+flagName, traversalPayload)
		} else {
			args = append(args, traversalPayload)
		}

		res, err := input.Prober.Run(ctx, probe.Opts{Args: args})
		if err != nil {
			probeErrors = append(probeErrors, fmt.Sprintf("%s: %v", strings.Join(cmd.FullPath, " "), err))
			continue
		}

		// Non-zero exit code means the input was rejected — pass.
		if res.ExitCode != 0 {
			return PassResult(c, fmt.Sprintf("command %q rejected path traversal input (exit code %d)",
				strings.Join(cmd.FullPath, " "), res.ExitCode))
		}

		// Exit code 0 — check if output contains error/warning about path traversal.
		combined := strings.ToLower(res.StdoutStr() + " " + res.StderrStr())
		for _, indicator := range []string{"path traversal", "traversal", "invalid path", "outside", "not allowed", "denied", "forbidden"} {
			if strings.Contains(combined, indicator) {
				return PassResult(c, fmt.Sprintf("command %q reported path traversal error in output",
					strings.Join(cmd.FullPath, " ")))
			}
		}

		// Exit 0 with no traversal warning — fail.
		return FailResult(c, fmt.Sprintf("command %q accepted path traversal input %q without error (exit code 0)",
			strings.Join(cmd.FullPath, " "), traversalPayload))
	}

	detail := "could not successfully probe any file-path accepting commands"
	if len(probeErrors) > 0 {
		detail += fmt.Sprintf("; probe errors: %s", strings.Join(probeErrors, "; "))
	}
	return SkipResult(c, detail)
}

// SA-3: Rejects control characters

type checkSA3 struct{ BaseCheck }

func newCheckSA3() *checkSA3 {
	return &checkSA3{BaseCheck{
		CheckID:             "SA-3",
		CheckName:           "Rejects control characters",
		CheckCategory:       CatAutomationSafety,
		CheckSeverity:       Warn,
		CheckMethod:         Active,
		CheckRecommendation: "Reject input containing control characters (ASCII < 0x20) to prevent injection.",
	}}
}

func (c *checkSA3) Run(ctx context.Context, input *Input) *Result {
	if r := skipIfNoProber(c, input); r != nil {
		return r
	}

	idx := input.GetIndex()
	stringCmds := idx.StringInput()
	var cmd *discovery.Command
	var flagName string
	if len(stringCmds) > 0 {
		cmd = stringCmds[0]
		flagName = idx.StringInputFlag(cmd)
	}
	if cmd == nil {
		return SkipResult(c, "no commands with string input detected")
	}

	controlPayload := "test\x01\x02\x03value"

	args := make([]string, 0, len(cmd.FullPath)+2)
	args = append(args, cmd.FullPath[1:]...)

	if flagName != "" {
		args = append(args, "--"+flagName, controlPayload)
	} else {
		args = append(args, controlPayload)
	}

	res, err := input.Prober.Run(ctx, probe.Opts{Args: args})
	if err != nil {
		return ErrorResult(c, fmt.Errorf("probing command %q: %w", strings.Join(cmd.FullPath, " "), err))
	}

	if res.ExitCode != 0 {
		return PassResult(c, fmt.Sprintf("command %q rejected control character input (exit code %d)",
			strings.Join(cmd.FullPath, " "), res.ExitCode))
	}

	return FailResult(c, fmt.Sprintf("command %q accepted control character input without error (exit code 0)",
		strings.Join(cmd.FullPath, " ")))
}

// SA-4: Dry-run support

type checkSA4 struct{ BaseCheck }

func newCheckSA4() *checkSA4 {
	return &checkSA4{BaseCheck{
		CheckID:             "SA-4",
		CheckName:           "Dry-run support",
		CheckCategory:       CatAutomationSafety,
		CheckSeverity:       Warn,
		CheckMethod:         Passive,
		CheckRecommendation: "Add `--dry-run` to all mutating commands so agents can validate before executing.",
	}}
}

func (c *checkSA4) Run(ctx context.Context, input *Input) *Result {
	idx := input.GetIndex()
	if idx == nil {
		return SkipResult(c, "no command index available")
	}
	mutating := idx.Mutating()

	if len(mutating) == 0 {
		return PassResult(c, "no mutating commands detected")
	}

	var missing []string
	for _, cmd := range mutating {
		if !idx.CmdHasFlag(cmd, dryRunFlagNames...) {
			missing = append(missing, strings.Join(cmd.FullPath, " "))
		}
	}

	if len(missing) == 0 {
		return PassResult(c, fmt.Sprintf("all %d mutating command(s) have a dry-run flag", len(mutating)))
	}

	detail := fmt.Sprintf("%d of %d mutating command(s) missing dry-run flag: %s",
		len(missing), len(mutating), strings.Join(missing, ", "))
	return FailResult(c, detail)
}


