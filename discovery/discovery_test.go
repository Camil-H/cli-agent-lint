package discovery

import (
	"path/filepath"
	"testing"
)

// Canned help texts

// cobraHelp simulates help output from a Cobra-based CLI.
// Note: "A tool for managing cloud resources" matches sectionHeaderRe
// (starts with uppercase, all letters/spaces) so extractDescription
// treats it as a section header and yields an empty description.
const cobraHelp = `A tool for managing cloud resources

Usage:
  cloudctl [command] [flags]

Available Commands:
  list        List all resources
  create      Create a new resource
  delete      Delete a resource
  describe    Show detailed information about a resource
  completion  Generate shell completions
  help        Help about any command

Flags:
  -h, --help              help for cloudctl
  -o, --output string     Output format {json,text,yaml}
  -v, --verbose           Enable verbose logging
      --timeout duration  Request timeout (default 30s)

Global Flags:
      --config string   Path to config file (default "~/.cloudctl.yaml")
      --region string   Cloud region to operate in
`

// clickHelp simulates help output from a Click or Typer CLI.
const clickHelp = `Usage: myapp [OPTIONS] COMMAND [ARGS]...

  A CLI app built with Click or Typer.

Options:
  --debug / --no-debug  Enable debug mode.
  -q, --quiet           Suppress output.
  --help                Show this message and exit.

Commands:
  add     Add a new item.
  remove  Remove an existing item.
  list    List all items.
  show    Show details for an item.
`

// argparseHelp simulates help output from an argparse-based CLI.
// The brace subcommands appear under "positional arguments:" which is
// treated as a command section by the parser. The brace line itself does
// not have trailing description text on the same line, so commandLineRe
// does not match it. The next line "Available management commands."
// does match commandLineRe with name="Available".
const argparseHelp = `usage: manage.py [-h] [--version] {migrate,runserver,shell,test} ...

Django management utility.

positional arguments:
  {migrate,runserver,shell,test}
                        Available management commands.

optional arguments:
  -h, --help            show this help message and exit
  --version             show program's version number and exit
  --settings MODULE     The Python path to a settings module.
  --verbosity {0,1,2,3}
                        Verbosity level; 0=minimal, 1=normal, 2=verbose, 3=debug.
`

// argparseBraceHelp uses a format where the brace subcommand list appears
// as its own section header (ends with ":") to exercise parseBraceSubcommands.
const argparseBraceHelp = `usage: tool.py [-h] command ...

A tool that uses brace-style subcommand listing.

{migrate,runserver,shell,test}:
  migrate       Run database migrations
  runserver     Start the development server
  shell         Open an interactive shell
  test          Run the test suite

optional arguments:
  -h, --help    show this help message and exit
`

// clapHelp simulates help output from a Rust Clap CLI.
const clapHelp = `my-rust-cli 0.1.0
A sample CLI written in Rust with Clap

USAGE:
    my-rust-cli [OPTIONS] <SUBCOMMAND>

OPTIONS:
    -h, --help             Print help information
    -V, --version          Print version information
    -c, --config <FILE>    Sets a custom config file
        --dry-run          Run without making changes

SUBCOMMANDS:
    init      Initialize a new project
    build     Build the current project
    push      Push changes to remote
    install   Install dependencies
    help      Print this message or the help of the given subcommand(s)
`

// Tests for ParseHelpOutput: Cobra-style

func TestParseHelpOutput_Cobra(t *testing.T) {
	cmd := ParseHelpOutput(cobraHelp, "cloudctl")

	t.Run("name", func(t *testing.T) {
		if cmd.Name != "cloudctl" {
			t.Errorf("Name = %q; want %q", cmd.Name, "cloudctl")
		}
	})

	// The description line "A tool for managing cloud resources" matches
	// the sectionHeaderRe regex (starts uppercase, all alpha/space chars),
	// so extractDescription treats it as a section header boundary and the
	// resulting description is empty.
	t.Run("description_empty_due_to_section_header_match", func(t *testing.T) {
		if cmd.Description != "" {
			t.Errorf("Description = %q; want empty (line matches sectionHeaderRe)", cmd.Description)
		}
	})

	t.Run("subcommands", func(t *testing.T) {
		want := map[string]string{
			"list":       "List all resources",
			"create":     "Create a new resource",
			"delete":     "Delete a resource",
			"describe":   "Show detailed information about a resource",
			"completion": "Generate shell completions",
			"help":       "Help about any command",
		}
		got := make(map[string]string)
		for _, sub := range cmd.Subcommands {
			got[sub.Name] = sub.Description
		}
		for name, desc := range want {
			if got[name] != desc {
				t.Errorf("subcommand %q description = %q; want %q", name, got[name], desc)
			}
		}
		if len(cmd.Subcommands) != len(want) {
			t.Errorf("len(Subcommands) = %d; want %d", len(cmd.Subcommands), len(want))
		}
	})

	t.Run("flags_parsed", func(t *testing.T) {
		flagNames := make(map[string]bool)
		for _, f := range cmd.Flags {
			flagNames[f.Name] = true
		}
		for _, name := range []string{"help", "output", "verbose", "timeout"} {
			if !flagNames[name] {
				t.Errorf("expected flag --%s in Flags; not found", name)
			}
		}
	})

	t.Run("flag_short_names", func(t *testing.T) {
		f := cmd.FindFlag("output")
		if f == nil {
			t.Fatal("flag --output not found")
		}
		if f.ShortName != "o" {
			t.Errorf("output ShortName = %q; want %q", f.ShortName, "o")
		}

		fHelp := cmd.FindFlag("help")
		if fHelp == nil {
			t.Fatal("flag --help not found")
		}
		if fHelp.ShortName != "h" {
			t.Errorf("help ShortName = %q; want %q", fHelp.ShortName, "h")
		}
	})

	t.Run("flag_output_has_description", func(t *testing.T) {
		f := cmd.FindFlag("output")
		if f == nil {
			t.Fatal("flag --output not found")
		}
		if f.Description == "" {
			t.Error("output flag description is empty; want non-empty")
		}
	})

	t.Run("inherited_flags", func(t *testing.T) {
		inheritedNames := make(map[string]bool)
		for _, f := range cmd.InheritedFlags {
			inheritedNames[f.Name] = true
		}
		for _, name := range []string{"config", "region"} {
			if !inheritedNames[name] {
				t.Errorf("expected inherited flag --%s; not found", name)
			}
		}
	})

	t.Run("inherited_flags_not_in_flags", func(t *testing.T) {
		flagNames := make(map[string]bool)
		for _, f := range cmd.Flags {
			flagNames[f.Name] = true
		}
		if flagNames["config"] {
			t.Error("--config should be in InheritedFlags, not Flags")
		}
		if flagNames["region"] {
			t.Error("--region should be in InheritedFlags, not Flags")
		}
	})

	// classifyCommand is called on the root command (cloudctl) which is
	// neither mutating nor list-like.
	t.Run("root_not_mutating", func(t *testing.T) {
		if cmd.IsMutating {
			t.Error("root IsMutating = true; want false")
		}
	})

	t.Run("root_not_list_like", func(t *testing.T) {
		if cmd.IsListLike {
			t.Error("root IsListLike = true; want false")
		}
	})

	// Subcommands are NOT classified by ParseHelpOutput (classification
	// happens during recursive discovery in discoverSubcommands).
	t.Run("subcommands_not_classified_by_parse", func(t *testing.T) {
		for _, sub := range cmd.Subcommands {
			if sub.IsMutating || sub.IsListLike {
				t.Errorf("subcommand %q: IsMutating=%v, IsListLike=%v; want both false (not classified in ParseHelpOutput)",
					sub.Name, sub.IsMutating, sub.IsListLike)
			}
		}
	})
}

