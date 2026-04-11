package discovery

import (
	"strings"
	"sync"
)

type FlagHit struct {
	Cmd  *Command
	Flag *Flag
}

// CommandIndex provides pre-computed lookups over a command tree,
// eliminating redundant full-tree walks across checks.
type CommandIndex struct {
	all         []*Command
	mutating    []*Command
	destructive []*Command
	listLike    []*Command
	readOnly    []*Command
	fileAccepting []*Command
	stringInput   []*Command
	flags         map[string][]FlagHit
	cmds          map[string][]*Command
	lowerHelp   map[*Command]string // lazily populated by LowerHelp()
	lowerHelpMu sync.RWMutex
	cmdFlags    map[*Command]map[string]bool
}

func NewIndex(root *Command) *CommandIndex {
	idx := &CommandIndex{
		flags:    make(map[string][]FlagHit),
		cmds:     make(map[string][]*Command),
		lowerHelp: make(map[*Command]string),
		cmdFlags: make(map[*Command]map[string]bool),
	}

	root.Walk(func(cmd *Command) {
		idx.all = append(idx.all, cmd)

		if cmd.IsMutating {
			idx.mutating = append(idx.mutating, cmd)
		}
		if cmd.IsDestructive {
			idx.destructive = append(idx.destructive, cmd)
		}
		if cmd.IsListLike {
			idx.listLike = append(idx.listLike, cmd)
		}
		if cmd.IsReadOnly {
			idx.readOnly = append(idx.readOnly, cmd)
		}

		// Pre-compute file-accepting and string-input classifications.
		if looksLikeFileCommand(cmd) || fileArgFlag(cmd) != "" {
			idx.fileAccepting = append(idx.fileAccepting, cmd)
		}
		if cmdFlag, _ := stringInputFlag(cmd); cmdFlag != "" {
			idx.stringInput = append(idx.stringInput, cmd)
		} else if len(cmd.Subcommands) == 0 && len(cmd.FullPath) > 1 {
			// Leaf command that might take positional arguments.
			idx.stringInput = append(idx.stringInput, cmd)
		}

		cmdNameLower := strings.ToLower(cmd.Name)
		idx.cmds[cmdNameLower] = append(idx.cmds[cmdNameLower], cmd)

		flagSet := make(map[string]bool, len(cmd.Flags)+len(cmd.InheritedFlags))
		for _, flags := range [][]*Flag{cmd.Flags, cmd.InheritedFlags} {
			for _, f := range flags {
				hit := FlagHit{Cmd: cmd, Flag: f}
				if f.Name != "" {
					nameLower := strings.ToLower(f.Name)
					idx.flags[nameLower] = append(idx.flags[nameLower], hit)
					flagSet[nameLower] = true
				}
				if f.ShortName != "" {
					shortLower := strings.ToLower(f.ShortName)
					idx.flags[shortLower] = append(idx.flags[shortLower], hit)
					flagSet[shortLower] = true
				}
			}
		}
		idx.cmdFlags[cmd] = flagSet
	})

	return idx
}

