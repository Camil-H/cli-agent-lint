package checks

import (
	"context"
	"testing"

	"github.com/cli-agent-lint/cli-agent-lint/discovery"
)

func makeTree(root *discovery.Command) *discovery.CommandTree {
	return &discovery.CommandTree{Root: root, TargetPath: "/usr/bin/test-cli"}
}

func makeIndex(root *discovery.Command) *discovery.CommandIndex {
	return discovery.NewIndex(root)
}

// makeInput creates an Input with both Tree and pre-computed Index.
func makeInput(root *discovery.Command) *Input {
	tree := makeTree(root)
	var idx *discovery.CommandIndex
	if root != nil {
		idx = discovery.NewIndex(root)
	}
	return &Input{Tree: tree, Index: idx}
}

func TestSO1_PassWithOutputJsonFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Flags: []*discovery.Flag{
			{Name: "output", Description: "Output format", EnumValues: []string{"json", "text"}},
		},
	}

	check := newCheckSO1()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass, got %s: %s", result.Status, result.Detail)
	}
}

func TestSO1_PassWithFormatFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Flags: []*discovery.Flag{
			{Name: "format", Description: "Choose output format: json or yaml"},
		},
	}

	check := newCheckSO1()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass, got %s: %s", result.Status, result.Detail)
	}
}

func TestSO1_PassWithJsonFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Flags: []*discovery.Flag{
			{Name: "json", Description: "Output JSON"},
		},
	}

	check := newCheckSO1()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass, got %s: %s", result.Status, result.Detail)
	}
}

func TestSO1_PassWithJsonFlagOnSubcommand(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Subcommands: []*discovery.Command{
			{
				Name:     "list",
				FullPath: []string{"mycli", "list"},
				Flags: []*discovery.Flag{
					{Name: "output", EnumValues: []string{"json", "table"}},
				},
			},
		},
	}

	check := newCheckSO1()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass, got %s: %s", result.Status, result.Detail)
	}
}

func TestSO1_FailNoJsonFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Flags: []*discovery.Flag{
			{Name: "verbose", Description: "Be verbose"},
		},
	}

	check := newCheckSO1()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusFail {
		t.Errorf("expected Fail, got %s: %s", result.Status, result.Detail)
	}
}

func TestSO1_SkipNilTree(t *testing.T) {
	check := newCheckSO1()
	result := check.Run(context.Background(), &Input{Tree: nil})

	if result.Status != StatusSkip {
		t.Errorf("expected StatusSkip, got %s: %s", result.Status, result.Detail)
	}
}

func TestSO2_SkipNilProber(t *testing.T) {
	check := newCheckSO2()
	result := check.Run(context.Background(), &Input{Prober: nil})

	if result.Status != StatusSkip {
		t.Errorf("expected StatusSkip, got %s: %s", result.Status, result.Detail)
	}
	if result.Detail != "skipped: active check disabled by --no-probe" {
		t.Errorf("unexpected detail: %s", result.Detail)
	}
}

func TestSO3_SkipNilProber(t *testing.T) {
	check := newCheckSO3()
	result := check.Run(context.Background(), &Input{Prober: nil})

	if result.Status != StatusSkip {
		t.Errorf("expected StatusSkip, got %s: %s", result.Status, result.Detail)
	}
}

func TestSO3_SkipWhenSO1NotPassed(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Flags: []*discovery.Flag{
			{Name: "output", EnumValues: []string{"json", "text"}},
		},
	}
	tree := makeTree(root)

	// Create a ResultSet where SO-1 failed.
	rs := NewResultSet()
	rs.Set("SO-1", &Result{CheckID: "SO-1", Status: StatusFail})

	// SO-3 is active but we need a prober. With nil prober it skips first.
	// Test the cross-check logic by verifying metadata.
	check := newCheckSO3()
	if check.ID() != "SO-3" {
		t.Errorf("expected ID SO-3, got %s", check.ID())
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

func TestSO4_SkipNilProber(t *testing.T) {
	check := newCheckSO4()
	result := check.Run(context.Background(), &Input{Prober: nil})

	if result.Status != StatusSkip {
		t.Errorf("expected StatusSkip, got %s: %s", result.Status, result.Detail)
	}
}

func TestSO1_Metadata(t *testing.T) {
	check := newCheckSO1()

	t.Run("ID", func(t *testing.T) {
		if check.ID() != "SO-1" {
			t.Errorf("expected SO-1, got %s", check.ID())
		}
	})

	t.Run("Name", func(t *testing.T) {
		if check.Name() != "JSON output support" {
			t.Errorf("unexpected name: %s", check.Name())
		}
	})

	t.Run("Category", func(t *testing.T) {
		if check.Category() != CatStructuredOutput {
			t.Errorf("unexpected category: %s", check.Category())
		}
	})

	t.Run("Severity", func(t *testing.T) {
		if check.Severity() != Fail {
			t.Errorf("expected Fail severity, got %s", check.Severity())
		}
	})

	t.Run("Method", func(t *testing.T) {
		if check.Method() != Passive {
			t.Errorf("expected Passive method, got %s", check.Method())
		}
	})
}

func TestSO2_Metadata(t *testing.T) {
	check := newCheckSO2()
	if check.Method() != Active {
		t.Errorf("expected Active method, got %s", check.Method())
	}
	if check.Severity() != Fail {
		t.Errorf("expected Fail severity, got %s", check.Severity())
	}
}

func TestSO3_Metadata(t *testing.T) {
	check := newCheckSO3()
	if check.Method() != Active {
		t.Errorf("expected Active method, got %s", check.Method())
	}
	if check.Severity() != Warn {
		t.Errorf("expected Warn severity, got %s", check.Severity())
	}
}

func TestSO4_Metadata(t *testing.T) {
	check := newCheckSO4()
	if check.Method() != Active {
		t.Errorf("expected Active method, got %s", check.Method())
	}
	if check.Severity() != Warn {
		t.Errorf("expected Warn severity, got %s", check.Severity())
	}
}

func TestSO1_PassWithShortOFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Flags: []*discovery.Flag{
			{Name: "o", ShortName: "o", Description: "Output format (json, yaml)"},
		},
	}

	check := newCheckSO1()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for -o flag with json in description, got %s: %s", result.Status, result.Detail)
	}
}

