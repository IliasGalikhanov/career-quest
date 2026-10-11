package main

import (
	"careerquest/internal/evaluation"
	"encoding/json"
	"os"
)

func main() {
	results := evaluation.Run()
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(results); err != nil {
		panic(err)
	}
	for _, r := range results {
		if !r.Passed {
			os.Exit(1)
		}
	}
}
