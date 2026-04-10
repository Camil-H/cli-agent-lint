package checks

import (
	"context"
	"testing"

	"github.com/Camil-H/cli-agent-lint/discovery"
)

// PV-1: Timeout flag (passive check)

func TestPV1_PassWithTimeoutFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Flags: []*discovery.Flag{
			{Name: "timeout", Description: "Request timeout in seconds"},
		},
	}

	check := newCheckPV1()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for --timeout flag, got %s: %s", result.Status, result.Detail)
	}
}

func TestPV1_PassWithRequestTimeoutFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Flags: []*discovery.Flag{
			{Name: "request-timeout", Description: "Request timeout duration"},
		},
	}

	check := newCheckPV1()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for --request-timeout flag, got %s: %s", result.Status, result.Detail)
	}
}

func TestPV1_PassWithTimeoutFlagOnSubcommand(t *testing.T) {
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

	check := newCheckPV1()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for --timeout on subcommand, got %s: %s", result.Status, result.Detail)
	}
}

func TestPV1_PassWithTimeoutInHelp(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli [options]\n\n  --timeout duration  Set request timeout",
	}

	check := newCheckPV1()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for --timeout in help text, got %s: %s", result.Status, result.Detail)
	}
}

func TestPV1_FailNoTimeoutFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli [options]\n\nA CLI tool.",
		Flags: []*discovery.Flag{
			{Name: "verbose", Description: "Verbose output"},
		},
	}

	check := newCheckPV1()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusFail {
		t.Errorf("expected Fail, got %s: %s", result.Status, result.Detail)
	}
}

func TestPV1_SkipNilTree(t *testing.T) {
	check := newCheckPV1()
	result := check.Run(context.Background(), &Input{Tree: nil})

	if result.Status != StatusSkip {
		t.Errorf("expected StatusSkip for nil tree, got %s: %s", result.Status, result.Detail)
	}
}

func TestPV1_Metadata(t *testing.T) {
	check := newCheckPV1()

	if check.Severity() != Info {
		t.Errorf("expected Info, got %s", check.Severity())
	}
	if check.Method() != Passive {
		t.Errorf("expected Passive, got %s", check.Method())
	}
}

// PV-2: Retry / rate-limit hints (passive check)

func TestPV2_PassWithRetryFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Flags: []*discovery.Flag{
			{Name: "retry", Description: "Number of retries"},
		},
	}

	check := newCheckPV2()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for --retry flag, got %s: %s", result.Status, result.Detail)
	}
}

func TestPV2_PassWithMaxRetriesFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Flags: []*discovery.Flag{
			{Name: "max-retries", Description: "Maximum number of retries"},
		},
	}

	check := newCheckPV2()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for --max-retries flag, got %s: %s", result.Status, result.Detail)
	}
}

func TestPV2_PassWithRetryInHelp(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli\n\nWill retry failed requests up to 3 times.",
	}

	check := newCheckPV2()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for retry in help, got %s: %s", result.Status, result.Detail)
	}
}

func TestPV2_PassWithRateLimitInHelp(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli\n\nRespects rate-limit headers from the server.",
	}

	check := newCheckPV2()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for rate-limit in help, got %s: %s", result.Status, result.Detail)
	}
}

func TestPV2_PassWithThrottleInHelp(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli\n\nThrottle requests to avoid API limits.",
	}

	check := newCheckPV2()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for throttle in help, got %s: %s", result.Status, result.Detail)
	}
}

func TestPV2_PassWithRetryFlagOnSubcommand(t *testing.T) {
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

	check := newCheckPV2()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for --retry-count on subcommand, got %s: %s", result.Status, result.Detail)
	}
}

func TestPV2_PassNoNetworkCommands(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli\n\nA CLI tool.",
		Flags: []*discovery.Flag{
			{Name: "verbose", Description: "Verbose output"},
		},
	}

	check := newCheckPV2()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass (no network commands), got %s: %s", result.Status, result.Detail)
	}
}

func TestPV2_FailNetworkCLINoRetry(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli\n\nA CLI tool for https://api.example.com.",
		Flags: []*discovery.Flag{
			{Name: "verbose", Description: "Verbose output"},
		},
	}

	check := newCheckPV2()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusFail {
		t.Errorf("expected Fail (network CLI without retry), got %s: %s", result.Status, result.Detail)
	}
}

func TestPV2_SkipNilTree(t *testing.T) {
	check := newCheckPV2()
	result := check.Run(context.Background(), &Input{Tree: nil})

	if result.Status != StatusSkip {
		t.Errorf("expected StatusSkip for nil tree, got %s: %s", result.Status, result.Detail)
	}
}

