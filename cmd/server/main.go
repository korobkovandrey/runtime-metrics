package main

import (
	"github.com/korobkovandrey/runtime-metrics/internal/server"
	"github.com/korobkovandrey/runtime-metrics/internal/server/config"

	"log"
)

func main() {
	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatal(err)
	}
	if err = server.New(cfg).Run(); err != nil {
		log.Fatal(err)
	}
}
