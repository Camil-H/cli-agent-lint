package checks

import (
	"context"
	"fmt"
	"testing"

	"github.com/Camil-H/cli-agent-lint/discovery"
)

func TestSD1_SkipNilProber(t *testing.T) {
	check := newCheckSD1()
	result := check.Run(context.Background(), &Input{Prober: nil})

	if result.Status != StatusSkip {
		t.Errorf("expected StatusSkip, got %s: %s", result.Status, result.Detail)
	}
}

func TestSD1_SkipWhenSO1NotPassed(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Flags: []*discovery.Flag{
			{Name: "output", EnumValues: []string{"json", "text"}},
		},
	}
	tree := makeTree(root)

	// Create a ResultSet where TE-1 failed.
	rs := NewResultSet()
	rs.Set("TE-1", &Result{CheckID: "TE-1", Status: StatusFail})

	// SD-1 is active but we need a prober. With nil prober it skips first.
	// Test the cross-check logic by verifying metadata.
	check := newCheckSD1()
	if check.ID() != "SD-1" {
		t.Errorf("expected ID SD-1, got %s", check.ID())
	}
	if check.Method() != Active {
		t.Errorf("expected Active method, got %s", check.Method())
	}

	// With nil prober, active check is skipped.
	result := check.Run(context.Background(), &Input{Tree: tree, Index: makeIndex(root), Prober: nil, ResultSet: rs})
	if result.Status != StatusSkip {
		t.Errorf("expected StatusSkip for nil prober, got %s", result.Status)
	}
}

func TestSD1_Metadata(t *testing.T) {
	check := newCheckSD1()
	if check.Method() != Active {
		t.Errorf("expected Active method, got %s", check.Method())
	}
	if check.Severity() != Warn {
		t.Errorf("expected Warn severity, got %s", check.Severity())
	}
}

func TestSD2_SkipNilProber(t *testing.T) {
	check := newCheckSD2()
	result := check.Run(context.Background(), &Input{Prober: nil})

	if result.Status != StatusSkip {
		t.Errorf("expected StatusSkip, got %s: %s", result.Status, result.Detail)
	}
}

func TestSD2_Metadata(t *testing.T) {
	check := newCheckSD2()
	if check.Method() != Active {
		t.Errorf("expected Active method, got %s", check.Method())
	}
	if check.Severity() != Warn {
		t.Errorf("expected Warn severity, got %s", check.Severity())
	}
}

// SD-3: Shell completions available (passive check)

func TestSD3_PassWithCompletionSubcommand(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Subcommands: []*discovery.Command{
			{Name: "completion", FullPath: []string{"mycli", "completion"}},
			{Name: "help", FullPath: []string{"mycli", "help"}},
		},
	}

	r := DefaultRegistry()
	check := r.Get("SD-3")
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for completion subcommand, got %s: %s", result.Status, result.Detail)
	}
}

func TestSD3_PassWithCompletionsSubcommand(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Subcommands: []*discovery.Command{
			{Name: "completions", FullPath: []string{"mycli", "completions"}},
		},
	}

	r := DefaultRegistry()
	check := r.Get("SD-3")
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for completions subcommand, got %s: %s", result.Status, result.Detail)
	}
}

func TestSD3_PassWithGenerateCompletionFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Flags: []*discovery.Flag{
			{Name: "generate-completion", Description: "Generate shell completion script"},
		},
	}

	r := DefaultRegistry()
	check := r.Get("SD-3")
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for --generate-completion flag, got %s: %s", result.Status, result.Detail)
	}
}

func TestSD3_PassWithCompletionsFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Flags: []*discovery.Flag{
			{Name: "completions", Description: "Generate shell completions"},
		},
	}

	r := DefaultRegistry()
	check := r.Get("SD-3")
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for --completions flag, got %s: %s", result.Status, result.Detail)
	}
}

func TestSD3_PassWithNestedCompletionSubcommand(t *testing.T) {
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
	check := r.Get("SD-3")
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for nested completion subcommand, got %s: %s", result.Status, result.Detail)
	}
}

func TestSD3_FailNoCompletion(t *testing.T) {
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
	check := r.Get("SD-3")
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusFail {
		t.Errorf("expected Fail, got %s: %s", result.Status, result.Detail)
	}
}

