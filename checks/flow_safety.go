package checks

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/Camil-H/cli-agent-lint/discovery"
	"github.com/Camil-H/cli-agent-lint/probe"
)

// FS-1: Stderr vs stdout discipline

type checkFS1 struct {
	BaseCheck
}

func newCheckFS1() *checkFS1 {
	return &checkFS1{
		BaseCheck: BaseCheck{
			CheckID:             "FS-1",
			CheckName:           "Stderr vs stdout discipline",
			CheckCategory:       CatFlowSafety,
			CheckSeverity:       Fail,
			CheckMethod:         Active,
			CheckRecommendation: "Send errors and diagnostics to stderr. Reserve stdout for data output only.",
		},
	}
}

func (c *checkFS1) Run(ctx context.Context, input *Input) *Result {
	if r := skipIfNoProber(c, input); r != nil {
		return r
	}

	result, err := input.Prober.Run(ctx, probe.Opts{
		Args: []string{"__nonexistent_subcommand__"},
	})
	if err != nil {
		return ErrorResult(c, fmt.Errorf("running nonexistent subcommand: %w", err))
	}

	stdout := result.StdoutStr()
	stderr := result.StderrStr()

	if stdout != "" && stderr == "" {
		return FailResult(c, fmt.Sprintf("error text appeared on stdout only: %q", truncate(stdout, 200)))
	}

	if stdout != "" && stderr != "" {
		lower := strings.ToLower(stdout)
		if strings.Contains(lower, "error") || strings.Contains(lower, "unknown") ||
			strings.Contains(lower, "not found") || strings.Contains(lower, "invalid") ||
			strings.Contains(lower, "usage") {
			return FailResult(c, fmt.Sprintf("error/diagnostic text leaked to stdout: %q", truncate(stdout, 200)))
		}
	}

	if stderr != "" {
		return PassResult(c, "error output correctly sent to stderr")
	}

	return PassResult(c, "no output on stdout for invalid subcommand (exit code signals error)")
}

// FS-2: Non-TTY detection (no ANSI in pipes)

type checkFS2 struct {
	BaseCheck
}

func newCheckFS2() *checkFS2 {
	return &checkFS2{
		BaseCheck: BaseCheck{
			CheckID:             "FS-2",
			CheckName:           "Non-TTY detection (no ANSI in pipes)",
			CheckCategory:       CatFlowSafety,
			CheckSeverity:       Fail,
			CheckMethod:         Active,
			CheckRecommendation: "Detect non-TTY stdout and disable color/formatting automatically.",
		},
	}
}

