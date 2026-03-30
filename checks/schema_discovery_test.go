package checks

import (
	"context"
	"testing"

	"github.com/cli-agent-lint/cli-agent-lint/discovery"
)

// ---------------------------------------------------------------------------
// SD-1: Shell completions available (passive check)
// ---------------------------------------------------------------------------

func TestSD1_PassWithCompletionSubcommand(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Subcommands: []*discovery.Command{
			{Name: "completion", FullPath: []string{"mycli", "completion"}},
			{Name: "help", FullPath: []string{"mycli", "help"}},
		},
	}

	r := DefaultRegistry()
	check := r.Get("SD-1")
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for completion subcommand, got %s: %s", result.Status, result.Detail)
	}
}

func TestSD1_PassWithCompletionsSubcommand(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Subcommands: []*discovery.Command{
			{Name: "completions", FullPath: []string{"mycli", "completions"}},
		},
	}

	r := DefaultRegistry()
	check := r.Get("SD-1")
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for completions subcommand, got %s: %s", result.Status, result.Detail)
	}
}

func TestSD1_PassWithGenerateCompletionFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Flags: []*discovery.Flag{
			{Name: "generate-completion", Description: "Generate shell completion script"},
		},
	}

	r := DefaultRegistry()
	check := r.Get("SD-1")
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for --generate-completion flag, got %s: %s", result.Status, result.Detail)
	}
}

func TestSD1_PassWithCompletionsFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Flags: []*discovery.Flag{
			{Name: "completions", Description: "Generate shell completions"},
		},
	}

	r := DefaultRegistry()
	check := r.Get("SD-1")
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for --completions flag, got %s: %s", result.Status, result.Detail)
	}
}

func TestSD1_PassWithNestedCompletionSubcommand(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Subcommands: []*discovery.Command{
			{
				Name:     "tools",
				FullPath: []string{"mycli", "tools"},
				Subcommands: []*discovery.Command{
					{Name: "completion", FullPath: []string{"mycli", "tools", "completion"}},
				},
			},
		},
	}

	r := DefaultRegistry()
	check := r.Get("SD-1")
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for nested completion subcommand, got %s: %s", result.Status, result.Detail)
	}
}

func TestSD1_FailNoCompletion(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Subcommands: []*discovery.Command{
			{Name: "list", FullPath: []string{"mycli", "list"}},
			{Name: "create", FullPath: []string{"mycli", "create"}},
		},
		Flags: []*discovery.Flag{
			{Name: "verbose"},
		},
	}

	r := DefaultRegistry()
	check := r.Get("SD-1")
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusFail {
		t.Errorf("expected Fail, got %s: %s", result.Status, result.Detail)
	}
}

func TestSD1_SkipNilTree(t *testing.T) {
	r := DefaultRegistry()
	check := r.Get("SD-1")
	result := check.Run(context.Background(), &Input{Tree: nil})

	if result.Status != StatusSkip {
		t.Errorf("expected StatusSkip for nil tree, got %s: %s", result.Status, result.Detail)
	}
}

func TestSD1_Metadata(t *testing.T) {
	r := DefaultRegistry()
	check := r.Get("SD-1")

	if check.Method() != Passive {
		t.Errorf("expected Passive, got %s", check.Method())
	}
	if check.Severity() != Info {
		t.Errorf("expected Info, got %s", check.Severity())
	}
	if check.Category() != CatSchemaDiscovery {
		t.Errorf("expected schema-discovery, got %s", check.Category())
	}
}

// ---------------------------------------------------------------------------
// SD-2: Schema / describe introspection (passive check)
// ---------------------------------------------------------------------------

func TestSD2_PassWithSchemaSubcommand(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Subcommands: []*discovery.Command{
			{Name: "schema", FullPath: []string{"mycli", "schema"}},
		},
	}

	r := DefaultRegistry()
	check := r.Get("SD-2")
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for schema subcommand, got %s: %s", result.Status, result.Detail)
	}
}

func TestSD2_PassWithDescribeSubcommand(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Subcommands: []*discovery.Command{
			{Name: "describe", FullPath: []string{"mycli", "describe"}},
		},
	}

	r := DefaultRegistry()
	check := r.Get("SD-2")
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for describe subcommand, got %s: %s", result.Status, result.Detail)
	}
}

