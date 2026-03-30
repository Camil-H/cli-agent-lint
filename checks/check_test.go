package checks

import (
	"fmt"
	"sync"
	"testing"
)

func TestResultSet_ConcurrentAccess(t *testing.T) {
	rs := NewResultSet()
	const goroutines = 100

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(i int) {
			defer wg.Done()
			id := fmt.Sprintf("CHECK-%d", i)
			rs.Set(id, &Result{CheckID: id, Status: StatusPass})
			// Read back our own and a neighbor's.
			_ = rs.Get(id)
			_ = rs.Get(fmt.Sprintf("CHECK-%d", (i+1)%goroutines))
		}(i)
	}

	wg.Wait()

	// Verify all results were stored.
	for i := 0; i < goroutines; i++ {
		id := fmt.Sprintf("CHECK-%d", i)
		if r := rs.Get(id); r == nil {
			t.Errorf("missing result for %s", id)
		}
	}
}
