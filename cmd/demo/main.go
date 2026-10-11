package main

import (
	"careerquest/internal/demo"
	"encoding/json"
	"log"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("usage: go run ./cmd/demo <output.json>")
	}
	data, err := json.MarshalIndent(demo.Dataset(), "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(os.Args[1], append(data, '\n'), 0644); err != nil {
		log.Fatal(err)
	}
}
