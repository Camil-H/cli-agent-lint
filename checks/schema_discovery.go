package checks

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Camil-H/cli-agent-lint/discovery"
	"github.com/Camil-H/cli-agent-lint/probe"
)

// SD-3: Shell completions available

type checkSD3 struct {
	BaseCheck
}

func newCheckSD3() *checkSD3 {
	return &checkSD3{
		BaseCheck: BaseCheck{
			CheckID:             "SD-3",
			CheckName:           "Shell completions available",
			CheckCategory:       CatSelfDescribing,
			CheckSeverity:       Info,
			CheckMethod:         Passive,
			CheckRecommendation: "Ship shell completions (bash/zsh/fish) via a `completion` subcommand.",
		},
	}
}

func (c *checkSD3) Run(ctx context.Context, input *Input) *Result {
	idx := input.GetIndex()
	if idx == nil {
		return SkipResult(c, "no command tree available")
	}

	var found []string
	for _, name := range []string{"completion", "completions"} {
		for _, cmd := range idx.CommandsByName(name) {
			found = append(found, strings.Join(cmd.FullPath, " "))
		}
	}

	if idx.HasFlag("generate-completion", "completions") {
		found = append(found, "--generate-completion / --completions flag")
	}

	if len(found) > 0 {
		return PassResult(c, fmt.Sprintf("shell completion support found: %s", strings.Join(found, ", ")))
	}

	return FailResult(c, "no completion subcommand or flag found")
}

// SD-4: Schema / describe introspection

type checkSD4 struct {
	BaseCheck
}

func newCheckSD4() *checkSD4 {
	return &checkSD4{
		BaseCheck: BaseCheck{
			CheckID:             "SD-4",
			CheckName:           "Schema / describe introspection",
			CheckCategory:       CatSelfDescribing,
			CheckSeverity:       Warn,
			CheckMethod:         Passive,
			CheckRecommendation: "Expose command schemas at runtime so agents can discover parameters, types, and constraints.",
		},
	}
}

func (c *checkSD4) Run(ctx context.Context, input *Input) *Result {
	idx := input.GetIndex()
	if idx == nil {
		return SkipResult(c, "no command tree available")
	}

	var found []string

	for _, name := range []string{"schema", "describe", "api", "checks", "inspect"} {
		for _, cmd := range idx.CommandsByName(name) {
			found = append(found, strings.Join(cmd.FullPath, " "))
		}
	}

	for _, hit := range idx.FindFlagAll("describe", "schema") {
		found = append(found, fmt.Sprintf("%s has --describe/--schema flag", strings.Join(hit.Cmd.FullPath, " ")))
	}

	for _, cmd := range idx.All() {
		h := idx.LowerHelp(cmd)
		if strings.Contains(h, "schema") || strings.Contains(h, "introspection") {
			found = append(found, fmt.Sprintf("%s help mentions schema/introspection", strings.Join(cmd.FullPath, " ")))
		}
	}

	if len(found) > 0 {
		seen := make(map[string]bool)
		var unique []string
		for _, f := range found {
			if !seen[f] {
				seen[f] = true
				unique = append(unique, f)
			}
		}
		return PassResult(c, fmt.Sprintf("introspection support found: %s", strings.Join(unique, "; ")))
	}

	return FailResult(c, "no schema/describe introspection subcommands or flags found")
}

// SD-5: Skill / context files

type checkSD5 struct {
	BaseCheck
}

func newCheckSD5() *checkSD5 {
	return &checkSD5{
		BaseCheck: BaseCheck{
			CheckID:             "SD-5",
			CheckName:           "Skill / context files",
			CheckCategory:       CatSelfDescribing,
			CheckSeverity:       Info,
			CheckMethod:         Passive,
			CheckRecommendation: "Ship skill/context files (AGENTS.md, CONTEXT.md) encoding invariants agents cannot intuit.",
		},
	}
}

func (c *checkSD5) Run(ctx context.Context, input *Input) *Result {
	if input.Tree == nil {
		return SkipResult(c, "no command tree available")
	}

	targetPath := input.Tree.TargetPath
	if targetPath == "" {
		return SkipResult(c, "no target path available")
	}

	targetDir := filepath.Dir(targetPath)
	parentDir := filepath.Dir(targetDir)
	searchDirs := []string{targetDir}
	if parentDir != targetDir {
		searchDirs = append(searchDirs, parentDir)
	}

	candidates := []string{
		"AGENTS.md",
		"CONTEXT.md",
		"CLAUDE.md",
		"llms.txt",
		".claude",
		".mcp.json",
		"mcp.json",
	}

	var found []string
	for _, dir := range searchDirs {
		for _, name := range candidates {
			path := filepath.Join(dir, name)
			if _, err := os.Stat(path); err == nil {
				found = append(found, path)
			}
		}
	}

	var helpMentions []string
	if input.Tree.Root != nil {
		rawHelp := input.Tree.Root.RawHelp
		lowerHelp := strings.ToLower(rawHelp)
		for _, name := range candidates {
			if strings.Contains(lowerHelp, strings.ToLower(name)) {
				helpMentions = append(helpMentions, name)
			}
		}
	}

	if len(found) > 0 || len(helpMentions) > 0 {
		var details []string
		if len(found) > 0 {
			details = append(details, fmt.Sprintf("files found: %s", strings.Join(found, ", ")))
		}
		if len(helpMentions) > 0 {
			details = append(details, fmt.Sprintf("mentioned in help: %s", strings.Join(helpMentions, ", ")))
		}
		return PassResult(c, strings.Join(details, "; "))
	}

	return FailResult(c, "no skill/context files found near the binary")
}

