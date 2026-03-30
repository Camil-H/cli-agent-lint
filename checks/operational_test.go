package checks

import (
	"context"
	"testing"

	"github.com/cli-agent-lint/cli-agent-lint/discovery"
)

// ---------------------------------------------------------------------------
// OR-1: Exit codes (active check)
// ---------------------------------------------------------------------------

func TestOR1_SkipNilProber(t *testing.T) {
	check := newCheckOR1()
	result := check.Run(context.Background(), &Input{Prober: nil})

	if result.Status != StatusSkip {
		t.Errorf("expected StatusSkip for nil prober, got %s: %s", result.Status, result.Detail)
	}
	if result.Detail != "skipped: active check disabled by --no-probe" {
		t.Errorf("unexpected detail: %s", result.Detail)
	}
}

func TestOR1_Metadata(t *testing.T) {
	check := newCheckOR1()

	t.Run("ID", func(t *testing.T) {
		if check.ID() != "OR-1" {
			t.Errorf("expected OR-1, got %s", check.ID())
		}
	})

	t.Run("Category", func(t *testing.T) {
		if check.Category() != CatOperational {
			t.Errorf("expected operational, got %s", check.Category())
		}
	})

	t.Run("Severity", func(t *testing.T) {
		if check.Severity() != Fail {
			t.Errorf("expected Fail, got %s", check.Severity())
		}
	})

	t.Run("Method", func(t *testing.T) {
		if check.Method() != Active {
			t.Errorf("expected Active, got %s", check.Method())
		}
	})
}

// ---------------------------------------------------------------------------
// OR-2: Timeout flag (passive check)
// ---------------------------------------------------------------------------

func TestOR2_PassWithTimeoutFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Flags: []*discovery.Flag{
			{Name: "timeout", Description: "Request timeout in seconds"},
		},
	}

	check := newCheckOR2()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for --timeout flag, got %s: %s", result.Status, result.Detail)
	}
}

func TestOR2_PassWithRequestTimeoutFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Flags: []*discovery.Flag{
			{Name: "request-timeout", Description: "Request timeout duration"},
		},
	}

	check := newCheckOR2()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for --request-timeout flag, got %s: %s", result.Status, result.Detail)
	}
}

func TestOR2_PassWithTimeoutFlagOnSubcommand(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Subcommands: []*discovery.Command{
			{
				Name:     "query",
				FullPath: []string{"mycli", "query"},
				Flags: []*discovery.Flag{
					{Name: "timeout", Description: "Query timeout"},
				},
			},
		},
	}

	check := newCheckOR2()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for --timeout on subcommand, got %s: %s", result.Status, result.Detail)
	}
}

func TestOR2_PassWithTimeoutInHelp(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli [options]\n\n  --timeout duration  Set request timeout",
	}

	check := newCheckOR2()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for --timeout in help text, got %s: %s", result.Status, result.Detail)
	}
}

func TestOR2_FailNoTimeoutFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli [options]\n\nA CLI tool.",
		Flags: []*discovery.Flag{
			{Name: "verbose", Description: "Verbose output"},
		},
	}

	check := newCheckOR2()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusFail {
		t.Errorf("expected Fail, got %s: %s", result.Status, result.Detail)
	}
}

func TestOR2_SkipNilTree(t *testing.T) {
	check := newCheckOR2()
	result := check.Run(context.Background(), &Input{Tree: nil})

	if result.Status != StatusSkip {
		t.Errorf("expected StatusSkip for nil tree, got %s: %s", result.Status, result.Detail)
	}
}

func TestOR2_Metadata(t *testing.T) {
	check := newCheckOR2()

	if check.Severity() != Info {
		t.Errorf("expected Info, got %s", check.Severity())
	}
	if check.Method() != Passive {
		t.Errorf("expected Passive, got %s", check.Method())
	}
}

// ---------------------------------------------------------------------------
// OR-3: Pagination support (passive check)
// ---------------------------------------------------------------------------

