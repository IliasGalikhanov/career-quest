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
			b.Run("goal_update", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					d, err = s.Update(ctx, d.ID, d.Revision, func(bundle *domain.Bundle) error {
						bundle.Employees[0].Goal = &domain.Goal{Role: "Platform Engineer", Grade: "Lead"}
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
