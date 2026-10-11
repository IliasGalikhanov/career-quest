package domain_test

import (
	"careerquest/internal/demo"
	"careerquest/internal/domain"
	"encoding/json"
	"math/rand"
	"testing"
)

func FuzzGainBounds(f *testing.F) {
	f.Add(uint8(2), uint8(1), uint8(4))
	f.Fuzz(func(t *testing.T, l, g, c uint8) {
		level := int(l % 6)
		cap := int(c % 6)
		gain := int(g % 6)
		got := domain.ApplyGain(level, domain.Gain{Gain: gain, MaxLevel: cap})
		if got < level || got > 5 || got > max(level, cap) {
			t.Fatalf("invalid level %d", got)
		}
	})
}
func TestRankingIndependentOfCatalogOrder(t *testing.T) {
	b := demo.Dataset()
	a, _ := domain.Analyze(b, b.Employees[0].ID)
	want, _ := json.Marshal(a.Recommendations)
	rng := rand.New(rand.NewSource(89))
	for i := 0; i < 50; i++ {
		rng.Shuffle(len(b.Events), func(i, j int) { b.Events[i], b.Events[j] = b.Events[j], b.Events[i] })
		actual, _ := domain.Analyze(b, b.Employees[0].ID)
		got, _ := json.Marshal(actual.Recommendations)
		if string(got) != string(want) {
			t.Fatal("catalog order changed ranking")
		}
	}
}