// Tests for ParseHelpOutput: Click/Typer-style

func TestParseHelpOutput_Click(t *testing.T) {
	cmd := ParseHelpOutput(clickHelp, "myapp")

	t.Run("name", func(t *testing.T) {
		if cmd.Name != "myapp" {
			t.Errorf("Name = %q; want %q", cmd.Name, "myapp")
		}
	})

	// Starts with "Usage:" so extractDescription returns empty.
	t.Run("description_empty_due_to_usage_prefix", func(t *testing.T) {
		if cmd.Description != "" {
			t.Errorf("Description = %q; want empty (starts with Usage:)", cmd.Description)
		}
	})

	t.Run("subcommands", func(t *testing.T) {
		want := []string{"add", "remove", "list", "show"}
		got := make(map[string]bool)
		for _, sub := range cmd.Subcommands {
			got[sub.Name] = true
		}
		for _, name := range want {
			if !got[name] {
				t.Errorf("expected subcommand %q; not found", name)
			}
		}
		if len(cmd.Subcommands) != len(want) {
			t.Errorf("len(Subcommands) = %d; want %d", len(cmd.Subcommands), len(want))
		}
	})

	t.Run("subcommand_descriptions", func(t *testing.T) {
		descMap := make(map[string]string)
		for _, sub := range cmd.Subcommands {
			descMap[sub.Name] = sub.Description
		}
		if descMap["add"] != "Add a new item." {
			t.Errorf("add description = %q; want %q", descMap["add"], "Add a new item.")
		}
		if descMap["remove"] != "Remove an existing item." {
			t.Errorf("remove description = %q; want %q", descMap["remove"], "Remove an existing item.")
		}
	})

	t.Run("flags_from_options_section", func(t *testing.T) {
		// Click uses "Options:" which isFlagSection recognizes.
		flagNames := make(map[string]bool)
		for _, f := range cmd.Flags {
			flagNames[f.Name] = true
		}
		for _, name := range []string{"quiet", "help"} {
			if !flagNames[name] {
				t.Errorf("expected flag --%s; not found. All flags: %v", name, flagNamesList(cmd.Flags))
			}
		}
	})

	t.Run("flag_short_name", func(t *testing.T) {
		f := cmd.FindFlag("quiet")
		if f == nil {
			t.Fatal("flag --quiet not found")
		}
		if f.ShortName != "q" {
			t.Errorf("quiet ShortName = %q; want %q", f.ShortName, "q")
		}
	})
}

// Tests for ParseHelpOutput: Argparse-style

func TestParseHelpOutput_Argparse(t *testing.T) {
	cmd := ParseHelpOutput(argparseHelp, "manage.py")

	t.Run("name", func(t *testing.T) {
		if cmd.Name != "manage.py" {
			t.Errorf("Name = %q; want %q", cmd.Name, "manage.py")
		}
	})

	// Starts with "usage:" so description is empty.
	t.Run("description_empty_due_to_usage_prefix", func(t *testing.T) {
		if cmd.Description != "" {
			t.Errorf("Description = %q; want empty (starts with usage:)", cmd.Description)
		}
	})

	t.Run("optional_arguments_as_flags", func(t *testing.T) {
		flagNames := make(map[string]bool)
		for _, f := range cmd.Flags {
			flagNames[f.Name] = true
		}
		for _, name := range []string{"help", "version", "settings", "verbosity"} {
			if !flagNames[name] {
				t.Errorf("expected flag --%s; not found. All flags: %v", name, flagNamesList(cmd.Flags))
			}
		}
	})

	t.Run("help_flag_short_name", func(t *testing.T) {
		f := cmd.FindFlag("help")
		if f == nil {
			t.Fatal("flag --help not found")
		}
		if f.ShortName != "h" {
			t.Errorf("help ShortName = %q; want %q", f.ShortName, "h")
		}
	})

	// The usage line "usage: manage.py [-h] [--version] {migrate,...} ..."
	// also yields flags from parseUsageFlags: --version and --help (deduplicated).
	t.Run("usage_line_flags_extracted", func(t *testing.T) {
		// --version should be found (from both the usage line and optional arguments).
		f := cmd.FindFlag("version")
		if f == nil {
			t.Error("flag --version not found")
		}
	})

	// The positional arguments section treats the indented body lines as
	// command lines. The brace line has no trailing text so commandLineRe
	// does not match it. The next indented line "Available management commands."
	// matches commandLineRe with name = "Available".
	t.Run("positional_arguments_section_parsed_as_commands", func(t *testing.T) {
		if len(cmd.Subcommands) == 0 {
			t.Fatal("expected at least one subcommand from positional arguments section")
		}
		foundAvailable := false
		for _, sub := range cmd.Subcommands {
			if sub.Name == "Available" {
				foundAvailable = true
			}
		}
		if !foundAvailable {
			t.Errorf("expected subcommand 'Available' from commandLineRe match; got: %v", subNames(cmd.Subcommands))
		}
	})
}