func TestSD2_PassWithInspectSubcommand(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Subcommands: []*discovery.Command{
			{Name: "inspect", FullPath: []string{"mycli", "inspect"}},
		},
	}

	r := DefaultRegistry()
	check := r.Get("SD-2")
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for inspect subcommand, got %s: %s", result.Status, result.Detail)
	}
}

func TestSD2_PassWithApiSubcommand(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Subcommands: []*discovery.Command{
			{Name: "api", FullPath: []string{"mycli", "api"}},
		},
	}

	r := DefaultRegistry()
	check := r.Get("SD-2")
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for api subcommand, got %s: %s", result.Status, result.Detail)
	}
}

func TestSD2_PassWithDescribeFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Flags: []*discovery.Flag{
			{Name: "describe", Description: "Describe the schema"},
		},
	}

	r := DefaultRegistry()
	check := r.Get("SD-2")
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for --describe flag, got %s: %s", result.Status, result.Detail)
	}
}

func TestSD2_PassWithSchemaInHelp(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli\n\nThis tool supports schema introspection via the API.",
	}

	r := DefaultRegistry()
	check := r.Get("SD-2")
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for schema in help, got %s: %s", result.Status, result.Detail)
	}
}

func TestSD2_FailNoIntrospection(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "A basic CLI tool.",
		Subcommands: []*discovery.Command{
			{Name: "list", FullPath: []string{"mycli", "list"}},
		},
		Flags: []*discovery.Flag{
			{Name: "verbose"},
		},
	}

	r := DefaultRegistry()
	check := r.Get("SD-2")
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusFail {
		t.Errorf("expected Fail, got %s: %s", result.Status, result.Detail)
	}
}

func TestSD2_SkipNilTree(t *testing.T) {
	r := DefaultRegistry()
	check := r.Get("SD-2")
	result := check.Run(context.Background(), &Input{Tree: nil})

	if result.Status != StatusSkip {
		t.Errorf("expected StatusSkip for nil tree, got %s: %s", result.Status, result.Detail)
	}
}

func TestSD2_Metadata(t *testing.T) {
	r := DefaultRegistry()
	check := r.Get("SD-2")

	if check.Method() != Passive {
		t.Errorf("expected Passive, got %s", check.Method())
	}
	if check.Severity() != Warn {
		t.Errorf("expected Warn, got %s", check.Severity())
	}
}

// ---------------------------------------------------------------------------
// SD-3: Skill / context files (passive check)
// ---------------------------------------------------------------------------

func TestSD3_SkipNilTree(t *testing.T) {
	r := DefaultRegistry()
	check := r.Get("SD-3")
	result := check.Run(context.Background(), &Input{Tree: nil})

	if result.Status != StatusSkip {
		t.Errorf("expected StatusSkip for nil tree, got %s: %s", result.Status, result.Detail)
	}
}

func TestSD3_SkipEmptyTargetPath(t *testing.T) {
	root := &discovery.Command{Name: "mycli", FullPath: []string{"mycli"}}
	tree := &discovery.CommandTree{
		Root:       root,
		TargetPath: "",
	}

	r := DefaultRegistry()
	check := r.Get("SD-3")
	result := check.Run(context.Background(), &Input{Tree: tree, Index: discovery.NewIndex(root)})

	if result.Status != StatusSkip {
		t.Errorf("expected StatusSkip for empty target path, got %s: %s", result.Status, result.Detail)
	}
}

func TestSD3_PassWithMentionInHelp(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli\n\nPlace an AGENTS.md file in your project root.",
	}
	tree := &discovery.CommandTree{
		Root:       root,
		TargetPath: "/tmp/nonexistent/path/mycli",
	}

	r := DefaultRegistry()
	check := r.Get("SD-3")
	result := check.Run(context.Background(), &Input{Tree: tree, Index: discovery.NewIndex(root)})

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for AGENTS.md mention in help, got %s: %s", result.Status, result.Detail)
	}
}

