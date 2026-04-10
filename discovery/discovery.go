package discovery

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/Camil-H/cli-agent-lint/probe"
)

// Maximum parallel --help invocations during subcommand discovery.
const defaultDiscoveryConcurrency = 8

type Flag struct {
	Name        string
	ShortName   string
	Description string
	ValueType   string   // e.g. "string", "bool", "int"
	EnumValues  []string // e.g. ["json", "text"]
	Required    bool
}

type Command struct {
	Name           string
	FullPath       []string // e.g. ["git", "remote", "add"]
	Description    string
	RawHelp        string
	Flags          []*Flag
	InheritedFlags []*Flag
	Subcommands    []*Command
	IsMutating     bool
	IsDestructive  bool
	IsListLike     bool
	IsReadOnly     bool
}

// HasFlag returns true if the command has a flag matching any of the given
// names (long without --, or short without -).
func (c *Command) HasFlag(names ...string) bool {
	return c.FindFlag(names...) != nil
}

func (c *Command) FindFlag(names ...string) *Flag {
	for _, f := range c.Flags {
		for _, n := range names {
			if f.Name == n || f.ShortName == n {
				return f
			}
		}
	}
	for _, f := range c.InheritedFlags {
		for _, n := range names {
			if f.Name == n || f.ShortName == n {
				return f
			}
		}
	}
	return nil
}

// Walk calls fn for this command and all descendants depth-first.
func (c *Command) Walk(fn func(*Command)) {
	fn(c)
	for _, sub := range c.Subcommands {
		sub.Walk(fn)
	}
}

func (c *Command) AllCommands() []*Command {
	var result []*Command
	c.Walk(func(cmd *Command) {
		result = append(result, cmd)
	})
	return result
}

type CommandTree struct {
	Root            *Command
	TargetPath      string
	DiscoveryErrors []error
}

type DiscoverOpts struct {
	MaxDepth    int      // default 5
	MaxCommands int      // max total subcommands to discover; 0 defaults to 1000
	Subcommands []string // only discover these top-level subcommands
}

// Discover builds a CommandTree by recursively running --help on the target.
// defaultMaxCommands is the maximum number of subcommands to discover.
const defaultMaxCommands = 1000

func Discover(ctx context.Context, p *probe.Prober, opts DiscoverOpts) (*CommandTree, error) {
	if opts.MaxDepth <= 0 {
		opts.MaxDepth = 5
	}
	if opts.MaxCommands <= 0 {
		opts.MaxCommands = defaultMaxCommands
	}

	tree := &CommandTree{
		TargetPath: p.TargetPath(),
	}

	result, err := p.RunHelp(ctx)
	if err != nil {
		return nil, fmt.Errorf("running --help on target: %w", err)
	}

	helpText := result.StdoutStr()
	if helpText == "" {
		helpText = result.StderrStr()
	}

	rootName := filepath.Base(p.TargetPath())
	root := ParseHelpOutput(helpText, rootName)
	root.FullPath = []string{rootName}
	root.RawHelp = helpText

	if len(opts.Subcommands) > 0 {
		filtered := make([]*Command, 0)
		for _, sub := range root.Subcommands {
			for _, name := range opts.Subcommands {
				if sub.Name == name {
					filtered = append(filtered, sub)
					break
				}
			}
		}
		root.Subcommands = filtered
	}

	tree.Root = root
	var cmdCount atomic.Int32
	discoverSubcommands(ctx, p, root, 1, opts.MaxDepth, opts.MaxCommands, &cmdCount, tree)

	return tree, nil
}