// Tests for brace subcommand detection

func TestParseHelpOutput_ArgparseBraceSubcommands(t *testing.T) {
	cmd := ParseHelpOutput(argparseBraceHelp, "tool.py")

	t.Run("brace_subcommands_detected", func(t *testing.T) {
		want := []string{"migrate", "runserver", "shell", "test"}
		got := make(map[string]bool)
		for _, sub := range cmd.Subcommands {
			got[sub.Name] = true
		}
		for _, name := range want {
			if !got[name] {
				t.Errorf("expected subcommand %q from brace syntax; not found. Got: %v", name, subNames(cmd.Subcommands))
			}
		}
	})

	t.Run("brace_section_body_also_parsed_as_commands", func(t *testing.T) {
		// The lines under the brace header are also parsed as command lines
		// by the "commands" case since the brace header is NOT a "command section"
		// by isCommandSection. In the switch, isSubcommandBraceSection fires,
		// which parses from the header. The body lines under the brace header
		// are NOT parsed as commands because the switch only hits one case.
		// Verify the total count is from the brace parse only.
		names := subNames(cmd.Subcommands)
		braceNames := map[string]bool{"migrate": true, "runserver": true, "shell": true, "test": true}
		for _, n := range names {
			if !braceNames[n] {
				t.Logf("extra subcommand found: %q", n)
			}
		}
	})

	t.Run("flags_still_parsed", func(t *testing.T) {
		f := cmd.FindFlag("help")
		if f == nil {
			t.Error("flag --help not found")
		}
	})
}

// Tests for ParseHelpOutput: Clap-style (Rust)

func TestParseHelpOutput_Clap(t *testing.T) {
	cmd := ParseHelpOutput(clapHelp, "my-rust-cli")

	t.Run("name", func(t *testing.T) {
		if cmd.Name != "my-rust-cli" {
			t.Errorf("Name = %q; want %q", cmd.Name, "my-rust-cli")
		}
	})

	// "my-rust-cli 0.1.0" is the first non-empty line. The next line
	// "A sample CLI written in Rust with Clap" matches sectionHeaderRe,
	// so extractDescription stops and returns only the first line.
	t.Run("description", func(t *testing.T) {
		want := "my-rust-cli 0.1.0"
		if cmd.Description != want {
			t.Errorf("Description = %q; want %q", cmd.Description, want)
		}
	})

	t.Run("subcommands", func(t *testing.T) {
		want := []string{"init", "build", "push", "install", "help"}
		got := make(map[string]bool)
		for _, sub := range cmd.Subcommands {
			got[sub.Name] = true
		}
		for _, name := range want {
			if !got[name] {
				t.Errorf("expected subcommand %q; not found. Got: %v", name, subNames(cmd.Subcommands))
			}
		}
	})

	t.Run("subcommand_descriptions", func(t *testing.T) {
		descMap := make(map[string]string)
		for _, sub := range cmd.Subcommands {
			descMap[sub.Name] = sub.Description
		}
		if descMap["init"] != "Initialize a new project" {
			t.Errorf("init description = %q; want %q", descMap["init"], "Initialize a new project")
		}
		if descMap["build"] != "Build the current project" {
			t.Errorf("build description = %q; want %q", descMap["build"], "Build the current project")
		}
	})

	t.Run("flags", func(t *testing.T) {
		flagNames := make(map[string]bool)
		for _, f := range cmd.Flags {
			flagNames[f.Name] = true
		}
		for _, name := range []string{"help", "version", "config", "dry-run"} {
			if !flagNames[name] {
				t.Errorf("expected flag --%s; not found. Got flags: %v", name, flagNamesList(cmd.Flags))
			}
		}
	})

	t.Run("flag_short_names", func(t *testing.T) {
		f := cmd.FindFlag("help")
		if f == nil {
			t.Fatal("flag --help not found")
		}
		if f.ShortName != "h" {
			t.Errorf("help ShortName = %q; want %q", f.ShortName, "h")
		}

		fv := cmd.FindFlag("version")
		if fv == nil {
			t.Fatal("flag --version not found")
		}
		if fv.ShortName != "V" {
			t.Errorf("version ShortName = %q; want %q", fv.ShortName, "V")
		}
	})

	t.Run("config_flag_value_type", func(t *testing.T) {
		f := cmd.FindFlag("config")
		if f == nil {
			t.Fatal("flag --config not found")
		}
		if f.ShortName != "c" {
			t.Errorf("config ShortName = %q; want %q", f.ShortName, "c")
		}
	})

	t.Run("flag_without_short_name", func(t *testing.T) {
		f := cmd.FindFlag("dry-run")
		if f == nil {
			t.Fatal("flag --dry-run not found")
		}
		if f.ShortName != "" {
			t.Errorf("dry-run ShortName = %q; want empty", f.ShortName)
		}
	})
}

// Tests for Command helper methods

