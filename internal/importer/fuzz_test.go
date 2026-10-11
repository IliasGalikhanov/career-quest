package importer

import (
	"careerquest/internal/demo"
	"careerquest/internal/domain"
	"encoding/json"
	"testing"
)

func FuzzBundle(f *testing.F) {
	seed, _ := json.Marshal(demo.Dataset())
	f.Add(seed)
	f.Add([]byte("{broken"))
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 65536 {
			t.Skip()
		}
		bundle, issues := Parse(map[string][]byte{"bundle.json": data}, nil)
		if len(issues) == 0 && len(domain.Validate(bundle)) != 0 {
			t.Fatal("parser accepted an invalid bundle")
		}
	})
}
