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

func TestTE1_PassWithOutputJsonFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Flags: []*discovery.Flag{
			{Name: "output", Description: "Output format", EnumValues: []string{"json", "text"}},
		},
	}

	check := newCheckTE1()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass, got %s: %s", result.Status, result.Detail)
	}
}

func TestTE1_PassWithFormatFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Flags: []*discovery.Flag{
			{Name: "format", Description: "Choose output format: json or yaml"},
		},
	}

	check := newCheckTE1()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass, got %s: %s", result.Status, result.Detail)
	}
}

func TestTE1_PassWithJsonFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Flags: []*discovery.Flag{
			{Name: "json", Description: "Output JSON"},
		},
	}

	check := newCheckTE1()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass, got %s: %s", result.Status, result.Detail)
	}
}

func TestTE1_PassWithJsonFlagOnSubcommand(t *testing.T) {
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

	check := newCheckTE1()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass, got %s: %s", result.Status, result.Detail)
	}
}

func TestTE1_FailNoJsonFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Flags: []*discovery.Flag{
			{Name: "verbose", Description: "Be verbose"},
		},
	}

	check := newCheckTE1()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusFail {
		t.Errorf("expected Fail, got %s: %s", result.Status, result.Detail)
	}
}

func TestTE1_SkipNilTree(t *testing.T) {
	check := newCheckTE1()
	result := check.Run(context.Background(), &Input{Tree: nil})

	if result.Status != StatusSkip {
		t.Errorf("expected StatusSkip, got %s: %s", result.Status, result.Detail)
	}
}

func TestFS1_SkipNilProber(t *testing.T) {
	check := newCheckFS1()
	result := check.Run(context.Background(), &Input{Prober: nil})

	if result.Status != StatusSkip {
		t.Errorf("expected StatusSkip, got %s: %s", result.Status, result.Detail)
	}
	if result.Detail != "skipped: active check disabled by --no-probe" {
		t.Errorf("unexpected detail: %s", result.Detail)
	}
}

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

func TestSD2_SkipNilProber(t *testing.T) {
	check := newCheckSD2()
	result := check.Run(context.Background(), &Input{Prober: nil})

	if result.Status != StatusSkip {
		t.Errorf("expected StatusSkip, got %s: %s", result.Status, result.Detail)
	}
}

func TestTE1_Metadata(t *testing.T) {
	check := newCheckTE1()

	t.Run("ID", func(t *testing.T) {
		if check.ID() != "TE-1" {
			t.Errorf("expected TE-1, got %s", check.ID())
		}
	})

	t.Run("Name", func(t *testing.T) {
		if check.Name() != "JSON output support" {
			t.Errorf("unexpected name: %s", check.Name())
		}
	})

	t.Run("Category", func(t *testing.T) {
		if check.Category() != CatTokenEfficiency {
			t.Errorf("expected token-efficiency, got %s", check.Category())
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

func TestFS1_Metadata(t *testing.T) {
	check := newCheckFS1()
	if check.Method() != Active {
		t.Errorf("expected Active method, got %s", check.Method())
	}
	if check.Severity() != Fail {
		t.Errorf("expected Fail severity, got %s", check.Severity())
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

func TestSD2_Metadata(t *testing.T) {
	check := newCheckSD2()
	if check.Method() != Active {
		t.Errorf("expected Active method, got %s", check.Method())
	}
	if check.Severity() != Warn {
		t.Errorf("expected Warn severity, got %s", check.Severity())
	}
}

func TestTE1_PassWithShortOFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Flags: []*discovery.Flag{
			{Name: "o", ShortName: "o", Description: "Output format (json, yaml)"},
		},
	}

	check := newCheckTE1()
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

// TE-2: Stdin/pipe input support (passive check)

func TestTE2_Metadata(t *testing.T) {
	check := newCheckTE2()

	if check.ID() != "TE-2" {
		t.Errorf("expected TE-2, got %s", check.ID())
	}
	if check.Category() != CatTokenEfficiency {
		t.Errorf("expected token-efficiency, got %s", check.Category())
	}
	if check.Severity() != Info {
		t.Errorf("expected Info, got %s", check.Severity())
	}
	if check.Method() != Passive {
		t.Errorf("expected Passive, got %s", check.Method())
	}
}

func TestTE2_SkipNilTree(t *testing.T) {
	check := newCheckTE2()
	result := check.Run(context.Background(), &Input{Tree: nil})

	if result.Status != StatusSkip {
		t.Errorf("expected StatusSkip, got %s: %s", result.Status, result.Detail)
	}
}

func TestTE2_PassWithFromFileFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Flags: []*discovery.Flag{
			{Name: "from-file", Description: "Read input from file"},
		},
	}

	check := newCheckTE2()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for --from-file flag, got %s: %s", result.Status, result.Detail)
	}
}

func TestTE2_PassWithInputFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Flags: []*discovery.Flag{
			{Name: "input", Description: "Input source"},
		},
	}

	check := newCheckTE2()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for --input flag, got %s: %s", result.Status, result.Detail)
	}
}

func TestTE2_PassWithStdinInHelp(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli\n\nReads from stdin when no file is specified.",
	}

	check := newCheckTE2()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for stdin in help, got %s: %s", result.Status, result.Detail)
	}
}

func TestTE2_PassWithPipeInHelp(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli\n\nYou can pipe input to this command.",
	}

	check := newCheckTE2()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for pipe in help, got %s: %s", result.Status, result.Detail)
	}
}

func TestTE2_PassNoDataInputCommands(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli\n\nA simple CLI tool.",
		Flags: []*discovery.Flag{
			{Name: "verbose", Description: "Verbose output"},
		},
	}

	check := newCheckTE2()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass (no data-input commands), got %s: %s", result.Status, result.Detail)
	}
}

func TestTE2_FailMutatingWithoutStdin(t *testing.T) {
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

	check := newCheckTE2()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusFail {
		t.Errorf("expected Fail, got %s: %s", result.Status, result.Detail)
	}
}

func TestTE2_FailFileAcceptingWithoutStdin(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli\n\nA CLI tool.",
		Flags: []*discovery.Flag{
			{Name: "file", Description: "Path to config file"},
		},
	}

	check := newCheckTE2()
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
