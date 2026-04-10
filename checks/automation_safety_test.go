package checks

import (
	"context"
	"testing"

	"github.com/Camil-H/cli-agent-lint/discovery"
)

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

// SA-2: Rejects path traversal (active check)

func TestSA2_SkipNilProber(t *testing.T) {
	check := newCheckSA2()
	result := check.Run(context.Background(), &Input{
		Prober: nil,
		Tree:   makeTree(&discovery.Command{Name: "mycli", FullPath: []string{"mycli"}}),
	})

	if result.Status != StatusSkip {
		t.Errorf("expected StatusSkip for nil prober, got %s: %s", result.Status, result.Detail)
	}
	if result.Detail != "skipped: active check disabled by --no-probe" {
		t.Errorf("unexpected detail: %s", result.Detail)
	}
}

func TestSA2_Metadata(t *testing.T) {
	check := newCheckSA2()

	t.Run("ID", func(t *testing.T) {
		if check.ID() != "SA-2" {
			t.Errorf("expected SA-2, got %s", check.ID())
		}
	})

	t.Run("Category", func(t *testing.T) {
		if check.Category() != CatAutomationSafety {
			t.Errorf("expected automation-safety, got %s", check.Category())
		}
	})

	t.Run("Severity", func(t *testing.T) {
		if check.Severity() != Warn {
			t.Errorf("expected Warn, got %s", check.Severity())
		}
	})

	t.Run("Method", func(t *testing.T) {
		if check.Method() != Active {
			t.Errorf("expected Active, got %s", check.Method())
		}
	})
}

func TestLooksLikeFileCommand(t *testing.T) {
	tests := []struct {
		name     string
		cmdName  string
		expected bool
	}{
		{"open command", "open", true},
		{"read command", "read", true},
		{"load command", "load", true},
		{"import command", "import", true},
		{"file-upload", "file-upload", true},
		{"list command", "list", false},
		{"delete command", "delete", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := &discovery.Command{
				Name:     "mycli",
				FullPath: []string{"mycli"},
				Subcommands: []*discovery.Command{
					{Name: tt.cmdName, FullPath: []string{"mycli", tt.cmdName}},
				},
			}
			idx := makeIndex(root)
			fileCmds := idx.FileAccepting()
			got := false
			for _, c := range fileCmds {
				if c.Name == tt.cmdName {
					got = true
					break
				}
			}
			if got != tt.expected {
				t.Errorf("FileAccepting for %q: got %v, want %v", tt.cmdName, got, tt.expected)
			}
		})
	}
}

func TestFileArgFlag_Detection(t *testing.T) {
	makeIdxFor := func(cmd *discovery.Command) *discovery.CommandIndex {
		root := &discovery.Command{
			Name:        "mycli",
			FullPath:    []string{"mycli"},
			Subcommands: []*discovery.Command{cmd},
		}
		cmd.FullPath = []string{"mycli", cmd.Name}
		return makeIndex(root)
	}

	t.Run("has file flag", func(t *testing.T) {
		cmd := &discovery.Command{
			Name: "upload",
			Flags: []*discovery.Flag{
				{Name: "file", Description: "Path to the file"},
			},
		}
		if makeIdxFor(cmd).FileArgFlag(cmd) == "" {
			t.Error("expected FileArgFlag to return non-empty for file flag")
		}
	})

	t.Run("has path in description", func(t *testing.T) {
		cmd := &discovery.Command{
			Name: "upload",
			Flags: []*discovery.Flag{
				{Name: "input", Description: "The file path to upload"},
			},
		}
		if makeIdxFor(cmd).FileArgFlag(cmd) == "" {
			t.Error("expected FileArgFlag to return non-empty for path in description")
		}
	})

	t.Run("has dir flag", func(t *testing.T) {
		cmd := &discovery.Command{
			Name: "scan",
			Flags: []*discovery.Flag{
				{Name: "directory", Description: "The dir to scan"},
			},
		}
		if makeIdxFor(cmd).FileArgFlag(cmd) == "" {
			t.Error("expected FileArgFlag to return non-empty for dir flag")
		}
	})

	t.Run("inherited file flag", func(t *testing.T) {
		cmd := &discovery.Command{
			Name: "deploy",
			InheritedFlags: []*discovery.Flag{
				{Name: "config-file", Description: "Configuration file"},
			},
		}
		if makeIdxFor(cmd).FileArgFlag(cmd) == "" {
			t.Error("expected FileArgFlag to return non-empty for inherited file flag")
		}
	})

	t.Run("no file flag", func(t *testing.T) {
		cmd := &discovery.Command{
			Name: "status",
			Flags: []*discovery.Flag{
				{Name: "verbose", Description: "Show verbose output"},
			},
		}
		if makeIdxFor(cmd).FileArgFlag(cmd) != "" {
			t.Error("expected FileArgFlag to return empty")
		}
	})
}

