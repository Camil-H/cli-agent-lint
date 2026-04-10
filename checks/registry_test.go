package checks

import (
	"context"
	"testing"
)

type dummyCheck struct {
	BaseCheck
}

func (d *dummyCheck) Run(_ context.Context, _ *Input) *Result { return nil }

func newDummy(id string, cat Category, sev Severity, method Method) *dummyCheck {
	return &dummyCheck{
		BaseCheck: BaseCheck{
			CheckID:       id,
			CheckName:     "Dummy " + id,
			CheckCategory: cat,
			CheckSeverity: sev,
			CheckMethod:   method,
		},
	}
}

func TestRegistry_Register_And_Get(t *testing.T) {
	r := NewRegistry()
	c := newDummy("T-1", CatTokenEfficiency, Warn, Passive)
	r.Register(c)

	got := r.Get("T-1")
	if got == nil {
		t.Fatal("expected to get check T-1 but got nil")
	}
	if got.ID() != "T-1" {
		t.Errorf("expected ID T-1, got %s", got.ID())
	}
}

func TestRegistry_Get_NotFound(t *testing.T) {
	r := NewRegistry()
	if got := r.Get("nonexistent"); got != nil {
		t.Errorf("expected nil for nonexistent ID, got %v", got)
	}
}

func TestRegistry_All_ReturnsRegistrationOrder(t *testing.T) {
	r := NewRegistry()
	r.Register(newDummy("A-1", CatTokenEfficiency, Warn, Passive))
	r.Register(newDummy("B-2", CatFlowSafety, Fail, Active))
	r.Register(newDummy("C-3", CatSelfDescribing, Info, Passive))

	all := r.All()
	if len(all) != 3 {
		t.Fatalf("expected 3 checks, got %d", len(all))
	}
	expectedIDs := []string{"A-1", "B-2", "C-3"}
	for i, c := range all {
		if c.ID() != expectedIDs[i] {
			t.Errorf("index %d: expected ID %s, got %s", i, expectedIDs[i], c.ID())
		}
	}
}

func TestRegistry_All_ReturnsCopy(t *testing.T) {
	r := NewRegistry()
	r.Register(newDummy("X-1", CatTokenEfficiency, Warn, Passive))

	all1 := r.All()
	all2 := r.All()

	// Mutating the first slice should not affect the second.
	all1[0] = nil
	if all2[0] == nil {
		t.Error("All() should return a copy, but mutation propagated")
	}
}

func TestRegistry_Len(t *testing.T) {
	r := NewRegistry()
	if r.Len() != 0 {
		t.Errorf("expected empty registry, got len %d", r.Len())
	}
	r.Register(newDummy("L-1", CatFlowSafety, Warn, Passive))
	r.Register(newDummy("L-2", CatFlowSafety, Fail, Active))
	if r.Len() != 2 {
		t.Errorf("expected len 2, got %d", r.Len())
	}
}

func TestRegistry_IDs(t *testing.T) {
	r := NewRegistry()
	r.Register(newDummy("ID-A", CatTokenEfficiency, Warn, Passive))
	r.Register(newDummy("ID-B", CatFlowSafety, Fail, Active))

	ids := r.IDs()
	if len(ids) != 2 {
		t.Fatalf("expected 2 IDs, got %d", len(ids))
	}
	if ids[0] != "ID-A" || ids[1] != "ID-B" {
		t.Errorf("unexpected IDs: %v", ids)
	}
}

func TestRegistry_DuplicatePanics(t *testing.T) {
	r := NewRegistry()
	r.Register(newDummy("DUP-1", CatTokenEfficiency, Warn, Passive))

	defer func() {
		rec := recover()
		if rec == nil {
			t.Fatal("expected panic on duplicate registration, but did not panic")
		}
		msg, ok := rec.(string)
		if !ok {
			t.Fatalf("expected string panic, got %T: %v", rec, rec)
		}
		if msg != "duplicate check ID: DUP-1" {
			t.Errorf("unexpected panic message: %s", msg)
		}
	}()

	r.Register(newDummy("DUP-1", CatFlowSafety, Fail, Active))
}

func TestRegistry_Filter_Nil(t *testing.T) {
	r := NewRegistry()
	r.Register(newDummy("F-1", CatTokenEfficiency, Warn, Passive))
	r.Register(newDummy("F-2", CatFlowSafety, Fail, Active))

	result := r.Filter(nil)
	if len(result) != 2 {
		t.Errorf("nil filter should return all, got %d", len(result))
	}
}

func TestRegistry_Filter_ByCategory(t *testing.T) {
	r := NewRegistry()
	r.Register(newDummy("FC-1", CatTokenEfficiency, Warn, Passive))
	r.Register(newDummy("FC-2", CatFlowSafety, Fail, Active))
	r.Register(newDummy("FC-3", CatTokenEfficiency, Info, Passive))

	result := r.Filter(&Filter{Category: CatTokenEfficiency})
	if len(result) != 2 {
		t.Fatalf("expected 2 token-efficiency checks, got %d", len(result))
	}
	for _, c := range result {
		if c.Category() != CatTokenEfficiency {
			t.Errorf("expected category token-efficiency, got %s", c.Category())
		}
	}
}

