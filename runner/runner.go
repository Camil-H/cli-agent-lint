package runner

import (
	"context"
	"runtime"
	"sync"
	"time"

	"github.com/cli-agent-lint/cli-agent-lint/checks"
	"github.com/cli-agent-lint/cli-agent-lint/discovery"
	"github.com/cli-agent-lint/cli-agent-lint/probe"
	"github.com/cli-agent-lint/cli-agent-lint/report"
)

type Config struct {
	TargetPath   string
	Subcommands  []string
	ProbeTimeout time.Duration
	NoProbe      bool
	Concurrency  int // max parallel active checks; 0 defaults to runtime.NumCPU()
	Filter       *checks.Filter
	Registry     *checks.Registry
}

type ProgressFunc func(phase string, current, total int, detail string)

// Runner orchestrates discovery and check execution.
type Runner struct {
	cfg Config
}

func New(cfg Config) *Runner {
	return &Runner{cfg: cfg}
}

func (r *Runner) Run(ctx context.Context, progressFn ProgressFunc) (*report.Report, error) {
	if progressFn == nil {
		progressFn = func(string, int, int, string) {}
	}

	start := time.Now()

	// Always needed for --help discovery, even in no-probe mode.
	discoveryProber, err := probe.New(r.cfg.TargetPath, r.cfg.ProbeTimeout)
	if err != nil {
		return nil, err
	}
	defer discoveryProber.Close()

	// Use the Prober interface type to avoid the nil-pointer-in-interface pitfall:
	// a nil *probe.Prober assigned to an interface field is not == nil.
	var prober checks.Prober
	if !r.cfg.NoProbe {
		prober = discoveryProber
	}

	progressFn("discovery", 0, 0, "discovering command tree")
	tree, err := discovery.Discover(ctx, discoveryProber, discovery.DiscoverOpts{
		Subcommands: r.cfg.Subcommands,
	})
	if err != nil {
		return nil, err
	}

	// Pre-computed index for O(1) lookups in checks.
	var index *discovery.CommandIndex
	if tree.Root != nil {
		index = discovery.NewIndex(tree.Root)
	}

	selected := r.cfg.Registry.Filter(r.cfg.Filter)
	total := len(selected)

	resultSet := checks.NewResultSet()
	input := &checks.Input{
		Tree:      tree,
		Index:     index,
		Prober:    prober,
		ResultSet: resultSet,
	}

	var passive, active []checks.Check
	for _, c := range selected {
		if c.Method() == checks.Passive {
			passive = append(passive, c)
		} else {
			active = append(active, c)
		}
	}

	results := make([]*checks.Result, 0, total)
	var mu sync.Mutex
	current := 0

	addResult := func(res *checks.Result) {
		mu.Lock()
		defer mu.Unlock()
		results = append(results, res)
		resultSet.Set(res.CheckID, res)
		current++
		progressFn("checks", current, total, res.CheckID)
	}

	var wg sync.WaitGroup
	for _, c := range passive {
		wg.Add(1)
		go func(c checks.Check) {
			defer wg.Done()
			res := c.Run(ctx, input)
			addResult(res)
		}(c)
	}
	wg.Wait()

	// Active checks run after passive ones so cross-check dependencies
	// (e.g., SD-1 reads TE-1) are already satisfied.
	concurrency := r.cfg.Concurrency
	if concurrency <= 0 {
		concurrency = max(4, runtime.NumCPU())
	}
	sem := make(chan struct{}, concurrency)

	for _, c := range active {
		wg.Add(1)
		go func(c checks.Check) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			res := c.Run(ctx, input)
			addResult(res)
		}(c)
	}
	wg.Wait()

	duration := time.Since(start)

	var version string
	if so4 := resultSet.Get("SD-2"); so4 != nil && so4.Status == checks.StatusPass {
		version = so4.Detail
	}

	return report.NewReport(results, r.cfg.TargetPath, version, duration), nil
}
