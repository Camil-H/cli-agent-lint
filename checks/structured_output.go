package checks

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/cli-agent-lint/cli-agent-lint/discovery"
	"github.com/cli-agent-lint/cli-agent-lint/probe"
)

// ---------------------------------------------------------------------------
// SO-1: JSON output support
// ---------------------------------------------------------------------------

type checkSO1 struct {
	BaseCheck
}

func newCheckSO1() *checkSO1 {
	return &checkSO1{
		BaseCheck: BaseCheck{
			CheckID:             "SO-1",
			CheckName:           "JSON output support",
			CheckCategory:       CatStructuredOutput,
			CheckSeverity:       Fail,
			CheckMethod:         Passive,
			CheckRecommendation: "Add `--output json` flag to all commands that produce output.",
		},
	}
}

func findJSONOutputFlag(idx *discovery.CommandIndex) (*discovery.Flag, *discovery.Command) {
	for _, hit := range idx.FindFlagAll(jsonOutputFlagNames...) {
		f := hit.Flag
		for _, v := range f.EnumValues {
			if strings.EqualFold(v, "json") {
				return f, hit.Cmd
			}
		}
		if strings.Contains(strings.ToLower(f.Description), "json") {
			return f, hit.Cmd
		}
		if f.Name == "json" {
			return f, hit.Cmd
		}
	}
	return nil, nil
}

func (c *checkSO1) Run(ctx context.Context, input *Input) *Result {
	idx := input.GetIndex()
	if idx == nil {
		return SkipResult(c, "no command tree available")
	}

	flag, cmd := findJSONOutputFlag(idx)
	if flag != nil {
		detail := fmt.Sprintf("found flag --%s on command %q", flag.Name, strings.Join(cmd.FullPath, " "))
		return PassResult(c, detail)
	}

	return FailResult(c, "no --output/--format/--json/-o flag with JSON support found in any command")
}

// ---------------------------------------------------------------------------
// SO-2: Stderr vs stdout discipline
// ---------------------------------------------------------------------------

type checkSO2 struct {
	BaseCheck
}

func newCheckSO2() *checkSO2 {
	return &checkSO2{
		BaseCheck: BaseCheck{
			CheckID:             "SO-2",
			CheckName:           "Stderr vs stdout discipline",
			CheckCategory:       CatStructuredOutput,
			CheckSeverity:       Fail,
			CheckMethod:         Active,
			CheckRecommendation: "Send errors and diagnostics to stderr. Reserve stdout for data output only.",
		},
	}
}