func TestPV2_Metadata(t *testing.T) {
	check := newCheckPV2()

	if check.ID() != "PV-2" {
		t.Errorf("expected PV-2, got %s", check.ID())
	}
	if check.Severity() != Info {
		t.Errorf("expected Info, got %s", check.Severity())
	}
	if check.Method() != Passive {
		t.Errorf("expected Passive, got %s", check.Method())
	}
}

// PV-3: Deterministic output (active check)

func TestPV3_SkipNilProber(t *testing.T) {
	check := newCheckPV3()
	result := check.Run(context.Background(), &Input{Prober: nil})

	if result.Status != StatusSkip {
		t.Errorf("expected StatusSkip for nil prober, got %s: %s", result.Status, result.Detail)
	}
}

func TestPV3_Metadata(t *testing.T) {
	check := newCheckPV3()

	if check.ID() != "PV-3" {
		t.Errorf("expected PV-3, got %s", check.ID())
	}
	if check.Severity() != Info {
		t.Errorf("expected Info, got %s", check.Severity())
	}
	if check.Method() != Active {
		t.Errorf("expected Active, got %s", check.Method())
	}
	if check.Category() != CatPredictability {
		t.Errorf("expected predictability, got %s", check.Category())
	}
}

// PV-4: Distinct exit codes for error classes (passive check)

func TestPV4_Metadata(t *testing.T) {
	check := newCheckPV4()

	if check.ID() != "PV-4" {
		t.Errorf("expected PV-4, got %s", check.ID())
	}
	if check.Category() != CatPredictability {
		t.Errorf("expected predictability, got %s", check.Category())
	}
	if check.Severity() != Info {
		t.Errorf("expected Info, got %s", check.Severity())
	}
	if check.Method() != Passive {
		t.Errorf("expected Passive, got %s", check.Method())
	}
}

func TestPV4_SkipNilTree(t *testing.T) {
	check := newCheckPV4()
	result := check.Run(context.Background(), &Input{Tree: nil})

	if result.Status != StatusSkip {
		t.Errorf("expected StatusSkip, got %s: %s", result.Status, result.Detail)
	}
}

func TestPV4_PassWithExitCodesSection(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli\n\nEXIT CODES\n  0  Success\n  1  General error\n  2  Usage error\n",
	}

	check := newCheckPV4()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for EXIT CODES section, got %s: %s", result.Status, result.Detail)
	}
}

func TestPV4_PassWithExitStatusSection(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli\n\nEXIT STATUS\n  0  OK\n  1  Error\n",
	}

	check := newCheckPV4()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for EXIT STATUS section, got %s: %s", result.Status, result.Detail)
	}
}

func TestPV4_PassWithExitCodeMentionInHelp(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli\n\nReturns exit code 1 on failure.",
	}

	check := newCheckPV4()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for exit code mention, got %s: %s", result.Status, result.Detail)
	}
}

func TestPV4_PassWithReturnCodeMention(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli\n\nThe return code indicates the result.",
	}

	check := newCheckPV4()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for return code mention, got %s: %s", result.Status, result.Detail)
	}
}

func TestPV4_FailNoDocumentation(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli\n\nA simple CLI tool.",
		Flags: []*discovery.Flag{
			{Name: "verbose", Description: "Verbose output"},
		},
	}

	check := newCheckPV4()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusFail {
		t.Errorf("expected Fail, got %s: %s", result.Status, result.Detail)
	}
}

// PV-5: Reports actual effects

func TestPV5_PassNoMutating(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Subcommands: []*discovery.Command{
			{Name: "list", FullPath: []string{"mycli", "list"}, IsListLike: true},
		},
	}
	check := newCheckPV5()
	result := check.Run(context.Background(), makeInput(root))
	if result.Status != StatusPass {
		t.Errorf("expected pass (no mutating), got %s: %s", result.Status, result.Detail)
	}
}

func TestPV5_PassWithEffectTerms(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Subcommands: []*discovery.Command{
			{Name: "create", FullPath: []string{"mycli", "create"}, IsMutating: true,
				RawHelp: "Create a resource.\n\nReports the count of created and skipped items."},
		},
	}
	check := newCheckPV5()
	result := check.Run(context.Background(), makeInput(root))
	if result.Status != StatusPass {
		t.Errorf("expected pass for effect terms, got %s: %s", result.Status, result.Detail)
	}
}

func TestPV5_FailNoEffectTerms(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Subcommands: []*discovery.Command{
			{Name: "create", FullPath: []string{"mycli", "create"}, IsMutating: true,
				RawHelp: "Create a new resource."},
		},
	}
	check := newCheckPV5()
	result := check.Run(context.Background(), makeInput(root))
	if result.Status != StatusFail {
		t.Errorf("expected fail for missing effect terms, got %s: %s", result.Status, result.Detail)
	}
}
