package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Camil-H/cli-agent-lint/checks"
)

func newChecksCmd(opts *GlobalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "checks [check-id]",
		Short: "List all checks or describe a specific check",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			registry := checks.DefaultRegistry()

			if len(args) == 1 {
				return describeCheck(opts, registry, args[0])
			}

			if opts.Out.IsJSON() {
				return listChecksJSON(opts, registry)
			}
			return listChecksText(opts, registry)
		},
	}
	return cmd
}

func listChecksText(opts *GlobalOptions, registry *checks.Registry) error {
	w := opts.Out.DataWriter()

	byCategory := registry.ByCategory()
	for _, cat := range registry.CategoryNames() {
		fmt.Fprintf(w, "%s\n", cat)
		for _, c := range byCategory[cat] {
			fmt.Fprintf(w, "  %-6s %-8s [%s] %s\n", c.ID(), c.Severity(), c.Method(), c.Name())
		}
		fmt.Fprintln(w)
	}
	return nil
}

type jsonCheckInfo struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Category       string `json:"category"`
	Severity       string `json:"severity"`
	Method         string `json:"method"`
	Recommendation string `json:"recommendation"`
}

func listChecksJSON(opts *GlobalOptions, registry *checks.Registry) error {
	w := opts.Out.DataWriter()

	var items []jsonCheckInfo
	for _, c := range registry.All() {
		items = append(items, jsonCheckInfo{
			ID:             c.ID(),
			Name:           c.Name(),
			Category:       string(c.Category()),
			Severity:       c.Severity().String(),
			Method:         c.Method().String(),
			Recommendation: c.Recommendation(),
		})
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(items)
}

func describeCheck(opts *GlobalOptions, registry *checks.Registry, id string) error {
	w := opts.Out.DataWriter()

	c := registry.Get(id)
	if c == nil {
		return fmt.Errorf("unknown check ID: %q", id)
	}

	if opts.Out.IsJSON() {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(jsonCheckInfo{
			ID:             c.ID(),
			Name:           c.Name(),
			Category:       string(c.Category()),
			Severity:       c.Severity().String(),
			Method:         c.Method().String(),
			Recommendation: c.Recommendation(),
		})
	}

	fmt.Fprintf(w, "ID:             %s\n", c.ID())
	fmt.Fprintf(w, "Name:           %s\n", c.Name())
	fmt.Fprintf(w, "Category:       %s\n", c.Category())
	fmt.Fprintf(w, "Severity:       %s\n", c.Severity())
	fmt.Fprintf(w, "Method:         %s\n", c.Method())
	fmt.Fprintf(w, "Recommendation: %s\n", c.Recommendation())

	return nil
}