func TestFindFileCommands(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Subcommands: []*discovery.Command{
			{
				Name:     "open",
				FullPath: []string{"mycli", "open"},
			},
			{
				Name:     "list",
				FullPath: []string{"mycli", "list"},
			},
			{
				Name:     "upload",
				FullPath: []string{"mycli", "upload"},
				Flags: []*discovery.Flag{
					{Name: "file", Description: "File to upload"},
				},
			},
		},
	}
	tree := makeTree(root)

	cmds := makeIndex(tree.Root).FileAccepting()
	if len(cmds) != 2 {
		t.Fatalf("expected 2 file commands (open, upload), got %d", len(cmds))
	}
}

func TestFileArgFlag(t *testing.T) {
	makeIdxFor := func(cmd *discovery.Command) *discovery.CommandIndex {
		root := &discovery.Command{
			Name:        "mycli",
			FullPath:    []string{"mycli"},
			Subcommands: []*discovery.Command{cmd},
		}
		cmd.FullPath = []string{"mycli", cmd.Name}
		return makeIndex(root)
	}

	t.Run("returns file flag name", func(t *testing.T) {
		cmd := &discovery.Command{
			Name: "upload",
			Flags: []*discovery.Flag{
				{Name: "file", Description: "Path to file"},
			},
		}
		got := makeIdxFor(cmd).FileArgFlag(cmd)
		if got != "file" {
			t.Errorf("expected 'file', got %q", got)
		}
	})

	t.Run("returns empty for no file flag", func(t *testing.T) {
		cmd := &discovery.Command{
			Name: "open",
			Flags: []*discovery.Flag{
				{Name: "verbose", Description: "Verbose mode"},
			},
		}
		got := makeIdxFor(cmd).FileArgFlag(cmd)
		if got != "" {
			t.Errorf("expected empty string, got %q", got)
		}
	})
}

// SA-3: Rejects control characters (active check)

func TestSA3_SkipNilProber(t *testing.T) {
	check := newCheckSA3()
	result := check.Run(context.Background(), &Input{
		Prober: nil,
		Tree:   makeTree(&discovery.Command{Name: "mycli", FullPath: []string{"mycli"}}),
	})

	if result.Status != StatusSkip {
		t.Errorf("expected StatusSkip for nil prober, got %s: %s", result.Status, result.Detail)
	}
}

func TestSA3_Metadata(t *testing.T) {
	check := newCheckSA3()

	if check.ID() != "SA-3" {
		t.Errorf("expected SA-3, got %s", check.ID())
	}
	if check.Severity() != Warn {
		t.Errorf("expected Warn, got %s", check.Severity())
	}
	if check.Method() != Active {
		t.Errorf("expected Active, got %s", check.Method())
	}
}

