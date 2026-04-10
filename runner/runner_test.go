package runner

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cli-agent-lint/cli-agent-lint/checks"
)

// goodCLIPath returns the absolute path to the good-cli.sh test fixture.
func goodCLIPath(t *testing.T) string {
	t.Helper()
	// Navigate from the runner package to the project root's testdata directory.
	path, err := filepath.Abs(filepath.Join("..", "testdata", "fake-cli", "good-cli.sh"))
	if err != nil {
		t.Fatalf("failed to resolve good-cli.sh path: %v", err)
	}
	// Ensure the script is executable.
	if err := os.Chmod(path, 0o755); err != nil {
		t.Fatalf("failed to chmod good-cli.sh: %v", err)
	}
	return path
}

func TestNew_ReturnsNonNilRunner(t *testing.T) {
	cfg := Config{
		TargetPath: "/some/path",
		Registry:   checks.DefaultRegistry(),
	}
	r := New(cfg)
	if r == nil {
		t.Fatal("expected New() to return a non-nil Runner")
	}
}

func TestNew_StoresConfig(t *testing.T) {
	cfg := Config{
		TargetPath:   "/usr/bin/some-cli",
		NoProbe:      true,
		Concurrency:  4,
		ProbeTimeout: 10 * time.Second,
		Registry:     checks.DefaultRegistry(),
	}
	r := New(cfg)
	if r.cfg.TargetPath != cfg.TargetPath {
		t.Errorf("expected TargetPath %q, got %q", cfg.TargetPath, r.cfg.TargetPath)
	}
	if r.cfg.NoProbe != cfg.NoProbe {
		t.Errorf("expected NoProbe %v, got %v", cfg.NoProbe, r.cfg.NoProbe)
	}
	if r.cfg.Concurrency != cfg.Concurrency {
		t.Errorf("expected Concurrency %d, got %d", cfg.Concurrency, r.cfg.Concurrency)
	}
	if r.cfg.ProbeTimeout != cfg.ProbeTimeout {
		t.Errorf("expected ProbeTimeout %v, got %v", cfg.ProbeTimeout, r.cfg.ProbeTimeout)
	}
}