func TestOR3_PassNoListCommands(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Subcommands: []*discovery.Command{
			{Name: "create", FullPath: []string{"mycli", "create"}, IsMutating: true},
			{Name: "delete", FullPath: []string{"mycli", "delete"}, IsMutating: true},
		},
	}

	check := newCheckOR3()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass (no list commands), got %s: %s", result.Status, result.Detail)
	}
}

func TestOR3_PassAllListCommandsHavePagination(t *testing.T) {
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

	check := newCheckOR3()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass, got %s: %s", result.Status, result.Detail)
	}
}

func TestOR3_FailMissingPagination(t *testing.T) {
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

	check := newCheckOR3()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusFail {
		t.Errorf("expected Fail (search missing pagination), got %s: %s", result.Status, result.Detail)
	}
}

func TestOR3_PassWithVariousPaginationFlags(t *testing.T) {
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

			check := newCheckOR3()
			result := check.Run(context.Background(), makeInput(root))

			if result.Status != StatusPass {
				t.Errorf("expected StatusPass for --%s, got %s: %s", flagName, result.Status, result.Detail)
			}
		})
	}
}

func TestOR3_SkipNilTree(t *testing.T) {
	check := newCheckOR3()
	result := check.Run(context.Background(), &Input{Tree: nil})

	if result.Status != StatusSkip {
		t.Errorf("expected StatusSkip for nil tree, got %s: %s", result.Status, result.Detail)
	}
}

func TestOR3_Metadata(t *testing.T) {
	check := newCheckOR3()

	if check.ID() != "OR-3" {
		t.Errorf("expected OR-3, got %s", check.ID())
	}
	if check.Severity() != Warn {
		t.Errorf("expected Warn, got %s", check.Severity())
	}
	if check.Method() != Passive {
		t.Errorf("expected Passive, got %s", check.Method())
	}
}

// ---------------------------------------------------------------------------
// OR-4: Retry / rate-limit hints (passive check)
// ---------------------------------------------------------------------------

func TestOR4_PassWithRetryFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Flags: []*discovery.Flag{
			{Name: "retry", Description: "Number of retries"},
		},
	}

	check := newCheckOR4()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for --retry flag, got %s: %s", result.Status, result.Detail)
	}
}

func TestOR4_PassWithMaxRetriesFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Flags: []*discovery.Flag{
			{Name: "max-retries", Description: "Maximum number of retries"},
		},
	}

	check := newCheckOR4()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for --max-retries flag, got %s: %s", result.Status, result.Detail)
	}
}

func TestOR4_PassWithRetryInHelp(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli\n\nWill retry failed requests up to 3 times.",
	}

	check := newCheckOR4()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for retry in help, got %s: %s", result.Status, result.Detail)
	}
}

func TestOR4_PassWithRateLimitInHelp(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli\n\nRespects rate-limit headers from the server.",
	}

	check := newCheckOR4()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for rate-limit in help, got %s: %s", result.Status, result.Detail)
	}
}

func TestOR4_PassWithThrottleInHelp(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli\n\nThrottle requests to avoid API limits.",
	}

	check := newCheckOR4()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for throttle in help, got %s: %s", result.Status, result.Detail)
	}
}

func TestOR4_PassWithRetryFlagOnSubcommand(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Subcommands: []*discovery.Command{
			{
				Name:     "fetch",
				FullPath: []string{"mycli", "fetch"},
				Flags: []*discovery.Flag{
					{Name: "retry-count", Description: "Number of retry attempts"},
				},
			},
		},
	}

	check := newCheckOR4()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for --retry-count on subcommand, got %s: %s", result.Status, result.Detail)
	}
}

func TestOR4_PassNoNetworkCommands(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli\n\nA CLI tool.",
		Flags: []*discovery.Flag{
			{Name: "verbose", Description: "Verbose output"},
		},
	}

	check := newCheckOR4()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass (no network commands), got %s: %s", result.Status, result.Detail)
	}
}

