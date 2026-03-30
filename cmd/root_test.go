package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/cli-agent-lint/cli-agent-lint/output"
)

// ---------------------------------------------------------------------------
// containsControlChars
// ---------------------------------------------------------------------------

func TestContainsControlChars_TrueForControlCharacters(t *testing.T) {
	cases := []struct {
		name  string
		input string
	}{
		{"SOH (0x01)", "hello\x01world"},
		{"NUL (0x00)", "hello\x00world"},
		{"US (0x1f)", "hello\x1fworld"},
		{"BEL (0x07)", "\x07"},
		{"standalone NUL", "\x00"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if !containsControlChars(tc.input) {
				t.Errorf("expected true for input containing %q", tc.input)
			}
		})
	}
}

func TestContainsControlChars_FalseForSafeStrings(t *testing.T) {
	cases := []struct {
		name  string
		input string
	}{
		{"normal text", "hello world"},
		{"tab", "hello\tworld"},
		{"newline", "hello\nworld"},
		{"carriage return", "hello\rworld"},
		{"mixed whitespace", "hello\t\n\rworld"},
		{"empty string", ""},
		{"printable ASCII", "!@#$%^&*()_+-=[]{}|;':\",./<>?"},
		{"unicode text", "caf\u00e9 na\u00efve"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if containsControlChars(tc.input) {
				t.Errorf("expected false for input %q", tc.input)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// rejectControlChars
// ---------------------------------------------------------------------------

func TestRejectControlChars_ErrorForDirtyArgs(t *testing.T) {
	// rejectControlChars also checks os.Args[1:], but the explicit args
	// parameter is checked first. We test with dirty explicit args so the
	// error is triggered before os.Args scanning.
	dirty := []string{"clean", "di\x01rty"}
	err := rejectControlChars(dirty)
	if err == nil {
		t.Fatal("expected error for args containing control characters")
	}
	if !strings.Contains(err.Error(), "control characters") {
		t.Errorf("error message should mention control characters, got: %s", err.Error())
	}
}

func TestRejectControlChars_NilForCleanArgs(t *testing.T) {
	// Clean explicit args — os.Args may contain anything from the test runner,
	// but typical go test invocations use only printable characters, so this
	// should pass.
	clean := []string{"hello", "world", "--flag=value"}
	err := rejectControlChars(clean)
	if err != nil {
		t.Fatalf("expected nil for clean args, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// NewRootCmd
// ---------------------------------------------------------------------------

func TestNewRootCmd_Use(t *testing.T) {
	opts := &GlobalOptions{}
	cmd := NewRootCmd(opts)
	if cmd.Use != "cli-agent-lint" {
		t.Errorf("expected Use = %q, got %q", "cli-agent-lint", cmd.Use)
	}
}

func TestNewRootCmd_HasExpectedSubcommands(t *testing.T) {
	opts := &GlobalOptions{}
	cmd := NewRootCmd(opts)

	expected := map[string]bool{
		"check":      false,
		"checks":     false,
		"completion": false,
	}

	for _, sub := range cmd.Commands() {
		if _, ok := expected[sub.Name()]; ok {
			expected[sub.Name()] = true
		}
	}

	for name, found := range expected {
		if !found {
			t.Errorf("expected subcommand %q not found", name)
		}
	}
}

func TestNewRootCmd_HasExpectedPersistentFlags(t *testing.T) {
	opts := &GlobalOptions{}
	cmd := NewRootCmd(opts)

	flags := []string{"output", "no-color", "quiet"}
	for _, name := range flags {
		if cmd.PersistentFlags().Lookup(name) == nil {
			t.Errorf("expected persistent flag %q not found", name)
		}
	}
}

func TestNewRootCmd_HasVersionFlag(t *testing.T) {
	opts := &GlobalOptions{}
	cmd := NewRootCmd(opts)
	if cmd.Flags().Lookup("version") == nil {
		t.Error("expected local flag \"version\" not found")
	}
}

// ---------------------------------------------------------------------------
// printVersion
// ---------------------------------------------------------------------------

func newTestOutput(buf *bytes.Buffer, format string) *output.Output {
	return output.New(output.Config{
		Format: format,
		Stdout: buf,
		Stderr: buf,
	})
}

// ---------------------------------------------------------------------------
// detectOutputFormatFromArgs
// ---------------------------------------------------------------------------

func TestDetectOutputFormatFromArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"--output json", []string{"check", "--output", "json"}, "json"},
		{"--output=json", []string{"check", "--output=json"}, "json"},
		{"-o json", []string{"check", "-o", "json"}, "json"},
		{"-o=json", []string{"check", "-o=json"}, "json"},
		{"no flag", []string{"check", "target"}, ""},
		{"--output at end", []string{"check", "--output"}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &GlobalOptions{}
			detectOutputFormatFromArgs(opts, tt.args)
			if opts.OutputFormat != tt.want {
				t.Errorf("OutputFormat = %q, want %q", opts.OutputFormat, tt.want)
			}
		})
	}
}

func TestPrintVersion_TextMode(t *testing.T) {
	var buf bytes.Buffer
	opts := &GlobalOptions{
		Out: newTestOutput(&buf, "text"),
	}

	err := printVersion(opts)
	if err != nil {
		t.Fatalf("printVersion returned error: %v", err)
	}

	got := strings.TrimSpace(buf.String())
	if got != Version {
		t.Errorf("expected %q, got %q", Version, got)
	}
}

func TestPrintVersion_JSONMode(t *testing.T) {
	var buf bytes.Buffer
	opts := &GlobalOptions{
		Out: newTestOutput(&buf, "json"),
	}

	err := printVersion(opts)
	if err != nil {
		t.Fatalf("printVersion returned error: %v", err)
	}

	var result map[string]string
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nraw: %s", err, buf.String())
	}

	if result["version"] != Version {
		t.Errorf("expected version %q, got %q", Version, result["version"])
	}
}
