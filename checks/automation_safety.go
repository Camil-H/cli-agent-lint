package checks

import (
	"context"
	"fmt"
	"strings"

	"github.com/Camil-H/cli-agent-lint/discovery"
	"github.com/Camil-H/cli-agent-lint/probe"
)

// SA-1: Confirmation bypass for destructive commands

type checkSA1 struct {
	BaseCheck
}

func newCheckSA1() *checkSA1 {
	return &checkSA1{
		BaseCheck: BaseCheck{
			CheckID:             "SA-1",
			CheckName:           "Confirmation bypass for destructive commands",
			CheckCategory:       CatAutomationSafety,
			CheckSeverity:       Warn,
			CheckMethod:         Passive,
			CheckRecommendation: "Add a --yes or --force flag to destructive commands so agents can skip interactive confirmation.",
		},
	}
}

func (c *checkSA1) Run(ctx context.Context, input *Input) *Result {
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

// SA-5: Idempotency indicators

type checkSA5 struct{ BaseCheck }

func newCheckSA5() *checkSA5 {
	return &checkSA5{BaseCheck{
		CheckID:             "SA-5",
		CheckName:           "Idempotency indicators",
		CheckCategory:       CatAutomationSafety,
		CheckSeverity:       Info,
		CheckMethod:         Passive,
		CheckRecommendation: "Support idempotent operations via flags like `--if-not-exists`, `--upsert`, or `--create-or-update` so agent retries don't create duplicates.",
	}}
}

var idempotencyFlagNames = []string{
	"if-not-exists", "if-exists", "upsert", "idempotency-key",
	"create-or-update", "replace", "force-create",
}
var idempotencyHelpTerms = []string{"idempotent", "if already exists", "no-op if", "already exists"}
var createCommandNames = []string{"create", "add", "new", "insert", "put"}

func (c *checkSA5) Run(ctx context.Context, input *Input) *Result {
	idx := input.GetIndex()
	if idx == nil {
		return SkipResult(c, "no command index available")
	}

	mutating := idx.Mutating()
	if len(mutating) == 0 {
		return PassResult(c, "no mutating commands detected; idempotency not applicable")
	}

	// Check for idempotency flags
	if idx.HasFlag(idempotencyFlagNames...) {
		return PassResult(c, "found idempotency-related flag")
	}

	// Check for idempotency terms in help text
	if _, ok := idx.HelpContainsAny(idempotencyHelpTerms...); ok {
		return PassResult(c, "found idempotency reference in help text")
	}

	// Only fail if there are create-like commands (updates/deletes are often inherently idempotent)
	hasCreate := false
	for _, cmd := range mutating {
		for _, name := range createCommandNames {
			if strings.EqualFold(cmd.Name, name) {
				hasCreate = true
				break
			}
		}
	}
	if !hasCreate {
		return PassResult(c, "no create-like commands detected; mutating commands may be inherently idempotent")
	}

	return FailResult(c, "create-like commands found but no idempotency flags or help text detected")
}

// SA-6: Read vs write command separation

type checkSA6 struct{ BaseCheck }

func newCheckSA6() *checkSA6 {
	return &checkSA6{BaseCheck{
		CheckID:             "SA-6",
		CheckName:           "Read/write command separation",
		CheckCategory:       CatAutomationSafety,
		CheckSeverity:       Info,
		CheckMethod:         Passive,
		CheckRecommendation: "Clearly separate read-only commands (get, list, describe) from mutating ones (create, delete, update) so agent frameworks can apply different approval policies.",
	}}
}

func (c *checkSA6) Run(ctx context.Context, input *Input) *Result {
	idx := input.GetIndex()
	if idx == nil {
		return SkipResult(c, "no command index available")
	}

	readOnly := idx.ListLike()
	mutating := idx.Mutating()

	if len(readOnly) == 0 && len(mutating) == 0 {
		return PassResult(c, "no classifiable commands detected")
	}

	if len(readOnly) == 0 && len(mutating) > 0 {
		return FailResult(c, fmt.Sprintf("found %d mutating command(s) but no read-only commands (list, get, describe)", len(mutating)))
	}

	if len(mutating) == 0 && len(readOnly) > 0 {
		return PassResult(c, fmt.Sprintf("all %d classifiable command(s) are read-only", len(readOnly)))
	}

	return PassResult(c, fmt.Sprintf("clear separation: %d read-only and %d mutating command(s)",
		len(readOnly), len(mutating)))
}
