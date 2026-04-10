package checks

import (
	"context"
	"testing"

	"github.com/Camil-H/cli-agent-lint/discovery"
)

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

// TE-3: --no-color flag (passive check)

func TestTE3_PassWithNoColorFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Flags: []*discovery.Flag{
			{Name: "no-color", Description: "Disable color output"},
		},
	}

	r := DefaultRegistry()
	check := r.Get("TE-3")
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for --no-color flag, got %s: %s", result.Status, result.Detail)
	}
}

func TestTE3_PassWithColorFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Flags: []*discovery.Flag{
			{Name: "color", Description: "Control color output (auto, always, never)"},
		},
	}

	r := DefaultRegistry()
	check := r.Get("TE-3")
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for --color flag, got %s: %s", result.Status, result.Detail)
	}
}

func TestTE3_PassWithNoColorInHelp(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli [options]\n\nSet NO_COLOR=1 to disable colors.",
	}

	r := DefaultRegistry()
	check := r.Get("TE-3")
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for NO_COLOR in help, got %s: %s", result.Status, result.Detail)
	}
}

func TestTE3_PassWithColorNeverInHelp(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli [options]\n\n  --color=never  Disable ANSI colors",
	}

	r := DefaultRegistry()
	check := r.Get("TE-3")
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for --color=never in help, got %s: %s", result.Status, result.Detail)
	}
}

func TestTE3_FailNoColorSupport(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli [options]\n\nA simple CLI tool.",
		Flags: []*discovery.Flag{
			{Name: "verbose", Description: "Be verbose"},
		},
	}

	r := DefaultRegistry()
	check := r.Get("TE-3")
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusFail {
		t.Errorf("expected Fail, got %s: %s", result.Status, result.Detail)
	}
}

func TestTE3_Metadata(t *testing.T) {
	r := DefaultRegistry()
	check := r.Get("TE-3")

	if check.Method() != Passive {
		t.Errorf("expected Passive, got %s", check.Method())
	}
	if check.Severity() != Warn {
		t.Errorf("expected Warn, got %s", check.Severity())
	}
}

// TE-4: --quiet / --silent flag (passive check)

func TestTE4_PassWithQuietFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Flags: []*discovery.Flag{
			{Name: "quiet", Description: "Suppress output"},
		},
	}

	r := DefaultRegistry()
	check := r.Get("TE-4")
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for --quiet flag, got %s: %s", result.Status, result.Detail)
	}
}

func TestTE4_PassWithSilentFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Flags: []*discovery.Flag{
			{Name: "silent", Description: "Suppress output"},
		},
	}

	r := DefaultRegistry()
	check := r.Get("TE-4")
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for --silent flag, got %s: %s", result.Status, result.Detail)
	}
}

func TestTE4_PassWithShortQFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Flags: []*discovery.Flag{
			{Name: "q", ShortName: "q", Description: "Quiet mode"},
		},
	}

	r := DefaultRegistry()
	check := r.Get("TE-4")
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for -q flag, got %s: %s", result.Status, result.Detail)
	}
}

func TestTE4_PassWithQuietInHelp(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli [options]\n\n  --quiet       Suppress output",
	}

	r := DefaultRegistry()
	check := r.Get("TE-4")
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for --quiet in help, got %s: %s", result.Status, result.Detail)
	}
}

func TestTE4_FailNoQuietSupport(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli [options]\n\nA CLI tool.",
		Flags: []*discovery.Flag{
			{Name: "verbose", Description: "Be verbose"},
		},
	}

	r := DefaultRegistry()
	check := r.Get("TE-4")
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusFail {
		t.Errorf("expected Fail, got %s: %s", result.Status, result.Detail)
	}
}

func TestTE4_Metadata(t *testing.T) {
	r := DefaultRegistry()
	check := r.Get("TE-4")

	if check.Method() != Passive {
		t.Errorf("expected Passive, got %s", check.Method())
	}
	if check.Severity() != Info {
		t.Errorf("expected Info, got %s", check.Severity())
	}
}

// TE-5: Pagination support (passive check)

func TestTE5_PassNoListCommands(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Subcommands: []*discovery.Command{
			{Name: "create", FullPath: []string{"mycli", "create"}, IsMutating: true},
			{Name: "delete", FullPath: []string{"mycli", "delete"}, IsMutating: true},
		},
	}

	check := newCheckTE5()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass (no list commands), got %s: %s", result.Status, result.Detail)
	}
}

func TestTE5_PassAllListCommandsHavePagination(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Subcommands: []*discovery.Command{
			{
				Name:       "list",
				FullPath:   []string{"mycli", "list"},
				IsListLike: true,
				Flags: []*discovery.Flag{
					{Name: "page-size", Description: "Number of items per page"},
					{Name: "page", Description: "Page number"},
				},
			},
			{
				Name:       "search",
				FullPath:   []string{"mycli", "search"},
				IsListLike: true,
				Flags: []*discovery.Flag{
					{Name: "limit", Description: "Maximum number of results"},
				},
			},
		},
	}

	check := newCheckTE5()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass, got %s: %s", result.Status, result.Detail)
	}
}