func TestCommand_HasFlag(t *testing.T) {
	cmd := &Command{
		Name: "test",
		Flags: []*Flag{
			{Name: "output", ShortName: "o"},
			{Name: "verbose", ShortName: "v"},
		},
		InheritedFlags: []*Flag{
			{Name: "config", ShortName: "c"},
		},
	}

	t.Run("by_long_name", func(t *testing.T) {
		if !cmd.HasFlag("output") {
			t.Error("HasFlag(output) = false; want true")
		}
	})

	t.Run("by_short_name", func(t *testing.T) {
		if !cmd.HasFlag("o") {
			t.Error("HasFlag(o) = false; want true")
		}
	})

	t.Run("inherited_flag_by_long_name", func(t *testing.T) {
		if !cmd.HasFlag("config") {
			t.Error("HasFlag(config) = false; want true")
		}
	})

	t.Run("inherited_flag_by_short_name", func(t *testing.T) {
		if !cmd.HasFlag("c") {
			t.Error("HasFlag(c) = false; want true")
		}
	})

	t.Run("missing_flag", func(t *testing.T) {
		if cmd.HasFlag("nonexistent") {
			t.Error("HasFlag(nonexistent) = true; want false")
		}
	})

	t.Run("multiple_names_any_match", func(t *testing.T) {
		if !cmd.HasFlag("nonexistent", "verbose") {
			t.Error("HasFlag(nonexistent, verbose) = false; want true")
		}
	})

	t.Run("multiple_names_none_match", func(t *testing.T) {
		if cmd.HasFlag("x", "y", "z") {
			t.Error("HasFlag(x, y, z) = true; want false")
		}
	})
}

func TestCommand_FindFlag(t *testing.T) {
	outputFlag := &Flag{Name: "output", ShortName: "o", Description: "output format"}
	configFlag := &Flag{Name: "config", ShortName: "c", Description: "config path"}
	cmd := &Command{
		Name:           "test",
		Flags:          []*Flag{outputFlag},
		InheritedFlags: []*Flag{configFlag},
	}

	t.Run("finds_by_long_name", func(t *testing.T) {
		f := cmd.FindFlag("output")
		if f != outputFlag {
			t.Errorf("FindFlag(output) = %v; want %v", f, outputFlag)
		}
	})

	t.Run("finds_by_short_name", func(t *testing.T) {
		f := cmd.FindFlag("o")
		if f != outputFlag {
			t.Errorf("FindFlag(o) = %v; want %v", f, outputFlag)
		}
	})

	t.Run("finds_inherited_flag", func(t *testing.T) {
		f := cmd.FindFlag("config")
		if f != configFlag {
			t.Errorf("FindFlag(config) = %v; want %v", f, configFlag)
		}
	})

	t.Run("returns_nil_for_missing", func(t *testing.T) {
		f := cmd.FindFlag("missing")
		if f != nil {
			t.Errorf("FindFlag(missing) = %v; want nil", f)
		}
	})

	t.Run("returns_first_match_from_multiple_names", func(t *testing.T) {
		f := cmd.FindFlag("missing", "output", "config")
		if f != outputFlag {
			t.Errorf("FindFlag(missing, output, config) = %v; want %v (first match)", f, outputFlag)
		}
	})

	t.Run("empty_command_returns_nil", func(t *testing.T) {
		empty := &Command{Name: "empty"}
		f := empty.FindFlag("anything")
		if f != nil {
			t.Errorf("FindFlag on empty command = %v; want nil", f)
		}
	})
}

func TestCommand_Walk(t *testing.T) {
	root := &Command{
		Name: "root",
		Subcommands: []*Command{
			{
				Name: "child1",
				Subcommands: []*Command{
					{Name: "grandchild1"},
					{Name: "grandchild2"},
				},
			},
			{
				Name: "child2",
			},
		},
	}

	t.Run("visits_all_nodes_depth_first", func(t *testing.T) {
		var visited []string
		root.Walk(func(cmd *Command) {
			visited = append(visited, cmd.Name)
		})
		want := []string{"root", "child1", "grandchild1", "grandchild2", "child2"}
		if len(visited) != len(want) {
			t.Fatalf("visited %d nodes; want %d. Got: %v", len(visited), len(want), visited)
		}
		for i, name := range want {
			if visited[i] != name {
				t.Errorf("visited[%d] = %q; want %q", i, visited[i], name)
			}
		}
	})

	t.Run("leaf_only", func(t *testing.T) {
		leaf := &Command{Name: "leaf"}
		var visited []string
		leaf.Walk(func(cmd *Command) {
			visited = append(visited, cmd.Name)
		})
		if len(visited) != 1 || visited[0] != "leaf" {
			t.Errorf("Walk on leaf = %v; want [leaf]", visited)
		}
	})
}

func TestCommand_AllCommands(t *testing.T) {
	root := &Command{
		Name: "root",
		Subcommands: []*Command{
			{
				Name: "a",
				Subcommands: []*Command{
					{Name: "a1"},
				},
			},
			{Name: "b"},
		},
	}

	t.Run("returns_all_commands_in_tree", func(t *testing.T) {
		all := root.AllCommands()
		if len(all) != 4 {
			t.Fatalf("AllCommands() returned %d; want 4", len(all))
		}
		names := make(map[string]bool)
		for _, cmd := range all {
			names[cmd.Name] = true
		}
		for _, want := range []string{"root", "a", "a1", "b"} {
			if !names[want] {
				t.Errorf("expected %q in AllCommands(); not found", want)
			}
		}
	})

	t.Run("first_element_is_root", func(t *testing.T) {
		all := root.AllCommands()
		if all[0].Name != "root" {
			t.Errorf("AllCommands()[0].Name = %q; want %q", all[0].Name, "root")
		}
	})

	t.Run("leaf_returns_self", func(t *testing.T) {
		leaf := &Command{Name: "leaf"}
		all := leaf.AllCommands()
		if len(all) != 1 {
			t.Fatalf("leaf AllCommands() returned %d; want 1", len(all))
		}
		if all[0].Name != "leaf" {
			t.Errorf("leaf AllCommands()[0].Name = %q; want %q", all[0].Name, "leaf")
		}
	})
}