func (c *checkSO2) Run(ctx context.Context, input *Input) *Result {
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

// ---------------------------------------------------------------------------
// SO-3: Error format is structured
// ---------------------------------------------------------------------------

type checkSO3 struct {
	BaseCheck
}

func newCheckSO3() *checkSO3 {
	return &checkSO3{
		BaseCheck: BaseCheck{
			CheckID:             "SO-3",
			CheckName:           "Error format is structured",
			CheckCategory:       CatStructuredOutput,
			CheckSeverity:       Warn,
			CheckMethod:         Active,
			CheckRecommendation: "Emit structured JSON errors to stderr when `--output json` is set.",
		},
	}
}

func (c *checkSO3) Run(ctx context.Context, input *Input) *Result {
	if r := skipIfNoProber(c, input); r != nil {
		return r
	}

	// Cross-check dependency: SO-1 must have passed.
	so1Result := input.ResultSet.Get("SO-1")
	if so1Result == nil || so1Result.Status != StatusPass {
		return SkipResult(c, "skipped: no JSON output flag detected (SO-1 not passed)")
	}

	flag, _ := findJSONOutputFlag(input.GetIndex())
	if flag == nil {
		return SkipResult(c, "skipped: could not locate JSON output flag in command tree")
	}

	flagArg := buildJSONFlagArg(flag)

	args := []string{flagArg, "__nonexistent__"}
	argParts := strings.SplitN(flagArg, " ", 2)
	if len(argParts) == 2 {
		args = []string{argParts[0], argParts[1], "__nonexistent__"}
	}

	result, err := input.Prober.Run(ctx, probe.Opts{
		Args: args,
	})
	if err != nil {
		return ErrorResult(c, fmt.Errorf("running with JSON flag: %w", err))
	}

	stderr := result.StderrStr()
	if stderr == "" {
		return FailResult(c, "no stderr output when running with JSON flag and invalid subcommand")
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(stderr), &parsed); err != nil {
		return FailResult(c, fmt.Sprintf("stderr is not valid JSON: %s", truncate(stderr, 200)))
	}

	_, hasError := parsed["error"]
	_, hasMessage := parsed["message"]
	if hasError || hasMessage {
		return PassResult(c, "stderr contains structured JSON error with error/message field")
	}

	return FailResult(c, fmt.Sprintf("stderr is valid JSON but lacks 'error' or 'message' field: %s", truncate(stderr, 200)))
}

// buildJSONFlagArg returns the CLI argument to enable JSON output,
// e.g. "--output=json", "--format=json", or "--json".
func buildJSONFlagArg(f *discovery.Flag) string {
	switch f.Name {
	case "json":
		return "--json"
	default:
		return fmt.Sprintf("--%s=json", f.Name)
	}
}

// ---------------------------------------------------------------------------
// SO-4: Version output is parseable
// ---------------------------------------------------------------------------

type checkSO4 struct {
	BaseCheck
}

func newCheckSO4() *checkSO4 {
	return &checkSO4{
		BaseCheck: BaseCheck{
			CheckID:             "SO-4",
			CheckName:           "Version output is parseable",
			CheckCategory:       CatStructuredOutput,
			CheckSeverity:       Warn,
			CheckMethod:         Active,
			CheckRecommendation: "Emit a clean semver string from `--version`. Support `--output json` for `{\"version\": \"x.y.z\"}`.",
		},
	}
}

var semverRe = regexp.MustCompile(`v?(\d+\.\d+\.\d+)(?:[-+][a-zA-Z0-9.]+)*`)
var versionWordRe = regexp.MustCompile(`(?i)\bversion\b`)

func (c *checkSO4) Run(ctx context.Context, input *Input) *Result {
	if r := skipIfNoProber(c, input); r != nil {
		return r
	}

	result, err := input.Prober.Run(ctx, probe.Opts{
		Args: []string{"--version"},
	})
	if err != nil {
		return ErrorResult(c, fmt.Errorf("running --version: %w", err))
	}

	output := result.StdoutStr()
	if output == "" {
		output = result.StderrStr()
	}
	if output == "" {
		return FailResult(c, "--version produced no output")
	}

	match := semverRe.FindString(output)
	if match == "" {
		return FailResult(c, fmt.Sprintf("no semver pattern found in --version output: %q", truncate(output, 200)))
	}

	cleaned := output
	cleaned = strings.Replace(cleaned, match, "", 1)

	if input.Tree != nil && input.Tree.Root != nil {
		progName := input.Tree.Root.Name
		cleaned = strings.ReplaceAll(cleaned, progName, "")
	}

	cleaned = strings.TrimSpace(cleaned)
	cleaned = strings.Trim(cleaned, "/ \t\n\r")
	cleaned = versionWordRe.ReplaceAllString(cleaned, "")
	cleaned = strings.TrimSpace(cleaned)

	if cleaned == "" {
		return PassResult(c, fmt.Sprintf("clean version output: %s", match))
	}

	return FailResult(c, fmt.Sprintf("version output contains decorative text beyond semver: %q (extracted: %s)", truncate(output, 200), match))
}

// ---------------------------------------------------------------------------
// SO-5: Stdin/pipe input support
// ---------------------------------------------------------------------------

type checkSO5 struct {
	BaseCheck
}

func newCheckSO5() *checkSO5 {
	return &checkSO5{
		BaseCheck: BaseCheck{
			CheckID:             "SO-5",
			CheckName:           "Stdin/pipe input support",
			CheckCategory:       CatStructuredOutput,
			CheckSeverity:       Info,
			CheckMethod:         Passive,
			CheckRecommendation: "Support reading input from stdin or --from-file to enable composable pipelines between tools.",
		},
	}
}

func hasDataInputCommands(idx *discovery.CommandIndex) bool {
	if len(idx.Mutating()) > 0 {
		return true
	}
	return idx.HasFlag(dataInputFlagNames...)
}

func (c *checkSO5) Run(ctx context.Context, input *Input) *Result {
	idx := input.GetIndex()
	if idx == nil {
		return SkipResult(c, "no command tree available")
	}

	if idx.HasFlag(stdinFlagNames...) {
		return PassResult(c, "found stdin-related flag (e.g. --from-file, --input, --stdin)")
	}

	if _, ok := idx.HelpContainsAny(stdinHelpTerms...); ok {
		return PassResult(c, "found stdin/pipe reference in help text")
	}

	if !hasDataInputCommands(idx) {
		return PassResult(c, "no data-input commands detected; stdin not applicable")
	}

	return FailResult(c, "no stdin/pipe input support detected")
}

// ---------------------------------------------------------------------------
// Registration
// ---------------------------------------------------------------------------

func registerStructuredOutputChecks(r *Registry) {
	r.Register(newCheckSO1())
	r.Register(newCheckSO2())
	r.Register(newCheckSO3())
	r.Register(newCheckSO4())
	r.Register(newCheckSO5())
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}