func TestTE5_FailMissingPagination(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Subcommands: []*discovery.Command{
			{
				Name:       "list",
				FullPath:   []string{"mycli", "list"},
				IsListLike: true,
				Flags: []*discovery.Flag{
					{Name: "limit", Description: "Max results"},
				},
			},
			{
				Name:       "search",
				FullPath:   []string{"mycli", "search"},
				IsListLike: true,
				Flags: []*discovery.Flag{
					{Name: "verbose", Description: "Verbose output"},
				},
			},
		},
	}

	check := newCheckTE5()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusFail {
		t.Errorf("expected Fail (search missing pagination), got %s: %s", result.Status, result.Detail)
	}
}

func TestTE5_PassWithVariousPaginationFlags(t *testing.T) {
	paginationFlags := []string{"page-size", "page", "per-page", "limit", "cursor", "offset", "page-all", "paginate", "all"}

	for _, flagName := range paginationFlags {
		t.Run(flagName, func(t *testing.T) {
			root := &discovery.Command{
				Name:     "mycli",
				FullPath: []string{"mycli"},
				Subcommands: []*discovery.Command{
					{
						Name:       "list",
						FullPath:   []string{"mycli", "list"},
						IsListLike: true,
						Flags: []*discovery.Flag{
							{Name: flagName, Description: "Pagination flag"},
						},
					},
				},
			}

			check := newCheckTE5()
			result := check.Run(context.Background(), makeInput(root))

			if result.Status != StatusPass {
				t.Errorf("expected StatusPass for --%s, got %s: %s", flagName, result.Status, result.Detail)
			}
		})
	}
}

func TestTE5_SkipNilTree(t *testing.T) {
	check := newCheckTE5()
	result := check.Run(context.Background(), &Input{Tree: nil})

	if result.Status != StatusSkip {
		t.Errorf("expected StatusSkip for nil tree, got %s: %s", result.Status, result.Detail)
	}
}

func TestTE5_Metadata(t *testing.T) {
	check := newCheckTE5()

	if check.ID() != "TE-5" {
		t.Errorf("expected TE-5, got %s", check.ID())
	}
	if check.Severity() != Warn {
		t.Errorf("expected Warn, got %s", check.Severity())
	}
	if check.Method() != Passive {
		t.Errorf("expected Passive, got %s", check.Method())
	}
}

// TE-6: Field masks / response filtering (passive check)

func TestTE6_PassWithFieldsFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Flags: []*discovery.Flag{
			{Name: "fields", Description: "Comma-separated list of fields to return"},
		},
	}

	check := newCheckTE6()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for --fields flag, got %s: %s", result.Status, result.Detail)
	}
}

func TestTE6_PassWithJqFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Flags: []*discovery.Flag{
			{Name: "jq", Description: "jq expression to filter output"},
		},
	}

	check := newCheckTE6()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for --jq flag, got %s: %s", result.Status, result.Detail)
	}
}

func TestTE6_PassWithFilterFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Flags: []*discovery.Flag{
			{Name: "filter", Description: "Filter output by expression"},
		},
	}

	check := newCheckTE6()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for --filter flag, got %s: %s", result.Status, result.Detail)
	}
}

func TestTE6_PassWithSelectFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Flags: []*discovery.Flag{
			{Name: "select", Description: "Select specific fields"},
		},
	}

	check := newCheckTE6()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for --select flag, got %s: %s", result.Status, result.Detail)
	}
}

func TestTE6_PassWithColumnsFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Flags: []*discovery.Flag{
			{Name: "columns", Description: "Columns to display"},
		},
	}

	check := newCheckTE6()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for --columns flag, got %s: %s", result.Status, result.Detail)
	}
}

func TestTE6_PassWithQueryFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Flags: []*discovery.Flag{
			{Name: "query", Description: "JMESPath query"},
		},
	}

	check := newCheckTE6()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for --query flag, got %s: %s", result.Status, result.Detail)
	}
}

func TestTE6_PassWithFieldFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Flags: []*discovery.Flag{
			{Name: "field", Description: "Single field to extract"},
		},
	}

	check := newCheckTE6()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for --field flag, got %s: %s", result.Status, result.Detail)
	}
}

func TestTE6_PassWithFilterFlagOnSubcommand(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Subcommands: []*discovery.Command{
			{
				Name:     "list",
				FullPath: []string{"mycli", "list"},
				Flags: []*discovery.Flag{
					{Name: "jq", Description: "Filter with jq"},
				},
			},
		},
	}

	check := newCheckTE6()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for --jq on subcommand, got %s: %s", result.Status, result.Detail)
	}
}