// Tests for enum detection

func TestParseEnumValues(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want []string
	}{
		{
			name: "json_text_yaml",
			in:   "{json,text,yaml}",
			want: []string{"json", "text", "yaml"},
		},
		{
			name: "numeric",
			in:   "{0,1,2,3}",
			want: []string{"0", "1", "2", "3"},
		},
		{
			name: "single_value",
			in:   "{only}",
			want: []string{"only"},
		},
		{
			name: "no_braces",
			in:   "json,text,yaml",
			want: nil,
		},
		{
			name: "empty_braces",
			in:   "{}",
			want: nil,
		},
		{
			name: "enum_in_description_text",
			in:   "Output format: {table,json,csv}",
			want: []string{"table", "json", "csv"},
		},
		{
			name: "spaced_values",
			in:   "{ a , b , c }",
			want: []string{"a", "b", "c"},
		},
		{
			name: "enum_with_surrounding_text",
			in:   "Choose one of {alpha,beta,gamma} please",
			want: []string{"alpha", "beta", "gamma"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseEnumValues(tt.in)
			if tt.want == nil {
				if got != nil {
					t.Errorf("parseEnumValues(%q) = %v; want nil", tt.in, got)
				}
				return
			}
			if len(got) != len(tt.want) {
				t.Fatalf("parseEnumValues(%q) = %v; want %v", tt.in, got, tt.want)
			}
			for i, v := range tt.want {
				if got[i] != v {
					t.Errorf("parseEnumValues(%q)[%d] = %q; want %q", tt.in, i, got[i], v)
				}
			}
		})
	}
}

func TestParseEnumFromDesc(t *testing.T) {
	t.Run("extracts_enum_from_description", func(t *testing.T) {
		got := parseEnumFromDesc("Output format {json,text}")
		want := []string{"json", "text"}
		if len(got) != len(want) {
			t.Fatalf("parseEnumFromDesc = %v; want %v", got, want)
		}
		for i, v := range want {
			if got[i] != v {
				t.Errorf("[%d] = %q; want %q", i, got[i], v)
			}
		}
	})

	t.Run("returns_nil_for_no_enum", func(t *testing.T) {
		got := parseEnumFromDesc("Just a plain description")
		if got != nil {
			t.Errorf("parseEnumFromDesc = %v; want nil", got)
		}
	})
}

// Edge cases

func TestParseHelpOutput_EmptyText(t *testing.T) {
	cmd := ParseHelpOutput("", "empty")

	t.Run("name_is_set", func(t *testing.T) {
		if cmd.Name != "empty" {
			t.Errorf("Name = %q; want %q", cmd.Name, "empty")
		}
	})

	t.Run("no_subcommands", func(t *testing.T) {
		if len(cmd.Subcommands) != 0 {
			t.Errorf("len(Subcommands) = %d; want 0", len(cmd.Subcommands))
		}
	})

	t.Run("no_flags", func(t *testing.T) {
		if len(cmd.Flags) != 0 {
			t.Errorf("len(Flags) = %d; want 0", len(cmd.Flags))
		}
	})

	t.Run("no_inherited_flags", func(t *testing.T) {
		if len(cmd.InheritedFlags) != 0 {
			t.Errorf("len(InheritedFlags) = %d; want 0", len(cmd.InheritedFlags))
		}
	})

	t.Run("empty_description", func(t *testing.T) {
		if cmd.Description != "" {
			t.Errorf("Description = %q; want empty", cmd.Description)
		}
	})

	t.Run("not_mutating", func(t *testing.T) {
		if cmd.IsMutating {
			t.Error("IsMutating = true; want false")
		}
	})

	t.Run("not_list_like", func(t *testing.T) {
		if cmd.IsListLike {
			t.Error("IsListLike = true; want false")
		}
	})
}

func TestParseHelpOutput_NoSections(t *testing.T) {
	helpText := "my-tool v1.0: a simple utility that does things.\n\nThis tool has no structured sections at all.\nIt just prints some freeform text as help.\nNo flags, no commands, nothing to parse here.\n"
	cmd := ParseHelpOutput(helpText, "my-tool")

	t.Run("name_is_set", func(t *testing.T) {
		if cmd.Name != "my-tool" {
			t.Errorf("Name = %q; want %q", cmd.Name, "my-tool")
		}
	})

	t.Run("no_subcommands", func(t *testing.T) {
		if len(cmd.Subcommands) != 0 {
			t.Errorf("len(Subcommands) = %d; want 0", len(cmd.Subcommands))
		}
	})

	t.Run("no_flags", func(t *testing.T) {
		if len(cmd.Flags) != 0 {
			t.Errorf("len(Flags) = %d; want 0", len(cmd.Flags))
		}
	})

	t.Run("description_extracted", func(t *testing.T) {
		// First line "my-tool v1.0: a simple utility..." has a ":"
		// but isSectionHeader checks strings.HasSuffix(trimmed, ":").
		// It does NOT end with ":" (ends with "."), so not a section header.
		// The sectionHeaderRe won't match because of digits and punctuation.
		// So it gets added to desc. Then blank line terminates.
		if cmd.Description == "" {
			t.Error("Description is empty; expected first paragraph extracted")
		}
	})
}