func TestSD3_SkipNilTree(t *testing.T) {
	r := DefaultRegistry()
	check := r.Get("SD-3")
	result := check.Run(context.Background(), &Input{Tree: nil})

	if result.Status != StatusSkip {
		t.Errorf("expected StatusSkip for nil tree, got %s: %s", result.Status, result.Detail)
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
	if check.Category() != CatSelfDescribing {
		t.Errorf("expected self-describing, got %s", check.Category())
	}
}

// SD-4: Schema / describe introspection (passive check)

func TestSD4_PassWithSchemaSubcommand(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Subcommands: []*discovery.Command{
			{Name: "schema", FullPath: []string{"mycli", "schema"}},
		},
	}

	r := DefaultRegistry()
	check := r.Get("SD-4")
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for schema subcommand, got %s: %s", result.Status, result.Detail)
	}
}

func TestSD4_PassWithDescribeSubcommand(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Subcommands: []*discovery.Command{
			{Name: "describe", FullPath: []string{"mycli", "describe"}},
		},
	}

	r := DefaultRegistry()
	check := r.Get("SD-4")
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for describe subcommand, got %s: %s", result.Status, result.Detail)
	}
}

func TestSD4_PassWithInspectSubcommand(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Subcommands: []*discovery.Command{
			{Name: "inspect", FullPath: []string{"mycli", "inspect"}},
		},
	}

	r := DefaultRegistry()
	check := r.Get("SD-4")
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for inspect subcommand, got %s: %s", result.Status, result.Detail)
	}
}

func TestSD4_PassWithApiSubcommand(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Subcommands: []*discovery.Command{
			{Name: "api", FullPath: []string{"mycli", "api"}},
		},
	}

	r := DefaultRegistry()
	check := r.Get("SD-4")
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for api subcommand, got %s: %s", result.Status, result.Detail)
	}
}

func TestSD4_PassWithDescribeFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Flags: []*discovery.Flag{
			{Name: "describe", Description: "Describe the schema"},
		},
	}

	r := DefaultRegistry()
	check := r.Get("SD-4")
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for --describe flag, got %s: %s", result.Status, result.Detail)
	}
}

func TestSD4_PassWithSchemaInHelp(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli\n\nThis tool supports schema introspection via the API.",
	}

	r := DefaultRegistry()
	check := r.Get("SD-4")
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for schema in help, got %s: %s", result.Status, result.Detail)
	}
}

func TestSD4_FailNoIntrospection(t *testing.T) {
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
	check := r.Get("SD-4")
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusFail {
		t.Errorf("expected Fail, got %s: %s", result.Status, result.Detail)
	}
}

func TestSD4_SkipNilTree(t *testing.T) {
	r := DefaultRegistry()
	check := r.Get("SD-4")
	result := check.Run(context.Background(), &Input{Tree: nil})

	if result.Status != StatusSkip {
		t.Errorf("expected StatusSkip for nil tree, got %s: %s", result.Status, result.Detail)
	}
}

func TestSD4_Metadata(t *testing.T) {
	r := DefaultRegistry()
	check := r.Get("SD-4")

	if check.Method() != Passive {
		t.Errorf("expected Passive, got %s", check.Method())
	}
	if check.Severity() != Warn {
		t.Errorf("expected Warn, got %s", check.Severity())
	}
}

// SD-5: Skill / context files (passive check)

func TestSD5_SkipNilTree(t *testing.T) {
	r := DefaultRegistry()
	check := r.Get("SD-5")
	result := check.Run(context.Background(), &Input{Tree: nil})

	if result.Status != StatusSkip {
		t.Errorf("expected StatusSkip for nil tree, got %s: %s", result.Status, result.Detail)
	}
}

func TestSD5_SkipEmptyTargetPath(t *testing.T) {
	root := &discovery.Command{Name: "mycli", FullPath: []string{"mycli"}}
	tree := &discovery.CommandTree{
		Root:       root,
		TargetPath: "",
	}

	r := DefaultRegistry()
	check := r.Get("SD-5")
	result := check.Run(context.Background(), &Input{Tree: tree, Index: discovery.NewIndex(root)})

	if result.Status != StatusSkip {
		t.Errorf("expected StatusSkip for empty target path, got %s: %s", result.Status, result.Detail)
	}
}