func TestTE6_PassWithFilterInHelp(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli [options]\n\n  --fields  Comma-separated list of fields to include",
	}

	check := newCheckTE6()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for --fields in help, got %s: %s", result.Status, result.Detail)
	}
}

func TestTE6_PassNoListCommands(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli\n\nA CLI tool.",
		Flags: []*discovery.Flag{
			{Name: "verbose", Description: "Verbose output"},
		},
	}

	check := newCheckTE6()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass (no list commands), got %s: %s", result.Status, result.Detail)
	}
}

func TestTE6_FailListCommandNoFilter(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli\n\nA CLI tool.",
		Subcommands: []*discovery.Command{
			{
				Name:       "list",
				FullPath:   []string{"mycli", "list"},
				RawHelp:    "List resources",
				IsListLike: true,
			},
		},
	}

	check := newCheckTE6()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusFail {
		t.Errorf("expected Fail (list command without filter), got %s: %s", result.Status, result.Detail)
	}
}

func TestTE6_SkipNilTree(t *testing.T) {
	check := newCheckTE6()
	result := check.Run(context.Background(), &Input{Tree: nil})

	if result.Status != StatusSkip {
		t.Errorf("expected StatusSkip for nil tree, got %s: %s", result.Status, result.Detail)
	}
}

func TestTE6_Metadata(t *testing.T) {
	check := newCheckTE6()

	if check.ID() != "TE-6" {
		t.Errorf("expected TE-6, got %s", check.ID())
	}
	if check.Severity() != Info {
		t.Errorf("expected Info, got %s", check.Severity())
	}
	if check.Method() != Passive {
		t.Errorf("expected Passive, got %s", check.Method())
	}
	if check.Category() != CatTokenEfficiency {
		t.Errorf("expected token-efficiency, got %s", check.Category())
	}
}

// TE-7: Help output size

func TestTE7_PassSmallHelp(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli [command]\n\nAvailable commands:\n  list\n  create\n",
	}
	check := newCheckTE7()
	result := check.Run(context.Background(), makeInput(root))
	if result.Status != StatusPass {
		t.Errorf("expected pass for small help, got %s: %s", result.Status, result.Detail)
	}
}

func TestTE7_FailLargeHelp(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  string(make([]byte, 101*1024)), // 101 KB
	}
	check := newCheckTE7()
	result := check.Run(context.Background(), makeInput(root))
	if result.Status != StatusFail {
		t.Errorf("expected fail for 101KB help, got %s: %s", result.Status, result.Detail)
	}
}

func TestTE7_WarnMediumHelp(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  string(make([]byte, 50*1024)), // 50 KB — between warn and fail thresholds
	}
	check := newCheckTE7()
	result := check.Run(context.Background(), makeInput(root))
	if result.Status != StatusFail {
		t.Errorf("expected fail (warn-severity) for 50KB help, got %s: %s", result.Status, result.Detail)
	}
}

func TestTE7_SkipNilTree(t *testing.T) {
	check := newCheckTE7()
	result := check.Run(context.Background(), &Input{Tree: nil})
	if result.Status != StatusSkip {
		t.Errorf("expected skip, got %s", result.Status)
	}
}

// TE-8: Concise output mode

func TestTE8_PassWithBriefFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Flags:    []*discovery.Flag{{Name: "brief", Description: "Brief output"}},
		Subcommands: []*discovery.Command{
			{Name: "list", FullPath: []string{"mycli", "list"}},
		},
	}
	check := newCheckTE8()
	result := check.Run(context.Background(), makeInput(root))
	if result.Status != StatusPass {
		t.Errorf("expected pass for --brief flag, got %s: %s", result.Status, result.Detail)
	}
}

func TestTE8_PassWithFormatShortEnum(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Flags: []*discovery.Flag{
			{Name: "output", Description: "Output format", EnumValues: []string{"json", "table", "short"}},
		},
		Subcommands: []*discovery.Command{
			{Name: "list", FullPath: []string{"mycli", "list"}},
		},
	}
	check := newCheckTE8()
	result := check.Run(context.Background(), makeInput(root))
	if result.Status != StatusPass {
		t.Errorf("expected pass for --output=short enum value, got %s: %s", result.Status, result.Detail)
	}
}

func TestTE8_PassNoDataCommands(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
	}
	check := newCheckTE8()
	result := check.Run(context.Background(), makeInput(root))
	if result.Status != StatusPass {
		t.Errorf("expected pass (no data commands), got %s: %s", result.Status, result.Detail)
	}
}

func TestTE8_FailNoConciseFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Subcommands: []*discovery.Command{
			{Name: "list", FullPath: []string{"mycli", "list"}, IsListLike: true},
		},
	}
	check := newCheckTE8()
	result := check.Run(context.Background(), makeInput(root))
	if result.Status != StatusFail {
		t.Errorf("expected fail for missing concise flag, got %s: %s", result.Status, result.Detail)
	}
}