func TestOR4_FailNetworkCLINoRetry(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli\n\nA CLI tool for https://api.example.com.",
		Flags: []*discovery.Flag{
			{Name: "verbose", Description: "Verbose output"},
		},
	}

	check := newCheckOR4()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusFail {
		t.Errorf("expected Fail (network CLI without retry), got %s: %s", result.Status, result.Detail)
	}
}

func TestOR4_SkipNilTree(t *testing.T) {
	check := newCheckOR4()
	result := check.Run(context.Background(), &Input{Tree: nil})

	if result.Status != StatusSkip {
		t.Errorf("expected StatusSkip for nil tree, got %s: %s", result.Status, result.Detail)
	}
}

func TestOR4_Metadata(t *testing.T) {
	check := newCheckOR4()

	if check.ID() != "OR-4" {
		t.Errorf("expected OR-4, got %s", check.ID())
	}
	if check.Severity() != Info {
		t.Errorf("expected Info, got %s", check.Severity())
	}
	if check.Method() != Passive {
		t.Errorf("expected Passive, got %s", check.Method())
	}
}

// ---------------------------------------------------------------------------
// OR-5: Deterministic output (active check)
// ---------------------------------------------------------------------------

func TestOR5_SkipNilProber(t *testing.T) {
	check := newCheckOR5()
	result := check.Run(context.Background(), &Input{Prober: nil})

	if result.Status != StatusSkip {
		t.Errorf("expected StatusSkip for nil prober, got %s: %s", result.Status, result.Detail)
	}
}

func TestOR5_Metadata(t *testing.T) {
	check := newCheckOR5()

	if check.ID() != "OR-5" {
		t.Errorf("expected OR-5, got %s", check.ID())
	}
	if check.Severity() != Info {
		t.Errorf("expected Info, got %s", check.Severity())
	}
	if check.Method() != Active {
		t.Errorf("expected Active, got %s", check.Method())
	}
	if check.Category() != CatOperational {
		t.Errorf("expected operational, got %s", check.Category())
	}
}

// ---------------------------------------------------------------------------
// OR-6: Field masks / response filtering (passive check)
// ---------------------------------------------------------------------------

func TestOR6_PassWithFieldsFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Flags: []*discovery.Flag{
			{Name: "fields", Description: "Comma-separated list of fields to return"},
		},
	}

	check := newCheckOR6()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for --fields flag, got %s: %s", result.Status, result.Detail)
	}
}

func TestOR6_PassWithJqFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Flags: []*discovery.Flag{
			{Name: "jq", Description: "jq expression to filter output"},
		},
	}

	check := newCheckOR6()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for --jq flag, got %s: %s", result.Status, result.Detail)
	}
}

func TestOR6_PassWithFilterFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Flags: []*discovery.Flag{
			{Name: "filter", Description: "Filter output by expression"},
		},
	}

	check := newCheckOR6()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for --filter flag, got %s: %s", result.Status, result.Detail)
	}
}

func TestOR6_PassWithSelectFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Flags: []*discovery.Flag{
			{Name: "select", Description: "Select specific fields"},
		},
	}

	check := newCheckOR6()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for --select flag, got %s: %s", result.Status, result.Detail)
	}
}

func TestOR6_PassWithColumnsFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Flags: []*discovery.Flag{
			{Name: "columns", Description: "Columns to display"},
		},
	}

	check := newCheckOR6()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for --columns flag, got %s: %s", result.Status, result.Detail)
	}
}

func TestOR6_PassWithQueryFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Flags: []*discovery.Flag{
			{Name: "query", Description: "JMESPath query"},
		},
	}

	check := newCheckOR6()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for --query flag, got %s: %s", result.Status, result.Detail)
	}
}

func TestOR6_PassWithFieldFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Flags: []*discovery.Flag{
			{Name: "field", Description: "Single field to extract"},
		},
	}

	check := newCheckOR6()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for --field flag, got %s: %s", result.Status, result.Detail)
	}
}