func TestFindStringInputCommand(t *testing.T) {
	t.Run("finds command with string flag", func(t *testing.T) {
		root := &discovery.Command{
			Name:     "mycli",
			FullPath: []string{"mycli"},
			Subcommands: []*discovery.Command{
				{
					Name:     "create",
					FullPath: []string{"mycli", "create"},
					Flags: []*discovery.Flag{
						{Name: "name", ValueType: "string", Description: "Resource name"},
					},
				},
			},
		}
		tree := makeTree(root)
		idx := makeIndex(tree.Root)

		stringCmds := idx.StringInput()
		if len(stringCmds) == 0 {
			t.Fatal("expected to find a command with string input")
		}
		flagName := idx.StringInputFlag(stringCmds[0])
		if flagName != "name" {
			t.Errorf("expected flag 'name', got %q", flagName)
		}
	})

	t.Run("finds leaf command without string flag", func(t *testing.T) {
		root := &discovery.Command{
			Name:     "mycli",
			FullPath: []string{"mycli"},
			Subcommands: []*discovery.Command{
				{
					Name:     "echo",
					FullPath: []string{"mycli", "echo"},
					// No flags, but it is a leaf with depth > 1
				},
			},
		}
		tree := makeTree(root)
		idx := makeIndex(tree.Root)

		stringCmds := idx.StringInput()
		if len(stringCmds) == 0 {
			t.Fatal("expected to find a leaf command")
		}
		flagName := idx.StringInputFlag(stringCmds[0])
		if flagName != "" {
			t.Errorf("expected empty flag name for leaf command, got %q", flagName)
		}
	})

	t.Run("returns empty for root-only tree", func(t *testing.T) {
		root := &discovery.Command{
			Name:     "mycli",
			FullPath: []string{"mycli"},
			// No string flags, no subcommands
		}
		tree := makeTree(root)
		idx := makeIndex(tree.Root)

		stringCmds := idx.StringInput()
		if len(stringCmds) != 0 {
			t.Errorf("expected empty for root-only tree, got %d commands", len(stringCmds))
		}
	})
}

// SA-4: Dry-run support (passive check)

func TestSA4_PassNoMutatingCommands(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Subcommands: []*discovery.Command{
			{Name: "list", FullPath: []string{"mycli", "list"}, IsListLike: true},
			{Name: "show", FullPath: []string{"mycli", "show"}, IsListLike: true},
		},
	}

	check := newCheckSA4()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass (no mutating commands), got %s: %s", result.Status, result.Detail)
	}
}

func TestSA4_PassAllMutatingHaveDryRun(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Subcommands: []*discovery.Command{
			{
				Name:       "create",
				FullPath:   []string{"mycli", "create"},
				IsMutating: true,
				Flags: []*discovery.Flag{
					{Name: "dry-run", Description: "Simulate the operation"},
				},
			},
			{
				Name:       "delete",
				FullPath:   []string{"mycli", "delete"},
				IsMutating: true,
				Flags: []*discovery.Flag{
					{Name: "dry-run", Description: "Simulate deletion"},
				},
			},
		},
	}

	check := newCheckSA4()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass, got %s: %s", result.Status, result.Detail)
	}
}

func TestSA4_FailMissingDryRun(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Subcommands: []*discovery.Command{
			{
				Name:       "create",
				FullPath:   []string{"mycli", "create"},
				IsMutating: true,
				Flags: []*discovery.Flag{
					{Name: "dry-run", Description: "Simulate the operation"},
				},
			},
			{
				Name:       "delete",
				FullPath:   []string{"mycli", "delete"},
				IsMutating: true,
				Flags: []*discovery.Flag{
					{Name: "force", Description: "Force deletion"},
				},
			},
		},
	}

	check := newCheckSA4()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusFail {
		t.Errorf("expected Fail, got %s: %s", result.Status, result.Detail)
	}
}

