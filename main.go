package main

import (
	"fmt"
	"log"
)

func main() {
	cfg := parseFlags()

	yaml, err := buildKindYAML(cfg)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print(yaml)
}