func TestSD3_PassWithLlmsTxtMentionInHelp(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli\n\nReads llms.txt for context.",
	}
	tree := &discovery.CommandTree{
		Root:       root,
		TargetPath: "/tmp/nonexistent/path/mycli",
	}

	r := DefaultRegistry()
	check := r.Get("SD-3")
	result := check.Run(context.Background(), &Input{Tree: tree, Index: discovery.NewIndex(root)})

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for llms.txt mention in help, got %s: %s", result.Status, result.Detail)
	}
}

func TestSD3_FailNoContextFiles(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli\n\nA simple tool.",
	}
	// Use a non-existent path so no files will be found on disk.
	tree := &discovery.CommandTree{
		Root:       root,
		TargetPath: "/tmp/nonexistent-abcdef123456/mycli",
	}

	r := DefaultRegistry()
	check := r.Get("SD-3")
	result := check.Run(context.Background(), &Input{Tree: tree, Index: discovery.NewIndex(root)})

	if result.Status != StatusFail {
		t.Errorf("expected Fail, got %s: %s", result.Status, result.Detail)
	}
}

func TestSD3_Metadata(t *testing.T) {
	r := DefaultRegistry()
	check := r.Get("SD-3")

	if check.Method() != Passive {
		t.Errorf("expected Passive, got %s", check.Method())
	}
	if check.Severity() != Info {
		t.Errorf("expected Info, got %s", check.Severity())
	}
	if check.Category() != CatSchemaDiscovery {
		t.Errorf("expected schema-discovery, got %s", check.Category())
	}
}

// ---------------------------------------------------------------------------
// SD-4: Help text with usage examples (passive check)
// ---------------------------------------------------------------------------

func TestSD4_Metadata(t *testing.T) {
	check := newCheckSD4()

	if check.ID() != "SD-4" {
		t.Errorf("expected SD-4, got %s", check.ID())
	}
	if check.Category() != CatSchemaDiscovery {
		t.Errorf("expected schema-discovery, got %s", check.Category())
	}
	if check.Severity() != Warn {
		t.Errorf("expected Warn, got %s", check.Severity())
	}
	if check.Method() != Passive {
		t.Errorf("expected Passive, got %s", check.Method())
	}
}

func TestSD4_SkipNilTree(t *testing.T) {
	check := newCheckSD4()
	result := check.Run(context.Background(), &Input{Tree: nil})

	if result.Status != StatusSkip {
		t.Errorf("expected StatusSkip, got %s: %s", result.Status, result.Detail)
	}
}

func TestSD4_PassWithExamplesSection(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli [command]\n\nExamples:\n  mycli list --output json\n  mycli create --name foo\n",
	}

	check := newCheckSD4()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for Examples: section, got %s: %s", result.Status, result.Detail)
	}
}

func TestSD4_PassWithExampleSection(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli\n\nExample:\n  mycli run\n",
	}

	check := newCheckSD4()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for Example: section, got %s: %s", result.Status, result.Detail)
	}
}

func TestSD4_PassWithEXAMPLESSection(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "USAGE: mycli\n\nEXAMPLES:\n  mycli run\n",
	}

	check := newCheckSD4()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for EXAMPLES: section, got %s: %s", result.Status, result.Detail)
	}
}

func TestSD4_PassWithExamplesOnSubcommand(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli [command]",
		Subcommands: []*discovery.Command{
			{
				Name:     "list",
				FullPath: []string{"mycli", "list"},
				RawHelp:  "List items\n\nExamples:\n  mycli list --all\n",
			},
		},
	}

	check := newCheckSD4()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for Examples: on subcommand, got %s: %s", result.Status, result.Detail)
	}
}

func TestSD4_FailNoExamples(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli [command]\n\nA simple tool.\n\nFlags:\n  --verbose\n",
		Subcommands: []*discovery.Command{
			{
				Name:     "list",
				FullPath: []string{"mycli", "list"},
				RawHelp:  "List items\n\nFlags:\n  --all\n",
			},
		},
	}

	check := newCheckSD4()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusFail {
		t.Errorf("expected Fail, got %s: %s", result.Status, result.Detail)
	}
}