func discoverSubcommands(ctx context.Context, p *probe.Prober, parent *Command, depth, maxDepth, maxCommands int, cmdCount *atomic.Int32, tree *CommandTree) {
	if depth >= maxDepth {
		return
	}

	type subResult struct {
		sub      *Command
		helpText string
		parsed   *Command
		err      error
	}

	// Skip help/completion subcommands.
	var toProcess []*Command
	for _, sub := range parent.Subcommands {
		sub.FullPath = append(append([]string{}, parent.FullPath...), sub.Name)
		if sub.Name == "help" || sub.Name == "completion" || sub.Name == "completions" {
			continue
		}
		toProcess = append(toProcess, sub)
	}

	results := make([]subResult, len(toProcess))
	var wg sync.WaitGroup
	sem := make(chan struct{}, defaultDiscoveryConcurrency)

	for i, sub := range toProcess {
		// Check command count limit before launching each goroutine.
		if int(cmdCount.Add(1)) > maxCommands {
			tree.DiscoveryErrors = append(tree.DiscoveryErrors,
				fmt.Errorf("discovery limit reached: exceeded %d commands", maxCommands))
			break
		}

		wg.Add(1)
		go func(i int, sub *Command) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			args := make([]string, 0, len(sub.FullPath))
			args = append(args, sub.FullPath[1:]...)
			result, err := p.RunHelp(ctx, args...)
			if err != nil {
				results[i] = subResult{sub: sub, err: fmt.Errorf("help %s: %w", strings.Join(sub.FullPath, " "), err)}
				return
			}

			helpText := result.StdoutStr()
			if helpText == "" {
				helpText = result.StderrStr()
			}
			parsed := ParseHelpOutput(helpText, sub.Name)
			results[i] = subResult{sub: sub, helpText: helpText, parsed: parsed}
		}(i, sub)
	}
	wg.Wait()

	// Apply results sequentially (safe mutation of tree).
	for _, r := range results {
		if r.err != nil {
			tree.DiscoveryErrors = append(tree.DiscoveryErrors, r.err)
			continue
		}
		if r.parsed == nil {
			continue
		}
		r.sub.RawHelp = r.helpText
		r.sub.Flags = mergeFlags(r.sub.Flags, r.parsed.Flags)
		r.sub.InheritedFlags = r.parsed.InheritedFlags
		r.sub.Subcommands = r.parsed.Subcommands
		r.sub.Description = firstNonEmpty(r.sub.Description, r.parsed.Description)
		classifyCommand(r.sub)
	}

	// Recurse sequentially — depth adds complexity to parallelize safely.
	for _, r := range results {
		if r.err == nil && r.parsed != nil {
			discoverSubcommands(ctx, p, r.sub, depth+1, maxDepth, maxCommands, cmdCount, tree)
		}
	}
}

// ParseHelpOutput parses CLI help text and extracts commands and flags.
// Handles Cobra, Click/Typer, argparse, and Clap formats.
func ParseHelpOutput(helpText string, name string) *Command {
	cmd := &Command{Name: name}
	lines := strings.Split(helpText, "\n")

	cmd.Description = extractDescription(lines)

	var currentSection string
	var sectionLines []string

	flushSection := func() {
		switch {
		case isCommandSection(currentSection):
			cmd.Subcommands = append(cmd.Subcommands, parseCommandLines(sectionLines)...)
		case isFlagSection(currentSection):
			inherited := strings.Contains(strings.ToLower(currentSection), "global") ||
				strings.Contains(strings.ToLower(currentSection), "inherited") ||
				strings.Contains(strings.ToLower(currentSection), "persistent")
			flags := parseFlagLines(sectionLines)
			if inherited {
				cmd.InheritedFlags = append(cmd.InheritedFlags, flags...)
			} else {
				cmd.Flags = append(cmd.Flags, flags...)
			}
		case isSubcommandBraceSection(currentSection):
			cmd.Subcommands = append(cmd.Subcommands, parseBraceSubcommands(currentSection)...)
		}
		sectionLines = nil
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if isSectionHeader(trimmed) {
			flushSection()
			currentSection = trimmed
			continue
		}

		if currentSection != "" {
			if trimmed == "" {
				continue
			}
			sectionLines = append(sectionLines, line)
		}
	}
	flushSection()

	cmd.Flags = append(cmd.Flags, parseUsageFlags(lines)...)
	cmd.Flags = deduplicateFlags(cmd.Flags)

	classifyCommand(cmd)
	return cmd
}

// --- Section detection ---

var sectionHeaderRe = regexp.MustCompile(`^[A-Z][A-Za-z /\-]+:?\s*$`)

func isSectionHeader(line string) bool {
	if line == "" {
		return false
	}
	// Must end with ":" or be all-caps style.
	if strings.HasSuffix(line, ":") || sectionHeaderRe.MatchString(line) {
		return true
	}
	return false
}

