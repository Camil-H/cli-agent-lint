package checks

import (
	"context"
	"fmt"
	"sync"

	"github.com/cli-agent-lint/cli-agent-lint/discovery"
	"github.com/cli-agent-lint/cli-agent-lint/probe"
)

// Prober is the consumer-side type alias so checks depend on the probe.Runner interface.
type Prober = probe.Runner

type Severity int

const (
	Info Severity = iota
	Warn
	Fail
)

func (s Severity) String() string {
	switch s {
	case Info:
		return "info"
	case Warn:
		return "warn"
	case Fail:
		return "fail"
	default:
		return "unknown"
	}
}

// Points returns the score contribution for a passing check at this severity.
func (s Severity) Points() int {
	switch s {
	case Info:
		return 1
	case Warn:
		return 2
	case Fail:
		return 3
	default:
		return 0
	}
}

func ParseSeverity(s string) (Severity, error) {
	switch s {
	case "info":
		return Info, nil
	case "warn":
		return Warn, nil
	case "fail":
		return Fail, nil
	default:
		return Info, fmt.Errorf("unknown severity: %q (valid: info, warn, fail)", s)
	}
}

type Category string

const (
	CatFlowSafety      Category = "flow-safety"
	CatTokenEfficiency  Category = "token-efficiency"
	CatSelfDescribing   Category = "self-describing"
	CatAutomationSafety Category = "automation-safety"
	CatPredictability   Category = "predictability"
)

// AllCategories returns all valid categories in display order.
func AllCategories() []Category {
	return []Category{
		CatFlowSafety,
		CatTokenEfficiency,
		CatSelfDescribing,
		CatAutomationSafety,
		CatPredictability,
	}
}

type Status int

const (
	StatusPass Status = iota
	StatusFail
	StatusSkip
	StatusError
)

func (s Status) String() string {
	switch s {
	case StatusPass:
		return "pass"
	case StatusFail:
		return "fail"
	case StatusSkip:
		return "skip"
	case StatusError:
		return "error"
	default:
		return "unknown"
	}
}

type Method int

const (
	Passive Method = iota
	Active
)

func (m Method) String() string {
	switch m {
	case Passive:
		return "passive"
	case Active:
		return "active"
	default:
		return "unknown"
	}
}

type Result struct {
	CheckID        string
	CheckName      string
	Category       Category
	Severity       Severity
	Status         Status
	Method         Method
	Recommendation string
	Detail         string
	Error          error
}

func (r *Result) Points() int {
	if r.Status == StatusPass {
		return r.Severity.Points()
	}
	return 0
}

// MaxPoints returns the maximum possible points.
// Skipped and errored checks contribute 0 to both earned and possible.
func (r *Result) MaxPoints() int {
	if r.Status == StatusSkip || r.Status == StatusError {
		return 0
	}
	return r.Severity.Points()
}

type Input struct {
	Tree      *discovery.CommandTree
	Index     *discovery.CommandIndex
	Prober    Prober     // nil when --no-probe
	ResultSet *ResultSet // for cross-check dependencies
}

func (inp *Input) GetIndex() *discovery.CommandIndex {
	return inp.Index
}

// ResultSet provides thread-safe access to results from other checks.
type ResultSet struct {
	mu      sync.RWMutex
	results map[string]*Result
}

func NewResultSet() *ResultSet {
	return &ResultSet{results: make(map[string]*Result)}
}

func (rs *ResultSet) Set(id string, r *Result) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.results[id] = r
}

func (rs *ResultSet) Get(id string) *Result {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return rs.results[id]
}

type Check interface {
	ID() string
	Name() string
	Category() Category
	Severity() Severity
	Method() Method
	Recommendation() string
	Run(ctx context.Context, input *Input) *Result
}

// BaseCheck provides default implementations for Check metadata methods.
type BaseCheck struct {
	CheckID             string
	CheckName           string
	CheckCategory       Category
	CheckSeverity       Severity
	CheckMethod         Method
	CheckRecommendation string
}

func (b *BaseCheck) ID() string             { return b.CheckID }
func (b *BaseCheck) Name() string           { return b.CheckName }
func (b *BaseCheck) Category() Category     { return b.CheckCategory }
func (b *BaseCheck) Severity() Severity     { return b.CheckSeverity }
func (b *BaseCheck) Method() Method         { return b.CheckMethod }
func (b *BaseCheck) Recommendation() string { return b.CheckRecommendation }

func PassResult(c Check, detail string) *Result {
	return &Result{
		CheckID:        c.ID(),
		CheckName:      c.Name(),
		Category:       c.Category(),
		Severity:       c.Severity(),
		Status:         StatusPass,
		Method:         c.Method(),
		Recommendation: c.Recommendation(),
		Detail:         detail,
	}
}

func FailResult(c Check, detail string) *Result {
	return &Result{
		CheckID:        c.ID(),
		CheckName:      c.Name(),
		Category:       c.Category(),
		Severity:       c.Severity(),
		Status:         StatusFail,
		Method:         c.Method(),
		Recommendation: c.Recommendation(),
		Detail:         detail,
	}
}

func SkipResult(c Check, detail string) *Result {
	return &Result{
		CheckID:        c.ID(),
		CheckName:      c.Name(),
		Category:       c.Category(),
		Severity:       c.Severity(),
		Status:         StatusSkip,
		Method:         c.Method(),
		Recommendation: c.Recommendation(),
		Detail:         detail,
	}
}

func ErrorResult(c Check, err error) *Result {
	return &Result{
		CheckID:        c.ID(),
		CheckName:      c.Name(),
		Category:       c.Category(),
		Severity:       c.Severity(),
		Status:         StatusError,
		Method:         c.Method(),
		Recommendation: c.Recommendation(),
		Error:          err,
	}
}

// SkipActiveResult returns a skip result for active checks when --no-probe is set.
func SkipActiveResult(c Check) *Result {
	return SkipResult(c, "skipped: active check disabled by --no-probe")
}

// skipIfNoProber returns a SkipActiveResult if the prober is nil.
// Returns nil if the prober is available and the check should proceed.
func skipIfNoProber(c Check, input *Input) *Result {
	if input.Prober == nil {
		return SkipActiveResult(c)
	}
	return nil
}
