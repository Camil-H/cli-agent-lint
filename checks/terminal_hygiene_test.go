package checks

import (
	"context"
	"testing"

	"github.com/cli-agent-lint/cli-agent-lint/discovery"
)

// FS-2: Non-TTY detection (active check)

func TestFS2_SkipNilProber(t *testing.T) {
	check := &checkFS2{
		BaseCheck: BaseCheck{
			CheckID:     "FS-2",
			CheckName:   "Non-TTY detection (no ANSI in pipes)",
			CheckCategory: CatFlowSafety,
			CheckSeverity: Fail,
			CheckMethod:   Active,
		},
	}
	result := check.Run(context.Background(), &Input{Prober: nil})

	if result.Status != StatusSkip {
		t.Errorf("expected StatusSkip for nil prober, got %s: %s", result.Status, result.Detail)
	}
	if result.Detail != "skipped: active check disabled by --no-probe" {
		t.Errorf("unexpected detail: %s", result.Detail)
	}
}

func TestFS2_Metadata(t *testing.T) {
	r := DefaultRegistry()
	check := r.Get("FS-2")
	if check == nil {
		t.Fatal("FS-2 not found in registry")
	}

	t.Run("ID", func(t *testing.T) {
		if check.ID() != "FS-2" {
			t.Errorf("expected FS-2, got %s", check.ID())
		}
	})

	t.Run("Category", func(t *testing.T) {
		if check.Category() != CatFlowSafety {
			t.Errorf("expected flow-safety, got %s", check.Category())
		}
	})

	t.Run("Method", func(t *testing.T) {
		if check.Method() != Active {
			t.Errorf("expected Active, got %s", check.Method())
		}
	})

	t.Run("Severity", func(t *testing.T) {
		if check.Severity() != Fail {
			t.Errorf("expected Fail, got %s", check.Severity())
		}
	})
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

// SA-1: Confirmation bypass for destructive commands (passive check)

func TestSA1_Metadata(t *testing.T) {
	check := newCheckSA1()

	if check.ID() != "SA-1" {
		t.Errorf("expected SA-1, got %s", check.ID())
	}
	if check.Category() != CatAutomationSafety {
		t.Errorf("expected automation-safety, got %s", check.Category())
	}
	if check.Severity() != Warn {
		t.Errorf("expected Warn, got %s", check.Severity())
	}
	if check.Method() != Passive {
		t.Errorf("expected Passive, got %s", check.Method())
	}
}

func TestSA1_SkipNilTree(t *testing.T) {
	check := newCheckSA1()
	result := check.Run(context.Background(), &Input{Tree: nil})

	if result.Status != StatusSkip {
		t.Errorf("expected StatusSkip, got %s: %s", result.Status, result.Detail)
	}
}

func TestSA1_PassNoMutatingCommands(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Subcommands: []*discovery.Command{
			{Name: "list", FullPath: []string{"mycli", "list"}, IsListLike: true},
			{Name: "get", FullPath: []string{"mycli", "get"}},
		},
	}

	check := newCheckSA1()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass (no mutating commands), got %s: %s", result.Status, result.Detail)
	}
}

func TestSA1_PassAllHaveYesFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Subcommands: []*discovery.Command{
			{
				Name:       "create",
				FullPath:   []string{"mycli", "create"},
				IsMutating: true,
				Flags: []*discovery.Flag{
					{Name: "yes", ShortName: "y", Description: "Skip confirmation"},
				},
			},
			{
				Name:       "delete",
				FullPath:   []string{"mycli", "delete"},
				IsMutating: true,
				Flags: []*discovery.Flag{
					{Name: "yes", ShortName: "y", Description: "Skip confirmation"},
				},
			},
		},
	}

	check := newCheckSA1()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass, got %s: %s", result.Status, result.Detail)
	}
}

func TestSA1_PassWithForceFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Subcommands: []*discovery.Command{
			{
				Name:       "delete",
				FullPath:   []string{"mycli", "delete"},
				IsMutating: true,
				Flags: []*discovery.Flag{
					{Name: "force", Description: "Force without confirmation"},
				},
			},
		},
	}

	check := newCheckSA1()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for --force flag, got %s: %s", result.Status, result.Detail)
	}
}

func TestSA1_PassWithNonInteractiveFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Subcommands: []*discovery.Command{
			{
				Name:       "update",
				FullPath:   []string{"mycli", "update"},
				IsMutating: true,
				Flags: []*discovery.Flag{
					{Name: "non-interactive", Description: "Non-interactive mode"},
				},
			},
		},
	}

	check := newCheckSA1()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for --non-interactive flag, got %s: %s", result.Status, result.Detail)
	}
}

func TestSA1_FailMissingBypassFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Subcommands: []*discovery.Command{
			{
				Name:       "create",
				FullPath:   []string{"mycli", "create"},
				IsMutating: true,
				Flags: []*discovery.Flag{
					{Name: "yes", ShortName: "y", Description: "Skip confirmation"},
				},
			},
			{
				Name:       "delete",
				FullPath:   []string{"mycli", "delete"},
				IsMutating: true,
				Flags: []*discovery.Flag{
					{Name: "verbose", Description: "Verbose output"},
				},
			},
		},
	}

	check := newCheckSA1()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusFail {
		t.Errorf("expected Fail, got %s: %s", result.Status, result.Detail)
	}
}

func TestSA1_PassWithShortYFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Subcommands: []*discovery.Command{
			{
				Name:       "delete",
				FullPath:   []string{"mycli", "delete"},
				IsMutating: true,
				Flags: []*discovery.Flag{
					{Name: "y", ShortName: "y", Description: "Skip confirmation"},
				},
			},
		},
	}

	check := newCheckSA1()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for -y flag, got %s: %s", result.Status, result.Detail)
	}
}

// FS-3: No interactive prompts in non-TTY (active check)

func TestFS3_SkipNilProber(t *testing.T) {
	r := DefaultRegistry()
	check := r.Get("FS-3")
	result := check.Run(context.Background(), &Input{Prober: nil, Tree: makeTree(&discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
	})})

	if result.Status != StatusSkip {
		t.Errorf("expected StatusSkip for nil prober, got %s: %s", result.Status, result.Detail)
	}
}

func TestFS3_Metadata(t *testing.T) {
	r := DefaultRegistry()
	check := r.Get("FS-3")

	if check.Method() != Active {
		t.Errorf("expected Active, got %s", check.Method())
	}
	if check.Severity() != Fail {
		t.Errorf("expected Fail, got %s", check.Severity())
	}
	if check.Category() != CatFlowSafety {
		t.Errorf("expected flow-safety, got %s", check.Category())
	}
}
