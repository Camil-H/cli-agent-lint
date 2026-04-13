package checks

import (
	"context"
	"testing"

	"github.com/Camil-H/cli-agent-lint/discovery"
)

// FS-1: Stderr vs stdout discipline (active check)

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

func TestFS1_Metadata(t *testing.T) {
	check := newCheckFS1()
	if check.Method() != Active {
		t.Errorf("expected Active method, got %s", check.Method())
	}
	if check.Severity() != Fail {
		t.Errorf("expected Fail severity, got %s", check.Severity())
	}
}

// FS-2: Non-TTY detection (active check)

func TestFS2_SkipNilProber(t *testing.T) {
	check := &checkFS2{
		BaseCheck: BaseCheck{
			CheckID:       "FS-2",
			CheckName:     "Non-TTY detection (no ANSI in pipes)",
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

// FS-4: Env var auth support (passive check)

func TestFS4_PassWithAuthEnvVar(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli\n\nSet GITHUB_TOKEN to authenticate.",
	}

	check := newCheckFS4()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for GITHUB_TOKEN env var, got %s: %s", result.Status, result.Detail)
	}
}

func TestFS4_PassWithAPIKeyEnvVar(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli\n\nSet MY_SERVICE_API_KEY for access.",
	}

	check := newCheckFS4()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for MY_SERVICE_API_KEY env var, got %s: %s", result.Status, result.Detail)
	}
}

func TestFS4_PassWithTokenFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Flags: []*discovery.Flag{
			{Name: "token", Description: "Authentication token"},
		},
	}

	check := newCheckFS4()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for --token flag, got %s: %s", result.Status, result.Detail)
	}
}

func TestFS4_PassWithAPIKeyFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Flags: []*discovery.Flag{
			{Name: "api-key", Description: "API key for authentication"},
		},
	}

	check := newCheckFS4()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for --api-key flag, got %s: %s", result.Status, result.Detail)
	}
}

func TestFS4_PassWithAccessTokenFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Flags: []*discovery.Flag{
			{Name: "access-token", Description: "OAuth access token"},
		},
	}

	check := newCheckFS4()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for --access-token flag, got %s: %s", result.Status, result.Detail)
	}
}

func TestFS4_PassWithCredentialsFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Flags: []*discovery.Flag{
			{Name: "credentials", Description: "Path to credentials file"},
		},
	}

	check := newCheckFS4()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for --credentials flag, got %s: %s", result.Status, result.Detail)
	}
}

func TestFS4_PassWithAuthSubcommandMentioningToken(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Subcommands: []*discovery.Command{
			{
				Name:     "auth",
				FullPath: []string{"mycli", "auth"},
				RawHelp:  "Authenticate with a token or use env vars.",
			},
		},
	}

	check := newCheckFS4()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for auth subcommand mentioning token, got %s: %s", result.Status, result.Detail)
	}
}

func TestFS4_PassWithLoginSubcommandMentioningEnv(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Subcommands: []*discovery.Command{
			{
				Name:     "login",
				FullPath: []string{"mycli", "login"},
				RawHelp:  "Log in interactively or via env variables.",
			},
		},
	}

	check := newCheckFS4()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for login subcommand mentioning env, got %s: %s", result.Status, result.Detail)
	}
}

func TestFS4_FailAuthMentionedButNoEnvVar(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli\n\nYou must authenticate before using this tool.",
		Subcommands: []*discovery.Command{
			{
				Name:     "status",
				FullPath: []string{"mycli", "status"},
				RawHelp:  "Check auth status.",
			},
		},
	}

	check := newCheckFS4()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusFail {
		t.Errorf("expected Fail (auth mentioned but no env var), got %s: %s", result.Status, result.Detail)
	}
}

