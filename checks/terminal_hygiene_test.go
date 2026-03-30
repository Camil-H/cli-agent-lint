package checks

import (
	"context"
	"testing"

	"github.com/cli-agent-lint/cli-agent-lint/discovery"
)

// ---------------------------------------------------------------------------
// TH-1: Non-TTY detection (active check)
// ---------------------------------------------------------------------------

func TestTH1_SkipNilProber(t *testing.T) {
	check := &checkTH1{
		BaseCheck: BaseCheck{
			CheckID:     "TH-1",
			CheckName:   "Non-TTY detection (no ANSI in pipes)",
			CheckCategory: CatTerminalHygiene,
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

func TestTH1_Metadata(t *testing.T) {
	r := DefaultRegistry()
	check := r.Get("TH-1")
	if check == nil {
		t.Fatal("TH-1 not found in registry")
	}

	t.Run("ID", func(t *testing.T) {
		if check.ID() != "TH-1" {
			t.Errorf("expected TH-1, got %s", check.ID())
		}
	})

	t.Run("Category", func(t *testing.T) {
		if check.Category() != CatTerminalHygiene {
			t.Errorf("expected terminal-hygiene, got %s", check.Category())
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

// ---------------------------------------------------------------------------
// TH-2: --no-color flag (passive check)
// ---------------------------------------------------------------------------

func TestTH2_PassWithNoColorFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Flags: []*discovery.Flag{
			{Name: "no-color", Description: "Disable color output"},
		},
	}

	r := DefaultRegistry()
	check := r.Get("TH-2")
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for --no-color flag, got %s: %s", result.Status, result.Detail)
	}
}

func TestTH2_PassWithColorFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Flags: []*discovery.Flag{
			{Name: "color", Description: "Control color output (auto, always, never)"},
		},
	}

	r := DefaultRegistry()
	check := r.Get("TH-2")
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for --color flag, got %s: %s", result.Status, result.Detail)
	}
}

func TestTH2_PassWithNoColorInHelp(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli [options]\n\nSet NO_COLOR=1 to disable colors.",
	}

	r := DefaultRegistry()
	check := r.Get("TH-2")
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for NO_COLOR in help, got %s: %s", result.Status, result.Detail)
	}
}

func TestTH2_PassWithColorNeverInHelp(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli [options]\n\n  --color=never  Disable ANSI colors",
	}

	r := DefaultRegistry()
	check := r.Get("TH-2")
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for --color=never in help, got %s: %s", result.Status, result.Detail)
	}
}

func TestTH2_FailNoColorSupport(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli [options]\n\nA simple CLI tool.",
		Flags: []*discovery.Flag{
			{Name: "verbose", Description: "Be verbose"},
		},
	}

	r := DefaultRegistry()
	check := r.Get("TH-2")
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusFail {
		t.Errorf("expected Fail, got %s: %s", result.Status, result.Detail)
	}
}

func TestTH2_Metadata(t *testing.T) {
	r := DefaultRegistry()
	check := r.Get("TH-2")

	if check.Method() != Passive {
		t.Errorf("expected Passive, got %s", check.Method())
	}
	if check.Severity() != Warn {
		t.Errorf("expected Warn, got %s", check.Severity())
	}
}

// ---------------------------------------------------------------------------
// TH-3: --quiet / --silent flag (passive check)
// ---------------------------------------------------------------------------

func TestTH3_PassWithQuietFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Flags: []*discovery.Flag{
			{Name: "quiet", Description: "Suppress output"},
		},
	}

	r := DefaultRegistry()
	check := r.Get("TH-3")
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for --quiet flag, got %s: %s", result.Status, result.Detail)
	}
}

func TestTH3_PassWithSilentFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Flags: []*discovery.Flag{
			{Name: "silent", Description: "Suppress output"},
		},
	}

	r := DefaultRegistry()
	check := r.Get("TH-3")
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for --silent flag, got %s: %s", result.Status, result.Detail)
	}
}

func TestTH3_PassWithShortQFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Flags: []*discovery.Flag{
			{Name: "q", ShortName: "q", Description: "Quiet mode"},
		},
	}

	r := DefaultRegistry()
	check := r.Get("TH-3")
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for -q flag, got %s: %s", result.Status, result.Detail)
	}
}