func isCommandSection(header string) bool {
	h := strings.ToLower(strings.TrimRight(header, ": "))
	for _, kw := range []string{"available commands", "commands", "subcommands", "positional arguments"} {
		if strings.Contains(h, kw) {
			return true
		}
	}
	return false
}

func isFlagSection(header string) bool {
	h := strings.ToLower(strings.TrimRight(header, ": "))
	for _, kw := range []string{"flags", "options", "optional arguments", "global flags", "inherited flags", "persistent flags"} {
		if strings.Contains(h, kw) {
			return true
		}
	}
	return false
}

func isSubcommandBraceSection(header string) bool {
	return strings.Contains(header, "{") && strings.Contains(header, "}")
}

// --- Command parsing ---

var commandLineRe = regexp.MustCompile(`^\s{2,}(\S+)\s+(.*)$`)

func parseCommandLines(lines []string) []*Command {
	var cmds []*Command
	for _, line := range lines {
		m := commandLineRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		name := strings.TrimRight(m[1], ":")
		desc := strings.TrimSpace(m[2])
		if strings.HasPrefix(name, "-") {
			continue
		}
		cmds = append(cmds, &Command{
			Name:        name,
			Description: desc,
		})
	}
	return cmds
}

func parseBraceSubcommands(header string) []*Command {
	start := strings.Index(header, "{")
	end := strings.Index(header, "}")
	if start < 0 || end < 0 || end <= start {
		return nil
	}
	inner := header[start+1 : end]
	parts := strings.Split(inner, ",")
	var cmds []*Command
	for _, p := range parts {
		name := strings.TrimSpace(p)
		if name != "" {
			cmds = append(cmds, &Command{Name: name})
		}
	}
	return cmds
}

// --- Flag parsing ---

var (
	// Matches: -f, --flag, -f/--flag, --flag=<value>, --flag VALUE, etc.
	flagLineRe = regexp.MustCompile(`^\s+((-([a-zA-Z0-9]),?\s+)?--([a-zA-Z0-9][a-zA-Z0-9\-_]*))(?:[=\s]\s*[<\[]?([a-zA-Z0-9_\-|{}]+)[>\]]?)?\s+(.*)$`)
)

func parseFlagLines(lines []string) []*Flag {
	var flags []*Flag
	for _, line := range lines {
		f := parseFlagLine(line)
		if f != nil {
			flags = append(flags, f)
		}
	}
	return flags
}

func parseFlagLine(line string) *Flag {
	m := flagLineRe.FindStringSubmatch(line)
	if m != nil {
		f := &Flag{
			ShortName:   m[3],
			Name:        m[4],
			Description: strings.TrimSpace(m[6]),
		}
		if m[5] != "" {
			f.ValueType = m[5]
			f.EnumValues = parseEnumValues(m[5])
		}
		return f
	}

	// Fallback: more flexible matching for --flag with optional description.
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, "-") {
		return nil
	}

	parts := splitFlagLine(trimmed)
	if parts == nil {
		return nil
	}
	return parts
}

func splitFlagLine(line string) *Flag {
	f := &Flag{}
	rest := line

	// Parse short flag.
	if strings.HasPrefix(rest, "-") && !strings.HasPrefix(rest, "--") {
		if len(rest) >= 2 && rest[1] != '-' {
			f.ShortName = string(rest[1])
			rest = rest[2:]
			rest = strings.TrimLeft(rest, ", ")
		}
	}

	// Parse long flag.
	if strings.HasPrefix(rest, "--") {
		rest = rest[2:]
		endIdx := strings.IndexAny(rest, " =\t")
		if endIdx < 0 {
			f.Name = rest
			return f
		}
		f.Name = rest[:endIdx]
		rest = rest[endIdx:]
		rest = strings.TrimLeft(rest, "= ")

		// Check for value hint.
		if strings.HasPrefix(rest, "<") || strings.HasPrefix(rest, "[") {
			closeIdx := strings.IndexAny(rest, ">]")
			if closeIdx > 0 {
				f.ValueType = rest[1:closeIdx]
				rest = rest[closeIdx+1:]
			}
		} else if len(rest) > 0 {
			spaceIdx := strings.IndexAny(rest, " \t")
			if spaceIdx > 0 {
				candidate := rest[:spaceIdx]
				if looksLikeValueType(candidate) {
					f.ValueType = candidate
					rest = rest[spaceIdx:]
				}
			}
		}
	}

	if f.Name == "" && f.ShortName == "" {
		return nil
	}

	f.Description = strings.TrimSpace(rest)
	f.EnumValues = parseEnumFromDesc(f.Description)
	if len(f.EnumValues) == 0 && f.ValueType != "" {
		f.EnumValues = parseEnumValues(f.ValueType)
	}

	return f
}