func TestFS4_SkipNoAuthContent(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli\n\nA data processing tool.",
		Flags: []*discovery.Flag{
			{Name: "verbose", Description: "Enable verbose output"},
		},
	}

	check := newCheckFS4()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass (no auth content, not applicable), got %s: %s", result.Status, result.Detail)
	}
}

func TestFS4_SkipNilTree(t *testing.T) {
	check := newCheckFS4()
	result := check.Run(context.Background(), &Input{Tree: nil})

	if result.Status != StatusSkip {
		t.Errorf("expected StatusSkip for nil tree, got %s: %s", result.Status, result.Detail)
	}
}

func TestFS4_PassWithEnvVarOnSubcommand(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Subcommands: []*discovery.Command{
			{
				Name:     "deploy",
				FullPath: []string{"mycli", "deploy"},
				RawHelp:  "Deploy the app. Reads DEPLOY_SECRET for authentication.",
			},
		},
	}

	check := newCheckFS4()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for DEPLOY_SECRET in subcommand help, got %s: %s", result.Status, result.Detail)
	}
}

func TestFS4_Metadata(t *testing.T) {
	check := newCheckFS4()

	if check.ID() != "FS-4" {
		t.Errorf("expected FS-4, got %s", check.ID())
	}
	if check.Category() != CatFlowSafety {
		t.Errorf("expected flow-safety, got %s", check.Category())
	}
	if check.Severity() != Warn {
		t.Errorf("expected Warn, got %s", check.Severity())
	}
	if check.Method() != Passive {
		t.Errorf("expected Passive, got %s", check.Method())
	}
}

// FS-5: No mandatory interactive auth (passive check)

func TestFS5_PassNoLoginCommand(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Subcommands: []*discovery.Command{
			{Name: "list", FullPath: []string{"mycli", "list"}},
			{Name: "create", FullPath: []string{"mycli", "create"}},
		},
	}

	check := newCheckFS5()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass (no login command), got %s: %s", result.Status, result.Detail)
	}
}

func TestFS5_PassLoginWithTokenFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Subcommands: []*discovery.Command{
			{
				Name:     "login",
				FullPath: []string{"mycli", "login"},
				Flags: []*discovery.Flag{
					{Name: "token", Description: "Use a token instead of interactive login"},
				},
			},
		},
	}

	check := newCheckFS5()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass (login has --token flag), got %s: %s", result.Status, result.Detail)
	}
}

func TestFS5_PassLoginWithWithTokenFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Subcommands: []*discovery.Command{
			{
				Name:     "login",
				FullPath: []string{"mycli", "login"},
				Flags: []*discovery.Flag{
					{Name: "with-token", Description: "Read token from stdin"},
				},
			},
		},
	}

	check := newCheckFS5()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass (login has --with-token flag), got %s: %s", result.Status, result.Detail)
	}
}

func TestFS5_PassSigninWithEnvVarMention(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Subcommands: []*discovery.Command{
			{
				Name:     "signin",
				FullPath: []string{"mycli", "signin"},
				RawHelp:  "Sign in to the service.",
			},
			{
				Name:     "config",
				FullPath: []string{"mycli", "config"},
				RawHelp:  "Set MY_CLI_TOKEN env var for non-interactive auth.",
			},
		},
	}

	check := newCheckFS5()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass (signin exists but env var alternative found), got %s: %s", result.Status, result.Detail)
	}
}

func TestFS5_PassSignInWithServiceAccountFlag(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		Subcommands: []*discovery.Command{
			{
				Name:     "sign-in",
				FullPath: []string{"mycli", "sign-in"},
			},
		},
		Flags: []*discovery.Flag{
			{Name: "service-account", Description: "Service account key file"},
		},
	}

	check := newCheckFS5()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass (sign-in exists but --service-account alternative), got %s: %s", result.Status, result.Detail)
	}
}

