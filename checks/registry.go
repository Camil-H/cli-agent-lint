package checks

import (
	"fmt"
	"sort"
)

type Registry struct {
	byID    map[string]Check
	ordered []Check
}

func NewRegistry() *Registry {
	return &Registry{
		byID: make(map[string]Check),
	}
}

// Register adds a check to the registry. Panics on duplicate ID.
func (r *Registry) Register(c Check) {
	if _, exists := r.byID[c.ID()]; exists {
		panic(fmt.Sprintf("duplicate check ID: %s", c.ID()))
	}
	r.byID[c.ID()] = c
	r.ordered = append(r.ordered, c)
}

// All returns checks in registration order.
func (r *Registry) All() []Check {
	result := make([]Check, len(r.ordered))
	copy(result, r.ordered)
	return result
}

func (r *Registry) Get(id string) Check {
	return r.byID[id]
}

func (r *Registry) Filter(f *Filter) []Check {
	if f == nil {
		return r.All()
	}
	var result []Check
	for _, c := range r.ordered {
		if f.matches(c) {
			result = append(result, c)
		}
	}
	return result
}

func (r *Registry) ByCategory() map[Category][]Check {
	result := make(map[Category][]Check)
	for _, c := range r.ordered {
		result[c.Category()] = append(result[c.Category()], c)
	}
	return result
}

// CategoryNames returns categories that have checks, in stable order.
func (r *Registry) CategoryNames() []Category {
	seen := make(map[Category]bool)
	var cats []Category
	for _, c := range r.ordered {
		if !seen[c.Category()] {
			seen[c.Category()] = true
			cats = append(cats, c.Category())
		}
	}
	return cats
}

func (r *Registry) IDs() []string {
	ids := make([]string, len(r.ordered))
	for i, c := range r.ordered {
		ids[i] = c.ID()
	}
	return ids
}

func (r *Registry) Len() int { return len(r.ordered) }

type Filter struct {
	Category    Category // empty means all
	MinSeverity *Severity
	Method      *Method
	SkipIDs     map[string]bool
}

func (f *Filter) matches(c Check) bool {
	if f.Category != "" && c.Category() != f.Category {
		return false
	}
	if f.MinSeverity != nil && c.Severity() < *f.MinSeverity {
		return false
	}
	if f.Method != nil && c.Method() != *f.Method {
		return false
	}
	if f.SkipIDs != nil && f.SkipIDs[c.ID()] {
		return false
	}
	return true
}

func DefaultRegistry() *Registry {
	r := NewRegistry()
	registerBuiltinChecks(r)
	return r
}

// registerBuiltinChecks registers all check implementations.
// Each category file provides its own registration function.
func registerBuiltinChecks(r *Registry) {
	registerStructuredOutputChecks(r)
	registerTerminalHygieneChecks(r)
	registerInputValidationChecks(r)
	registerSchemaDiscoveryChecks(r)
	registerAuthChecks(r)
	registerOperationalChecks(r)
}

// SortResultsByCategory sorts results by category order, then by check ID.
func SortResultsByCategory(results []*Result) {
	catOrder := make(map[Category]int)
	for i, c := range AllCategories() {
		catOrder[c] = i
	}
	sort.SliceStable(results, func(i, j int) bool {
		ci := catOrder[results[i].Category]
		cj := catOrder[results[j].Category]
		if ci != cj {
			return ci < cj
		}
		return results[i].CheckID < results[j].CheckID
	})
}
