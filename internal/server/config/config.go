package config

import (
	"flag"
	"strings"
)

type Config struct {
	Addr       string
	UpdatePath string
	ValuePath  string
}

func GetConfig() Config {
	cfg := Config{}
	flag.StringVar(&cfg.Addr, `a`, `localhost:8080`, `server host`)
	flag.StringVar(&cfg.UpdatePath, `updatePath`, `update`, `update path`)
	flag.StringVar(&cfg.ValuePath, `valuePath`, `value`, `value path`)

	flag.Parse()

	cfg.UpdatePath = `/` + strings.Trim(cfg.UpdatePath, `/`)
	cfg.ValuePath = `/` + strings.Trim(cfg.ValuePath, `/`)
	return cfg
}