func TestFS5_FailLoginNoAlternative(t *testing.T) {
	root := &discovery.Command{
		Name:     "mycli",
		FullPath: []string{"mycli"},
		RawHelp:  "Usage: mycli\n\nA CLI tool that requires login.",
		Subcommands: []*discovery.Command{
			{
				Name:     "login",
				FullPath: []string{"mycli", "login"},
				RawHelp:  "Log in interactively using your browser.",
				Flags: []*discovery.Flag{
					{Name: "browser", Description: "Open browser for login"},
				},
			},
			{
				Name:     "list",
				FullPath: []string{"mycli", "list"},
				RawHelp:  "List resources.",
			},
		},
	}

	check := newCheckFS5()
	result := check.Run(context.Background(), makeInput(root))

	if result.Status != StatusFail {
		t.Errorf("expected Fail (login exists with no non-interactive alternative), got %s: %s", result.Status, result.Detail)
	}
}

func TestFS5_SkipNilTree(t *testing.T) {
	check := newCheckFS5()
	result := check.Run(context.Background(), &Input{Tree: nil})

	if result.Status != StatusSkip {
		t.Errorf("expected StatusSkip for nil tree, got %s: %s", result.Status, result.Detail)
	}
}

func TestFS5_Metadata(t *testing.T) {
	check := newCheckFS5()

	if check.ID() != "FS-5" {
		t.Errorf("expected FS-5, got %s", check.ID())
	}
	if check.Category() != CatFlowSafety {
		t.Errorf("expected flow-safety, got %s", check.Category())
	}
	if check.Severity() != Fail {
		t.Errorf("expected Fail, got %s", check.Severity())
	}
	if check.Method() != Passive {
		t.Errorf("expected Passive, got %s", check.Method())
	}
}

func TestHasAuthEnvVarMention(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected bool
	}{
		{"GITHUB_TOKEN", "Set GITHUB_TOKEN to authenticate", true},
		{"API_KEY suffix", "Use MY_SERVICE_API_KEY", true},
		{"CREDENTIALS suffix", "Set AWS_CREDENTIALS env var", true},
		{"SECRET suffix", "Use DATABASE_SECRET for auth", true},
		{"PASSWORD suffix", "DB_PASSWORD is required", true},
		{"no match", "Use verbose mode for debugging", false},
		{"lowercase token", "set github_token", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasAuthEnvVarMention(tt.text)
			if got != tt.expected {
				t.Errorf("hasAuthEnvVarMention(%q) = %v, want %v", tt.text, got, tt.expected)
			}
		})
	}
}

func TestFindLoginCommand(t *testing.T) {
	t.Run("finds login", func(t *testing.T) {
		root := &discovery.Command{
			Name:     "mycli",
			FullPath: []string{"mycli"},
			Subcommands: []*discovery.Command{
				{Name: "login", FullPath: []string{"mycli", "login"}},
			},
		}
		cmd := findLoginCommand(makeIndex(root))
		if cmd == nil {
			t.Fatal("expected to find login command")
		}
		if cmd.Name != "login" {
			t.Errorf("expected login, got %s", cmd.Name)
		}
	})

	t.Run("finds signin", func(t *testing.T) {
		root := &discovery.Command{
			Name:     "mycli",
			FullPath: []string{"mycli"},
			Subcommands: []*discovery.Command{
				{Name: "signin", FullPath: []string{"mycli", "signin"}},
			},
		}
		cmd := findLoginCommand(makeIndex(root))
		if cmd == nil {
			t.Fatal("expected to find signin command")
		}
	})

	t.Run("finds sign-in", func(t *testing.T) {
		root := &discovery.Command{
			Name:     "mycli",
			FullPath: []string{"mycli"},
			Subcommands: []*discovery.Command{
				{Name: "sign-in", FullPath: []string{"mycli", "sign-in"}},
			},
		}
		cmd := findLoginCommand(makeIndex(root))
		if cmd == nil {
			t.Fatal("expected to find sign-in command")
		}
	})

	t.Run("returns nil when no login", func(t *testing.T) {
		root := &discovery.Command{
			Name:     "mycli",
			FullPath: []string{"mycli"},
			Subcommands: []*discovery.Command{
				{Name: "list", FullPath: []string{"mycli", "list"}},
			},
		}
		cmd := findLoginCommand(makeIndex(root))
		if cmd != nil {
			t.Errorf("expected nil, got %s", cmd.Name)
		}
	})

	t.Run("finds nested login", func(t *testing.T) {
		root := &discovery.Command{
			Name:     "mycli",
			FullPath: []string{"mycli"},
			Subcommands: []*discovery.Command{
				{
					Name:     "auth",
					FullPath: []string{"mycli", "auth"},
					Subcommands: []*discovery.Command{
						{Name: "login", FullPath: []string{"mycli", "auth", "login"}},
					},
				},
			},
		}
		cmd := findLoginCommand(makeIndex(root))
		if cmd == nil {
			t.Fatal("expected to find nested login command")
		}
		if cmd.Name != "login" {
			t.Errorf("expected login, got %s", cmd.Name)
		}
	})
}