// looksLikeFileCommand returns true if the command name suggests file input.
func looksLikeFileCommand(cmd *Command) bool {
	lower := strings.ToLower(cmd.Name)
	for _, kw := range []string{"file", "open", "read", "load", "import"} {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

// fileArgFlag returns the flag name for a file-accepting flag, or "".
func fileArgFlag(cmd *Command) string {
	for _, flags := range [][]*Flag{cmd.Flags, cmd.InheritedFlags} {
		for _, f := range flags {
			nameLower := strings.ToLower(f.Name)
			descLower := strings.ToLower(f.Description)
			for _, kw := range []string{"file", "path", "dir"} {
				if strings.Contains(nameLower, kw) || strings.Contains(descLower, kw) {
					return f.Name
				}
			}
		}
	}
	return ""
}

// stringInputFlag returns the flag name for a string-typed flag, or "".
func stringInputFlag(cmd *Command) (string, string) {
	for _, flags := range [][]*Flag{cmd.Flags, cmd.InheritedFlags} {
		for _, f := range flags {
			vt := strings.ToLower(f.ValueType)
			if vt == "string" || vt == "str" || vt == "text" || vt == "name" || vt == "value" {
				return f.Name, f.Name
			}
		}
	}
	return "", ""
}

func (idx *CommandIndex) All() []*Command           { return idx.all }
func (idx *CommandIndex) Mutating() []*Command       { return idx.mutating }
func (idx *CommandIndex) Destructive() []*Command    { return idx.destructive }
func (idx *CommandIndex) ListLike() []*Command        { return idx.listLike }
func (idx *CommandIndex) ReadOnly() []*Command        { return idx.readOnly }
func (idx *CommandIndex) FileAccepting() []*Command   { return idx.fileAccepting }
func (idx *CommandIndex) StringInput() []*Command     { return idx.stringInput }

// FileArgFlag returns the flag name for a file-accepting flag on cmd, or "".
func (idx *CommandIndex) FileArgFlag(cmd *Command) string {
	return fileArgFlag(cmd)
}

// StringInputFlag returns the flag name for a string-typed flag on cmd, or "".
func (idx *CommandIndex) StringInputFlag(cmd *Command) string {
	name, _ := stringInputFlag(cmd)
	return name
}

// CmdHasFlag returns true if cmd has a flag matching any of the given names (O(1) per name).
func (idx *CommandIndex) CmdHasFlag(cmd *Command, names ...string) bool {
	fs := idx.cmdFlags[cmd]
	if fs == nil {
		return false
	}
	for _, n := range names {
		if fs[strings.ToLower(n)] {
			return true
		}
	}
	return false
}

func (idx *CommandIndex) HasFlag(names ...string) bool {
	for _, n := range names {
		if len(idx.flags[strings.ToLower(n)]) > 0 {
			return true
		}
	}
	return false
}

func (idx *CommandIndex) FindFlag(names ...string) (*Flag, *Command) {
	for _, n := range names {
		hits := idx.flags[strings.ToLower(n)]
		if len(hits) > 0 {
			return hits[0].Flag, hits[0].Cmd
		}
	}
	return nil, nil
}

func (idx *CommandIndex) FindFlagAll(names ...string) []FlagHit {
	var result []FlagHit
	seen := make(map[*Flag]bool)
	for _, n := range names {
		for _, hit := range idx.flags[strings.ToLower(n)] {
			if !seen[hit.Flag] {
				seen[hit.Flag] = true
				result = append(result, hit)
			}
		}
	}
	return result
}

// CommandsByName returns all commands matching the name (case-insensitive).
func (idx *CommandIndex) CommandsByName(name string) []*Command {
	return idx.cmds[strings.ToLower(name)]
}

func (idx *CommandIndex) LowerHelp(cmd *Command) string {
	idx.lowerHelpMu.RLock()
	if h, ok := idx.lowerHelp[cmd]; ok {
		idx.lowerHelpMu.RUnlock()
		return h
	}
	idx.lowerHelpMu.RUnlock()

	h := strings.ToLower(cmd.RawHelp)

	idx.lowerHelpMu.Lock()
	idx.lowerHelp[cmd] = h
	idx.lowerHelpMu.Unlock()
	return h
}

func (idx *CommandIndex) HelpContains(keyword string) (*Command, bool) {
	lower := strings.ToLower(keyword)
	for _, cmd := range idx.all {
		if strings.Contains(idx.LowerHelp(cmd), lower) {
			return cmd, true
		}
	}
	return nil, false
}

func (idx *CommandIndex) HelpContainsAny(keywords ...string) (*Command, bool) {
	lowered := make([]string, len(keywords))
	for i, kw := range keywords {
		lowered[i] = strings.ToLower(kw)
	}
	for _, cmd := range idx.all {
		h := idx.LowerHelp(cmd)
		for _, kw := range lowered {
			if strings.Contains(h, kw) {
				return cmd, true
			}
		}
	}
	return nil, false
}