func TestSA4_PassWithAlternateDryRunNames(t *testing.T) {
	alternateNames := []string{"dryrun", "whatif", "simulate", "dry_run"}

	for _, flagName := range alternateNames {
		t.Run(flagName, func(t *testing.T) {
			root := &discovery.Command{
				Name:     "mycli",
				FullPath: []string{"mycli"},
				Subcommands: []*discovery.Command{
					{
						Name:       "create",
						FullPath:   []string{"mycli", "create"},
						IsMutating: true,
						Flags: []*discovery.Flag{
							{Name: flagName, Description: "Simulate"},
						},
					},
				},
			}

			check := newCheckSA4()
			result := check.Run(context.Background(), makeInput(root))

			if result.Status != StatusPass {
				t.Errorf("expected StatusPass for flag --%s, got %s: %s", flagName, result.Status, result.Detail)
			}
		})
	}
}

func TestSA4_Metadata(t *testing.T) {
	check := newCheckSA4()

	if check.ID() != "SA-4" {
		t.Errorf("expected SA-4, got %s", check.ID())
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

// SA-5: Idempotency indicators

func TestSA5_PassNoMutating(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Subcommands: []*discovery.Command{
			{Name: "list", FullPath: []string{"mycli", "list"}, IsListLike: true},
		},
	}
	check := newCheckSA5()
	result := check.Run(context.Background(), makeInput(root))
	if result.Status != StatusPass {
		t.Errorf("expected pass (no mutating), got %s: %s", result.Status, result.Detail)
	}
}

func TestSA5_PassWithIdempotencyFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Subcommands: []*discovery.Command{
			{Name: "create", FullPath: []string{"mycli", "create"}, IsMutating: true,
				Flags: []*discovery.Flag{{Name: "if-not-exists", Description: "Skip if already exists"}}},
		},
	}
	check := newCheckSA5()
	result := check.Run(context.Background(), makeInput(root))
	if result.Status != StatusPass {
		t.Errorf("expected pass for --if-not-exists, got %s: %s", result.Status, result.Detail)
	}
}

func TestSA5_FailCreateWithoutIdempotency(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Subcommands: []*discovery.Command{
			{Name: "create", FullPath: []string{"mycli", "create"}, IsMutating: true},
		},
	}
	check := newCheckSA5()
	result := check.Run(context.Background(), makeInput(root))
	if result.Status != StatusFail {
		t.Errorf("expected fail for create without idempotency, got %s: %s", result.Status, result.Detail)
	}
}

func TestSA5_PassDeleteOnly(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Subcommands: []*discovery.Command{
			{Name: "delete", FullPath: []string{"mycli", "delete"}, IsMutating: true},
		},
	}
	check := newCheckSA5()
	result := check.Run(context.Background(), makeInput(root))
	if result.Status != StatusPass {
		t.Errorf("expected pass (delete is inherently idempotent), got %s: %s", result.Status, result.Detail)
	}
}

// SA-6: Read/write command separation

func TestSA6_PassClearSeparation(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Subcommands: []*discovery.Command{
			{Name: "list", FullPath: []string{"mycli", "list"}, IsListLike: true},
			{Name: "create", FullPath: []string{"mycli", "create"}, IsMutating: true},
		},
	}
	check := newCheckSA6()
	result := check.Run(context.Background(), makeInput(root))
	if result.Status != StatusPass {
		t.Errorf("expected pass for clear separation, got %s: %s", result.Status, result.Detail)
	}
}

func TestSA6_FailMutatingOnly(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Subcommands: []*discovery.Command{
			{Name: "create", FullPath: []string{"mycli", "create"}, IsMutating: true},
			{Name: "delete", FullPath: []string{"mycli", "delete"}, IsMutating: true},
		},
	}
	check := newCheckSA6()
	result := check.Run(context.Background(), makeInput(root))
	if result.Status != StatusFail {
		t.Errorf("expected fail for mutating-only, got %s: %s", result.Status, result.Detail)
	}
}

func TestSA6_PassNoCommands(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
	}
	check := newCheckSA6()
	result := check.Run(context.Background(), makeInput(root))
	if result.Status != StatusPass {
		t.Errorf("expected pass (no commands), got %s: %s", result.Status, result.Detail)
	}
}