// FS-6: Exit codes (active check)

func TestFS6_SkipNilProber(t *testing.T) {
	check := newCheckFS6()
	result := check.Run(context.Background(), &Input{Prober: nil})

	if result.Status != StatusSkip {
		t.Errorf("expected StatusSkip for nil prober, got %s: %s", result.Status, result.Detail)
	}
	if result.Detail != "skipped: active check disabled by --no-probe" {
		t.Errorf("unexpected detail: %s", result.Detail)
	}
}

func TestFS6_Metadata(t *testing.T) {
	check := newCheckFS6()

	t.Run("ID", func(t *testing.T) {
		if check.ID() != "FS-6" {
			t.Errorf("expected FS-6, got %s", check.ID())
		}
	})

	t.Run("Category", func(t *testing.T) {
		if check.Category() != CatFlowSafety {
			t.Errorf("expected flow-safety, got %s", check.Category())
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

// Active execution tests

func TestFS1_Active_GoodCLI(t *testing.T) {
	input := probeInput(t, "good-cli.sh")
	result := newCheckFS1().Run(context.Background(), input)
	if result.Status != StatusPass {
		t.Errorf("expected pass, got %s: %s", result.Status, result.Detail)
	}
}

func TestFS1_Active_BadCLI(t *testing.T) {
	input := probeInput(t, "bad-cli.sh")
	result := newCheckFS1().Run(context.Background(), input)
	if result.Status != StatusFail {
		t.Errorf("expected fail (error on stdout), got %s: %s", result.Status, result.Detail)
	}
}

func TestFS2_Active_GoodCLI(t *testing.T) {
	input := probeInput(t, "good-cli.sh")
	result := newCheckFS2().Run(context.Background(), input)
	if result.Status != StatusPass {
		t.Errorf("expected pass, got %s: %s", result.Status, result.Detail)
	}
}

func TestFS2_Active_BadCLI(t *testing.T) {
	input := probeInput(t, "bad-cli.sh")
	result := newCheckFS2().Run(context.Background(), input)
	if result.Status != StatusFail {
		t.Errorf("expected fail (ANSI in piped output), got %s: %s", result.Status, result.Detail)
	}
}

func TestFS6_Active_GoodCLI(t *testing.T) {
	input := probeInput(t, "good-cli.sh")
	result := newCheckFS6().Run(context.Background(), input)
	if result.Status != StatusPass {
		t.Errorf("expected pass, got %s: %s", result.Status, result.Detail)
	}
}

func TestFS6_Active_BadCLI(t *testing.T) {
	input := probeInput(t, "bad-cli.sh")
	result := newCheckFS6().Run(context.Background(), input)
	if result.Status != StatusFail {
		t.Errorf("expected fail (exit 0 on error), got %s: %s", result.Status, result.Detail)
	}
}
