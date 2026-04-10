package checks

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/Camil-H/cli-agent-lint/discovery"
)

// TE-1: JSON output support

type checkTE1 struct {
	BaseCheck
}

func newCheckTE1() *checkTE1 {
	return &checkTE1{
		BaseCheck: BaseCheck{
			CheckID:             "TE-1",
			CheckName:           "JSON output support",
			CheckCategory:       CatTokenEfficiency,
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

func (c *checkTE1) Run(ctx context.Context, input *Input) *Result {
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

// TE-2: Stdin/pipe input support

type checkTE2 struct {
	BaseCheck
}

func newCheckTE2() *checkTE2 {
	return &checkTE2{
		BaseCheck: BaseCheck{
			CheckID:             "TE-2",
			CheckName:           "Stdin/pipe input support",
			CheckCategory:       CatTokenEfficiency,
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

func (c *checkTE2) Run(ctx context.Context, input *Input) *Result {
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

// TE-3: --no-color flag

type checkTE3 struct {
	BaseCheck
}

func newCheckTE3() *checkTE3 {
	return &checkTE3{
		BaseCheck: BaseCheck{
			CheckID:             "TE-3",
			CheckName:           "--no-color flag",
			CheckCategory:       CatTokenEfficiency,
			CheckSeverity:       Warn,
			CheckMethod:         Passive,
			CheckRecommendation: "Support `--no-color` flag and/or the `NO_COLOR` env var (see https://no-color.org).",
		},
	}
}

var noColorHelpRe = regexp.MustCompile(`(?i)(--no-color|--color[= ]never|NO_COLOR)`)

func (c *checkTE3) Run(ctx context.Context, input *Input) *Result {
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

// TE-4: --quiet / --silent flag

type checkTE4 struct {
	BaseCheck
}

func newCheckTE4() *checkTE4 {
	return &checkTE4{
		BaseCheck: BaseCheck{
			CheckID:             "TE-4",
			CheckName:           "--quiet / --silent flag",
			CheckCategory:       CatTokenEfficiency,
			CheckSeverity:       Info,
			CheckMethod:         Passive,
			CheckRecommendation: "Add `--quiet` flag to suppress informational output, leaving only essential data.",
		},
	}
}

var quietHelpRe = regexp.MustCompile(`(?i)(--quiet|--silent|-q\b)`)

func (c *checkTE4) Run(ctx context.Context, input *Input) *Result {
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

// TE-5: Pagination support

type checkTE5 struct {
	BaseCheck
}

func newCheckTE5() *checkTE5 {
	return &checkTE5{
		BaseCheck: BaseCheck{
			CheckID:             "TE-5",
			CheckName:           "Pagination support",
			CheckCategory:       CatTokenEfficiency,
			CheckSeverity:       Warn,
			CheckMethod:         Passive,
			CheckRecommendation: "Support `--page-all` or NDJSON streaming for list commands to avoid silent truncation.",
		},
	}
}

func (c *checkTE5) Run(ctx context.Context, input *Input) *Result {
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

// TE-6: Field masks / response filtering

type checkTE6 struct {
	BaseCheck
}

func newCheckTE6() *checkTE6 {
	return &checkTE6{
		BaseCheck: BaseCheck{
			CheckID:             "TE-6",
			CheckName:           "Field masks / response filtering",
			CheckCategory:       CatTokenEfficiency,
			CheckSeverity:       Info,
			CheckMethod:         Passive,
			CheckRecommendation: "Support field masks or response filtering to limit output size and protect agent context windows.",
		},
	}
}

func (c *checkTE6) Run(ctx context.Context, input *Input) *Result {
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

// TE-7: Help output size

const (
	helpSizeWarnBytes = 40 * 1024  // 40 KB
	helpSizeFailBytes = 100 * 1024 // 100 KB
)

type checkTE7 struct {
	BaseCheck
}

func newCheckTE7() *checkTE7 {
	return &checkTE7{
		BaseCheck: BaseCheck{
			CheckID:             "TE-7",
			CheckName:           "Help output size",
			CheckCategory:       CatTokenEfficiency,
			CheckSeverity:       Fail,
			CheckMethod:         Passive,
			CheckRecommendation: "Keep `--help` output concise. Large help text fills agent context windows and degrades performance.",
		},
	}
}

func (c *checkTE7) Run(ctx context.Context, input *Input) *Result {
	if input.Tree == nil || input.Tree.Root == nil {
		return SkipResult(c, "no command tree available")
	}

	size := len(input.Tree.Root.RawHelp)
	detail := fmt.Sprintf("root --help output is %s", formatBytes(size))

	if size >= helpSizeFailBytes {
		return FailResult(c, detail)
	}
	if size >= helpSizeWarnBytes {
		return FailResult(c, detail)
	}
	return PassResult(c, detail)
}

func formatBytes(n int) string {
	switch {
	case n >= 1024*1024:
		return fmt.Sprintf("%.1f MB", float64(n)/(1024*1024))
	case n >= 1024:
		return fmt.Sprintf("%.1f KB", float64(n)/1024)
	default:
		return fmt.Sprintf("%d bytes", n)
	}
}

// TE-8: Concise output mode

type checkTE8 struct {
	BaseCheck
}

func newCheckTE8() *checkTE8 {
	return &checkTE8{
		BaseCheck: BaseCheck{
			CheckID:             "TE-8",
			CheckName:           "Concise output mode",
			CheckCategory:       CatTokenEfficiency,
			CheckSeverity:       Warn,
			CheckMethod:         Passive,
			CheckRecommendation: "Add a `--brief` or `--concise` flag that returns essential data without decorative formatting or verbose detail.",
		},
	}
}

var conciseFlagNames = []string{"brief", "concise", "short", "summary", "compact", "terse"}

func (c *checkTE8) Run(ctx context.Context, input *Input) *Result {
	idx := input.GetIndex()
	if idx == nil {
		return SkipResult(c, "no command tree available")
	}

	if idx.HasFlag(conciseFlagNames...) {
		return PassResult(c, "found concise output flag (e.g. --brief, --concise, --short)")
	}

	// Check for format enum values like --output=short or --format=brief
	for _, hit := range idx.FindFlagAll(jsonOutputFlagNames...) {
		for _, v := range hit.Flag.EnumValues {
			lower := strings.ToLower(v)
			for _, name := range conciseFlagNames {
				if lower == name {
					return PassResult(c, fmt.Sprintf("found concise format value %q on --%s flag", v, hit.Flag.Name))
				}
			}
		}
	}

	// No data-producing commands means concise mode isn't applicable
	if len(idx.ListLike()) == 0 && len(idx.Mutating()) == 0 {
		return PassResult(c, "no data-producing commands detected; concise mode not applicable")
	}

	return FailResult(c, "no --brief, --concise, or --short flag found")
}
