package main

import (
	"github.com/korobkovandrey/runtime-metrics/internal/server"
	"github.com/korobkovandrey/runtime-metrics/internal/server/config"

	"log"
)

func main() {
	if err := server.New(config.GetConfig()).Run(); err != nil {
		log.Fatal(err)
	}
}