func TestSD5_PassWithMentionInHelp(t *testing.T) {
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
	check := r.Get("SD-5")
	result := check.Run(context.Background(), &Input{Tree: tree, Index: discovery.NewIndex(root)})

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for AGENTS.md mention in help, got %s: %s", result.Status, result.Detail)
	}
}

func TestSD5_PassWithLlmsTxtMentionInHelp(t *testing.T) {
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
	check := r.Get("SD-5")
	result := check.Run(context.Background(), &Input{Tree: tree, Index: discovery.NewIndex(root)})

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for llms.txt mention in help, got %s: %s", result.Status, result.Detail)
	}
}

func TestSD5_FailNoContextFiles(t *testing.T) {
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
	check := r.Get("SD-5")
	result := check.Run(context.Background(), &Input{Tree: tree, Index: discovery.NewIndex(root)})

	if result.Status != StatusFail {
		t.Errorf("expected Fail, got %s: %s", result.Status, result.Detail)
	}
}

func TestSD5_Metadata(t *testing.T) {
	r := DefaultRegistry()
	check := r.Get("SD-5")

	if check.Method() != Passive {
		t.Errorf("expected Passive, got %s", check.Method())
	}
	if check.Severity() != Info {
		t.Errorf("expected Info, got %s", check.Severity())
	}
	if check.Category() != CatSelfDescribing {
		t.Errorf("expected self-describing, got %s", check.Category())
	}
}

// SD-6: Help text with usage examples (passive check)

func TestSD6_Metadata(t *testing.T) {
	check := newCheckSD6()

	if check.ID() != "SD-6" {
		t.Errorf("expected SD-6, got %s", check.ID())
	}
	if check.Category() != CatSelfDescribing {
		t.Errorf("expected self-describing, got %s", check.Category())
	}
	if check.Severity() != Warn {
		t.Errorf("expected Warn, got %s", check.Severity())
	}
	if check.Method() != Passive {
		t.Errorf("expected Passive, got %s", check.Method())
	}
}

func TestSD6_SkipNilTree(t *testing.T) {
	check := newCheckSD6()
	result := check.Run(context.Background(), &Input{Tree: nil})

	if result.Status != StatusSkip {
		t.Errorf("expected StatusSkip, got %s: %s", result.Status, result.Detail)
	}
}

func TestSD6_PassWithExamplesSection(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli [command]\n\nExamples:\n  mycli list --output json\n  mycli create --name foo\n",
	}

	check := newCheckSD6()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for Examples: section, got %s: %s", result.Status, result.Detail)
	}
}

func TestSD6_PassWithExampleSection(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli\n\nExample:\n  mycli run\n",
	}

	check := newCheckSD6()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for Example: section, got %s: %s", result.Status, result.Detail)
	}
}

func TestSD6_PassWithEXAMPLESSection(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "USAGE: mycli\n\nEXAMPLES:\n  mycli run\n",
	}

	check := newCheckSD6()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for EXAMPLES: section, got %s: %s", result.Status, result.Detail)
	}
}

func TestSD6_PassWithExamplesOnSubcommand(t *testing.T) {
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

	check := newCheckSD6()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for Examples: on subcommand, got %s: %s", result.Status, result.Detail)
	}
}

func TestSD6_FailNoExamples(t *testing.T) {
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

	check := newCheckSD6()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusFail {
		t.Errorf("expected Fail, got %s: %s", result.Status, result.Detail)
	}
}

// SD-7: Actionable error messages (active check)

func TestSD7_SkipNilProber(t *testing.T) {
	check := newCheckSD7()
	result := check.Run(context.Background(), &Input{Prober: nil})
	if result.Status != StatusSkip {
		t.Errorf("expected skip, got %s", result.Status)
	}
}

// SD-8: Subcommand fan-out (passive check)

func TestSD8_PassFewSubcommands(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Subcommands: []*discovery.Command{
			{Name: "list", FullPath: []string{"mycli", "list"}},
			{Name: "create", FullPath: []string{"mycli", "create"}},
			{Name: "delete", FullPath: []string{"mycli", "delete"}},
		},
	}
	check := newCheckSD8()
	result := check.Run(context.Background(), makeInput(root))
	if result.Status != StatusPass {
		t.Errorf("expected pass for 3 subcommands, got %s: %s", result.Status, result.Detail)
	}
}

