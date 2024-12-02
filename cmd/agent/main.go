package main

import (
	"github.com/korobkovandrey/runtime-metrics/internal/agent"
	"github.com/korobkovandrey/runtime-metrics/internal/agent/config"

	"log"
)

func main() {
	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatal(err)
	}
	if err := agent.New(cfg).Run(); err != nil {
		log.Fatal(err)
	}
}