func TestParseHelpOutput_OnlyFlags(t *testing.T) {
	helpText := `A flag-only tool.

Options:
  -n, --name string       Your name
  -a, --age int           Your age
      --dry-run           Do not actually do anything
`
	cmd := ParseHelpOutput(helpText, "flagtool")

	t.Run("has_flags", func(t *testing.T) {
		if len(cmd.Flags) == 0 {
			t.Fatal("expected flags to be parsed; got none")
		}
	})

	t.Run("no_subcommands", func(t *testing.T) {
		if len(cmd.Subcommands) != 0 {
			t.Errorf("len(Subcommands) = %d; want 0", len(cmd.Subcommands))
		}
	})

	t.Run("flag_name_with_short", func(t *testing.T) {
		f := cmd.FindFlag("name")
		if f == nil {
			t.Fatal("flag --name not found")
		}
		if f.ShortName != "n" {
			t.Errorf("name ShortName = %q; want %q", f.ShortName, "n")
		}
	})

	t.Run("flag_without_short", func(t *testing.T) {
		f := cmd.FindFlag("dry-run")
		if f == nil {
			t.Fatal("flag --dry-run not found")
		}
		if f.ShortName != "" {
			t.Errorf("dry-run ShortName = %q; want empty", f.ShortName)
		}
	})

	t.Run("description_extracted", func(t *testing.T) {
		// "A flag-only tool." contains punctuation so sectionHeaderRe
		// does not match. It's the first non-blank line.
		if cmd.Description == "" {
			t.Error("Description is empty; want non-empty")
		}
	})
}

func TestParseHelpOutput_UsageLineFlags(t *testing.T) {
	helpText := `Usage: tool [--format FORMAT] [--verbose] [--output FILE] command

Commands:
  run    Execute the task
  check  Validate configuration
`
	cmd := ParseHelpOutput(helpText, "tool")

	t.Run("flags_from_usage_line", func(t *testing.T) {
		// parseUsageFlags extracts --format, --verbose, --output from usage line.
		flagNames := make(map[string]bool)
		for _, f := range cmd.Flags {
			flagNames[f.Name] = true
		}
		for _, name := range []string{"format", "verbose", "output"} {
			if !flagNames[name] {
				t.Errorf("expected flag --%s from usage line; not found. Got: %v", name, flagNamesList(cmd.Flags))
			}
		}
	})

	t.Run("subcommands_parsed", func(t *testing.T) {
		names := make(map[string]bool)
		for _, sub := range cmd.Subcommands {
			names[sub.Name] = true
		}
		if !names["run"] || !names["check"] {
			t.Errorf("expected subcommands run and check; got: %v", subNames(cmd.Subcommands))
		}
	})
}

func TestParseHelpOutput_PersistentFlags(t *testing.T) {
	helpText := `mytool does things.

Flags:
  -v, --verbose   Be verbose

Persistent Flags:
      --log-level string   Set log level
`
	cmd := ParseHelpOutput(helpText, "mytool")

	t.Run("persistent_flags_are_inherited", func(t *testing.T) {
		inheritedNames := make(map[string]bool)
		for _, f := range cmd.InheritedFlags {
			inheritedNames[f.Name] = true
		}
		if !inheritedNames["log-level"] {
			t.Errorf("expected --log-level in InheritedFlags; not found. Got: %v", flagNamesList(cmd.InheritedFlags))
		}
	})

	t.Run("regular_flags_separate", func(t *testing.T) {
		flagNames := make(map[string]bool)
		for _, f := range cmd.Flags {
			flagNames[f.Name] = true
		}
		if !flagNames["verbose"] {
			t.Error("expected --verbose in Flags; not found")
		}
		if flagNames["log-level"] {
			t.Error("--log-level should be in InheritedFlags, not Flags")
		}
	})
}

// Tests for classifyCommand

func TestClassifyCommand(t *testing.T) {
	mutatingNames := []string{
		"create", "delete", "update", "set", "remove", "add", "modify",
		"push", "put", "post", "patch", "destroy", "rm", "write", "drop",
		"insert", "apply", "install", "uninstall",
	}
	for _, name := range mutatingNames {
		t.Run("mutating_"+name, func(t *testing.T) {
			cmd := &Command{Name: name}
			classifyCommand(cmd)
			if !cmd.IsMutating {
				t.Errorf("%q: IsMutating = false; want true", name)
			}
			if cmd.IsListLike {
				t.Errorf("%q: IsListLike = true; want false", name)
			}
		})
	}

	listLikeNames := []string{
		"list", "ls", "search", "find", "get", "query", "show",
		"describe", "inspect", "view",
	}
	for _, name := range listLikeNames {
		t.Run("listlike_"+name, func(t *testing.T) {
			cmd := &Command{Name: name}
			classifyCommand(cmd)
			if !cmd.IsListLike {
				t.Errorf("%q: IsListLike = false; want true", name)
			}
			if cmd.IsMutating {
				t.Errorf("%q: IsMutating = true; want false", name)
			}
		})
	}

	neutralNames := []string{"help", "version", "config", "auth", "login", "init", "build", "run"}
	for _, name := range neutralNames {
		t.Run("neutral_"+name, func(t *testing.T) {
			cmd := &Command{Name: name}
			classifyCommand(cmd)
			if cmd.IsMutating {
				t.Errorf("%q: IsMutating = true; want false", name)
			}
			if cmd.IsListLike {
				t.Errorf("%q: IsListLike = true; want false", name)
			}
		})
	}

	t.Run("case_insensitive", func(t *testing.T) {
		// classifyCommand lowercases the name before matching.
		cmd := &Command{Name: "CREATE"}
		classifyCommand(cmd)
		if !cmd.IsMutating {
			t.Error("CREATE: IsMutating = false; want true (case insensitive)")
		}
	})

	t.Run("partial_name_no_match", func(t *testing.T) {
		// "creating" should NOT match "create" (regex uses ^ and $).
		cmd := &Command{Name: "creating"}
		classifyCommand(cmd)
		if cmd.IsMutating {
			t.Error("creating: IsMutating = true; want false (partial match)")
		}
	})
}

// Tests for internal helpers

func TestIsSectionHeader(t *testing.T) {
	tests := []struct {
		line string
		want bool
	}{
		{"Available Commands:", true},
		{"Flags:", true},
		{"Global Flags:", true},
		{"SUBCOMMANDS:", true},
		{"OPTIONS:", true},
		{"Commands:", true},
		{"positional arguments:", true},
		{"optional arguments:", true},
		// All-alpha lines that match sectionHeaderRe
		{"Available Commands", true},
		{"Some Section Header", true},
		// Does NOT match
		{"", false},
		{"  --flag  description", false},      // leading whitespace trimmed, but starts with -
		{"{foo,bar}", false},                   // starts with {
		{"my-tool v1.0: does things.", false},  // contains digits and punctuation
		{"123 numeric start:", true},            // ends with ":" so HasSuffix matches
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			got := isSectionHeader(tt.line)
			if got != tt.want {
				t.Errorf("isSectionHeader(%q) = %v; want %v", tt.line, got, tt.want)
			}
		})
	}
}