// SD-6: Help text with usage examples

type checkSD6 struct {
	BaseCheck
}

func newCheckSD6() *checkSD6 {
	return &checkSD6{
		BaseCheck: BaseCheck{
			CheckID:             "SD-6",
			CheckName:           "Help text with usage examples",
			CheckCategory:       CatSelfDescribing,
			CheckSeverity:       Warn,
			CheckMethod:         Passive,
			CheckRecommendation: "Include an Examples section in --help output so agents can learn correct usage patterns.",
		},
	}
}

var exampleSectionRe = regexp.MustCompile(`(?m)^(?:Examples?|EXAMPLES?|Usage [Ee]xamples?):\s*$`)

func (c *checkSD6) Run(ctx context.Context, input *Input) *Result {
	idx := input.GetIndex()
	if idx == nil {
		return SkipResult(c, "no command tree available")
	}

	for _, cmd := range idx.All() {
		if exampleSectionRe.MatchString(cmd.RawHelp) {
			return PassResult(c, fmt.Sprintf("found examples section in %q help text", strings.Join(cmd.FullPath, " ")))
		}
	}

	return FailResult(c, "no Examples section found in any command's help text")
}

// SD-7: Actionable error messages

type checkSD7 struct {
	BaseCheck
}

func newCheckSD7() *checkSD7 {
	return &checkSD7{
		BaseCheck: BaseCheck{
			CheckID:             "SD-7",
			CheckName:           "Actionable error messages",
			CheckCategory:       CatSelfDescribing,
			CheckSeverity:       Warn,
			CheckMethod:         Active,
			CheckRecommendation: "Include a suggested fix or next step in error messages (e.g., \"Did you mean ...?\", \"Try ...\", \"Run ... to ...\").",
		},
	}
}

var actionablePatterns = regexp.MustCompile(`(?i)(did you mean|try |run |use |see |hint:|suggestion:|usage:|available commands|valid|possible)`)

func (c *checkSD7) Run(ctx context.Context, input *Input) *Result {
	if r := skipIfNoProber(c, input); r != nil {
		return r
	}

	result, err := input.Prober.Run(ctx, probe.Opts{
		Args: []string{"__nonexistent_subcommand__"},
	})
	if err != nil {
		return ErrorResult(c, fmt.Errorf("running nonexistent subcommand: %w", err))
	}

	combined := result.StdoutStr() + "\n" + result.StderrStr()
	if strings.TrimSpace(combined) == "" {
		return FailResult(c, "no output for invalid subcommand — agent gets no guidance on what went wrong")
	}

	if actionablePatterns.MatchString(combined) {
		return PassResult(c, "error output contains actionable guidance")
	}

	return FailResult(c, fmt.Sprintf("error output lacks actionable guidance: %q", truncate(strings.TrimSpace(combined), 200)))
}

// SD-8: Subcommand fan-out

const maxSubcommandsPerLevel = 15

type checkSD8 struct {
	BaseCheck
}

func newCheckSD8() *checkSD8 {
	return &checkSD8{
		BaseCheck: BaseCheck{
			CheckID:             "SD-8",
			CheckName:           "Subcommand fan-out",
			CheckCategory:       CatSelfDescribing,
			CheckSeverity:       Info,
			CheckMethod:         Passive,
			CheckRecommendation: "Group subcommands into categories or namespaces. Agents struggle to select the right command when more than ~15 are listed at one level.",
		},
	}
}

func (c *checkSD8) Run(ctx context.Context, input *Input) *Result {
	if input.Tree == nil || input.Tree.Root == nil {
		return SkipResult(c, "no command tree available")
	}

	var worst *discovery.Command
	worstCount := 0

	input.Tree.Root.Walk(func(cmd *discovery.Command) {
		n := len(cmd.Subcommands)
		if n > worstCount {
			worstCount = n
			worst = cmd
		}
	})

	if worstCount == 0 {
		return PassResult(c, "no subcommands detected")
	}

	if worstCount > maxSubcommandsPerLevel {
		return FailResult(c, fmt.Sprintf("%q has %d subcommands (threshold: %d)",
			strings.Join(worst.FullPath, " "), worstCount, maxSubcommandsPerLevel))
	}

	return PassResult(c, fmt.Sprintf("max subcommand fan-out is %d at %q",
		worstCount, strings.Join(worst.FullPath, " ")))
}