func looksLikeValueType(s string) bool {
	upper := strings.ToUpper(s)
	for _, t := range []string{"STRING", "INT", "BOOL", "FLOAT", "DURATION", "FILE", "PATH", "URL", "FORMAT"} {
		if upper == t {
			return true
		}
	}
	return false
}

var enumRe = regexp.MustCompile(`\{([^}]+)\}`)
var usageFlagRe = regexp.MustCompile(`--([a-zA-Z0-9][a-zA-Z0-9\-_]*)`)

func parseEnumValues(s string) []string {
	m := enumRe.FindStringSubmatch(s)
	if m == nil {
		return nil
	}
	parts := strings.Split(m[1], ",")
	var vals []string
	for _, p := range parts {
		v := strings.TrimSpace(p)
		if v != "" {
			vals = append(vals, v)
		}
	}
	return vals
}

func parseEnumFromDesc(desc string) []string {
	return parseEnumValues(desc)
}

func parseUsageFlags(lines []string) []*Flag {
	var flags []*Flag
	for _, line := range lines {
		lower := strings.ToLower(strings.TrimSpace(line))
		if !strings.HasPrefix(lower, "usage:") && !strings.HasPrefix(lower, "usage :") {
			continue
		}
		matches := usageFlagRe.FindAllStringSubmatch(line, -1)
		for _, m := range matches {
			flags = append(flags, &Flag{Name: m[1]})
		}
		break
	}
	return flags
}

func deduplicateFlags(flags []*Flag) []*Flag {
	seen := make(map[string]bool)
	var result []*Flag
	for _, f := range flags {
		key := f.Name
		if key == "" {
			key = f.ShortName
		}
		if seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, f)
	}
	return result
}

// --- Description extraction ---

func extractDescription(lines []string) string {
	var desc []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			if len(desc) > 0 {
				break
			}
			continue
		}
		// Stop at section headers or usage lines.
		if isSectionHeader(trimmed) || strings.HasPrefix(strings.ToLower(trimmed), "usage") {
			break
		}
		desc = append(desc, trimmed)
	}
	return strings.Join(desc, " ")
}

// --- Command classification ---

var mutatingNames = regexp.MustCompile(`^(create|delete|update|set|remove|add|modify|push|put|post|patch|destroy|rm|write|drop|insert|apply|install|uninstall)$`)
var destructiveNames = regexp.MustCompile(`^(delete|destroy|rm|remove|drop|uninstall|purge)$`)
var listLikeNames = regexp.MustCompile(`^(list|ls|search|find|query)$`)
var readOnlyNames = regexp.MustCompile(`^(list|ls|search|find|query|get|show|view|describe|inspect|status|info|cat|read|dump|export|print|check|verify|validate|diff|log|history|whoami)$`)

func classifyCommand(cmd *Command) {
	lower := strings.ToLower(cmd.Name)
	cmd.IsMutating = mutatingNames.MatchString(lower)
	cmd.IsDestructive = destructiveNames.MatchString(lower)
	cmd.IsListLike = listLikeNames.MatchString(lower)
	cmd.IsReadOnly = readOnlyNames.MatchString(lower)
}

func mergeFlags(existing, parsed []*Flag) []*Flag {
	if len(existing) == 0 {
		return parsed
	}
	if len(parsed) == 0 {
		return existing
	}
	combined := make([]*Flag, 0, len(existing)+len(parsed))
	combined = append(combined, existing...)
	combined = append(combined, parsed...)
	return deduplicateFlags(combined)
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
