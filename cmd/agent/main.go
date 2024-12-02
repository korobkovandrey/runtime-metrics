package main

import (
	"github.com/korobkovandrey/runtime-metrics/internal/agent"
	"github.com/korobkovandrey/runtime-metrics/internal/agent/config"

	"log"
)

func main() {
	if err := agent.New(config.GetConfig()).Run(); err != nil {
		log.Fatal(err)
	}
}