func TestTH3_PassWithQuietInHelp(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli [options]\n\n  --quiet       Suppress output",
	}

	r := DefaultRegistry()
	check := r.Get("TH-3")
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for --quiet in help, got %s: %s", result.Status, result.Detail)
	}
}

func TestTH3_FailNoQuietSupport(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli [options]\n\nA CLI tool.",
		Flags: []*discovery.Flag{
			{Name: "verbose", Description: "Be verbose"},
		},
	}

	r := DefaultRegistry()
	check := r.Get("TH-3")
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusFail {
		t.Errorf("expected Fail, got %s: %s", result.Status, result.Detail)
	}
}

func TestTH3_Metadata(t *testing.T) {
	r := DefaultRegistry()
	check := r.Get("TH-3")

	if check.Method() != Passive {
		t.Errorf("expected Passive, got %s", check.Method())
	}
	if check.Severity() != Info {
		t.Errorf("expected Info, got %s", check.Severity())
	}
}

// ---------------------------------------------------------------------------
// TH-5: Confirmation bypass for destructive commands (passive check)
// ---------------------------------------------------------------------------

func TestTH5_Metadata(t *testing.T) {
	check := newCheckTH5()

	if check.ID() != "TH-5" {
		t.Errorf("expected TH-5, got %s", check.ID())
	}
	if check.Category() != CatTerminalHygiene {
		t.Errorf("expected terminal-hygiene, got %s", check.Category())
	}
	if check.Severity() != Warn {
		t.Errorf("expected Warn, got %s", check.Severity())
	}
	if check.Method() != Passive {
		t.Errorf("expected Passive, got %s", check.Method())
	}
}

func TestTH5_SkipNilTree(t *testing.T) {
	check := newCheckTH5()
	result := check.Run(context.Background(), &Input{Tree: nil})

	if result.Status != StatusSkip {
		t.Errorf("expected StatusSkip, got %s: %s", result.Status, result.Detail)
	}
}

func TestTH5_PassNoMutatingCommands(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Subcommands: []*discovery.Command{
			{Name: "list", FullPath: []string{"mycli", "list"}, IsListLike: true},
			{Name: "get", FullPath: []string{"mycli", "get"}},
		},
	}

	check := newCheckTH5()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass (no mutating commands), got %s: %s", result.Status, result.Detail)
	}
}

func TestTH5_PassAllHaveYesFlag(t *testing.T) {
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

	check := newCheckTH5()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass, got %s: %s", result.Status, result.Detail)
	}
}

func TestTH5_PassWithForceFlag(t *testing.T) {
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

	check := newCheckTH5()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for --force flag, got %s: %s", result.Status, result.Detail)
	}
}

func TestTH5_PassWithNonInteractiveFlag(t *testing.T) {
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

	check := newCheckTH5()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for --non-interactive flag, got %s: %s", result.Status, result.Detail)
	}
}

func TestTH5_FailMissingBypassFlag(t *testing.T) {
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

	check := newCheckTH5()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusFail {
		t.Errorf("expected Fail, got %s: %s", result.Status, result.Detail)
	}
}

func TestTH5_PassWithShortYFlag(t *testing.T) {
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

	check := newCheckTH5()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for -y flag, got %s: %s", result.Status, result.Detail)
	}
}

// ---------------------------------------------------------------------------
// TH-4: No interactive prompts in non-TTY (active check)
// ---------------------------------------------------------------------------

func TestTH4_SkipNilProber(t *testing.T) {
	r := DefaultRegistry()
	check := r.Get("TH-4")
	result := check.Run(context.Background(), &Input{Prober: nil, Tree: makeTree(&discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
	})})

	if result.Status != StatusSkip {
		t.Errorf("expected StatusSkip for nil prober, got %s: %s", result.Status, result.Detail)
	}
}

func TestTH4_Metadata(t *testing.T) {
	r := DefaultRegistry()
	check := r.Get("TH-4")

	if check.Method() != Active {
		t.Errorf("expected Active, got %s", check.Method())
	}
	if check.Severity() != Fail {
		t.Errorf("expected Fail, got %s", check.Severity())
	}
	if check.Category() != CatTerminalHygiene {
		t.Errorf("expected terminal-hygiene, got %s", check.Category())
	}
}