func TestSD8_FailTooManySubcommands(t *testing.T) {
	subs := make([]*discovery.Command, 20)
	for i := range subs {
		name := fmt.Sprintf("cmd%d", i)
		subs[i] = &discovery.Command{Name: name, FullPath: []string{"mycli", name}}
	}
	root := &discovery.Command{
		Name:        "mycli",
		FullPath:    []string{"mycli"},
		Subcommands: subs,
	}
	check := newCheckSD8()
	result := check.Run(context.Background(), makeInput(root))
	if result.Status != StatusFail {
		t.Errorf("expected fail for 20 subcommands, got %s: %s", result.Status, result.Detail)
	}
}

func TestSD8_PassNoSubcommands(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
	}
	check := newCheckSD8()
	result := check.Run(context.Background(), makeInput(root))
	if result.Status != StatusPass {
		t.Errorf("expected pass for no subcommands, got %s: %s", result.Status, result.Detail)
	}
}

func TestSD8_ChecksNestedLevels(t *testing.T) {
	// Root has 3 subcommands, but one nested command has 20
	bigSubs := make([]*discovery.Command, 20)
	for i := range bigSubs {
		name := fmt.Sprintf("sub%d", i)
		bigSubs[i] = &discovery.Command{Name: name, FullPath: []string{"mycli", "admin", name}}
	}
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Subcommands: []*discovery.Command{
			{Name: "list", FullPath: []string{"mycli", "list"}},
			{Name: "admin", FullPath: []string{"mycli", "admin"}, Subcommands: bigSubs},
		},
	}
	check := newCheckSD8()
	result := check.Run(context.Background(), makeInput(root))
	if result.Status != StatusFail {
		t.Errorf("expected fail for nested 20 subcommands, got %s: %s", result.Status, result.Detail)
	}
}

func TestSD8_SkipNilTree(t *testing.T) {
	check := newCheckSD8()
	result := check.Run(context.Background(), &Input{Tree: nil})
	if result.Status != StatusSkip {
		t.Errorf("expected skip, got %s", result.Status)
	}
}

// Active execution tests

func TestSD1_Active_GoodCLI(t *testing.T) {
	input := probeInput(t, "good-cli.sh")
	te1Result := newCheckTE1().Run(context.Background(), input)
	input.ResultSet.Set("TE-1", te1Result)
	if te1Result.Status != StatusPass {
		t.Fatalf("TE-1 must pass first, got %s: %s", te1Result.Status, te1Result.Detail)
	}

	result := newCheckSD1().Run(context.Background(), input)
	if result.Status != StatusPass {
		t.Errorf("expected pass, got %s: %s", result.Status, result.Detail)
	}
}

func TestSD1_Active_BadCLI(t *testing.T) {
	input := probeInput(t, "bad-cli.sh")
	te1Result := newCheckTE1().Run(context.Background(), input)
	input.ResultSet.Set("TE-1", te1Result)

	result := newCheckSD1().Run(context.Background(), input)
	if result.Status != StatusSkip {
		t.Errorf("expected skip (TE-1 not passed), got %s: %s", result.Status, result.Detail)
	}
}

func TestSD2_Active_GoodCLI(t *testing.T) {
	input := probeInput(t, "good-cli.sh")
	result := newCheckSD2().Run(context.Background(), input)
	if result.Status != StatusPass {
		t.Errorf("expected pass, got %s: %s", result.Status, result.Detail)
	}
}

func TestSD2_Active_BadCLI(t *testing.T) {
	input := probeInput(t, "bad-cli.sh")
	result := newCheckSD2().Run(context.Background(), input)
	if result.Status != StatusFail {
		t.Errorf("expected fail (decorated version), got %s: %s", result.Status, result.Detail)
	}
}

func TestSD7_Active_GoodCLI(t *testing.T) {
	input := probeInput(t, "good-cli.sh")
	result := newCheckSD7().Run(context.Background(), input)
	if result.Status != StatusPass {
		t.Errorf("expected pass, got %s: %s", result.Status, result.Detail)
	}
}

func TestSD7_Active_BadCLI(t *testing.T) {
	input := probeInput(t, "bad-cli.sh")
	result := newCheckSD7().Run(context.Background(), input)
	if result.Status != StatusFail {
		t.Errorf("expected fail (no actionable guidance), got %s: %s", result.Status, result.Detail)
	}
}