func TestRegistry_Filter_ByMinSeverity(t *testing.T) {
	r := NewRegistry()
	r.Register(newDummy("MS-1", CatFlowSafety, Info, Passive))
	r.Register(newDummy("MS-2", CatFlowSafety, Warn, Passive))
	r.Register(newDummy("MS-3", CatFlowSafety, Fail, Active))

	sev := Warn
	result := r.Filter(&Filter{MinSeverity: &sev})
	if len(result) != 2 {
		t.Fatalf("expected 2 checks with severity >= Warn, got %d", len(result))
	}
	for _, c := range result {
		if c.Severity() < Warn {
			t.Errorf("expected severity >= Warn, got %s", c.Severity())
		}
	}
}

func TestRegistry_Filter_ByMethod(t *testing.T) {
	r := NewRegistry()
	r.Register(newDummy("M-1", CatFlowSafety, Warn, Passive))
	r.Register(newDummy("M-2", CatFlowSafety, Warn, Active))
	r.Register(newDummy("M-3", CatFlowSafety, Warn, Passive))

	method := Passive
	result := r.Filter(&Filter{Method: &method})
	if len(result) != 2 {
		t.Fatalf("expected 2 passive checks, got %d", len(result))
	}
	for _, c := range result {
		if c.Method() != Passive {
			t.Errorf("expected Passive method, got %s", c.Method())
		}
	}
}

func TestRegistry_Filter_BySkipIDs(t *testing.T) {
	r := NewRegistry()
	r.Register(newDummy("SK-1", CatFlowSafety, Warn, Passive))
	r.Register(newDummy("SK-2", CatFlowSafety, Warn, Passive))
	r.Register(newDummy("SK-3", CatFlowSafety, Warn, Passive))

	result := r.Filter(&Filter{SkipIDs: map[string]bool{"SK-2": true}})
	if len(result) != 2 {
		t.Fatalf("expected 2 checks after skipping SK-2, got %d", len(result))
	}
	for _, c := range result {
		if c.ID() == "SK-2" {
			t.Error("SK-2 should have been skipped")
		}
	}
}

func TestRegistry_ByCategory(t *testing.T) {
	r := NewRegistry()
	r.Register(newDummy("BC-1", CatTokenEfficiency, Warn, Passive))
	r.Register(newDummy("BC-2", CatFlowSafety, Fail, Active))
	r.Register(newDummy("BC-3", CatTokenEfficiency, Info, Passive))

	byCat := r.ByCategory()
	if len(byCat[CatTokenEfficiency]) != 2 {
		t.Errorf("expected 2 token-efficiency checks, got %d", len(byCat[CatTokenEfficiency]))
	}
	if len(byCat[CatFlowSafety]) != 1 {
		t.Errorf("expected 1 flow-safety check, got %d", len(byCat[CatFlowSafety]))
	}
	if len(byCat[CatSelfDescribing]) != 0 {
		t.Errorf("expected 0 self-describing checks, got %d", len(byCat[CatSelfDescribing]))
	}
}

func TestRegistry_CategoryNames(t *testing.T) {
	r := NewRegistry()
	r.Register(newDummy("CN-1", CatFlowSafety, Warn, Passive))
	r.Register(newDummy("CN-2", CatTokenEfficiency, Fail, Active))
	r.Register(newDummy("CN-3", CatFlowSafety, Info, Passive))

	cats := r.CategoryNames()
	if len(cats) != 2 {
		t.Fatalf("expected 2 category names, got %d", len(cats))
	}
	// Order should be registration order of first appearance.
	if cats[0] != CatFlowSafety {
		t.Errorf("expected first category CatFlowSafety, got %s", cats[0])
	}
	if cats[1] != CatTokenEfficiency {
		t.Errorf("expected second category CatTokenEfficiency, got %s", cats[1])
	}
}

func TestDefaultRegistry_HasAllChecks(t *testing.T) {
	r := DefaultRegistry()

	// We expect at least TE-1..TE-2, FS-2..SA-1, SA-2..SA-4, SD-3..SD-6, FS-4..FS-5, FS-6..PV-4 = 26 checks.
	if r.Len() < 26 {
		t.Errorf("expected at least 26 checks in default registry, got %d", r.Len())
	}

	expectedIDs := []string{
		"TE-1", "FS-1", "SD-1", "SD-2", "TE-2",
		"FS-2", "TE-3", "TE-4", "FS-3", "SA-1",
		"SA-2", "SA-3", "SA-4",
		"SD-3", "SD-4", "SD-5", "SD-6",
		"FS-4", "FS-5",
		"FS-6", "PV-1", "TE-5", "PV-2", "PV-3", "TE-6", "PV-4",
	}
	for _, id := range expectedIDs {
		if r.Get(id) == nil {
			t.Errorf("expected check %s in default registry but not found", id)
		}
	}
}

func TestSortResultsByCategory(t *testing.T) {
	results := []*Result{
		{CheckID: "FS-6", Category: CatPredictability},
		{CheckID: "TE-1", Category: CatTokenEfficiency},
		{CheckID: "FS-4", Category: CatFlowSafety},
		{CheckID: "FS-1", Category: CatTokenEfficiency},
		{CheckID: "FS-2", Category: CatFlowSafety},
	}

	SortResultsByCategory(results)

	expectedOrder := []string{"FS-2", "FS-4", "FS-1", "TE-1", "FS-6"}
	for i, r := range results {
		if r.CheckID != expectedOrder[i] {
			t.Errorf("index %d: expected %s, got %s", i, expectedOrder[i], r.CheckID)
		}
	}
}