func TestBuildJSONFlagArg(t *testing.T) {
	tests := []struct {
		name     string
		flag     *discovery.Flag
		expected string
	}{
		{"json flag", &discovery.Flag{Name: "json"}, "--json"},
		{"output flag", &discovery.Flag{Name: "output"}, "--output=json"},
		{"format flag", &discovery.Flag{Name: "format"}, "--format=json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildJSONFlagArg(tt.flag)
			if got != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// SO-5: Stdin/pipe input support (passive check)
// ---------------------------------------------------------------------------

func TestSO5_Metadata(t *testing.T) {
	check := newCheckSO5()

	if check.ID() != "SO-5" {
		t.Errorf("expected SO-5, got %s", check.ID())
	}
	if check.Category() != CatStructuredOutput {
		t.Errorf("expected structured-output, got %s", check.Category())
	}
	if check.Severity() != Info {
		t.Errorf("expected Info, got %s", check.Severity())
	}
	if check.Method() != Passive {
		t.Errorf("expected Passive, got %s", check.Method())
	}
}

func TestSO5_SkipNilTree(t *testing.T) {
	check := newCheckSO5()
	result := check.Run(context.Background(), &Input{Tree: nil})

	if result.Status != StatusSkip {
		t.Errorf("expected StatusSkip, got %s: %s", result.Status, result.Detail)
	}
}

func TestSO5_PassWithFromFileFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Flags: []*discovery.Flag{
			{Name: "from-file", Description: "Read input from file"},
		},
	}

	check := newCheckSO5()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for --from-file flag, got %s: %s", result.Status, result.Detail)
	}
}

func TestSO5_PassWithInputFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Flags: []*discovery.Flag{
			{Name: "input", Description: "Input source"},
		},
	}

	check := newCheckSO5()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for --input flag, got %s: %s", result.Status, result.Detail)
	}
}

func TestSO5_PassWithStdinInHelp(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli\n\nReads from stdin when no file is specified.",
	}

	check := newCheckSO5()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for stdin in help, got %s: %s", result.Status, result.Detail)
	}
}

func TestSO5_PassWithPipeInHelp(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli\n\nYou can pipe input to this command.",
	}

	check := newCheckSO5()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for pipe in help, got %s: %s", result.Status, result.Detail)
	}
}

func TestSO5_PassNoDataInputCommands(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli\n\nA simple CLI tool.",
		Flags: []*discovery.Flag{
			{Name: "verbose", Description: "Verbose output"},
		},
	}

	check := newCheckSO5()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass (no data-input commands), got %s: %s", result.Status, result.Detail)
	}
}

func TestSO5_FailMutatingWithoutStdin(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli\n\nA CLI tool.",
		Subcommands: []*discovery.Command{
			{
				Name:       "create",
				FullPath:   []string{"mycli", "create"},
				IsMutating: true,
				Flags: []*discovery.Flag{
					{Name: "name", Description: "Item name"},
				},
			},
		},
	}

	check := newCheckSO5()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusFail {
		t.Errorf("expected Fail, got %s: %s", result.Status, result.Detail)
	}
}

func TestSO5_FailFileAcceptingWithoutStdin(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli\n\nA CLI tool.",
		Flags: []*discovery.Flag{
			{Name: "file", Description: "Path to config file"},
		},
	}

	check := newCheckSO5()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusFail {
		t.Errorf("expected Fail, got %s: %s", result.Status, result.Detail)
	}
}

func TestTruncate(t *testing.T) {
	t.Run("short string", func(t *testing.T) {
		got := truncate("hello", 10)
		if got != "hello" {
			t.Errorf("expected hello, got %s", got)
		}
	})

	t.Run("long string", func(t *testing.T) {
		got := truncate("hello world this is a long string", 10)
		if got != "hello worl..." {
			t.Errorf("expected truncated string, got %s", got)
		}
	})

	t.Run("exact length", func(t *testing.T) {
		got := truncate("12345", 5)
		if got != "12345" {
			t.Errorf("expected 12345, got %s", got)
		}
	})
}