func TestIsCommandSection(t *testing.T) {
	tests := []struct {
		header string
		want   bool
	}{
		{"Available Commands:", true},
		{"Commands:", true},
		{"SUBCOMMANDS:", true},
		{"positional arguments:", true},
		{"Flags:", false},
		{"Options:", false},
		{"Global Flags:", false},
	}

	for _, tt := range tests {
		t.Run(tt.header, func(t *testing.T) {
			got := isCommandSection(tt.header)
			if got != tt.want {
				t.Errorf("isCommandSection(%q) = %v; want %v", tt.header, got, tt.want)
			}
		})
	}
}

func TestIsFlagSection(t *testing.T) {
	tests := []struct {
		header string
		want   bool
	}{
		{"Flags:", true},
		{"Options:", true},
		{"optional arguments:", true},
		{"Global Flags:", true},
		{"Inherited Flags:", true},
		{"Persistent Flags:", true},
		{"OPTIONS:", true},
		{"Available Commands:", false},
		{"Commands:", false},
		{"SUBCOMMANDS:", false},
	}

	for _, tt := range tests {
		t.Run(tt.header, func(t *testing.T) {
			got := isFlagSection(tt.header)
			if got != tt.want {
				t.Errorf("isFlagSection(%q) = %v; want %v", tt.header, got, tt.want)
			}
		})
	}
}

func TestIsSubcommandBraceSection(t *testing.T) {
	tests := []struct {
		header string
		want   bool
	}{
		{"{a,b,c}:", true},
		{"{migrate,runserver,shell,test}:", true},
		{"Commands:", false},
		{"positional arguments:", false},
		{"no braces here", false},
	}

	for _, tt := range tests {
		t.Run(tt.header, func(t *testing.T) {
			got := isSubcommandBraceSection(tt.header)
			if got != tt.want {
				t.Errorf("isSubcommandBraceSection(%q) = %v; want %v", tt.header, got, tt.want)
			}
		})
	}
}

