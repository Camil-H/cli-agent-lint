package cmd

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/Camil-H/cli-agent-lint/checks"
	"github.com/Camil-H/cli-agent-lint/output"
	"github.com/Camil-H/cli-agent-lint/report"
	"github.com/Camil-H/cli-agent-lint/runner"
)

type checkFlags struct {
	severity string
	category string
	skip     []string
	timeout  time.Duration
	noProbe  bool
}

func newCheckCmd(opts *GlobalOptions) *cobra.Command {
	f := &checkFlags{}

	cmd := &cobra.Command{
		Use:   "check <target-cli> [-- subcommand1 subcommand2 ...]",
		Short: "Audit a CLI tool for AI agent-readiness",
		Long:  "Run checks against a target CLI and produce a scorecard report.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCheck(opts, args, f)
		},
	}

	cmd.Flags().StringVar(&f.severity, "severity", "info", "Minimum severity to report: info, warn, fail")
	cmd.Flags().StringVar(&f.category, "category", "", "Run checks from a specific category only")
	cmd.Flags().StringSliceVar(&f.skip, "skip", nil, "Skip specific checks (repeatable)")
	cmd.Flags().DurationVar(&f.timeout, "timeout", 5*time.Second, "Timeout for each probe command")
	cmd.Flags().BoolVar(&f.noProbe, "no-probe", false, "Passive only: parse help text, don't run commands")

	return cmd
}

func runCheck(opts *GlobalOptions, args []string, f *checkFlags) error {
	targetPath := args[0]

	// Subcommands come after --.
	var subcommands []string
	if len(args) > 1 {
		subcommands = args[1:]
	}

	filter := &checks.Filter{}

	if f.severity != "" {
		sev, err := checks.ParseSeverity(f.severity)
		if err != nil {
			return err
		}
		filter.MinSeverity = &sev
	}

	if f.category != "" {
		filter.Category = checks.Category(f.category)
	}

	if len(f.skip) > 0 {
		filter.SkipIDs = make(map[string]bool)
		for _, id := range f.skip {
			filter.SkipIDs[id] = true
		}
	}

	registry := checks.DefaultRegistry()
	r := runner.New(runner.Config{
		TargetPath:   targetPath,
		Subcommands:  subcommands,
		ProbeTimeout: f.timeout,
		NoProbe:      f.noProbe,
		Filter:       filter,
		Registry:     registry,
	})

	out := opts.Out
	progressFn := func(phase string, current, total int, detail string) {
		if phase == "discovery" {
			out.Diag("Discovering command tree...")
		} else if phase == "checks" && total > 0 {
			out.Diag("Running checks... [%d/%d] %s", current, total, detail)
		}
	}

	ctx := context.Background()
	rpt, err := r.Run(ctx, progressFn)
	if err != nil {
		return err
	}

	if out.IsJSON() {
		jf := &report.JSONFormatter{}
		if err := jf.Format(out.DataWriter(), rpt); err != nil {
			return err
		}
	} else {
		tf := &report.TextFormatter{
			NoColor: out.NoColor(),
			Quiet:   out.IsQuiet(),
		}
		if err := tf.Format(out.DataWriter(), rpt); err != nil {
			return err
		}
	}

	// Exit code 1 if critical failures.
	if rpt.HasCriticalFailures() {
		return &output.CheckFailedError{}
	}

	return nil
}
