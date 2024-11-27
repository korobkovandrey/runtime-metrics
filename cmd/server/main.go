package main

import (
	"github.com/korobkovandrey/runtime-metrics/internal/server"

	"log"
)

func main() {
	if err := server.New().Run(); err != nil {
		log.Fatal(err)
	}
}
