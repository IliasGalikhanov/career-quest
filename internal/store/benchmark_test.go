package store_test

import (
	"careerquest/internal/demo"
	"careerquest/internal/domain"
	"careerquest/internal/store"
	"context"
	"fmt"
	"path/filepath"
	"testing"
)

func BenchmarkDataset(b *testing.B) {
	for _, size := range []int{12, 500, 2000} {
		b.Run(fmt.Sprint(size), func(b *testing.B) {
			s, err := store.Open(filepath.Join(b.TempDir(), "bench.db"))
			if err != nil {
				b.Fatal(err)
			}
			defer s.Close()
			ctx := context.Background()
			d, err := s.Create(ctx, demo.Scaled(size))
			if err != nil {
				b.Fatal(err)
			}
			b.Run("read", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					if _, err := s.Get(ctx, d.ID); err != nil {
						b.Fatal(err)
					}
				}
			})
			b.Run("recommendation_with_evidence", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					var selected store.Dataset
					if selective, ok := any(s).(interface {
						Select(context.Context, string, string, string, string) (store.Dataset, error)
					}); ok {
						selected, err = selective.Select(ctx, d.ID, "load-00000", "", "")
					} else {
						selected, err = s.Get(ctx, d.ID)
					}
					if err != nil {
						b.Fatal(err)
					}
					a, _ := domain.Analyze(selected.Bundle, "load-00000")
					if err = s.SaveRun(ctx, selected, a); err != nil {
						b.Fatal(err)
					}
				}
			})

			b.Run("goal_update", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					d, err = s.Update(ctx, d.ID, d.Revision, func(bundle *domain.Bundle) error {
						grade := "Lead"
						if d.Revision%2 == 0 {
							grade = "Senior"
						}
						bundle.Employees[0].Goal = &domain.Goal{Role: "Platform Engineer", Grade: grade}
						return nil
					})
					if err != nil {
						b.Fatal(err)
					}
				}
			})
		})
	}
}