func TestOR6_PassWithFilterFlagOnSubcommand(t *testing.T) {
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

	check := newCheckOR6()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for --jq on subcommand, got %s: %s", result.Status, result.Detail)
	}
}

func TestOR6_PassWithFilterInHelp(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli [options]\n\n  --fields  Comma-separated list of fields to include",
	}

	check := newCheckOR6()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for --fields in help, got %s: %s", result.Status, result.Detail)
	}
}

func TestOR6_PassNoListCommands(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli\n\nA CLI tool.",
		Flags: []*discovery.Flag{
			{Name: "verbose", Description: "Verbose output"},
		},
	}

	check := newCheckOR6()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass (no list commands), got %s: %s", result.Status, result.Detail)
	}
}

func TestOR6_FailListCommandNoFilter(t *testing.T) {
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

	check := newCheckOR6()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusFail {
		t.Errorf("expected Fail (list command without filter), got %s: %s", result.Status, result.Detail)
	}
}

func TestOR6_SkipNilTree(t *testing.T) {
	check := newCheckOR6()
	result := check.Run(context.Background(), &Input{Tree: nil})

	if result.Status != StatusSkip {
		t.Errorf("expected StatusSkip for nil tree, got %s: %s", result.Status, result.Detail)
	}
}

func TestOR6_Metadata(t *testing.T) {
	check := newCheckOR6()

	if check.ID() != "OR-6" {
		t.Errorf("expected OR-6, got %s", check.ID())
	}
	if check.Severity() != Info {
		t.Errorf("expected Info, got %s", check.Severity())
	}
	if check.Method() != Passive {
		t.Errorf("expected Passive, got %s", check.Method())
	}
	if check.Category() != CatOperational {
		t.Errorf("expected operational, got %s", check.Category())
	}
}

// ---------------------------------------------------------------------------
// OR-7: Distinct exit codes for error classes (passive check)
// ---------------------------------------------------------------------------

func TestOR7_Metadata(t *testing.T) {
	check := newCheckOR7()

	if check.ID() != "OR-7" {
		t.Errorf("expected OR-7, got %s", check.ID())
	}
	if check.Category() != CatOperational {
		t.Errorf("expected operational, got %s", check.Category())
	}
	if check.Severity() != Info {
		t.Errorf("expected Info, got %s", check.Severity())
	}
	if check.Method() != Passive {
		t.Errorf("expected Passive, got %s", check.Method())
	}
}

func TestOR7_SkipNilTree(t *testing.T) {
	check := newCheckOR7()
	result := check.Run(context.Background(), &Input{Tree: nil})

	if result.Status != StatusSkip {
		t.Errorf("expected StatusSkip, got %s: %s", result.Status, result.Detail)
	}
}

func TestOR7_PassWithExitCodesSection(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli\n\nEXIT CODES\n  0  Success\n  1  General error\n  2  Usage error\n",
	}

	check := newCheckOR7()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for EXIT CODES section, got %s: %s", result.Status, result.Detail)
	}
}

func TestOR7_PassWithExitStatusSection(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli\n\nEXIT STATUS\n  0  OK\n  1  Error\n",
	}

	check := newCheckOR7()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for EXIT STATUS section, got %s: %s", result.Status, result.Detail)
	}
}

func TestOR7_PassWithExitCodeMentionInHelp(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli\n\nReturns exit code 1 on failure.",
	}

	check := newCheckOR7()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for exit code mention, got %s: %s", result.Status, result.Detail)
	}
}

func TestOR7_PassWithReturnCodeMention(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli\n\nThe return code indicates the result.",
	}

	check := newCheckOR7()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for return code mention, got %s: %s", result.Status, result.Detail)
	}
}

func TestOR7_FailNoDocumentation(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli\n\nA simple CLI tool.",
		Flags: []*discovery.Flag{
			{Name: "verbose", Description: "Verbose output"},
		},
	}

	check := newCheckOR7()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusFail {
		t.Errorf("expected Fail, got %s: %s", result.Status, result.Detail)
	}
}