func TestRun_GoodCLI_NoError(t *testing.T) {
	target := goodCLIPath(t)
	registry := checks.DefaultRegistry()

	r := New(Config{
		TargetPath:   target,
		Registry:     registry,
		ProbeTimeout: 10 * time.Second,
	})

	var progressCalls int64
	progressFn := func(phase string, current, total int, detail string) {
		atomic.AddInt64(&progressCalls, 1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	report, err := r.Run(ctx, progressFn)
	if err != nil {
		t.Fatalf("Run() returned unexpected error: %v", err)
	}
	if report == nil {
		t.Fatal("Run() returned nil report")
	}
	if len(report.Results) == 0 {
		t.Fatal("expected report to contain results, got 0")
	}
	if atomic.LoadInt64(&progressCalls) == 0 {
		t.Error("expected progress callback to be invoked at least once")
	}
}

func TestRun_GoodCLI_ReportHasAllChecks(t *testing.T) {
	target := goodCLIPath(t)
	registry := checks.DefaultRegistry()
	expectedCount := registry.Len()

	r := New(Config{
		TargetPath:   target,
		Registry:     registry,
		ProbeTimeout: 10 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	report, err := r.Run(ctx, nil)
	if err != nil {
		t.Fatalf("Run() returned unexpected error: %v", err)
	}
	if len(report.Results) != expectedCount {
		t.Errorf("expected %d results (one per registered check), got %d", expectedCount, len(report.Results))
	}
}

func TestRun_GoodCLI_ProgressPhasesIncludeDiscoveryAndChecks(t *testing.T) {
	target := goodCLIPath(t)
	registry := checks.DefaultRegistry()

	r := New(Config{
		TargetPath:   target,
		Registry:     registry,
		ProbeTimeout: 10 * time.Second,
	})

	phases := make(map[string]bool)
	progressFn := func(phase string, current, total int, detail string) {
		phases[phase] = true
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	_, err := r.Run(ctx, progressFn)
	if err != nil {
		t.Fatalf("Run() returned unexpected error: %v", err)
	}
	if !phases["discovery"] {
		t.Error("expected 'discovery' phase in progress callbacks")
	}
	if !phases["checks"] {
		t.Error("expected 'checks' phase in progress callbacks")
	}
}

func TestRun_NilProgressFunc(t *testing.T) {
	target := goodCLIPath(t)
	registry := checks.DefaultRegistry()

	r := New(Config{
		TargetPath:   target,
		Registry:     registry,
		ProbeTimeout: 10 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Should not panic with nil progressFn.
	report, err := r.Run(ctx, nil)
	if err != nil {
		t.Fatalf("Run() with nil progressFn returned unexpected error: %v", err)
	}
	if report == nil {
		t.Fatal("Run() with nil progressFn returned nil report")
	}
}

func TestRun_NoProbe_ActiveChecksSkipped(t *testing.T) {
	target := goodCLIPath(t)
	registry := checks.DefaultRegistry()

	r := New(Config{
		TargetPath:   target,
		Registry:     registry,
		NoProbe:      true,
		ProbeTimeout: 10 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	report, err := r.Run(ctx, nil)
	if err != nil {
		t.Fatalf("Run() with NoProbe returned unexpected error: %v", err)
	}
	if report == nil {
		t.Fatal("Run() with NoProbe returned nil report")
	}

	// Build a set of active check IDs from the registry.
	activeIDs := make(map[string]bool)
	for _, c := range registry.All() {
		if c.Method() == checks.Active {
			activeIDs[c.ID()] = true
		}
	}

	if len(activeIDs) == 0 {
		t.Fatal("expected at least one active check in the default registry")
	}

	for _, result := range report.Results {
		if activeIDs[result.CheckID] {
			if result.Status != checks.StatusSkip {
				t.Errorf("active check %s: expected Status=Skip with NoProbe, got %s", result.CheckID, result.Status)
			}
		}
	}
}

func TestRun_NoProbe_PassiveChecksStillRun(t *testing.T) {
	target := goodCLIPath(t)
	registry := checks.DefaultRegistry()

	r := New(Config{
		TargetPath:   target,
		Registry:     registry,
		NoProbe:      true,
		ProbeTimeout: 10 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	report, err := r.Run(ctx, nil)
	if err != nil {
		t.Fatalf("Run() returned unexpected error: %v", err)
	}

	// Passive checks should not all be skipped; at least some should pass or fail.
	passiveNonSkip := 0
	for _, result := range report.Results {
		if result.Method == checks.Passive && result.Status != checks.StatusSkip {
			passiveNonSkip++
		}
	}
	if passiveNonSkip == 0 {
		t.Error("expected at least some passive checks to run (not Skip) in NoProbe mode")
	}
}

func TestRun_ConcurrencyDefaultsToNumCPU(t *testing.T) {
	// When Concurrency is 0, the runner should default to runtime.NumCPU().
	// We cannot directly observe the semaphore size, but we can verify the
	// runner completes successfully with Concurrency=0 and produces the
	// same number of results as with an explicit concurrency.
	target := goodCLIPath(t)
	registry := checks.DefaultRegistry()

	rDefault := New(Config{
		TargetPath:   target,
		Registry:     registry,
		Concurrency:  0, // should default to runtime.NumCPU()
		ProbeTimeout: 10 * time.Second,
	})

	rExplicit := New(Config{
		TargetPath:   target,
		Registry:     registry,
		Concurrency:  runtime.NumCPU(),
		ProbeTimeout: 10 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	reportDefault, err := rDefault.Run(ctx, nil)
	if err != nil {
		t.Fatalf("Run() with Concurrency=0 returned error: %v", err)
	}

	reportExplicit, err := rExplicit.Run(ctx, nil)
	if err != nil {
		t.Fatalf("Run() with Concurrency=%d returned error: %v", runtime.NumCPU(), err)
	}

	if len(reportDefault.Results) != len(reportExplicit.Results) {
		t.Errorf("expected same number of results: default=%d, explicit=%d",
			len(reportDefault.Results), len(reportExplicit.Results))
	}
}

func TestRun_InvalidTarget_ReturnsError(t *testing.T) {
	r := New(Config{
		TargetPath:   "/nonexistent/binary/that/does/not/exist",
		Registry:     checks.DefaultRegistry(),
		ProbeTimeout: 5 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	report, err := r.Run(ctx, nil)
	if err == nil {
		t.Fatal("expected Run() to return an error for a nonexistent target")
	}
	if report != nil {
		t.Error("expected nil report when Run() returns an error")
	}
}

func TestRun_GoodCLI_ReportTargetPath(t *testing.T) {
	target := goodCLIPath(t)
	registry := checks.DefaultRegistry()

	r := New(Config{
		TargetPath:   target,
		Registry:     registry,
		ProbeTimeout: 10 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	report, err := r.Run(ctx, nil)
	if err != nil {
		t.Fatalf("Run() returned unexpected error: %v", err)
	}
	if report.TargetPath != target {
		t.Errorf("expected report.TargetPath=%q, got %q", target, report.TargetPath)
	}
}

func TestRun_GoodCLI_ReportHasDuration(t *testing.T) {
	target := goodCLIPath(t)
	registry := checks.DefaultRegistry()

	r := New(Config{
		TargetPath:   target,
		Registry:     registry,
		ProbeTimeout: 10 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	report, err := r.Run(ctx, nil)
	if err != nil {
		t.Fatalf("Run() returned unexpected error: %v", err)
	}
	if report.Duration <= 0 {
		t.Errorf("expected positive Duration, got %v", report.Duration)
	}
}

func TestRun_WithFilter_SubsetOfResults(t *testing.T) {
	target := goodCLIPath(t)
	registry := checks.DefaultRegistry()

	// Run with a category filter to get only token-efficiency checks.
	r := New(Config{
		TargetPath:   target,
		Registry:     registry,
		ProbeTimeout: 10 * time.Second,
		Filter:       &checks.Filter{Category: checks.CatTokenEfficiency},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	report, err := r.Run(ctx, nil)
	if err != nil {
		t.Fatalf("Run() with filter returned unexpected error: %v", err)
	}

	// Every result should be in the token-efficiency category.
	for _, result := range report.Results {
		if result.Category != checks.CatTokenEfficiency {
			t.Errorf("expected all results to be in category %q, got %q for check %s",
				checks.CatTokenEfficiency, result.Category, result.CheckID)
		}
	}

	// Should be fewer results than the full registry.
	if len(report.Results) >= registry.Len() {
		t.Errorf("expected filtered results (%d) to be fewer than total checks (%d)",
			len(report.Results), registry.Len())
	}
}
