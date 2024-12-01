package main

import (
	"github.com/korobkovandrey/runtime-metrics/internal/server"

	"log"
)

func main() {
	if err := server.New(server.Config{
		Addr:       `localhost:8080`,
		UpdatePath: `/update`,
		ValuePath:  `/value`,
	}).Run(); err != nil {
		log.Fatal(err)
	}
}