func (c *checkFS2) Run(ctx context.Context, input *Input) *Result {
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

// FS-3: No interactive prompts in non-TTY

type checkFS3 struct {
	BaseCheck
}

func newCheckFS3() *checkFS3 {
	return &checkFS3{
		BaseCheck: BaseCheck{
			CheckID:             "FS-3",
			CheckName:           "No interactive prompts in non-TTY",
			CheckCategory:       CatFlowSafety,
			CheckSeverity:       Fail,
			CheckMethod:         Active,
			CheckRecommendation: "Never prompt for input when stdin is not a TTY. Fail fast with a clear error instead.",
		},
	}
}

var promptRe = regexp.MustCompile(`(?i)(enter\b|password:|press\b|continue\?|y/n|\(yes/no\))`)

func (c *checkFS3) Run(ctx context.Context, input *Input) *Result {
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

// FS-4: Env var auth support

type checkFS4 struct {
	BaseCheck
}

func newCheckFS4() *checkFS4 {
	return &checkFS4{
		BaseCheck: BaseCheck{
			CheckID:             "FS-4",
			CheckName:           "Env var auth support",
			CheckCategory:       CatFlowSafety,
			CheckSeverity:       Warn,
			CheckMethod:         Passive,
			CheckRecommendation: "Support authentication via environment variables for headless/agent usage.",
		},
	}
}

var envVarSuffixRe = regexp.MustCompile(`[A-Z][A-Z0-9_]*_(TOKEN|API_KEY|CREDENTIALS|SECRET|PASSWORD)\b`)

func hasAuthEnvVarMention(text string) bool {
	return envVarSuffixRe.MatchString(text)
}

func (c *checkFS4) Run(ctx context.Context, input *Input) *Result {
	idx := input.GetIndex()
	if idx == nil {
		return SkipResult(c, "no command tree available")
	}

	if f, cmd := idx.FindFlag(authTokenFlagNames...); f != nil {
		return PassResult(c, fmt.Sprintf("found auth flag --%s on command %q", f.Name, strings.Join(cmd.FullPath, " ")))
	}

	for _, cmd := range idx.All() {
		if hasAuthEnvVarMention(cmd.RawHelp) {
			match := envVarSuffixRe.FindString(cmd.RawHelp)
			return PassResult(c, fmt.Sprintf("found auth env var %s in help for %q", match, strings.Join(cmd.FullPath, " ")))
		}
	}

	for _, name := range []string{"auth", "login"} {
		for _, cmd := range idx.CommandsByName(name) {
			h := idx.LowerHelp(cmd)
			if strings.Contains(h, "token") || strings.Contains(h, "api_key") ||
				strings.Contains(h, "api-key") || strings.Contains(h, "env") {
				return PassResult(c, fmt.Sprintf("found token/env var reference in %q subcommand help", cmd.Name))
			}
		}
	}

	anyAuthMention := false
	if _, ok := idx.HelpContainsAny(authRelatedTerms...); ok {
		anyAuthMention = true
	}
	if !anyAuthMention {
		if len(idx.CommandsByName("auth")) > 0 || len(idx.CommandsByName("login")) > 0 {
			anyAuthMention = true
		}
	}

	if !anyAuthMention {
		return PassResult(c, "no auth-related commands or flags detected; auth not applicable")
	}

	return FailResult(c, "auth-related content found but no env var or token flag for non-interactive auth")
}

// FS-5: No mandatory interactive auth

type checkFS5 struct {
	BaseCheck
}

func newCheckFS5() *checkFS5 {
	return &checkFS5{
		BaseCheck: BaseCheck{
			CheckID:             "FS-5",
			CheckName:           "No mandatory interactive auth",
			CheckCategory:       CatFlowSafety,
			CheckSeverity:       Fail,
			CheckMethod:         Passive,
			CheckRecommendation: "Provide non-interactive auth paths (API keys, service account files, token env vars) alongside interactive flows.",
		},
	}
}

func findLoginCommand(idx *discovery.CommandIndex) *discovery.Command {
	for name := range loginCommandNames {
		if cmds := idx.CommandsByName(name); len(cmds) > 0 {
			return cmds[0]
		}
	}
	return nil
}

func hasNonInteractiveAlternative(idx *discovery.CommandIndex) (bool, string) {
	if f, cmd := idx.FindFlag(nonInteractiveAuthFlagNames...); f != nil {
		return true, fmt.Sprintf("found non-interactive auth flag --%s on command %q", f.Name, strings.Join(cmd.FullPath, " "))
	}

	for _, cmd := range idx.All() {
		if hasAuthEnvVarMention(cmd.RawHelp) {
			match := envVarSuffixRe.FindString(cmd.RawHelp)
			return true, fmt.Sprintf("found auth env var %s in help for %q", match, strings.Join(cmd.FullPath, " "))
		}
	}
	return false, ""
}

func (c *checkFS5) Run(ctx context.Context, input *Input) *Result {
	idx := input.GetIndex()
	if idx == nil {
		return SkipResult(c, "no command tree available")
	}

	loginCmd := findLoginCommand(idx)
	if loginCmd == nil {
		return PassResult(c, "no login/signin/sign-in command found; no mandatory interactive auth")
	}

	found, detail := hasNonInteractiveAlternative(idx)
	if found {
		return PassResult(c, fmt.Sprintf("login command %q exists but non-interactive alternative found: %s",
			strings.Join(loginCmd.FullPath, " "), detail))
	}

	return FailResult(c, fmt.Sprintf("login command %q exists with no non-interactive auth alternative (no token flags or auth env vars found)",
		strings.Join(loginCmd.FullPath, " ")))
}

// FS-6: Exit codes

type checkFS6 struct {
	BaseCheck
}

func newCheckFS6() *checkFS6 {
	return &checkFS6{
		BaseCheck: BaseCheck{
			CheckID:             "FS-6",
			CheckName:           "Exit codes",
			CheckCategory:       CatFlowSafety,
			CheckSeverity:       Fail,
			CheckMethod:         Active,
			CheckRecommendation: "Use distinct non-zero exit codes for different failure modes (bad input, auth failure, server error).",
		},
	}
}

func (c *checkFS6) Run(ctx context.Context, input *Input) *Result {
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
