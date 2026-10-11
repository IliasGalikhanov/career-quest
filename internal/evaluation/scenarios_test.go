package evaluation

import "testing"

func TestProductExpectations(t *testing.T) {
	for _, r := range Run() {
		t.Run(r.Name, func(t *testing.T) {
			if !r.Passed {
				t.Fatalf("want %q, got %q: %s", r.Expected, r.Actual, r.Rationale)
			}
		})
	}
}
