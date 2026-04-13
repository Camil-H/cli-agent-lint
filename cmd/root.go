package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Camil-H/cli-agent-lint/output"
)

// Version is set at build time via ldflags.
var Version = "0.3.0"

type GlobalOptions struct {
	OutputFormat string
	NoColor      bool
	Quiet        bool
	Out          *output.Output
}

func initOutput(opts *GlobalOptions) {
	opts.Out = output.New(output.Config{
		Format:  opts.OutputFormat,
		NoColor: opts.NoColor,
		Quiet:   opts.Quiet,
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
	})
}

func NewRootCmd(opts *GlobalOptions) *cobra.Command {
	var showVersion bool

	rootCmd := &cobra.Command{
		Use:           "cli-agent-lint",
		Short:         "Audit CLI tools for AI agent-readiness",
		Long: `Audit CLI tools for AI agent-readiness.

Exit Codes:
  0  Success (all checks passed or grade above threshold)
  1  Check failure (grade below threshold)
  2  Usage or runtime error`,
		Example: `  cli-agent-lint check ./my-tool
  cli-agent-lint check --output json ./my-tool
  cli-agent-lint checks --list`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := rejectControlChars(args); err != nil {
				return err
			}
			initOutput(opts)
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if showVersion {
				initOutput(opts)
				return printVersion(opts)
			}
			return cmd.Help()
		},
	}

	pf := rootCmd.PersistentFlags()
	pf.StringVarP(&opts.OutputFormat, "output", "o", "text", "Output format: text, json")
	pf.BoolVar(&opts.NoColor, "no-color", false, "Disable colored output")
	pf.BoolVarP(&opts.Quiet, "quiet", "q", false, "Suppress informational output")

	// Custom --version flag (not cobra's built-in, to avoid "appname version X" format).
	rootCmd.Flags().BoolVarP(&showVersion, "version", "v", false, "Print version")

	rootCmd.SuggestionsMinimumDistance = 2

	rootCmd.AddCommand(newCheckCmd(opts))
	rootCmd.AddCommand(newChecksCmd(opts))
	rootCmd.AddCommand(newCompletionCmd(opts))

	return rootCmd
}

func Execute() int {
	opts := &GlobalOptions{}

	rootCmd := NewRootCmd(opts)
	err := rootCmd.Execute()

	// PersistentPreRunE may not have run for unknown subcommands or flag parse
	// errors, so scan os.Args for --output flags to ensure JSON errors work.
	if opts.Out == nil {
		detectOutputFormat(opts)
		initOutput(opts)
	}

	if err == nil {
		return 0
	}

	var checkErr *output.CheckFailedError
	if errors.As(err, &checkErr) {
		return 1
	}
	errMsg := err.Error()
	if strings.Contains(errMsg, "unknown command") && !strings.Contains(errMsg, "Did you mean") {
		err = fmt.Errorf("%s\n\nRun 'cli-agent-lint --help' for available commands.", errMsg)
	}
	opts.Out.Error(err)
	return 2
}

// detectOutputFormat scans os.Args for --output or -o flags when cobra's
// flag parsing hasn't run (e.g. unknown subcommand errors).
func detectOutputFormat(opts *GlobalOptions) {
	detectOutputFormatFromArgs(opts, os.Args[1:])
}

func detectOutputFormatFromArgs(opts *GlobalOptions, args []string) {
	for i, a := range args {
		switch {
		case a == "--output" || a == "-o":
			if i+1 < len(args) {
				opts.OutputFormat = args[i+1]
			}
		case strings.HasPrefix(a, "--output="):
			opts.OutputFormat = strings.TrimPrefix(a, "--output=")
		case strings.HasPrefix(a, "-o="):
			opts.OutputFormat = strings.TrimPrefix(a, "-o=")
		}
	}
}

// containsControlChars returns true if s contains ASCII control characters
// (< 0x20) other than tab, newline, and carriage return.
func containsControlChars(s string) bool {
	for _, r := range s {
		if r < 0x20 && r != '\t' && r != '\n' && r != '\r' {
			return true
		}
	}
	return false
}

func rejectControlChars(args []string) error {
	for _, a := range args {
		if containsControlChars(a) {
			return fmt.Errorf("argument %q contains control characters (ASCII < 0x20)", a)
		}
	}
	// Also validate os.Args for flag values that cobra already parsed.
	for _, a := range os.Args[1:] {
		if containsControlChars(a) {
			return fmt.Errorf("input contains control characters (ASCII < 0x20)")
		}
	}
	return nil
}

func printVersion(opts *GlobalOptions) error {
	if opts.Out != nil && opts.Out.IsJSON() {
		enc := json.NewEncoder(opts.Out.DataWriter())
		enc.SetIndent("", "  ")
		return enc.Encode(map[string]string{"version": Version})
	}
	fmt.Fprintln(opts.Out.DataWriter(), Version)
	return nil
}