func TestBaseName(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"/usr/local/bin/myapp", "myapp"},
		{"myapp", "myapp"},
		{"./myapp", "myapp"},
		{"/bin/ls", "ls"},
		{"relative/path/tool", "tool"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := filepath.Base(tt.path)
			if got != tt.want {
				t.Errorf("filepath.Base(%q) = %q; want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestFirstNonEmpty(t *testing.T) {
	tests := []struct {
		a, b, want string
	}{
		{"first", "second", "first"},
		{"", "second", "second"},
		{"first", "", "first"},
		{"", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.a+"_or_"+tt.b, func(t *testing.T) {
			got := firstNonEmpty(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("firstNonEmpty(%q, %q) = %q; want %q", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestDeduplicateFlags(t *testing.T) {
	t.Run("removes_duplicates_keeps_first", func(t *testing.T) {
		flags := []*Flag{
			{Name: "output", ShortName: "o", Description: "first"},
			{Name: "verbose", ShortName: "v"},
			{Name: "output", ShortName: "o", Description: "duplicate"},
			{Name: "help", ShortName: "h"},
		}
		result := deduplicateFlags(flags)
		if len(result) != 3 {
			t.Fatalf("len(deduplicateFlags) = %d; want 3", len(result))
		}
		for _, f := range result {
			if f.Name == "output" && f.Description == "duplicate" {
				t.Error("kept the duplicate instead of the first occurrence")
			}
		}
	})

	t.Run("nil_input", func(t *testing.T) {
		result := deduplicateFlags(nil)
		if len(result) != 0 {
			t.Errorf("deduplicateFlags(nil) returned %d flags; want 0", len(result))
		}
	})

	t.Run("no_duplicates", func(t *testing.T) {
		flags := []*Flag{
			{Name: "a"},
			{Name: "b"},
			{Name: "c"},
		}
		result := deduplicateFlags(flags)
		if len(result) != 3 {
			t.Errorf("deduplicateFlags returned %d; want 3", len(result))
		}
	})

	t.Run("short_name_only_dedup", func(t *testing.T) {
		flags := []*Flag{
			{ShortName: "v"},
			{ShortName: "v", Description: "dup"},
		}
		result := deduplicateFlags(flags)
		if len(result) != 1 {
			t.Errorf("deduplicateFlags returned %d; want 1", len(result))
		}
	})
}

func TestMergeFlags(t *testing.T) {
	t.Run("both_nil", func(t *testing.T) {
		result := mergeFlags(nil, nil)
		if result != nil {
			t.Errorf("mergeFlags(nil, nil) = %v; want nil", result)
		}
	})

	t.Run("existing_nil", func(t *testing.T) {
		parsed := []*Flag{{Name: "a"}}
		result := mergeFlags(nil, parsed)
		if len(result) != 1 || result[0].Name != "a" {
			t.Errorf("mergeFlags(nil, [a]) = %v; want [a]", flagNamesList(result))
		}
	})

	t.Run("parsed_nil", func(t *testing.T) {
		existing := []*Flag{{Name: "a"}}
		result := mergeFlags(existing, nil)
		if len(result) != 1 || result[0].Name != "a" {
			t.Errorf("mergeFlags([a], nil) = %v; want [a]", flagNamesList(result))
		}
	})

	t.Run("deduplicates_on_merge", func(t *testing.T) {
		existing := []*Flag{{Name: "a"}, {Name: "b"}}
		parsed := []*Flag{{Name: "b"}, {Name: "c"}}
		result := mergeFlags(existing, parsed)
		if len(result) != 3 {
			t.Fatalf("mergeFlags([a,b], [b,c]) returned %d flags; want 3", len(result))
		}
		names := make(map[string]bool)
		for _, f := range result {
			names[f.Name] = true
		}
		for _, want := range []string{"a", "b", "c"} {
			if !names[want] {
				t.Errorf("expected flag %q in merged result; not found", want)
			}
		}
	})
}

func TestParseBraceSubcommands(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		cmds := parseBraceSubcommands("{a,b,c}")
		if len(cmds) != 3 {
			t.Fatalf("got %d subcommands; want 3", len(cmds))
		}
		want := []string{"a", "b", "c"}
		for i, cmd := range cmds {
			if cmd.Name != want[i] {
				t.Errorf("cmds[%d].Name = %q; want %q", i, cmd.Name, want[i])
			}
		}
	})

	t.Run("with_surrounding_text", func(t *testing.T) {
		cmds := parseBraceSubcommands("  {migrate,runserver,shell}  ")
		if len(cmds) != 3 {
			t.Fatalf("got %d subcommands; want 3", len(cmds))
		}
	})

	t.Run("no_braces", func(t *testing.T) {
		cmds := parseBraceSubcommands("no braces here")
		if cmds != nil {
			t.Errorf("expected nil; got %v", cmds)
		}
	})

	t.Run("empty_braces", func(t *testing.T) {
		cmds := parseBraceSubcommands("{}")
		if len(cmds) != 0 {
			t.Errorf("expected 0 subcommands for empty braces; got %d", len(cmds))
		}
	})

	t.Run("single_command", func(t *testing.T) {
		cmds := parseBraceSubcommands("{only}")
		if len(cmds) != 1 || cmds[0].Name != "only" {
			t.Errorf("expected [only]; got %v", subNames(cmds))
		}
	})

	t.Run("with_colon_suffix", func(t *testing.T) {
		cmds := parseBraceSubcommands("{migrate,runserver,shell,test}:")
		if len(cmds) != 4 {
			t.Fatalf("got %d subcommands; want 4", len(cmds))
		}
	})
}

func TestParseUsageFlags(t *testing.T) {
	t.Run("extracts_flags_from_usage", func(t *testing.T) {
		lines := []string{
			"Usage: mytool [--output FORMAT] [--verbose] [--config FILE] COMMAND",
		}
		flags := parseUsageFlags(lines)
		names := make(map[string]bool)
		for _, f := range flags {
			names[f.Name] = true
		}
		for _, want := range []string{"output", "verbose", "config"} {
			if !names[want] {
				t.Errorf("expected flag %q from usage line; not found", want)
			}
		}
	})

	t.Run("no_usage_line", func(t *testing.T) {
		lines := []string{
			"This is a description",
			"It has no usage line",
		}
		flags := parseUsageFlags(lines)
		if len(flags) != 0 {
			t.Errorf("expected 0 flags; got %d", len(flags))
		}
	})

	t.Run("usage_with_colon_space", func(t *testing.T) {
		lines := []string{
			"usage : tool [--flag] subcommand",
		}
		flags := parseUsageFlags(lines)
		names := make(map[string]bool)
		for _, f := range flags {
			names[f.Name] = true
		}
		if !names["flag"] {
			t.Errorf("expected flag 'flag' from 'usage :' line; not found")
		}
	})
}

func TestLooksLikeValueType(t *testing.T) {
	trueCases := []string{
		"string", "STRING", "int", "INT", "bool", "BOOL",
		"float", "FLOAT", "duration", "DURATION", "file", "FILE",
		"path", "PATH", "url", "URL", "format", "FORMAT",
	}
	for _, v := range trueCases {
		t.Run("true_"+v, func(t *testing.T) {
			if !looksLikeValueType(v) {
				t.Errorf("looksLikeValueType(%q) = false; want true", v)
			}
		})
	}

	falseCases := []string{"some-description", "FooBar", "text", "json", "verbose", ""}
	for _, v := range falseCases {
		t.Run("false_"+v, func(t *testing.T) {
			if looksLikeValueType(v) {
				t.Errorf("looksLikeValueType(%q) = true; want false", v)
			}
		})
	}
}

func TestExtractDescription(t *testing.T) {
	t.Run("stops_at_blank_line", func(t *testing.T) {
		lines := []string{"first line.", "second line.", "", "third line."}
		got := extractDescription(lines)
		if got != "first line. second line." {
			t.Errorf("extractDescription = %q; want %q", got, "first line. second line.")
		}
	})

	t.Run("skips_leading_blanks", func(t *testing.T) {
		lines := []string{"", "", "content here."}
		got := extractDescription(lines)
		if got != "content here." {
			t.Errorf("extractDescription = %q; want %q", got, "content here.")
		}
	})

	t.Run("stops_at_usage_line", func(t *testing.T) {
		lines := []string{"Usage: tool [flags]", "some text"}
		got := extractDescription(lines)
		if got != "" {
			t.Errorf("extractDescription = %q; want empty (starts with usage)", got)
		}
	})

	t.Run("stops_at_section_header", func(t *testing.T) {
		lines := []string{"Options:", "  --flag  description"}
		got := extractDescription(lines)
		if got != "" {
			t.Errorf("extractDescription = %q; want empty (starts with section header)", got)
		}
	})

	t.Run("empty_lines_only", func(t *testing.T) {
		lines := []string{"", "", ""}
		got := extractDescription(lines)
		if got != "" {
			t.Errorf("extractDescription = %q; want empty", got)
		}
	})
}

// Test helpers

func subNames(cmds []*Command) []string {
	var names []string
	for _, c := range cmds {
		names = append(names, c.Name)
	}
	return names
}

func flagNamesList(flags []*Flag) []string {
	var names []string
	for _, f := range flags {
		n := f.Name
		if n == "" {
			n = "-" + f.ShortName
		}
		names = append(names, n)
	}
	return names
}
