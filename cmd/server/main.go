package main

import (
	"log"

	"github.com/korobkovandrey/runtime-metrics/internal/server"
	"github.com/korobkovandrey/runtime-metrics/internal/server/config"
	"github.com/korobkovandrey/runtime-metrics/internal/server/logger"
)

func main() {
	zapLogger, err := logger.NewZapLogger()
	if err != nil {
		log.Fatal(err)
	}
	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatal(err)
	}
	if err = server.New(cfg, zapLogger).Run(); err != nil {
		log.Fatal(err)
	}
}
